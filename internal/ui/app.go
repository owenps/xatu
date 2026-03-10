package ui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xaws "github.com/owen/xatu/internal/aws"
	"github.com/owen/xatu/internal/config"
	logpkg "github.com/owen/xatu/internal/log"
	"github.com/owen/xatu/internal/ui/query"
	"github.com/owen/xatu/internal/ui/settings"
	"github.com/owen/xatu/internal/ui/setup"
	"github.com/owen/xatu/internal/ui/tile"
	"github.com/owen/xatu/internal/ui/tiles"
)

// RunSetupMsg can be sent from any screen (e.g. settings) to re-launch the setup wizard.
type RunSetupMsg struct{}

// pollTickMsg fires on each auto-poll interval.
type pollTickMsg time.Time

// fetchErrMsg surfaces AWS fetch errors.
type fetchErrMsg struct{ err error }

type Screen int

const (
	ScreenSetup Screen = iota
	ScreenDashboard
	ScreenSettings
	ScreenQuery
)

type App struct {
	screen      Screen
	setup       setup.Model
	settings    settings.Model
	query       query.Model
	grid        *tile.Grid
	buffer      *logpkg.Buffer
	statusBar   StatusBar
	spinner     spinner.Model
	config      *config.Config
	client      *xaws.Client
	theme       Theme
	width       int
	height      int
	activeCtx   int
	lastFetchAt time.Time
	fetching    bool
	lastErr     error
	showHelp    bool
}

func NewApp(cfg *config.Config, client *xaws.Client, needsSetup bool) App {
	theme := GetTheme(cfg.General.Theme)

	ctxName := "no context"
	if len(cfg.Contexts) > 0 {
		ctxName = cfg.Contexts[0].Name
	}

	buf := logpkg.NewBuffer(cfg.Dashboard.BufferSize)

	grid := tile.NewGrid(theme.TileBorder, theme.TileBorderFocused, theme.TileTitle)
	grid.AddTile(tiles.NewAttributes(buf))
	grid.AddTile(tiles.NewHeatMap(buf, parseAggInterval(cfg.Dashboard.AggregationInterval)))
	grid.AddTile(tiles.NewPatterns(buf))
	grid.AddTile(tiles.NewLogStream(buf))

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(theme.Primary)

	var logGroups []string
	if len(cfg.Contexts) > 0 {
		logGroups = cfg.Contexts[0].LogGroups
	}

	app := App{
		config:    cfg,
		client:    client,
		theme:     theme,
		buffer:    buf,
		grid:      grid,
		settings:  settings.New(cfg),
		query:     query.New(cfg, client, logGroups),
		spinner:   sp,
		statusBar: NewStatusBar(theme, ctxName),
	}

	if needsSetup {
		app.screen = ScreenSetup
		app.setup = setup.New(client, cfg.General.Region)
	} else {
		app.screen = ScreenDashboard
	}

	return app
}

func (a App) Init() tea.Cmd {
	if a.screen == ScreenSetup {
		return a.setup.Init()
	}
	cmds := []tea.Cmd{a.grid.Init(), a.fetchInitialLogs(), a.spinner.Tick}
	if a.config.Dashboard.AutoPoll {
		cmds = append(cmds, a.pollTickCmd())
	}
	return tea.Batch(cmds...)
}

