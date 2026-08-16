package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"

	tea "charm.land/bubbletea/v2"
	"github.com/sina96/ytunes/internal/player"
)

// Tea model + view + Update

type clockTickMsg time.Time

func tickClock() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return clockTickMsg(t)
	})
}

func initialModel(minimal bool) model {
	ti := newTextInput()

	player := player.New()
	loadingSpinner := spinner.New(spinner.WithSpinner(spinner.Dot))
	playingSpinner := spinner.New(spinner.WithSpinner(spinner.Pulse))

	progress := progress.New(getProgressBarOptions(ThemeDefault)...)

	tabs := []string{}
	tabs = append(tabs, "ytb-play")

	help := help.New()

	themeList := newThemeList(0, 0, "Pick a theme")
	theme := loadSavedTheme()

	return model{
		minimal:        minimal,
		now:            time.Now(),
		tabs:           tabs,
		activeTab:      0,
		textInput:      ti,
		player:         player,
		loadingSpinner: loadingSpinner,
		playingSpinner: playingSpinner,
		progress:       progress,
		help:           help,
		theme:          theme,
		themeList:      themeList,
	}
}

func main() {
	minimal := flag.Bool("minimal", false, "hide some features")
	flag.Parse()
	p := tea.NewProgram(initialModel(*minimal))
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
