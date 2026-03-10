package query

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xaws "github.com/owen/xatu/internal/aws"
	"github.com/owen/xatu/internal/config"
)

// focus tracks which panel has focus.
type focus int

const (
	focusEditor focus = iota
	focusSaved
	focusResults
)

var (
	green    = lipgloss.Color("#00FF00")
	dimGreen = lipgloss.Color("#007700")
	white    = lipgloss.Color("#CCCCCC")
	subtle   = lipgloss.Color("#555555")
	red      = lipgloss.Color("#FF0000")
)

// savedItem wraps a saved query for the list.
type savedItem struct {
	name  string
	query string
}

func (i savedItem) Title() string       { return i.name }
func (i savedItem) Description() string { return truncate(i.query, 40) }
func (i savedItem) FilterValue() string { return i.name }

// queryStartedMsg indicates a query was submitted.
type queryStartedMsg struct{ queryID string }

// queryPollMsg triggers polling for query results.
type queryPollMsg struct{}

// queryResultMsg carries completed query results.
type queryResultMsg struct {
	result *xaws.QueryResult
	err    error
}

// querySavedMsg indicates the query was saved to config.
type querySavedMsg struct{}

type Model struct {
	focus   focus
	editor  textarea.Model
	saved   list.Model
	results table.Model
	spinner spinner.Model

	client    *xaws.Client
	config    *config.Config
	logGroups []string

	queryID   string
	running   bool
	lastErr   error
	stats     *xaws.QueryStats
	rowCount  int

	width  int
	height int
}

func New(cfg *config.Config, client *xaws.Client, logGroups []string) Model {
	// Editor
	ta := textarea.New()
	ta.Placeholder = "fields @timestamp, @message | sort @timestamp desc | limit 20"
	ta.CharLimit = 4096
	ta.SetWidth(60)
	ta.SetHeight(6)
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.FocusedStyle.Base = lipgloss.NewStyle().Foreground(white)
	ta.BlurredStyle.Base = lipgloss.NewStyle().Foreground(subtle)
	ta.Focus()

	// Saved queries list
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = true
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(green).BorderForeground(green)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(white)
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(subtle)
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.Foreground(dimGreen).BorderForeground(green)

	items := savedItemsFromConfig(cfg)
	sl := list.New(items, delegate, 30, 10)
	sl.Title = "saved queries"
	sl.Styles.Title = lipgloss.NewStyle().Foreground(green).Bold(true)
	sl.SetShowStatusBar(false)
	sl.SetShowHelp(false)
	sl.SetFilteringEnabled(false)

	// Results table
	cols := []table.Column{
		{Title: "@timestamp", Width: 20},
		{Title: "@message", Width: 60},
	}
	t := table.New(
		table.WithColumns(cols),
		table.WithRows(nil),
		table.WithFocused(false),
		table.WithHeight(8),
	)
	s := table.DefaultStyles()
	s.Header = s.Header.Foreground(green).Bold(true)
	s.Selected = s.Selected.Foreground(green).Bold(false)
	s.Cell = s.Cell.Foreground(white)
	t.SetStyles(s)

	sp := spinner.New()
	sp.Spinner = spinner.Dot
	sp.Style = lipgloss.NewStyle().Foreground(green)

	return Model{
		focus:     focusEditor,
		editor:    ta,
		saved:     sl,
		results:   t,
		spinner:   sp,
		client:    client,
		config:    cfg,
		logGroups: logGroups,
	}
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height

	editorWidth := width*2/3 - 4
	if editorWidth < 30 {
		editorWidth = 30
	}
	m.editor.SetWidth(editorWidth)

	savedWidth := width/3 - 2
	if savedWidth < 20 {
		savedWidth = 20
	}
	savedHeight := height/2 - 2
	if savedHeight < 4 {
		savedHeight = 4
	}
	m.saved.SetSize(savedWidth, savedHeight)

	resultHeight := height/2 - 4
	if resultHeight < 3 {
		resultHeight = 3
	}
	m.results.SetHeight(resultHeight)
}

func (m *Model) SetLogGroups(groups []string) {
	m.logGroups = groups
}

