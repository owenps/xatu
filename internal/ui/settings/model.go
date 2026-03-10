package settings

import (
	"fmt"
	"strconv"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/owenps/xatu/internal/config"
	"github.com/owenps/xatu/internal/ui/shared"
)

// ConfigUpdatedMsg is sent when settings are saved.
type ConfigUpdatedMsg struct {
	Config *config.Config
}

type section int

const (
	sectionGeneral section = iota
	sectionContexts
	sectionDashboard
	sectionLogQuery
	sectionAbout
)

type navItem struct {
	name    string
	section section
}

func (i navItem) Title() string       { return i.name }
func (i navItem) Description() string { return "" }
func (i navItem) FilterValue() string { return i.name }

// focus tracks whether the nav or the pane has focus.
type focus int

const (
	focusNav focus = iota
	focusPane
)

type Model struct {
	nav      list.Model
	section  section
	focus    focus
	config   *config.Config
	dirty    bool
	width    int
	height   int

	// Pane inputs
	generalInputs   []labeledInput
	dashboardInputs []labeledInput
	activeInput     int
}

type labeledInput struct {
	label string
	input textinput.Model
}

var (
	green     = lipgloss.Color("#00FF00")
	dimGreen  = lipgloss.Color("#007700")
	white     = lipgloss.Color("#CCCCCC")
	subtle    = lipgloss.Color("#555555")
	navWidth  = 22
)

func New(cfg *config.Config) Model {
	// Nav list
	items := []list.Item{
		navItem{"general", sectionGeneral},
		navItem{"contexts", sectionContexts},
		navItem{"dashboard", sectionDashboard},
		navItem{"log query", sectionLogQuery},
		navItem{"about", sectionAbout},
	}
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(green).BorderForeground(green)
	delegate.Styles.NormalTitle = delegate.Styles.NormalTitle.Foreground(white)

	nav := list.New(items, delegate, navWidth, 20)
	nav.Title = "settings"
	nav.Styles.Title = lipgloss.NewStyle().Foreground(green).Bold(true)
	nav.SetShowStatusBar(false)
	nav.SetShowHelp(false)
	nav.SetFilteringEnabled(false)

	m := Model{
		nav:     nav,
		section: sectionGeneral,
		focus:   focusNav,
		config:  cfg,
	}
	m.buildGeneralInputs()
	m.buildDashboardInputs()

	return m
}

func (m *Model) buildGeneralInputs() {
	m.generalInputs = []labeledInput{
		m.newInput("Time Zone", m.config.General.TimeZone),
		m.newInput("Theme", m.config.General.Theme),
		m.newInput("Region", m.config.General.Region),
	}
}

func (m *Model) buildDashboardInputs() {
	m.dashboardInputs = []labeledInput{
		m.newInput("Auto Poll", fmt.Sprintf("%v", m.config.Dashboard.AutoPoll)),
		m.newInput("Poll Interval (s)", strconv.Itoa(m.config.Dashboard.PollIntervalSeconds)),
		m.newInput("Aggregation Interval", m.config.Dashboard.AggregationInterval),
		m.newInput("Buffer Size", strconv.Itoa(m.config.Dashboard.BufferSize)),
	}
}

func (m *Model) newInput(label, value string) labeledInput {
	ti := textinput.New()
	ti.SetValue(value)
	ti.PromptStyle = lipgloss.NewStyle().Foreground(green)
	ti.TextStyle = lipgloss.NewStyle().Foreground(white)
	ti.Prompt = ""
	ti.CharLimit = 128
	return labeledInput{label: label, input: ti}
}

// InputFocused returns true when a text input in the pane has focus,
// so the parent can avoid intercepting single-char keys.
func (m *Model) InputFocused() bool {
	return m.focus == focusPane
}

func (m *Model) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.nav.SetSize(navWidth, height-2)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			if m.focus == focusNav {
				m.focus = focusPane
				m.focusFirstInput()
			} else {
				m.focus = focusNav
				m.blurAllInputs()
			}
			return m, nil

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+s"))):
			m.applyInputsToConfig()
			err := config.Save(m.config)
			if err == nil {
				m.dirty = false
			}
			return m, func() tea.Msg { return ConfigUpdatedMsg{Config: m.config} }
		}

		if m.focus == focusNav {
			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("enter"))):
				if item, ok := m.nav.SelectedItem().(navItem); ok {
					m.section = item.section
					m.activeInput = 0
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.nav, cmd = m.nav.Update(msg)
			return m, cmd
		}

		// Pane focus: navigate inputs with j/k or up/down, tab to switch fields
		if m.focus == focusPane {
			inputs := m.activeInputs()
			if inputs == nil {
				return m, nil
			}

			switch {
			case key.Matches(msg, key.NewBinding(key.WithKeys("up", "shift+tab"))):
				m.blurAllInputs()
				m.activeInput--
				if m.activeInput < 0 {
					m.activeInput = len(inputs) - 1
				}
				m.focusActiveInput()
				return m, textinput.Blink
			case key.Matches(msg, key.NewBinding(key.WithKeys("down", "tab"))):
				// Within pane, tab cycles fields (not back to nav)
				m.blurAllInputs()
				m.activeInput++
				if m.activeInput >= len(inputs) {
					m.activeInput = 0
				}
				m.focusActiveInput()
				return m, textinput.Blink
			}

			// Update active text input
			if m.activeInput < len(inputs) {
				var cmd tea.Cmd
				inputs[m.activeInput].input, cmd = inputs[m.activeInput].input.Update(msg)
				m.dirty = true
				m.setActiveInputs(inputs)
				return m, cmd
			}
		}
	}

	return m, nil
}

