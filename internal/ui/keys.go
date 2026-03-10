package ui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Help        key.Binding
	Quit        key.Binding
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Tab         key.Binding
	ShiftTab    key.Binding
	Enter       key.Binding
	Escape      key.Binding
	QueryPage   key.Binding
	Home        key.Binding
	Settings    key.Binding
	SubmitQuery key.Binding
	Refresh     key.Binding
}

var Keys = KeyMap{
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Quit: key.NewBinding(
		key.WithKeys("ctrl+c"),
		key.WithHelp("ctrl+c", "quit"),
	),
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑/k", "up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓/j", "down"),
	),
	Left: key.NewBinding(
		key.WithKeys("left"),
		key.WithHelp("←", "left"),
	),
	Right: key.NewBinding(
		key.WithKeys("right"),
		key.WithHelp("→", "right"),
	),
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tile"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "switch context"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Escape: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	QueryPage: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "query"),
	),
	Home: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("h", "home"),
	),
	Settings: key.NewBinding(
		key.WithKeys("cmd+,", "ctrl+,"),
		key.WithHelp("cmd+,", "settings"),
	),
	SubmitQuery: key.NewBinding(
		key.WithKeys("shift+enter"),
		key.WithHelp("shift+enter", "submit query"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}

func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help, k.Tab, k.Enter, k.Escape, k.Refresh, k.QueryPage, k.Home, k.Quit}
}

func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Tab, k.ShiftTab, k.Enter, k.Escape},
		{k.QueryPage, k.Home, k.Settings},
		{k.Help, k.Quit},
	}
}
