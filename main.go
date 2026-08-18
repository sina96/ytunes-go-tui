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
	"github.com/charmbracelet/harmonica"
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

	themeList := newThemeList(0, 0, "Pick a theme")
	theme := loadSavedTheme()

	player := player.New()
	loadingSpinner := spinner.New(spinner.WithSpinner(spinner.Dot))
	playingSpinner := spinner.New(spinner.WithSpinner(spinner.Pulse))

	progress := progress.New(getProgressBarOptions()...)

	tabs := []string{}
	tabs = append(tabs, "ytb-play")

	help := help.New()

	visualizerBars := newVisualizerBars()
	visualizerSpring := harmonica.NewSpring(harmonica.FPS(10), 6.0, 1.0) //angularFrequency/dampingRatio

	img, _ := decodeSidebarImage()
	sidebarImageSrc := img

	return model{
		minimal:          minimal,
		now:              time.Now(),
		tabs:             tabs,
		activeTab:        0,
		textInput:        ti,
		player:           player,
		loadingSpinner:   loadingSpinner,
		playingSpinner:   playingSpinner,
		progress:         progress,
		help:             help,
		theme:            theme,
		themeList:        themeList,
		visualizerBars:   visualizerBars,
		visualizerSpring: visualizerSpring,
		sidebarImageSrc:  sidebarImageSrc,
	}
}

func main() {
	minimal := flag.Bool("minimal", false, "hide some features")
	version := flag.Bool("version", false, "print version and exit")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "ytunes — a retro-terminal YouTube audio player\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n  ytunes [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if *version {
		fmt.Println(appVersion)
		os.Exit(0)
	}

	p := tea.NewProgram(initialModel(*minimal))
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
