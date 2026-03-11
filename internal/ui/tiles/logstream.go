package tiles

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/owenps/xatu/internal/aws"
	logpkg "github.com/owenps/xatu/internal/log"
	"github.com/owenps/xatu/internal/ui/tile"
)

// NewLogsMsg is sent when new log entries are fetched.
type NewLogsMsg struct {
	Entries []aws.LogEntry
}

// LogStream displays log entries in a scrollable table.
type LogStream struct {
	table      table.Model
	buffer     *logpkg.Buffer
	focused    bool
	width      int
	height     int
	showDetail bool
	detail     viewport.Model
	detailIdx  int
}

var _ tile.Tile = (*LogStream)(nil)

var levelColors = map[aws.LogLevel]lipgloss.Color{
	aws.LevelFatal:   lipgloss.Color("#AA00FF"),
	aws.LevelError:   lipgloss.Color("#FF0000"),
	aws.LevelWarn:    lipgloss.Color("#FFFF00"),
	aws.LevelInfo:    lipgloss.Color("#0088FF"),
	aws.LevelDebug:   lipgloss.Color("#888888"),
	aws.LevelUnknown: lipgloss.Color("#CCCCCC"),
}

func NewLogStream(buf *logpkg.Buffer) *LogStream {
	columns := []table.Column{
		{Title: "TIME", Width: 12},
		{Title: "LEVEL", Width: 7},
		{Title: "GROUP", Width: 20},
		{Title: "MESSAGE", Width: 58},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithFocused(true),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("#555555")).
		BorderBottom(true).
		Bold(true).
		Foreground(lipgloss.Color("#00FF00"))
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("#00FF00")).
		Bold(true)
	t.SetStyles(s)

	return &LogStream{
		table:  t,
		buffer: buf,
	}
}

func (l *LogStream) Init() tea.Cmd {
	return nil
}

func (l *LogStream) Update(msg tea.Msg) (tile.Tile, tea.Cmd) {
	switch msg := msg.(type) {
	case NewLogsMsg:
		// Parse levels and attributes, append to buffer
		for i := range msg.Entries {
			msg.Entries[i].Level = logpkg.ParseLevel(msg.Entries[i].Message)
			msg.Entries[i].Attributes = logpkg.ExtractKV(msg.Entries[i].Message)
		}
		l.buffer.Append(msg.Entries)
		l.refreshRows()
		return l, nil

	case tea.KeyMsg:
		if l.focused {
			if l.showDetail {
				if key.Matches(msg, key.NewBinding(key.WithKeys("esc"))) {
					l.showDetail = false
					return l, nil
				}
				var cmd tea.Cmd
				l.detail, cmd = l.detail.Update(msg)
				return l, cmd
			}
			if key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
				entries := l.buffer.Entries()
				cursor := l.table.Cursor()
				if cursor >= 0 && cursor < len(entries) {
					l.showDetail = true
					l.detailIdx = cursor
					l.detail = viewport.New(l.width, l.height)
					l.detail.SetContent(l.renderDetail(entries[cursor]))
				}
				return l, nil
			}
			var cmd tea.Cmd
			l.table, cmd = l.table.Update(msg)
			return l, cmd
		}
	}

	return l, nil
}

func (l *LogStream) View() string {
	if l.showDetail {
		return l.detail.View()
	}
	return l.table.View()
}

func (l *LogStream) SetSize(width, height int) {
	l.width = width
	l.height = height
	l.table.SetWidth(width)
	l.table.SetHeight(height)
	l.recalcColumns()
	l.refreshRows()
}

func (l *LogStream) Title() string {
	return fmt.Sprintf("log stream (%d)", l.buffer.Len())
}

func (l *LogStream) GridSize() (int, int) {
	return 3, 1 // full width, 1 row
}

func (l *LogStream) Focused() bool {
	return l.focused
}

func (l *LogStream) HandlesEsc() bool {
	return l.showDetail
}

func (l *LogStream) SetFocused(focused bool) {
	l.focused = focused
	l.table.Focus()
	if !focused {
		l.table.Blur()
	}
}

func (l *LogStream) refreshRows() {
	entries := l.buffer.Entries()
	rows := make([]table.Row, len(entries))

	for i, e := range entries {
		rows[i] = table.Row{
			e.Timestamp.Format(time.TimeOnly),
			e.Level.String(),
			truncate(e.LogGroup, 20),
			truncate(e.Message, l.msgWidth()),
		}
	}

	l.table.SetRows(rows)

	// Scroll to bottom to show latest
	if len(rows) > 0 {
		l.table.GotoBottom()
	}
}

func (l *LogStream) recalcColumns() {
	msgW := l.msgWidth()
	if msgW < 10 {
		msgW = 10
	}

	l.table.SetColumns([]table.Column{
		{Title: "TIME", Width: 12},
		{Title: "LEVEL", Width: 7},
		{Title: "GROUP", Width: 20},
		{Title: "MESSAGE", Width: msgW},
	})
}

func (l *LogStream) msgWidth() int {
	// TIME(12) + LEVEL(7) + GROUP(20) + padding(8) = 47
	w := l.width - 47
	if w < 10 {
		w = 10
	}
	return w
}

func (l *LogStream) renderDetail(entry aws.LogEntry) string {
	green := lipgloss.Color("#00FF00")
	white := lipgloss.Color("#CCCCCC")
	subtle := lipgloss.Color("#555555")

	titleStyle := lipgloss.NewStyle().Foreground(green).Bold(true)
	labelStyle := lipgloss.NewStyle().Foreground(subtle).Width(16)
	valueStyle := lipgloss.NewStyle().Foreground(white)
	msgStyle := lipgloss.NewStyle().Foreground(white).Width(l.width - 4)

	levelColor := levelColors[entry.Level]

	var b strings.Builder
	b.WriteString(titleStyle.Render("Log Entry Detail") + "\n\n")
	b.WriteString(labelStyle.Render("Timestamp:") + " " + valueStyle.Render(entry.Timestamp.Format(time.RFC3339)) + "\n")
	b.WriteString(labelStyle.Render("Level:") + " " + lipgloss.NewStyle().Foreground(levelColor).Render(entry.Level.String()) + "\n")
	b.WriteString(labelStyle.Render("Log Group:") + " " + valueStyle.Render(entry.LogGroup) + "\n")
	b.WriteString(labelStyle.Render("Log Stream:") + " " + valueStyle.Render(entry.LogStream) + "\n")
	b.WriteString("\n" + titleStyle.Render("Message") + "\n\n")
	b.WriteString(msgStyle.Render(entry.Message) + "\n")

	if len(entry.Attributes) > 0 {
		b.WriteString("\n" + titleStyle.Render("Attributes") + "\n\n")
		// Sort keys for stable output
		keys := make([]string, 0, len(entry.Attributes))
		for k := range entry.Attributes {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			b.WriteString("  " + labelStyle.Render(k+":") + " " + valueStyle.Render(entry.Attributes[k]) + "\n")
		}
	}

	b.WriteString("\n" + lipgloss.NewStyle().Foreground(subtle).Render("esc: back to log stream"))

	return b.String()
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 3 {
		return s[:max]
	}
	return s[:max-3] + "..."
}
