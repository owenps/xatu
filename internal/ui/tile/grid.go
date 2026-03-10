package tile

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	narrowThreshold = 120 // below this, stack tiles vertically
	gridCols        = 3   // max grid columns in wide mode
)

// Grid manages a collection of tiles, their layout, focus, and expand/collapse.
type Grid struct {
	tiles       []Tile
	focusIndex  int
	expanded    bool // true when a tile is in full-screen mode
	width       int
	height      int
	borderStyle lipgloss.Style
	focusStyle  lipgloss.Style
	titleStyle  lipgloss.Style
}

func NewGrid(borderStyle, focusStyle, titleStyle lipgloss.Style) *Grid {
	return &Grid{
		borderStyle: borderStyle,
		focusStyle:  focusStyle,
		titleStyle:  titleStyle,
	}
}

func (g *Grid) AddTile(t Tile) {
	g.tiles = append(g.tiles, t)
	if len(g.tiles) == 1 {
		g.tiles[0].SetFocused(true)
	}
}

func (g *Grid) SetSize(width, height int) {
	g.width = width
	g.height = height
	g.recalcLayout()
}

func (g *Grid) Init() tea.Cmd {
	var cmds []tea.Cmd
	for _, t := range g.tiles {
		if cmd := t.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (g *Grid) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if key.Matches(msg, key.NewBinding(key.WithKeys("tab"))) && !g.expanded {
			return g.cycleFocus()
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) && !g.expanded {
			g.expanded = true
			g.recalcLayout()
			return nil
		}
		if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) && g.expanded {
			// Let the tile handle esc first if it has an inner view (e.g. log detail)
			if g.focusIndex < len(g.tiles) {
				if eh, ok := g.tiles[g.focusIndex].(EscHandler); ok && eh.HandlesEsc() {
					newTile, cmd := g.tiles[g.focusIndex].Update(msg)
					g.tiles[g.focusIndex] = newTile
					return cmd
				}
			}
			g.expanded = false
			g.recalcLayout()
			return nil
		}
	}

	// Delegate to focused tile
	if g.focusIndex < len(g.tiles) {
		newTile, cmd := g.tiles[g.focusIndex].Update(msg)
		g.tiles[g.focusIndex] = newTile
		return cmd
	}

	return nil
}

// UpdateAll sends a message to all tiles (used for data messages like NewLogsMsg).
func (g *Grid) UpdateAll(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	for i, t := range g.tiles {
		newTile, cmd := t.Update(msg)
		g.tiles[i] = newTile
		if cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (g *Grid) View() string {
	if len(g.tiles) == 0 {
		return ""
	}

	// Expanded mode: show only the focused tile
	if g.expanded && g.focusIndex < len(g.tiles) {
		t := g.tiles[g.focusIndex]
		return g.renderTile(t, g.width, g.height)
	}

	// Narrow mode: stack vertically
	if g.width < narrowThreshold {
		return g.renderStacked()
	}

	return g.renderGrid()
}

func (g *Grid) cycleFocus() tea.Cmd {
	if len(g.tiles) == 0 {
		return nil
	}
	g.tiles[g.focusIndex].SetFocused(false)
	g.focusIndex = (g.focusIndex + 1) % len(g.tiles)
	g.tiles[g.focusIndex].SetFocused(true)
	return nil
}

func (g *Grid) recalcLayout() {
	if g.expanded && g.focusIndex < len(g.tiles) {
		// Give full space to expanded tile (minus border)
		g.tiles[g.focusIndex].SetSize(g.width-2, g.height-2)
		return
	}

	if g.width < narrowThreshold {
		g.recalcStacked()
		return
	}

	g.recalcGrid()
}

func (g *Grid) recalcStacked() {
	if len(g.tiles) == 0 {
		return
	}
	tileHeight := g.height / len(g.tiles)
	tileWidth := g.width - 2 // border

	for _, t := range g.tiles {
		t.SetSize(tileWidth, tileHeight-2) // border top+bottom
	}
}

func (g *Grid) recalcGrid() {
	if len(g.tiles) == 0 {
		return
	}

	// Calculate row layout: each tile declares its GridSize (cols, rows)
	// We lay out left-to-right, wrapping when cols exceed gridCols
	type placed struct {
		tile       Tile
		col, row   int
		cols, rows int
	}

	var placements []placed
	curCol := 0
	curRow := 0
	maxRowHeight := 1

	for _, t := range g.tiles {
		cols, rows := t.GridSize()
		if curCol+cols > gridCols {
			curCol = 0
			curRow += maxRowHeight
			maxRowHeight = 1
		}
		placements = append(placements, placed{
			tile: t, col: curCol, row: curRow, cols: cols, rows: rows,
		})
		curCol += cols
		if rows > maxRowHeight {
			maxRowHeight = rows
		}
	}

	totalRows := curRow + maxRowHeight
	if totalRows == 0 {
		return
	}
	colWidth := g.width / gridCols
	rowHeight := g.height / totalRows

	for _, p := range placements {
		w := colWidth*p.cols - 2 // border
		h := rowHeight*p.rows - 2
		if w < 1 {
			w = 1
		}
		if h < 1 {
			h = 1
		}
		p.tile.SetSize(w, h)
	}
}

func (g *Grid) renderStacked() string {
	var rows []string
	for _, t := range g.tiles {
		rows = append(rows, g.renderTile(t, g.width, 0))
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (g *Grid) renderGrid() string {
	if len(g.tiles) == 0 {
		return ""
	}

	// Group tiles by grid row
	type placed struct {
		tile       Tile
		col, row   int
		cols, rows int
	}

	var placements []placed
	curCol := 0
	curRow := 0
	maxRowHeight := 1

	for _, t := range g.tiles {
		cols, rows := t.GridSize()
		if curCol+cols > gridCols {
			curCol = 0
			curRow += maxRowHeight
			maxRowHeight = 1
		}
		placements = append(placements, placed{
			tile: t, col: curCol, row: curRow, cols: cols, rows: rows,
		})
		curCol += cols
		if rows > maxRowHeight {
			maxRowHeight = rows
		}
	}

	colWidth := g.width / gridCols

	// Group by row
	rowMap := make(map[int][]placed)
	for _, p := range placements {
		rowMap[p.row] = append(rowMap[p.row], p)
	}

	var gridRows []string
	maxRow := curRow + maxRowHeight
	for r := 0; r < maxRow; r++ {
		tiles, ok := rowMap[r]
		if !ok {
			continue
		}
		var rowViews []string
		for _, p := range tiles {
			w := colWidth * p.cols
			rowViews = append(rowViews, g.renderTile(p.tile, w, 0))
		}
		gridRows = append(gridRows, lipgloss.JoinHorizontal(lipgloss.Top, rowViews...))
	}

	return lipgloss.JoinVertical(lipgloss.Left, gridRows...)
}

func (g *Grid) renderTile(t Tile, width, height int) string {
	style := g.borderStyle
	if t.Focused() {
		style = g.focusStyle
	}

	title := g.titleStyle.Render(t.Title())

	if width > 0 {
		style = style.Width(width - 2) // subtract border width
	}
	if height > 0 {
		style = style.Height(height - 2)
	}

	content := style.Render(t.View())

	return lipgloss.JoinVertical(lipgloss.Left, title, content)
}
