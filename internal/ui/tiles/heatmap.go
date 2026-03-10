package tiles

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/owen/xatu/internal/aws"
	logpkg "github.com/owen/xatu/internal/log"
	"github.com/owen/xatu/internal/ui/tile"
)

var shades = []string{" ", "░", "▒", "▓", "█"}

// severity rows in display order (top to bottom)
var severityRows = []aws.LogLevel{
	aws.LevelFatal,
	aws.LevelError,
	aws.LevelWarn,
	aws.LevelInfo,
	aws.LevelDebug,
}

var severityColors = map[aws.LogLevel]lipgloss.Color{
	aws.LevelFatal: lipgloss.Color("#AA00FF"),
	aws.LevelError: lipgloss.Color("#FF0000"),
	aws.LevelWarn:  lipgloss.Color("#FFFF00"),
	aws.LevelInfo:  lipgloss.Color("#0088FF"),
	aws.LevelDebug: lipgloss.Color("#888888"),
}

// HeatMap renders a severity × time heat map using dithered block characters.
type HeatMap struct {
	buffer     *logpkg.Buffer
	counts     map[aws.LogLevel]map[int64]int // level -> bucket timestamp -> count
	maxCount   int
	bucketSize time.Duration
	focused    bool
	width      int
	height     int
}

var _ tile.Tile = (*HeatMap)(nil)

func NewHeatMap(buf *logpkg.Buffer, bucketSize time.Duration) *HeatMap {
	return &HeatMap{
		buffer:     buf,
		counts:     make(map[aws.LogLevel]map[int64]int),
		bucketSize: bucketSize,
	}
}

func (h *HeatMap) Init() tea.Cmd {
	return nil
}

func (h *HeatMap) Update(msg tea.Msg) (tile.Tile, tea.Cmd) {
	switch msg := msg.(type) {
	case NewLogsMsg:
		h.ingestEntries(msg.Entries)
	}
	return h, nil
}

func (h *HeatMap) View() string {
	if h.width < 10 || h.height < 3 {
		return ""
	}

	now := time.Now()
	labelWidth := 6 // "FATAL " etc.
	cellsAvailable := (h.width - labelWidth) / 2
	if cellsAvailable < 1 {
		cellsAvailable = 1
	}

	// Time range: latest N buckets that fit
	latestBucket := now.Truncate(h.bucketSize).UnixMilli()
	startBucket := latestBucket - int64(cellsAvailable-1)*h.bucketSize.Milliseconds()

	var rows []string

	for _, level := range severityRows {
		color := severityColors[level]
		style := lipgloss.NewStyle().Foreground(color)

		label := style.Bold(true).Render(fmt.Sprintf("%-6s", level.String()))

		var cells strings.Builder
		bucketCounts, ok := h.counts[level]

		for i := range cellsAvailable {
			bucketTs := startBucket + int64(i)*h.bucketSize.Milliseconds()
			count := 0
			if ok {
				count = bucketCounts[bucketTs]
			}
			shade := h.countToShade(count)
			cells.WriteString(style.Render(shade + shade))
		}

		rows = append(rows, label+cells.String())
	}

	// Legend
	legendStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#555555"))
	legend := legendStyle.Render(fmt.Sprintf("  █ high  ▓ med  ▒ low  ░ min  · none  [%s buckets]", h.bucketSize))
	rows = append(rows, legend)

	// Time axis
	if cellsAvailable > 10 {
		startTime := time.UnixMilli(startBucket).Format("15:04")
		endTime := time.UnixMilli(latestBucket).Format("15:04")
		gap := (cellsAvailable*2 + labelWidth) - len(startTime) - len(endTime)
		if gap < 1 {
			gap = 1
		}
		timeAxis := legendStyle.Render(
			strings.Repeat(" ", labelWidth) + startTime + strings.Repeat(" ", gap) + endTime,
		)
		rows = append(rows, timeAxis)
	}

	return strings.Join(rows, "\n")
}

func (h *HeatMap) SetSize(width, height int) {
	h.width = width
	h.height = height
}

func (h *HeatMap) Title() string {
	return "severity heat map"
}

func (h *HeatMap) GridSize() (int, int) {
	return 2, 1 // 2 columns wide, 1 row tall
}

func (h *HeatMap) Focused() bool {
	return h.focused
}

func (h *HeatMap) SetFocused(focused bool) {
	h.focused = focused
}

func (h *HeatMap) ingestEntries(entries []aws.LogEntry) {
	for _, e := range entries {
		level := logpkg.ParseLevel(e.Message)
		bucket := e.Timestamp.Truncate(h.bucketSize).UnixMilli()

		if h.counts[level] == nil {
			h.counts[level] = make(map[int64]int)
		}
		h.counts[level][bucket]++

		count := h.counts[level][bucket]
		if count > h.maxCount {
			h.maxCount = count
		}
	}
}

func (h *HeatMap) countToShade(count int) string {
	if count == 0 {
		return " "
	}
	if h.maxCount == 0 {
		return shades[1]
	}

	// Map count to shade index 1-4
	ratio := float64(count) / float64(h.maxCount)
	switch {
	case ratio >= 0.75:
		return shades[4]
	case ratio >= 0.5:
		return shades[3]
	case ratio >= 0.25:
		return shades[2]
	default:
		return shades[1]
	}
}
