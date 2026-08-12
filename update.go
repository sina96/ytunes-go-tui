package main

import (
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/bubbles/v2/spinner"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/sina96/ytunes-go-tui/internal/player"
)

type metadataFetchedMsg struct {
	metadata player.Metadata
	err      error
}
type playbackFinishedMsg struct{ err error }

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

// positionTickMsg drives the on-screen elapsed-time display on a fixed,
// local cadence — independent of how long any given mpv IPC round-trip
// takes, so the display stays smooth even when a query is briefly slow.
type positionTickMsg time.Time

func tickPosition() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return positionTickMsg(t)
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

func (m model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, tickClock())
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.themeList.SetSize(msg.Width, msg.Height)
		return m, nil

	case clockTickMsg:
		m.now = time.Time(msg)
		return m, tickClock()
	case spinner.TickMsg:
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
	case positionTickMsg:
		if m.state != StatePlaying || m.player.IsPaused() || m.pausePending {
			return m, nil
		}
		m.elapsedSeconds++
		pcmd := m.progress.SetPercent(m.elapsedSeconds / float64(m.metadata.DurationSeconds))
		m.positionSeq++
		return m, tea.Batch(pcmd, tickPosition(), queryPosition(m, m.positionSeq))
	case positionMsg:
		if msg.seq != m.positionSeq || msg.err != nil {
			// stale (an older query resolving after a newer one already
			// superseded it) or failed — the local tick already covers
			// this second regardless, just skip the correction
			return m, nil
		}
		if m.state != StatePlaying || m.player.IsPaused() || m.pausePending {
			// arrived after we've since paused/stopped, or while a
			// pause/resume toggle is still in flight — don't let a late
			// correction move the frozen display out from under it
			return m, nil
		}
		m.elapsedSeconds = msg.seconds
		return m, m.progress.SetPercent(msg.seconds / float64(m.metadata.DurationSeconds))
	case pauseToggleMsg:
		m.pausePending = false
		if msg.err != nil {
			return m, nil
		}
		if m.state == StatePlaying && !m.player.IsPaused() {
			m.positionSeq++
			return m, tea.Batch(tickPosition(), queryPosition(m, m.positionSeq), m.playingSpinner.Tick)
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
	case metadataFetchedMsg:
		if msg.err != nil {
			m.err = msg.err
			m.state = StateIdle
			m, cmd = m.resetTextInput()
			return m, cmd
		}
		m.metadata = msg.metadata

		m.state = StatePlaying
		return m, tea.Batch(m.playingSpinner.Tick, tickPosition(),
			func() tea.Msg {
				err := m.player.Wait()
				if err != nil {
					return playbackFinishedMsg{err}
				}
				return playbackFinishedMsg{nil}
			})
	case playbackFinishedMsg:
		if m.userStopped {
			m.userStopped = false
		} else if msg.err != nil {
			m.err = msg.err
		}
		m.state = StateStopped
		m, cmd = m.resetTextInput()
		return m, cmd
	case tea.KeyPressMsg:

		if m.pickingTheme {
			key := msg.String()
			if key == "enter" {
				m.theme = Themes[m.themeList.Index()].Theme
				m.progress = progress.New(getProgressBarOptions(m.theme)...)
				m.pickingTheme = false
				return m, nil
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
			if m.state == StatePlaying {
				return m, nil
			}
			m.state = StateLoading
			m.textInput.Blur()
			url := m.textInput.Value()
			return m, tea.Batch(m.loadingSpinner.Tick, func() tea.Msg {
				meta, err := m.player.Play(url)
				if err != nil {
					return metadataFetchedMsg{err: err}
				}
				return metadataFetchedMsg{metadata: meta}
			})
		case "ctrl+v", "ctrl+p":
			m.confirmQuitting = false
			if m.state != StatePlaying {
				return m, textinput.Paste
			}
			return m, nil
		case "ctrl+c":
			if m.confirmQuitting {
				if m.state == StatePlaying {
					m.player.Stop()
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
		case "esc":
			m.confirmQuitting = false
			if m.state == StatePlaying {
				m.player.Stop()
				m.userStopped = true
			}
			return m, nil
		case "ctrl+t":
			m.pickingTheme = true
			return m, nil
		default:
			m.confirmQuitting = false
		}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}
