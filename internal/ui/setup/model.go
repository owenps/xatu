package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	xaws "github.com/owenps/xatu/internal/aws"
	"github.com/owenps/xatu/internal/config"
	"github.com/owenps/xatu/internal/ui/shared"
)

// AWS regions
var awsRegions = []string{
	"us-east-1", "us-east-2", "us-west-1", "us-west-2",
	"af-south-1",
	"ap-east-1", "ap-south-1", "ap-south-2",
	"ap-southeast-1", "ap-southeast-2", "ap-southeast-3", "ap-southeast-4",
	"ap-northeast-1", "ap-northeast-2", "ap-northeast-3",
	"ca-central-1", "ca-west-1",
	"eu-central-1", "eu-central-2",
	"eu-west-1", "eu-west-2", "eu-west-3",
	"eu-south-1", "eu-south-2", "eu-north-1",
	"il-central-1",
	"me-south-1", "me-central-1",
	"sa-east-1",
}

type step int

const (
	stepRegion step = iota
	stepDiscovering
	stepSelectGroups
	stepContextName
	stepSaving
	stepDone
)

// Styles
var (
	green      = lipgloss.Color("#00FF00")
	dimGreen   = lipgloss.Color("#007700")
	white      = lipgloss.Color("#CCCCCC")
	red        = lipgloss.Color("#FF0000")
	titleStyle = lipgloss.NewStyle().Foreground(green).Bold(true).MarginBottom(1)
	subtitleStyle = lipgloss.NewStyle().Foreground(white)
	errorStyle    = lipgloss.NewStyle().Foreground(red)
	hintStyle     = lipgloss.NewStyle().Foreground(dimGreen)
)

func themedDelegate() list.DefaultDelegate {
	delegate := themedDelegate()
	delegate.Styles.NormalDesc = delegate.Styles.NormalDesc.Foreground(white)
	delegate.Styles.DimmedTitle = delegate.Styles.DimmedTitle.Foreground(dimGreen)
	delegate.Styles.DimmedDesc = delegate.Styles.DimmedDesc.Foreground(dimGreen)
	delegate.Styles.FilterMatch = delegate.Styles.FilterMatch.Foreground(green)
	return delegate
}

// regionItem for the region selector list.
type regionItem string

func (i regionItem) Title() string       { return string(i) }
func (i regionItem) Description() string { return "" }
func (i regionItem) FilterValue() string { return string(i) }

// logGroupItem for the log group multi-select list.
type logGroupItem struct {
	name     string
	selected bool
}

func (i logGroupItem) Title() string {
	check := lipgloss.NewStyle().Foreground(dimGreen).Render("[ ] ")
	if i.selected {
		check = lipgloss.NewStyle().Foreground(green).Render("[✓] ")
	}
	return check + lipgloss.NewStyle().Foreground(green).Render(i.name)
}
func (i logGroupItem) Description() string { return "" }
func (i logGroupItem) FilterValue() string { return i.name }

// Messages
type discoverGroupsMsg struct {
	groups []xaws.LogGroup
	err    error
	client *xaws.Client
}

type configSavedMsg struct {
	err error
}

// SetupCompleteMsg is sent when setup finishes successfully.
type SetupCompleteMsg struct {
	Config *config.Config
}

type Model struct {
	step       step
	region     string
	client     *xaws.Client
	groups     []xaws.LogGroup
	selected   map[int]bool
	regionList list.Model
	groupList  list.Model
	nameInput  textinput.Model
	err        error
	width      int
	height     int
}

