package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// splitView renders two content areas with a horizontal divider.
type splitView struct {
	width     int
	topHeight int
	botHeight int
}

func newSplitView(width, totalHeight, splitPct int) splitView {
	topH := totalHeight * splitPct / 100
	botH := max(totalHeight-topH-1, 0) // 1 for divider
	return splitView{
		width:     width,
		topHeight: topH,
		botHeight: botH,
	}
}

func (sv splitView) render(top, bottom string, theme Theme) string {
	topStyle := lipgloss.NewStyle().Width(sv.width).Height(sv.topHeight)
	botStyle := lipgloss.NewStyle().Width(sv.width).Height(sv.botHeight)
	divider := lipgloss.NewStyle().
		Foreground(theme.PaneBorder).
		Render(strings.Repeat("─", sv.width))

	return lipgloss.JoinVertical(lipgloss.Left,
		topStyle.Render(top),
		divider,
		botStyle.Render(bottom),
	)
}
