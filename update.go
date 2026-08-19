package main

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/mosaic" //experimental

	"github.com/sina96/ytunes/internal/player"
)

type metadataFetchedMsg struct {
	metadata         player.Metadata
	queue            []player.QueueEntry
	err              error
	isNewPlayRequest bool // true for a fresh "enter" URL, false for playQueueEntry navigation — must replace m.queue even with nil
}
type playbackFinishedMsg struct {
	err error
	seq int
}

type quitTimeoutMsg struct {
	armedAt time.Time
}

type positionMsg struct {
	seconds float64
	err     error
	seq     int
}

type pauseToggleMsg struct {
	err error
}

type themeSavedMsg struct {
	err error
}

// positionTickMsg positionTickMsg now carries the generation
type positionTickMsg struct {
	seq int
}

type queueModeHideMsg struct {
	shownAt time.Time
}

const playbackStartPollInterval = 250 * time.Millisecond // fast poll while waiting for real playback to start, vs. the normal 1s cadence

func tickPosition(gen int, interval time.Duration) tea.Cmd {
	return tea.Tick(interval, func(t time.Time) tea.Msg {
		return positionTickMsg{seq: gen}
	})
}

func (m model) playQueueEntry(index int) (model, tea.Cmd) {
	if index < 0 || index >= len(m.queue) {
		return m, nil
	}
	m.queueIndex = index
	m.state = StateLoading
	m.elapsedSeconds = 0
	m.playGen++
	m.tickGen++

	return m, tea.Batch(m.loadingSpinner.Tick, m.progress.SetPercent(0), func() tea.Msg {
		_ = m.player.Stop()
		meta, err := m.player.Play("", index, m.queue)
		if err != nil {
			return metadataFetchedMsg{err: err}
		}
		return metadataFetchedMsg{metadata: meta, queue: m.queue}
	})
}

// queryPosition asks mpv for its real position, to correct any drift the
// local tick-based increment accumulates (buffering, seeking, etc.) — it
// does not drive the display's schedule, only its accuracy. seq tags this
// specific query so a stale reply (an older query resolving after a newer
// one) can be told apart and discarded rather than applied out of order.
func queryPosition(m model, seq int) tea.Cmd {
	return func() tea.Msg {
		seconds, err := m.player.Position()
		return positionMsg{seconds: seconds, err: err, seq: seq}
	}
}

func hideQueueModeLabel(shownAt time.Time) tea.Cmd {
	return tea.Tick(time.Second, func(_ time.Time) tea.Msg {
		return queueModeHideMsg{shownAt: shownAt}
	})
}

func (m model) handleVisualizerTickMsg(_ visualizerTickMsg) (model, tea.Cmd) {
	if m.state != StatePlaying {
		return m, nil
	}
	paused := m.player.IsPaused() || m.pausePending
	for i := range m.visualizerBars {
		bar := &m.visualizerBars[i]
		if paused {
			bar.target = 0
		} else if rand.Float64() < visualizerRetargetChance {
			bar.target = rand.Float64()
		}
		bar.pos, bar.vel = m.visualizerSpring.Update(bar.pos, bar.vel, bar.target)
		if math.Abs(bar.pos-bar.target) < 0.01 && math.Abs(bar.vel) < 0.01 {
			bar.pos, bar.vel = bar.target, 0
		}
	}
	return m, tickVisualizer()
}

