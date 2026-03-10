package tiles

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	logpkg "github.com/owen/xatu/internal/log"
	"github.com/owen/xatu/internal/ui/tile"
)

type attrItem struct {
	key   string
	count int
}

func (i attrItem) Title() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).
		Render(fmt.Sprintf("%-30s %d", i.key, i.count))
}
func (i attrItem) Description() string { return "" }
func (i attrItem) FilterValue() string { return i.key }

// Attributes shows the top attribute keys and their counts.
type Attributes struct {
	list    list.Model
	buffer  *logpkg.Buffer
	focused bool
	width   int
	height  int
}

var _ tile.Tile = (*Attributes)(nil)

func NewAttributes(buf *logpkg.Buffer) *Attributes {
	delegate := list.NewDefaultDelegate()
	delegate.ShowDescription = false
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#00FF00")).
		BorderForeground(lipgloss.Color("#00FF00"))

	l := list.New(nil, delegate, 30, 10)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetFilteringEnabled(false)

	return &Attributes{
		list:   l,
		buffer: buf,
	}
}

func (a *Attributes) Init() tea.Cmd { return nil }

func (a *Attributes) Update(msg tea.Msg) (tile.Tile, tea.Cmd) {
	switch msg := msg.(type) {
	case NewLogsMsg:
		a.refreshItems()
		return a, nil
	case tea.KeyMsg:
		if a.focused {
			var cmd tea.Cmd
			a.list, cmd = a.list.Update(msg)
			return a, cmd
		}
	}
	return a, nil
}

func (a *Attributes) View() string {
	return a.list.View()
}

func (a *Attributes) SetSize(width, height int) {
	a.width = width
	a.height = height
	a.list.SetSize(width, height)
}

func (a *Attributes) Title() string { return "top attributes" }

func (a *Attributes) GridSize() (int, int) { return 1, 1 }

func (a *Attributes) Focused() bool { return a.focused }

func (a *Attributes) SetFocused(focused bool) {
	a.focused = focused
}

func (a *Attributes) refreshItems() {
	entries := a.buffer.Entries()
	counts := make(map[string]int)

	for _, e := range entries {
		for k := range e.Attributes {
			counts[k]++
		}
	}

	type kv struct {
		key   string
		count int
	}
	var sorted []kv
	for k, c := range counts {
		sorted = append(sorted, kv{k, c})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	limit := 20
	if len(sorted) < limit {
		limit = len(sorted)
	}

	items := make([]list.Item, limit)
	for i := range limit {
		items[i] = attrItem{key: sorted[i].key, count: sorted[i].count}
	}
	a.list.SetItems(items)
}