func New(client *xaws.Client, region string) Model {
	ti := textinput.New()
	ti.Placeholder = "my-context"
	ti.CharLimit = 64
	ti.PromptStyle = lipgloss.NewStyle().Foreground(green)
	ti.TextStyle = lipgloss.NewStyle().Foreground(green)

	// Build region list
	items := make([]list.Item, len(awsRegions))
	for i, r := range awsRegions {
		items[i] = regionItem(r)
	}
	delegate := themedDelegate()

	regionList := list.New(items, delegate, 40, 20)
	regionList.Title = "Select AWS Region"
	regionList.Styles.Title = titleStyle
	regionList.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(green)
	regionList.Styles.FilterCursor = lipgloss.NewStyle().Foreground(green)
	regionList.Styles.DefaultFilterCharacterMatch = lipgloss.NewStyle().Foreground(green)
	regionList.Styles.ActivePaginationDot = lipgloss.NewStyle().Foreground(green)
	regionList.Styles.InactivePaginationDot = lipgloss.NewStyle().Foreground(dimGreen)
	regionList.SetFilteringEnabled(true)
	regionList.SetShowStatusBar(false)

	// Pre-select the current region if set
	if region != "" {
		for i, r := range awsRegions {
			if r == region {
				regionList.Select(i)
				break
			}
		}
	}

	return Model{
		step:       stepRegion,
		region:     region,
		client:     client,
		selected:   make(map[int]bool),
		regionList: regionList,
		nameInput:  ti,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.regionList.SetSize(msg.Width-4, msg.Height-12)
		if m.step == stepSelectGroups {
			m.groupList.SetSize(msg.Width-4, msg.Height-6)
		}
		return m, nil

	case tea.KeyMsg:
		// Back navigation with esc
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
			switch m.step {
			case stepSelectGroups:
				m.step = stepRegion
				return m, nil
			case stepContextName:
				m.step = stepSelectGroups
				m.nameInput.Blur()
				return m, nil
			}
		}

		if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
			return m, tea.Quit
		}

		switch m.step {
		case stepRegion:
			if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
				selected := m.regionList.SelectedItem()
				if selected != nil {
					m.region = string(selected.(regionItem))
					m.step = stepDiscovering
					return m, m.discoverGroups()
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.regionList, cmd = m.regionList.Update(msg)
			return m, cmd

		case stepSelectGroups:
			if key.Matches(msg, key.NewBinding(key.WithKeys(" "))) {
				idx := m.groupList.Index()
				m.selected[idx] = !m.selected[idx]
				items := m.groupList.Items()
				if idx < len(items) {
					item := items[idx].(logGroupItem)
					item.selected = m.selected[idx]
					cmd := m.groupList.SetItem(idx, item)
					return m, cmd
				}
				return m, nil
			}
			if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
				if m.hasSelections() {
					m.step = stepContextName
					m.nameInput.Focus()
					return m, textinput.Blink
				}
				return m, nil
			}
			var cmd tea.Cmd
			m.groupList, cmd = m.groupList.Update(msg)
			return m, cmd

		case stepContextName:
			if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
				name := m.nameInput.Value()
				if name == "" {
					name = "default"
				}
				m.step = stepSaving
				return m, m.saveConfig(name)
			}
			var cmd tea.Cmd
			m.nameInput, cmd = m.nameInput.Update(msg)
			return m, cmd
		}

	case discoverGroupsMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = stepRegion
			return m, nil
		}
		if len(msg.groups) == 0 {
			m.err = fmt.Errorf("no log groups found in region %s", m.region)
			m.step = stepRegion
			return m, nil
		}
		m.client = msg.client
		m.groups = msg.groups
		m.step = stepSelectGroups

		items := make([]list.Item, len(m.groups))
		for i, g := range m.groups {
			items[i] = logGroupItem{name: g.Name}
		}
		delegate := themedDelegate()

		m.groupList = list.New(items, delegate, m.width-4, m.height-6)
		m.groupList.Title = "Select log groups"
		m.groupList.Styles.Title = titleStyle
		m.groupList.Styles.FilterPrompt = lipgloss.NewStyle().Foreground(green)
		m.groupList.Styles.FilterCursor = lipgloss.NewStyle().Foreground(green)
		m.groupList.Styles.DefaultFilterCharacterMatch = lipgloss.NewStyle().Foreground(green)
		m.groupList.Styles.ActivePaginationDot = lipgloss.NewStyle().Foreground(green)
		m.groupList.Styles.InactivePaginationDot = lipgloss.NewStyle().Foreground(dimGreen)
		m.groupList.SetShowStatusBar(true)
		m.groupList.SetFilteringEnabled(true)
		return m, nil

	case configSavedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.step = stepContextName
			m.nameInput.Focus()
			return m, textinput.Blink
		}
		m.step = stepDone
		cfg := m.buildConfig()
		return m, func() tea.Msg { return SetupCompleteMsg{Config: cfg} }
	}

	return m, nil
}

func (m Model) View() string {
	switch m.step {
	case stepRegion:
		banner := shared.Banner(green, dimGreen, m.width)
		hint := hintStyle.Render("  ↕ navigate  / filter  enter select")
		var errMsg string
		if m.err != nil {
			errMsg = "\n" + errorStyle.Render(fmt.Sprintf("  Error: %v", m.err))
		}
		return fmt.Sprintf("\n%s\n%s\n%s\n%s", banner, hint, errMsg, m.regionList.View())

	case stepDiscovering:
		return fmt.Sprintf("\n  %s\n", titleStyle.Render("Discovering log groups..."))

	case stepSelectGroups:
		count := m.selectionCount()
		hint := hintStyle.Render(fmt.Sprintf("  space toggle  enter confirm (%d selected)  esc back", count))
		return fmt.Sprintf("\n%s\n%s", hint, m.groupList.View())

	case stepContextName:
		header := titleStyle.Render("Name your context")
		hint := hintStyle.Render("  enter confirm  esc back")
		return fmt.Sprintf("\n  %s\n%s\n\n  %s\n", header, hint, m.nameInput.View())

	case stepSaving:
		return subtitleStyle.Render("\n  Saving configuration...\n")

	case stepDone:
		return titleStyle.Render("\n  Setup complete! Loading dashboard...\n")
	}

	return ""
}

func (m Model) hasSelections() bool {
	for _, v := range m.selected {
		if v {
			return true
		}
	}
	return false
}

func (m Model) selectionCount() int {
	count := 0
	for _, v := range m.selected {
		if v {
			count++
		}
	}
	return count
}

func (m Model) discoverGroups() tea.Cmd {
	region := m.region
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		client, err := xaws.NewClient(ctx, region)
		if err != nil {
			return discoverGroupsMsg{err: err}
		}
		groups, err := client.DiscoverLogGroups(ctx)
		return discoverGroupsMsg{groups: groups, err: err, client: client}
	}
}

func (m Model) buildConfig() *config.Config {
	cfg := config.DefaultConfig()
	cfg.General.Region = m.region

	var logGroups []string
	for idx, sel := range m.selected {
		if sel && idx < len(m.groups) {
			logGroups = append(logGroups, m.groups[idx].Name)
		}
	}

	name := m.nameInput.Value()
	if name == "" {
		name = "default"
	}

	cfg.Contexts = []config.Context{{
		Name:      name,
		LogGroups: logGroups,
		Colour:    "#00FF00",
	}}

	return cfg
}

func (m Model) saveConfig(name string) tea.Cmd {
	cfg := m.buildConfig()
	cfg.Contexts[0].Name = name
	return func() tea.Msg {
		err := config.Save(cfg)
		return configSavedMsg{err: err}
	}
}
