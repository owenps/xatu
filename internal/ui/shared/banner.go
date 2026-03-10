package shared

import "github.com/charmbracelet/lipgloss"

// Banner renders "XATU" in large dithered block letters using Unicode shade characters.
// If width is too small, falls back to just "X".
func Banner(primary, secondary lipgloss.Color, width int) string {
	// Each letter is defined as a grid of intensity values:
	// 0 = empty, 1 = ░ light, 2 = ▒ medium, 3 = ▓ dark, 4 = █ full
	x := [][]int{
		{4, 3, 0, 0, 3, 4},
		{3, 4, 2, 2, 4, 3},
		{0, 2, 4, 4, 2, 0},
		{0, 2, 4, 4, 2, 0},
		{3, 4, 2, 2, 4, 3},
		{4, 3, 0, 0, 3, 4},
	}

	a := [][]int{
		{0, 2, 4, 4, 2, 0},
		{2, 4, 3, 3, 4, 2},
		{4, 3, 0, 0, 3, 4},
		{4, 4, 4, 4, 4, 4},
		{4, 3, 0, 0, 3, 4},
		{4, 2, 0, 0, 2, 4},
	}

	t := [][]int{
		{4, 4, 4, 4, 4, 4},
		{3, 2, 4, 4, 2, 3},
		{0, 0, 4, 4, 0, 0},
		{0, 0, 4, 4, 0, 0},
		{0, 0, 4, 4, 0, 0},
		{0, 0, 3, 3, 0, 0},
	}

	u := [][]int{
		{4, 3, 0, 0, 3, 4},
		{4, 3, 0, 0, 3, 4},
		{4, 3, 0, 0, 3, 4},
		{4, 3, 0, 0, 3, 4},
		{3, 4, 2, 2, 4, 3},
		{0, 3, 4, 4, 3, 0},
	}

	// Each letter is 6 cols * 2 chars = 12 wide, plus 2 gap = 14 per letter
	// Full "XATU" = 4*12 + 3*2 = 54 chars wide
	// Just "X" = 12 chars wide
	var letters [][][]int
	if width >= 54 {
		letters = [][][]int{x, a, t, u}
	} else {
		letters = [][][]int{x}
	}

	shades := []string{" ", "░", "▒", "▓", "█"}

	result := ""
	for row := range 6 {
		line := ""
		for li, letter := range letters {
			for _, val := range letter[row] {
				line += shades[val] + shades[val]
			}
			if li < len(letters)-1 {
				line += "  "
			}
		}
		result += line + "\n"
	}

	style := lipgloss.NewStyle().
		Foreground(primary).
		Bold(true)

	return style.Render(result)
}
