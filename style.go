package main

import (
	"strings"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
)

func getTitleStyle(theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
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

func getProgressBarOptions(theme Theme) []progress.Option {
	return []progress.Option{
		progress.WithColors(theme.Border, theme.Accent),
		progress.WithScaled(true), // gradient scales to fill, not full width
		progress.WithoutPercentage(),
	}
}

func panelHeightFor(termHeight int) int {
	return max(int(float64(termHeight-2)*0.75), sidebarMinHeight)
}

func sidebarStyleFor(termHeight int, theme Theme) lipgloss.Style {
	h := panelHeightFor(termHeight)
	return lipgloss.NewStyle().Width(26).Height(h).
		Border(lipgloss.RoundedBorder()).BorderForeground(theme.Border).Padding(1, 1)
}

func topPanelStyleFor(remainingHeight, availableWidth int, theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Height(remainingHeight).Width(max(availableWidth, 65)).
		Border(lipgloss.RoundedBorder()).BorderForeground(theme.Border).Padding(1, 2)
}

func playingPanelStyleFor(availableWidth int, theme Theme) lipgloss.Style {
	return lipgloss.NewStyle().Width(max(availableWidth, 65)).
		Border(lipgloss.RoundedBorder()).BorderForeground(theme.Border).Padding(1, 2)
	// no Height() — naturally small, bottom, fixed
}

func nameRule(width int) string {
	label := " " + appTitle + " "
	dashes := width - lipgloss.Width(label)
	left := dashes / 2
	right := dashes - left
	return strings.Repeat("⑉", left) + label + strings.Repeat("⑉", right)
}