func (m *Model) SetConfig(cfg *config.Config) {
	m.config = cfg
	m.saved.SetItems(savedItemsFromConfig(cfg))
}

// InputFocused returns true when the editor textarea has focus.
func (m *Model) InputFocused() bool {
	return m.focus == focusEditor
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			m.cycleFocus()
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+enter"))):
			if m.running || len(m.logGroups) == 0 {
				return m, nil
			}
			query := m.editor.Value()
			if strings.TrimSpace(query) == "" {
				return m, nil
			}
			m.running = true
			m.lastErr = nil
			m.stats = nil
			return m, m.submitQuery(query)

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
			return m, m.saveCurrentQuery()

		case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
			if m.focus == focusSaved {
				if item, ok := m.saved.SelectedItem().(savedItem); ok {
					m.editor.SetValue(item.query)
					m.focus = focusEditor
					m.editor.Focus()
				}
				return m, nil
			}
		}

		// Delegate to focused panel
		switch m.focus {
		case focusEditor:
			var cmd tea.Cmd
			m.editor, cmd = m.editor.Update(msg)
			return m, cmd
		case focusSaved:
			var cmd tea.Cmd
			m.saved, cmd = m.saved.Update(msg)
			return m, cmd
		case focusResults:
			var cmd tea.Cmd
			m.results, cmd = m.results.Update(msg)
			return m, cmd
		}

	case queryStartedMsg:
		m.queryID = msg.queryID
		return m, m.pollQueryCmd()

	case queryPollMsg:
		return m, m.fetchResults()

	case queryResultMsg:
		if msg.err != nil {
			m.running = false
			m.lastErr = msg.err
			return m, nil
		}
		if msg.result.Status == xaws.QueryStatusComplete ||
			msg.result.Status == xaws.QueryStatusFailed ||
			msg.result.Status == xaws.QueryStatusCancelled ||
			msg.result.Status == xaws.QueryStatusTimeout {
			m.running = false
			m.stats = &msg.result.Stats
			if msg.result.Status != xaws.QueryStatusComplete {
				m.lastErr = fmt.Errorf("query %s", msg.result.Status)
			}
			m.populateResults(msg.result)
			return m, nil
		}
		// Still running, poll again
		return m, m.pollQueryCmd()

	case spinner.TickMsg:
		if m.running {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}

	return m, nil
}

func (m Model) View() string {
	titleStyle := lipgloss.NewStyle().Foreground(green).Bold(true)
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(subtle)
	focusedBorder := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(green)

	// Editor panel
	editorBorder := borderStyle
	if m.focus == focusEditor {
		editorBorder = focusedBorder
	}
	editorWidth := m.width*2/3 - 4
	if editorWidth < 30 {
		editorWidth = 30
	}
	editorPanel := editorBorder.Width(editorWidth).Render(
		titleStyle.Render("query editor") + "\n\n" +
			m.editor.View() + "\n\n" +
			m.editorHints(),
	)

	// Saved queries panel
	savedBorder := borderStyle
	if m.focus == focusSaved {
		savedBorder = focusedBorder
	}
	savedWidth := m.width/3 - 4
	if savedWidth < 20 {
		savedWidth = 20
	}
	savedPanel := savedBorder.Width(savedWidth).Render(m.saved.View())

	topRow := lipgloss.JoinHorizontal(lipgloss.Top, editorPanel, " ", savedPanel)

	// Results panel
	resultsBorder := borderStyle
	if m.focus == focusResults {
		resultsBorder = focusedBorder
	}
	resultsContent := m.resultsView()
	resultsPanel := resultsBorder.Width(m.width - 4).Render(
		titleStyle.Render("results") + " " + m.statusIndicator() + "\n\n" +
			resultsContent,
	)

	return lipgloss.JoinVertical(lipgloss.Left, topRow, resultsPanel)
}

func (m Model) editorHints() string {
	hintStyle := lipgloss.NewStyle().Foreground(dimGreen)
	hints := "shift+enter: run  ctrl+s: save  tab: switch panel"
	if len(m.logGroups) > 0 {
		hints += fmt.Sprintf("  (%d log groups)", len(m.logGroups))
	}
	return hintStyle.Render(hints)
}

