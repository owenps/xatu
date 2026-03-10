package ui

import (
	"fmt"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

type StatusBar struct {
	help        help.Model
	keys        KeyMap
	contextName string
	width       int
	theme       Theme
}

func NewStatusBar(theme Theme, contextName string) StatusBar {
	h := help.New()
	h.ShortSeparator = "  "
	return StatusBar{
		help:        h,
		keys:        Keys,
		contextName: contextName,
		theme:       theme,
	}
}

func (s *StatusBar) SetWidth(w int) {
	s.width = w
	s.help.Width = w - 40
}

func (s *StatusBar) SetContext(name string) {
	s.contextName = name
}

func (s StatusBar) View() string {
	return s.ViewWithState(false, "", time.Time{}, nil)
}

func (s StatusBar) ViewWithState(fetching bool, spinnerView string, lastFetch time.Time, lastErr error) string {
	ctx := lipgloss.NewStyle().
		Foreground(s.theme.Primary).
		Bold(true).
		Render("[" + s.contextName + "]")

	// Status indicator
	var status string
	if fetching {
		status = lipgloss.NewStyle().
			Foreground(s.theme.Primary).
			Render(spinnerView + " fetching")
	} else if lastErr != nil {
		status = lipgloss.NewStyle().
			Foreground(s.theme.Error).
			Render(fmt.Sprintf("err: %v", lastErr))
	} else if !lastFetch.IsZero() {
		ago := time.Since(lastFetch).Truncate(time.Second)
		status = lipgloss.NewStyle().
			Foreground(s.theme.Subtle).
			Render(fmt.Sprintf("updated %s ago", ago))
	}

	helpView := s.help.View(s.keys)

	left := ctx + "  " + status
	right := helpView

	// If there's room, put help on the right
	gap := s.width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 2 {
		gap = 2
	}
	padding := lipgloss.NewStyle().Width(gap).Render("")

	bar := s.theme.StatusBar.
		Width(s.width).
		Render(left + padding + right)

	return bar
}
