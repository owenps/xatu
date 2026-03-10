package tile

import tea "github.com/charmbracelet/bubbletea"

// Tile is the interface all dashboard tiles must implement.
type Tile interface {
	Init() tea.Cmd
	Update(msg tea.Msg) (Tile, tea.Cmd)
	View() string
	SetSize(width, height int)
	Title() string
	GridSize() (cols, rows int) // e.g. (1,1), (2,1), (1,2)
	Focused() bool
	SetFocused(bool)
}

// EscHandler is an optional interface for tiles that have inner views
// (e.g. log detail) and want to handle esc before the grid collapses.
type EscHandler interface {
	HandlesEsc() bool
}
