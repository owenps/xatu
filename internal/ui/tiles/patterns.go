package tiles

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	logpkg "github.com/owen/xatu/internal/log"
	"github.com/owen/xatu/internal/ui/tile"
)

type patternItem struct {
	pattern string
	count   int
	pct     float64
}

func (i patternItem) Title() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#00FF00")).
		Render(fmt.Sprintf("%4.0f%%  %s", i.pct, i.pattern))
}
func (i patternItem) Description() string { return "" }
func (i patternItem) FilterValue() string { return i.pattern }

// Patterns shows the top recurring log message patterns and their frequency.
type Patterns struct {
	list    list.Model
	buffer  *logpkg.Buffer
	total   int
	focused bool
	width   int
	height  int
}

var _ tile.Tile = (*Patterns)(nil)

func NewPatterns(buf *logpkg.Buffer) *Patterns {
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

	return &Patterns{
		list:   l,
		buffer: buf,
	}
}

func (p *Patterns) Init() tea.Cmd { return nil }

func (p *Patterns) Update(msg tea.Msg) (tile.Tile, tea.Cmd) {
	switch msg := msg.(type) {
	case NewLogsMsg:
		p.refreshItems()
		return p, nil
	case tea.KeyMsg:
		if p.focused {
			var cmd tea.Cmd
			p.list, cmd = p.list.Update(msg)
			return p, cmd
		}
	}
	return p, nil
}

func (p *Patterns) View() string {
	footer := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#555555")).
		Render(fmt.Sprintf("from %d logs", p.total))
	return p.list.View() + "\n" + footer
}

func (p *Patterns) SetSize(width, height int) {
	p.width = width
	p.height = height
	p.list.SetSize(width, max(height-2, 1)) // leave room for footer
}

func (p *Patterns) Title() string { return "top patterns" }

func (p *Patterns) GridSize() (int, int) { return 1, 1 }

func (p *Patterns) Focused() bool { return p.focused }

func (p *Patterns) SetFocused(focused bool) {
	p.focused = focused
}

// normalizeMessage strips variable parts (numbers, UUIDs, timestamps) to find patterns.
func normalizeMessage(msg string) string {
	var result strings.Builder
	i := 0
	for i < len(msg) {
		ch := msg[i]
		// Replace runs of digits with <N>
		if ch >= '0' && ch <= '9' {
			result.WriteString("<N>")
			for i < len(msg) && ((msg[i] >= '0' && msg[i] <= '9') || msg[i] == '-' || msg[i] == '.') {
				i++
			}
			continue
		}
		result.WriteByte(ch)
		i++
	}
	s := result.String()
	// Truncate long patterns
	if len(s) > 80 {
		s = s[:77] + "..."
	}
	return s
}

func (p *Patterns) refreshItems() {
	entries := p.buffer.Entries()
	p.total = len(entries)

	counts := make(map[string]int)
	for _, e := range entries {
		pattern := normalizeMessage(e.Message)
		counts[pattern]++
	}

	type kv struct {
		pattern string
		count   int
	}
	var sorted []kv
	for pat, c := range counts {
		sorted = append(sorted, kv{pat, c})
	}
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].count > sorted[j].count
	})

	limit := 15
	if len(sorted) < limit {
		limit = len(sorted)
	}

	items := make([]list.Item, limit)
	for i := range limit {
		pct := 0.0
		if p.total > 0 {
			pct = float64(sorted[i].count) / float64(p.total) * 100
		}
		items[i] = patternItem{
			pattern: sorted[i].pattern,
			count:   sorted[i].count,
			pct:     pct,
		}
	}
	p.list.SetItems(items)
}