func (a App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.statusBar.SetWidth(msg.Width)

		if a.screen == ScreenSetup {
			newSetup, cmd := a.setup.Update(msg)
			a.setup = newSetup
			return a, cmd
		}

		statusHeight := lipgloss.Height(a.statusBar.View())
		contentHeight := msg.Height - statusHeight
		a.grid.SetSize(msg.Width, contentHeight)
		a.settings.SetSize(msg.Width, contentHeight)
		a.query.SetSize(msg.Width, contentHeight)
		return a, nil

	case settings.ConfigUpdatedMsg:
		a.config = msg.Config
		a.theme = GetTheme(a.config.General.Theme)
		a.query.SetConfig(a.config)
		return a, nil

	case setup.SetupCompleteMsg:
		a.config = msg.Config
		a.screen = ScreenDashboard
		a.theme = GetTheme(a.config.General.Theme)
		if len(a.config.Contexts) > 0 {
			a.statusBar.SetContext(a.config.Contexts[0].Name)
			a.query.SetLogGroups(a.config.Contexts[0].LogGroups)
		}
		a.query.SetConfig(a.config)
		cmds := []tea.Cmd{a.grid.Init(), a.fetchInitialLogs()}
		if a.config.Dashboard.AutoPoll {
			cmds = append(cmds, a.pollTickCmd())
		}
		return a, tea.Batch(cmds...)

	case RunSetupMsg:
		a.screen = ScreenSetup
		a.setup = setup.New(a.client, a.config.General.Region)
		return a, a.setup.Init()

	case tiles.NewLogsMsg:
		a.fetching = false
		a.lastErr = nil
		a.lastFetchAt = time.Now()
		cmd := a.grid.UpdateAll(msg)
		return a, cmd

	case fetchErrMsg:
		a.fetching = false
		a.lastErr = msg.err
		return a, nil

	case pollTickMsg:
		if !a.config.Dashboard.AutoPoll {
			return a, nil
		}
		var cmds []tea.Cmd
		if !a.fetching {
			cmds = append(cmds, a.fetchSinceLastCmd())
		}
		cmds = append(cmds, a.pollTickCmd())
		return a, tea.Batch(cmds...)

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

	case tea.KeyMsg:
		// Global keys
		if key.Matches(msg, Keys.Quit) {
			return a, tea.Quit
		}

		// Help overlay toggle (any screen except setup)
		if a.showHelp {
			// Any key dismisses help
			a.showHelp = false
			return a, nil
		}

		// Screen-specific global navigation (only when not in setup)
		if a.screen != ScreenSetup {
			// When a settings/query text input is focused, skip single-char global keys
			// so the user can type freely. Only allow modifier-based keys (ctrl+c, cmd+,).
			inputActive := (a.screen == ScreenSettings && a.settings.InputFocused()) ||
				(a.screen == ScreenQuery && a.query.InputFocused())

			switch {
			case !inputActive && key.Matches(msg, Keys.Help):
				a.showHelp = true
				return a, nil
			case !inputActive && key.Matches(msg, Keys.ShiftTab) && a.screen == ScreenDashboard:
				if len(a.config.Contexts) > 1 {
					a.activeCtx = (a.activeCtx + 1) % len(a.config.Contexts)
					ctx := a.config.Contexts[a.activeCtx]
					a.statusBar.SetContext(ctx.Name)
					a.query.SetLogGroups(ctx.LogGroups)
					a.buffer.Clear()
					a.lastFetchAt = time.Time{}
					return a, a.fetchInitialLogs()
				}
				return a, nil
			case !inputActive && key.Matches(msg, Keys.Home):
				a.screen = ScreenDashboard
				return a, nil
			case !inputActive && key.Matches(msg, Keys.QueryPage):
				a.screen = ScreenQuery
				return a, nil
			case key.Matches(msg, Keys.Settings):
				a.screen = ScreenSettings
				return a, nil
			case key.Matches(msg, Keys.Escape) && (a.screen == ScreenSettings || a.screen == ScreenQuery):
				a.screen = ScreenDashboard
				return a, nil
			case !inputActive && key.Matches(msg, Keys.Refresh) && a.screen == ScreenDashboard:
				if !a.fetching {
					a.fetching = true
					return a, a.fetchSinceLastCmd()
				}
				return a, nil
			}
		}
	}

	// Delegate to active screen
	switch a.screen {
	case ScreenSetup:
		newSetup, cmd := a.setup.Update(msg)
		a.setup = newSetup
		return a, cmd
	case ScreenDashboard:
		cmd := a.grid.Update(msg)
		return a, cmd
	case ScreenSettings:
		newSettings, cmd := a.settings.Update(msg)
		a.settings = newSettings
		return a, cmd
	case ScreenQuery:
		newQuery, cmd := a.query.Update(msg)
		a.query = newQuery
		return a, cmd
	}

	return a, nil
}