func (m Model) statusIndicator() string {
	if m.running {
		return m.spinner.View() + lipgloss.NewStyle().Foreground(dimGreen).Render(" running...")
	}
	if m.lastErr != nil {
		return lipgloss.NewStyle().Foreground(red).Render("✗ " + m.lastErr.Error())
	}
	if m.stats != nil {
		return lipgloss.NewStyle().Foreground(dimGreen).Render(
			fmt.Sprintf("✓ %d rows  (%.0f matched, %.0f scanned)",
				m.rowCount, m.stats.RecordsMatched, m.stats.RecordsScanned))
	}
	return ""
}

func (m Model) resultsView() string {
	if m.rowCount == 0 && !m.running && m.lastErr == nil {
		return lipgloss.NewStyle().Foreground(subtle).Render("  No results yet. Write a query and press shift+enter to run.")
	}
	return m.results.View()
}

func (m *Model) cycleFocus() {
	m.editor.Blur()
	switch m.focus {
	case focusEditor:
		m.focus = focusSaved
	case focusSaved:
		m.focus = focusResults
		m.results.Focus()
	case focusResults:
		m.focus = focusEditor
		m.editor.Focus()
		m.results.Blur()
	}
}

func (m *Model) submitQuery(query string) tea.Cmd {
	client := m.client
	groups := make([]string, len(m.logGroups))
	copy(groups, m.logGroups)
	return tea.Batch(
		m.spinner.Tick,
		func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			end := time.Now()
			start := end.Add(-1 * time.Hour)
			id, err := client.StartInsightsQuery(ctx, groups, query, start, end)
			if err != nil {
				return queryResultMsg{err: err}
			}
			return queryStartedMsg{queryID: id}
		},
	)
}

func (m *Model) pollQueryCmd() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return queryPollMsg{}
	})
}

func (m *Model) fetchResults() tea.Cmd {
	client := m.client
	queryID := m.queryID
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		result, err := client.GetQueryResults(ctx, queryID)
		return queryResultMsg{result: result, err: err}
	}
}

func (m *Model) populateResults(result *xaws.QueryResult) {
	if len(result.Results) == 0 {
		m.rowCount = 0
		return
	}

	// Discover columns from the result set
	colSet := make(map[string]bool)
	var colOrder []string
	for _, row := range result.Results {
		for k := range row {
			if k == "@ptr" {
				continue // skip internal pointer field
			}
			if !colSet[k] {
				colSet[k] = true
				colOrder = append(colOrder, k)
			}
		}
	}

	// Build table columns with proportional widths
	availWidth := m.width - 6
	colWidth := availWidth / max(len(colOrder), 1)
	if colWidth < 10 {
		colWidth = 10
	}
	if colWidth > 60 {
		colWidth = 60
	}

	cols := make([]table.Column, len(colOrder))
	for i, name := range colOrder {
		cols[i] = table.Column{Title: name, Width: colWidth}
	}

	// Build rows
	rows := make([]table.Row, len(result.Results))
	for i, row := range result.Results {
		r := make(table.Row, len(colOrder))
		for j, col := range colOrder {
			r[j] = row[col]
		}
		rows[i] = r
	}

	m.rowCount = len(rows)
	m.results.SetColumns(cols)
	m.results.SetRows(rows)
}

func (m *Model) saveCurrentQuery() tea.Cmd {
	query := m.editor.Value()
	if strings.TrimSpace(query) == "" {
		return nil
	}

	// Generate name from query prefix
	name := truncate(strings.TrimSpace(query), 30)

	cfg := m.config
	cfg.LogQuery.SavedQueries = append(cfg.LogQuery.SavedQueries, config.SavedQuery{
		Name:  name,
		Query: query,
	})

	return func() tea.Msg {
		_ = config.Save(cfg)
		return querySavedMsg{}
	}
}

func savedItemsFromConfig(cfg *config.Config) []list.Item {
	items := make([]list.Item, len(cfg.LogQuery.SavedQueries))
	for i, sq := range cfg.LogQuery.SavedQueries {
		items[i] = savedItem{name: sq.Name, query: sq.Query}
	}
	return items
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-3] + "..."
}
