package main

import (
	"math/rand"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	visualizerBarPool        = 24 // generously more than any realistic terminal width needs
	visualizerRows           = 4  // how tall the bars render, in text rows
	visualizerTickEvery      = 100 * time.Millisecond
	visualizerRetargetChance = 0.15 // per bar, per tick
)

type visualizerTickMsg struct{}

type visualizerBar struct {
	pos, vel, target float64 // target and pos both live in [0, 1]
}

func tickVisualizer() tea.Cmd {
	return tea.Tick(visualizerTickEvery, func(_ time.Time) tea.Msg {
		return visualizerTickMsg{}
	})
}

func renderVisualizer(bars []visualizerBar, width int, theme Theme) string {
	barCount := max(5, min(width/2, visualizerBarPool))
	visible := bars[:barCount]

	// bottom row (quiet) -> theme.Muted, theme.Accent in between, top row (loud) -> theme.Peak
	gradient := lipgloss.Blend1D(visualizerRows, theme.Muted, theme.Accent, theme.Peak)

	rows := make([]string, visualizerRows)
	for rowFromTop := range rows {
		rowFromBottom := visualizerRows - 1 - rowFromTop

		style := lipgloss.NewStyle().Foreground(gradient[rowFromBottom])
		var line strings.Builder
		for _, bar := range visible {
			filled := bar.pos*float64(visualizerRows) > float64(rowFromBottom)
			if filled {
				line.WriteString("█ ")
			} else {
				line.WriteString("  ")
			}
		}
		rows[rowFromTop] = style.Render(line.String())
	}
	return strings.Join(rows, "\n")
}

func newVisualizerBars() []visualizerBar {
	bars := make([]visualizerBar, visualizerBarPool)
	for i := range bars {
		bars[i].target = rand.Float64()
	}
	return bars
}
