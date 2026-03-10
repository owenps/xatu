package ui

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Name string

	// Base colors
	Primary   lipgloss.Color
	Secondary lipgloss.Color
	Accent    lipgloss.Color
	BG        lipgloss.Color
	FG        lipgloss.Color
	Subtle    lipgloss.Color

	// Log level colors
	Fatal lipgloss.Color
	Error lipgloss.Color
	Warn  lipgloss.Color
	Info  lipgloss.Color
	Debug lipgloss.Color

	// UI element styles
	TileBorder        lipgloss.Style
	TileBorderFocused lipgloss.Style
	TileTitle         lipgloss.Style
	StatusBar         lipgloss.Style
	SelectedItem      lipgloss.Style
}

var DarkTheme = Theme{
	Name:      "dark",
	Primary:   lipgloss.Color("#00FF00"),
	Secondary: lipgloss.Color("#0088FF"),
	Accent:    lipgloss.Color("#FFAA00"),
	BG:        lipgloss.Color("#000000"),
	FG:        lipgloss.Color("#CCCCCC"),
	Subtle:    lipgloss.Color("#555555"),

	Fatal: lipgloss.Color("#AA00FF"),
	Error: lipgloss.Color("#FF0000"),
	Warn:  lipgloss.Color("#FFFF00"),
	Info:  lipgloss.Color("#0088FF"),
	Debug: lipgloss.Color("#888888"),

	TileBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#555555")),

	TileBorderFocused: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#00FF00")),

	TileTitle: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true),

	StatusBar: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Background(lipgloss.Color("#222222")),

	SelectedItem: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true),
}

var LightTheme = Theme{
	Name:      "light",
	Primary:   lipgloss.Color("#007700"),
	Secondary: lipgloss.Color("#0055CC"),
	Accent:    lipgloss.Color("#CC7700"),
	BG:        lipgloss.Color("#FFFFFF"),
	FG:        lipgloss.Color("#333333"),
	Subtle:    lipgloss.Color("#999999"),

	Fatal: lipgloss.Color("#7700AA"),
	Error: lipgloss.Color("#CC0000"),
	Warn:  lipgloss.Color("#CC8800"),
	Info:  lipgloss.Color("#0055CC"),
	Debug: lipgloss.Color("#999999"),

	TileBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#CCCCCC")),

	TileBorderFocused: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#007700")),

	TileTitle: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#007700")).
		Bold(true),

	StatusBar: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#333333")).
		Background(lipgloss.Color("#EEEEEE")),

	SelectedItem: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#007700")).
		Bold(true),
}

var CatppuccinTheme = Theme{
	Name:      "catppuccin",
	Primary:   lipgloss.Color("#A6E3A1"), // green
	Secondary: lipgloss.Color("#89B4FA"), // blue
	Accent:    lipgloss.Color("#F9E2AF"), // yellow
	BG:        lipgloss.Color("#1E1E2E"), // base
	FG:        lipgloss.Color("#CDD6F4"), // text
	Subtle:    lipgloss.Color("#585B70"), // surface2

	Fatal: lipgloss.Color("#CBA6F7"), // mauve
	Error: lipgloss.Color("#F38BA8"), // red
	Warn:  lipgloss.Color("#F9E2AF"), // yellow
	Info:  lipgloss.Color("#89B4FA"), // blue
	Debug: lipgloss.Color("#6C7086"), // overlay0

	TileBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#585B70")),

	TileBorderFocused: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#A6E3A1")),

	TileTitle: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A6E3A1")).
		Bold(true),

	StatusBar: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CDD6F4")).
		Background(lipgloss.Color("#313244")),

	SelectedItem: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#A6E3A1")).
		Bold(true),
}

var EmberTheme = Theme{
	Name:      "ember",
	Primary:   lipgloss.Color("#FF8C00"),
	Secondary: lipgloss.Color("#CC5500"),
	Accent:    lipgloss.Color("#FFAA33"),
	BG:        lipgloss.Color("#000000"),
	FG:        lipgloss.Color("#CCCCCC"),
	Subtle:    lipgloss.Color("#554400"),

	Fatal: lipgloss.Color("#FF0000"),
	Error: lipgloss.Color("#FF3300"),
	Warn:  lipgloss.Color("#FFAA00"),
	Info:  lipgloss.Color("#FF8C00"),
	Debug: lipgloss.Color("#665500"),

	TileBorder: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#554400")),

	TileBorderFocused: lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#FF8C00")),

	TileTitle: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF8C00")).
		Bold(true),

	StatusBar: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#CCCCCC")).
		Background(lipgloss.Color("#1A1000")),

	SelectedItem: lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF8C00")).
		Bold(true),
}

func GetTheme(name string) Theme {
	switch name {
	case "dark":
		return DarkTheme
	case "light":
		return LightTheme
	case "catppuccin":
		return CatppuccinTheme
	case "ember":
		return EmberTheme
	default:
		return DarkTheme
	}
}
