# Changelog

One-liner summary of the biggest thing each development phase added.

## v0.1.1 (latest)

- Playlist & radio support: paste a playlist/radio URL with "Queue
  Mode" on (`ctrl+q`, off by default) to build a queue with Next/Prev
  and auto-advance; a track that fails to play is skipped instead of
  ending the whole queue.
- mpv process lifecycle hardening: fixed an orphaned mpv process when
  skipping tracks, plus a fallback YouTube player-client list that
  reduces (not eliminates) playback 403s from YouTube's current
  anti-bot enforcement.
- Terminal window/tab title now reflects live state (`y𝕋unes`,
  `y𝕋unes • Playing`, `y𝕋unes • Paused`).
- Theme choice now persists across restarts (`~/.config/ytunes/theme`).

## v0.1.0

- Real mpv IPC: live playback position and clean pause/resume, no more
  audible `SIGSTOP` artifact.
- Theming system: five built-in themes (Default, Terminal-aligned,
  Catppuccin, Gruvbox, Dracula) with an in-app picker (`ctrl+t`).
- Unified hints/footer via `bubbles/help` — one consistent look across
  every screen.
- Redesign: split main panel into a Top Panel + shared Playing Panel,
  reused across full and compact (`--minimal`) modes.
- Responsive layout — sidebar and panels react to real terminal size.
- Mouse click-to-focus support for the URL input.
- Sidebar + tab-strip redesign, `--minimal` compact mode, full
  alternate-screen rendering.
- Local elapsed-time display and progress bar against the track's
  known duration.
- Loading spinner and indeterminate progress bar while a URL resolves.
- Real `lipgloss` styling (borders, colors, padding) replacing plain
  string layout.
- Code reorganized into an `internal` package and multiple files.
- `Player` interface decouples the UI from mpv; clipboard paste added
  cleanly on top of it.
- Quit confirmation now expires automatically if not confirmed quickly.
- Custom playback screen — pause/stop in-app instead of handing the
  terminal to mpv.
- Idle / Playing / Stopped state machine with double-tap quit
  confirmation.
- First full Bubble Tea TUI: URL input, play, quit.
- Clipboard-based URL detection (later replaced by the TUI's own
  input field).
- Plays a YouTube URL from the command line via mpv.