func (m Model) View() string {
	navView := m.nav.View()

	paneView := m.renderPane()

	divider := lipgloss.NewStyle().
		Foreground(subtle).
		Render("│")

	navStyled := lipgloss.NewStyle().
		Width(navWidth).
		Height(m.height - 2).
		Render(navView)

	paneWidth := m.width - navWidth - 3
	if paneWidth < 10 {
		paneWidth = 10
	}
	paneStyled := lipgloss.NewStyle().
		Width(paneWidth).
		Height(m.height - 2).
		Render(paneView)

	content := lipgloss.JoinHorizontal(lipgloss.Top, navStyled, divider, " ", paneStyled)

	if m.dirty {
		saveHint := lipgloss.NewStyle().Foreground(dimGreen).
			Render("  ctrl+s: save  tab: switch focus")
		content += "\n" + saveHint
	}

	return content
}

func (m Model) renderPane() string {
	titleStyle := lipgloss.NewStyle().Foreground(green).Bold(true).MarginBottom(1)
	labelStyle := lipgloss.NewStyle().Foreground(subtle).Width(22)
	valueStyle := lipgloss.NewStyle().Foreground(white)

	switch m.section {
	case sectionGeneral:
		out := titleStyle.Render("General") + "\n\n"
		for _, inp := range m.generalInputs {
			out += labelStyle.Render(inp.label+":") + " " + inp.input.View() + "\n\n"
		}
		return out

	case sectionContexts:
		out := titleStyle.Render("Contexts") + "\n\n"
		for i, ctx := range m.config.Contexts {
			out += valueStyle.Render(fmt.Sprintf("  %d. %s (%d log groups)", i+1, ctx.Name, len(ctx.LogGroups))) + "\n"
			for _, lg := range ctx.LogGroups {
				out += lipgloss.NewStyle().Foreground(subtle).Render("     "+lg) + "\n"
			}
			out += "\n"
		}
		if len(m.config.Contexts) == 0 {
			out += valueStyle.Render("  No contexts configured.") + "\n"
		}
		runSetup := lipgloss.NewStyle().Foreground(dimGreen).Render("  Run setup wizard to add/modify contexts")
		out += "\n" + runSetup
		return out

	case sectionDashboard:
		out := titleStyle.Render("Dashboard") + "\n\n"
		for _, inp := range m.dashboardInputs {
			out += labelStyle.Render(inp.label+":") + " " + inp.input.View() + "\n\n"
		}
		return out

	case sectionLogQuery:
		out := titleStyle.Render("Saved Queries") + "\n\n"
		if len(m.config.LogQuery.SavedQueries) == 0 {
			out += valueStyle.Render("  No saved queries.") + "\n"
			out += lipgloss.NewStyle().Foreground(subtle).Render("  Save queries from the query editor (q)") + "\n"
		} else {
			for _, sq := range m.config.LogQuery.SavedQueries {
				out += valueStyle.Render("  "+sq.Name) + "\n"
				out += lipgloss.NewStyle().Foreground(subtle).Render("  "+sq.Query) + "\n\n"
			}
		}
		return out

	case sectionAbout:
		out := shared.Banner(green, dimGreen, m.width-navWidth-5) + "\n"
		out += titleStyle.Render("About") + "\n\n"
		out += valueStyle.Render("  Version:  0.1.0") + "\n"
		out += valueStyle.Render("  Author:   owen") + "\n"
		out += labelStyle.Render("  Repo:") + " " + valueStyle.Render("github.com/owenps/xatu") + "\n"
		return out
	}

	return ""
}

func (m *Model) activeInputs() []labeledInput {
	switch m.section {
	case sectionGeneral:
		return m.generalInputs
	case sectionDashboard:
		return m.dashboardInputs
	}
	return nil
}

func (m *Model) setActiveInputs(inputs []labeledInput) {
	switch m.section {
	case sectionGeneral:
		m.generalInputs = inputs
	case sectionDashboard:
		m.dashboardInputs = inputs
	}
}

func (m *Model) focusFirstInput() {
	m.activeInput = 0
	m.focusActiveInput()
}

func (m *Model) focusActiveInput() {
	inputs := m.activeInputs()
	if inputs != nil && m.activeInput < len(inputs) {
		inputs[m.activeInput].input.Focus()
		m.setActiveInputs(inputs)
	}
}

func (m *Model) blurAllInputs() {
	for i := range m.generalInputs {
		m.generalInputs[i].input.Blur()
	}
	for i := range m.dashboardInputs {
		m.dashboardInputs[i].input.Blur()
	}
}

func (m *Model) applyInputsToConfig() {
	// General
	if len(m.generalInputs) >= 3 {
		m.config.General.TimeZone = m.generalInputs[0].input.Value()
		m.config.General.Theme = m.generalInputs[1].input.Value()
		m.config.General.Region = m.generalInputs[2].input.Value()
	}

	// Dashboard
	if len(m.dashboardInputs) >= 4 {
		m.config.Dashboard.AutoPoll = m.dashboardInputs[0].input.Value() == "true"
		if v, err := strconv.Atoi(m.dashboardInputs[1].input.Value()); err == nil {
			m.config.Dashboard.PollIntervalSeconds = v
		}
		m.config.Dashboard.AggregationInterval = m.dashboardInputs[2].input.Value()
		if v, err := strconv.Atoi(m.dashboardInputs[3].input.Value()); err == nil {
			m.config.Dashboard.BufferSize = v
		}
	}
}
