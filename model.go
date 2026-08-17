package main

import (
	"time"

	"charm.land/bubbles/v2/help"
	"charm.land/bubbles/v2/list"
	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/charmbracelet/harmonica"
	"github.com/sina96/ytunes/internal/player"
)

type State int

const (
	StateIdle State = iota
	StateLoading
	StatePlaying
	StateStopped
)

type model struct {
	minimal            bool
	playlistMode       bool
	now                time.Time
	tabs               []string
	activeTab          int
	termWidth          int
	termHeight         int
	textInput          textinput.Model
	err                error
	state              State
	player             player.Player
	queue              []player.QueueEntry
	queueIndex         int
	metadata           player.Metadata
	userStopped        bool
	confirmQuitting    bool
	quitArmedAt        time.Time
	loadingSpinner     spinner.Model
	playingSpinner     spinner.Model
	elapsedSeconds     float64
	positionSeq        int
	pausePending       bool
	tickGen            int
	playGen            int
	progress           progress.Model
	help               help.Model
	theme              Theme
	pickingTheme       bool
	themeList          list.Model
	showQueueModeLabel bool
	queueModeShownAt   time.Time
	visualizerBars     []visualizerBar
	visualizerSpring   harmonica.Spring
}

func newTextInput() textinput.Model {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.SetVirtualCursor(false)
	ti.Focus()
	ti.CharLimit = 2048
	ti.SetWidth(60)
	return ti
}

func (m model) resetTextInput() (model, tea.Cmd) {
	m.textInput = newTextInput()
	cmd := m.textInput.Focus()
	m.elapsedSeconds = 0
	return m, tea.Batch(cmd, m.progress.SetPercent(0))
}

func (m model) availableMainWidth() int {
	if m.minimal {
		return m.termWidth // small margin, no sidebar to reserve for
	}
	return m.termWidth - lipgloss.Width(m.renderSideBar())
}
