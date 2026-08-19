package main

import (
	"math"
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
)

func getTitleStyle(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
}

func getBorderStyle(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Bold(false).Foreground(theme.Border)
}

func getLabelStyle(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Muted)
}

func getMutedLabelStyle(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Muted)
}

func getErrorStyle(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Error)
}

func getStatusStyle(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
}

func getTabStyle(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Foreground(theme.Muted)
}

func getActiveTabStyle(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Background(theme.Accent).Foreground(theme.Gray)
}

func centerTextStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().Width(width).Align(lipgloss.Center)
}

func getProgressBarOptions() []progress.Option {
	return []progress.Option{
		progress.WithSpringOptions(18.0, 1.0), // frequency, damping — tune fill snappiness
	}
}

// renderProgressBar draws a "[=======>---]" style bar
func renderProgressBar(p progress.Model, theme Theme, width int) string {
	innerWidth := width - 2 // reserve for the [ and ] brackets
	if innerWidth < 1 {
		return strings.Repeat("=", max(width, 0))
	}

	percent := p.Percent()
	filled := int(math.Round(float64(innerWidth) * percent))
	filled = max(0, min(innerWidth, filled))

	blend := lipgloss.Blend1D(max(filled, 1), theme.Muted, theme.Accent, theme.Peak)

	var b strings.Builder
	b.WriteString("[")
	for i := 0; i < filled; i++ {
		char := "="
		if i == filled-1 && filled < innerWidth {
			char = ">"
		}
		b.WriteString(lipgloss.NewStyle().Foreground(blend[i]).Render(char))
	}
	b.WriteString(strings.Repeat("-", innerWidth-filled))
	b.WriteString("]")
	return b.String()
}

func panelHeightFor(termHeight int) int {
	return max(termHeight-2, sidebarMinHeight)
}

func sidebarStyleFor(termHeight int, theme Theme) lipgloss.Style {
	h := panelHeightFor(termHeight) - 1 // -1: compensates for the topEdge row
	return lipgloss.NewStyle().Width(sidebarWidth).Height(h).
		Border(lipgloss.RoundedBorder(), false, true, true, true).
		BorderForeground(theme.Border).Padding(1, 1)
}

func topPanelStyleFor(remainingHeight, availableWidth int, theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Height(remainingHeight).Width(max(availableWidth, 65)).
		Border(lipgloss.RoundedBorder()).BorderForeground(theme.Border).Padding(1, 1)
}

func playingPanelStyleFor(availableWidth int, theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Width(max(availableWidth, 65)).
		Border(lipgloss.RoundedBorder()).BorderForeground(theme.Border).Padding(1, 1)
}

func minimalBoxStyleFor(width, height int, theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().
		Width(width).Height(height).
		Border(lipgloss.RoundedBorder(), false, true, true, true).
		BorderForeground(theme.Border).
		Padding(0, 1)
}

func labeledTopEdge(border lipgloss.Border, label string, width int) string {
	labelStr := " " + label + " "
	dashes := max(width-lipgloss.Width(labelStr)-lipgloss.Width(border.TopLeft)-lipgloss.Width(border.TopRight), 0)
	left := dashes / 2
	right := dashes - left
	return border.TopLeft + strings.Repeat(border.Top, left) + labelStr + strings.Repeat(border.Top, right) + border.TopRight
}

func renderLabeledBox(label string, width, height int, theme Theme, content string) string {
	border := lipgloss.RoundedBorder()
	topEdge := getBorderStyle(theme).Render(labeledTopEdge(border, label, width))
	box := minimalBoxStyleFor(width, height, theme).Render(content)
	return lipgloss.JoinVertical(lipgloss.Left, topEdge, box)
}