func (m model) handleSpinnerTickMsg(msg spinner.TickMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	cmds := make([]tea.Cmd, 0, 2)
	if m.state == StateLoading || (m.state == StatePlaying && m.pausePending) {
		m.loadingSpinner, cmd = m.loadingSpinner.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.state == StatePlaying && !m.player.IsPaused() {
		var pcmd tea.Cmd
		m.playingSpinner, pcmd = m.playingSpinner.Update(msg)
		cmds = append(cmds, pcmd)
	}
	return m, tea.Batch(cmds...)
}

func (m model) handlePositionTickMsg(msg positionTickMsg) (model, tea.Cmd) {
	if msg.seq != m.tickGen {
		return m, nil
	}

	if m.state != StatePlaying || m.player.IsPaused() || m.pausePending {
		return m, nil
	}

	if m.awaitingPlaybackStart {
		m.positionSeq++
		return m, tea.Batch(tickPosition(m.tickGen, playbackStartPollInterval), queryPosition(m, m.positionSeq))
	}

	m.elapsedSeconds++
	var pcmd tea.Cmd
	if m.metadata.DurationSeconds > 0 {
		pcmd = m.progress.SetPercent(m.elapsedSeconds / float64(m.metadata.DurationSeconds))
	}
	m.positionSeq++
	return m, tea.Batch(pcmd, tickPosition(m.tickGen, time.Second), queryPosition(m, m.positionSeq))
}

func (m model) handleMetadataFetchedMsg(msg metadataFetchedMsg) (model, tea.Cmd) {
	var cmd tea.Cmd
	if msg.isNewPlayRequest {
		m.queue = msg.queue // even nil — a fresh direct URL replaces any leftover playlist queue
		m.queueIndex = 0
	} else if msg.queue != nil {
		m.queue = msg.queue
	}
	if msg.err != nil {
		if m.queue != nil {
			if m.queueIndex+1 < len(m.queue) {
				return m.playQueueEntry(m.queueIndex + 1)
			}
		}
		m.err = msg.err
		m.state = StateIdle
		m, cmd = m.resetTextInput()
		return m, cmd
	}
	m.metadata = msg.metadata
	m.state = StatePlaying
	m.awaitingPlaybackStart = true
	m.elapsedSeconds = 0
	m.tickGen++
	m.playGen++
	gen := m.tickGen
	playGen := m.playGen
	return m, tea.Batch(m.playingSpinner.Tick, tickPosition(gen, playbackStartPollInterval), tickVisualizer(),
		func() tea.Msg {
			err := m.player.Wait()
			if err != nil {
				return playbackFinishedMsg{err, playGen}
			}
			return playbackFinishedMsg{nil, playGen}
		})
}

func (m model) handlePlaybackFinished(msg playbackFinishedMsg) (model, tea.Cmd) {
	var cmd tea.Cmd
	if msg.seq != m.playGen {
		return m, nil
	}
	if m.userStopped {
		m.userStopped = false
	} else if msg.err != nil {
		m.err = msg.err
	} else if m.queueIndex+1 < len(m.queue) {
		return m.playQueueEntry(m.queueIndex + 1)
	}
	m.queue = nil
	m.queueIndex = 0
	m.state = StateStopped
	m, cmd = m.resetTextInput()
	return m, cmd

}

func (m model) handleKeyPress(msg tea.KeyPressMsg) (model, tea.Cmd) {
	var cmd tea.Cmd

	if m.pickingTheme {
		key := msg.String()
		if key == "enter" {
			selected := Themes[m.themeList.Index()]
			m.theme = selected.Theme
			m.progress = progress.New(getProgressBarOptions()...)
			m.pickingTheme = false
			return m, func() tea.Msg {
				err := saveTheme(selected.Name)
				return themeSavedMsg{err}
			}
		}

		if key == "esc" {
			m.pickingTheme = false
			return m, nil
		}
		m.themeList, cmd = m.themeList.Update(msg)
		return m, cmd
	}

	switch msg.String() {
	case "enter":
		//playing logic
		m.confirmQuitting = false
		m.err = nil
		if m.state == StatePlaying || m.state == StateLoading {
			return m, nil
		}

		if m.textInput.Value() == "" {
			return m, nil
		}

		m.state = StateLoading
		m.textInput.Blur()
		url := m.textInput.Value()
		return m, tea.Batch(m.loadingSpinner.Tick, func() tea.Msg {
			var queue []player.QueueEntry
			var err error
			if m.playlistMode {
				queue, err = m.player.ResolveQueue(url)
				if err != nil {
					return metadataFetchedMsg{err: err, isNewPlayRequest: true}
				}
			}

			meta, err := m.player.Play(url, 0, queue)
			if err != nil {
				return metadataFetchedMsg{err: err, isNewPlayRequest: true}
			}
			return metadataFetchedMsg{metadata: meta, queue: queue, isNewPlayRequest: true}
		})
	case "ctrl+v", "ctrl+p":
		m.confirmQuitting = false
		if m.state != StatePlaying {
			return m, textinput.Paste
		}
		return m, nil
	case "ctrl+c":
		if m.state == StateLoading {
			return m, nil
		}

		if m.confirmQuitting {
			if m.state == StatePlaying {
				if err := m.player.Stop(); err != nil {
					m.err = err
				}
			}
			return m, tea.Quit
		}
		m.confirmQuitting = true
		timeQuitArmedAt := time.Now()
		m.quitArmedAt = timeQuitArmedAt
		return m, tea.Tick(5*time.Second, func(_ time.Time) tea.Msg {
			return quitTimeoutMsg{armedAt: timeQuitArmedAt}
		})
	case "space":
		m.confirmQuitting = false

		if m.state == StatePlaying && !m.pausePending {
			m.pausePending = true
			wasPaused := m.player.IsPaused()
			return m, tea.Batch(m.loadingSpinner.Tick, func() tea.Msg {
				var err error
				if wasPaused {
					err = m.player.Resume()
				} else {
					err = m.player.Pause()
				}
				return pauseToggleMsg{err: err}
			})
		}
		return m, nil
	case "right", "l":
		if m.state == StatePlaying && m.playlistMode {
			if m.queueIndex+1 < len(m.queue) {
				return m.playQueueEntry(m.queueIndex + 1)
			}
			return m, nil
		}
	case "left", "h":
		if m.state == StatePlaying && m.playlistMode {
			if m.queueIndex > 0 {
				return m.playQueueEntry(m.queueIndex - 1)
			}
			return m, nil
		}
	case "esc":
		m.confirmQuitting = false
		if m.state == StatePlaying {
			if err := m.player.Stop(); err != nil {
				m.err = err
			}
			m.userStopped = true
		}
		return m, nil
	case "ctrl+t":
		m.pickingTheme = true
		return m, nil
	case "ctrl+q":
		m.playlistMode = !m.playlistMode
		m.showQueueModeLabel = true
		m.queueModeShownAt = time.Now()
		return m, hideQueueModeLabel(m.queueModeShownAt)
	default:
		m.confirmQuitting = false
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (model, tea.Cmd) {
	m.termWidth = msg.Width
	m.termHeight = msg.Height
	m.themeList.SetSize(msg.Width, msg.Height)
	if m.sidebarImageSrc != nil {
		sidebarStyle := sidebarStyleFor(m.termHeight, m.theme)
		w := getContentWidth(sidebarStyle)

		headerContent := lipgloss.JoinVertical(lipgloss.Center,
			getTitleStyle(m.theme).Render(strings.TrimSpace(logo)),
			getTitleStyle(m.theme).Render(appTitle),
			getLabelStyle(m.theme).Render(appVersion))
		headerHeight := lipgloss.Height(headerContent)

		availableForImage := sidebarStyle.GetHeight() - sidebarStyle.GetVerticalFrameSize() - headerHeight - 1
		squareHeight := w / 2
		h := min(squareHeight, max(availableForImage, 0))

		if w != m.sidebarImageWidth || h != m.sidebarImageHeight {
			newMosaic := mosaic.New().Width(w * 2).Height(h * 2)
			m.sidebarImage = newMosaic.Render(m.sidebarImageSrc)
			m.sidebarImageWidth = w
			m.sidebarImageHeight = h
		}
	}
	return m, nil
}

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tickClock(), checkForUpdate())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case updateCheckMsg:
		if msg.tag != "" && msg.tag != appVersion {
			m.latestVersion = msg.tag
			return m, nil
		}

	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)

	case queueModeHideMsg:
		if msg.shownAt.Equal(m.queueModeShownAt) {
			m.showQueueModeLabel = false
			return m, nil
		}
		return m, nil

	case themeSavedMsg:
		if msg.err != nil {
			m.err = fmt.Errorf("could not save theme: %w", msg.err)
		}
		return m, nil

	case clockTickMsg:
		m.now = time.Time(msg)
		return m, tickClock()

	case spinner.TickMsg:
		return m.handleSpinnerTickMsg(msg)

	case positionTickMsg:
		return m.handlePositionTickMsg(msg)

	case positionMsg:
		if msg.seq != m.positionSeq {
			return m, nil // stale
		}
		if m.state != StatePlaying || m.player.IsPaused() || m.pausePending {
			return m, nil // paused
		}
		if m.awaitingPlaybackStart {
			if msg.err != nil {
				return m, nil // mpv hasn't loaded the file yet, keep polling
			}
			m.awaitingPlaybackStart = false
		} else if msg.err != nil {
			return m, nil // stale/transient IPC error
		}
		m.elapsedSeconds = msg.seconds
		var pcmd tea.Cmd
		if m.metadata.DurationSeconds > 0 {
			pcmd = m.progress.SetPercent(msg.seconds / float64(m.metadata.DurationSeconds))
		}
		return m, pcmd

	case pauseToggleMsg:
		m.pausePending = false
		if msg.err != nil {
			return m, nil
		}
		if m.state == StatePlaying && !m.player.IsPaused() {
			m.tickGen++
			m.positionSeq++
			m.awaitingPlaybackStart = true // resume has the same async gap as initial buffering — poll fast, don't tick optimistically
			return m, tea.Batch(tickPosition(m.tickGen, playbackStartPollInterval), queryPosition(m, m.positionSeq), m.playingSpinner.Tick)
		}
		return m, nil

	case progress.FrameMsg:
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd

	case quitTimeoutMsg:
		if msg.armedAt.Equal(m.quitArmedAt) {
			m.confirmQuitting = false
		}
		return m, nil

	case visualizerTickMsg:
		return m.handleVisualizerTickMsg(msg)

	case metadataFetchedMsg:
		return m.handleMetadataFetchedMsg(msg)

	case playbackFinishedMsg:
		return m.handlePlaybackFinished(msg)

	case tea.KeyPressMsg:
		return m.handleKeyPress(msg)
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}