func (a App) View() string {
	if a.width == 0 {
		return ""
	}

	if a.showHelp {
		return a.helpOverlay()
	}

	var content string

	switch a.screen {
	case ScreenSetup:
		return a.setup.View()
	case ScreenDashboard:
		content = a.grid.View()
	case ScreenSettings:
		content = a.settingsView()
	case ScreenQuery:
		content = a.queryView()
	}

	statusBar := a.statusBar.ViewWithState(a.fetching, a.spinner.View(), a.lastFetchAt, a.lastErr)

	contentHeight := a.height - lipgloss.Height(statusBar)

	styledContent := lipgloss.NewStyle().
		Height(contentHeight).
		Width(a.width).
		Render(content)

	return lipgloss.JoinVertical(lipgloss.Left, styledContent, statusBar)
}

func (a App) settingsView() string {
	return a.settings.View()
}

func (a App) queryView() string {
	return a.query.View()
}

func (a App) activeLogGroups() []string {
	if a.activeCtx < len(a.config.Contexts) {
		return a.config.Contexts[a.activeCtx].LogGroups
	}
	return nil
}

func (a App) pollTickCmd() tea.Cmd {
	interval := time.Duration(a.config.Dashboard.PollIntervalSeconds) * time.Second
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return pollTickMsg(t)
	})
}

func (a *App) fetchInitialLogs() tea.Cmd {
	groups := a.activeLogGroups()
	if len(groups) == 0 {
		return nil
	}
	a.fetching = true
	client := a.client
	return func() tea.Msg {
		end := time.Now()
		start := end.Add(-90 * time.Minute)
		entries, err := client.FetchLogs(context.Background(), groups, start, end)
		if err != nil {
			return fetchErrMsg{err: err}
		}
		return tiles.NewLogsMsg{Entries: entries}
	}
}

func (a *App) fetchSinceLastCmd() tea.Cmd {
	groups := a.activeLogGroups()
	if len(groups) == 0 {
		return nil
	}
	a.fetching = true
	client := a.client
	since := a.lastFetchAt
	if since.IsZero() {
		since = time.Now().Add(-1 * time.Minute)
	}
	return func() tea.Msg {
		entries, err := client.FetchLogs(context.Background(), groups, since, time.Now())
		if err != nil {
			return fetchErrMsg{err: err}
		}
		return tiles.NewLogsMsg{Entries: entries}
	}
}

func (a App) helpOverlay() string {
	w := a.width - 8
	if w < 40 {
		w = 40
	}
	if w > 70 {
		w = 70
	}

	title := lipgloss.NewStyle().
		Foreground(a.theme.Primary).
		Bold(true).
		Render("Keybindings")

	keyStyle := lipgloss.NewStyle().Foreground(a.theme.Primary).Width(16)
	descStyle := lipgloss.NewStyle().Foreground(a.theme.FG)

	bindings := []struct{ key, desc string }{
		{"?", "toggle this help"},
		{"ctrl+c", "quit"},
		{"↑/↓ or j/k", "navigate / scroll"},
		{"←/→", "navigate between panels"},
		{"tab", "cycle focus between tiles"},
		{"shift+tab", "switch context"},
		{"enter", "select / expand tile"},
		{"esc", "back / collapse / unfocus"},
		{"r", "refresh logs"},
		{"h", "home (dashboard)"},
		{"q", "query editor"},
		{"cmd+,", "settings"},
		{"shift+enter", "submit query"},
		{"ctrl+s", "save (in settings/query)"},
	}

	lines := "\n" + title + "\n\n"
	for _, b := range bindings {
		lines += "  " + keyStyle.Render(b.key) + descStyle.Render(b.desc) + "\n"
	}
	lines += "\n" + lipgloss.NewStyle().Foreground(a.theme.Subtle).Render("  press any key to dismiss")

	overlay := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(a.theme.Primary).
		Width(w).
		Padding(1, 2).
		Render(lines)

	// Center the overlay
	return lipgloss.Place(a.width, a.height, lipgloss.Center, lipgloss.Center, overlay)
}

func parseAggInterval(s string) time.Duration {
	switch s {
	case "5m":
		return 5 * time.Minute
	case "15m":
		return 15 * time.Minute
	case "1h":
		return time.Hour
	default:
		return time.Minute
	}
}
