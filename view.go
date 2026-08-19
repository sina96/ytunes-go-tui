package main

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/help"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	utils "github.com/sina96/ytunes/internal/utils"
)

func (m model) windowTitle() string {
	if m.state != StatePlaying {
		return windowTitleBase
	}
	if m.player.IsPaused() {
		return windowTitleBase + " • Paused"
	}
	return windowTitleBase + " • Playing"
}

func (m model) renderHints(keymap help.KeyMap) string {
	hints := m.help.View(keymap)
	if m.showQueueModeLabel {
		hints += getMutedLabelStyle(m.theme).Render(queueModeLabel(m.playlistMode))
	}
	return hints
}

func (m model) confirmQuitLabel() string {
	if !m.confirmQuitting {
		return ""
	}
	return getMutedLabelStyle(m.theme).Render("  " + labelConfirmQuit)
}

func queueModeLabel(on bool) string {
	str := "(queue mode is now: "
	if on {
		return str + "ON)"
	}
	return str + "OFF)"
}

func (m model) topPanelFooter(keymap help.KeyMap) string {
	hints := m.renderHints(keymap)
	if m.minimal {
		return ""
	}

	hints = hints + m.confirmQuitLabel()
	return hints
}

func (m model) renderPlayingPanel() string {
	width := getContentWidth(playingPanelStyleFor(m.availableMainWidth(), m.theme))

	confirmQuitLabelForCompact := ""
	if m.confirmQuitting && m.minimal {
		confirmQuitLabelForCompact = getMutedLabelStyle(m.theme).Render(labelConfirmQuit)
	}

	if m.err != nil {
		return centerTextStyle(width).Render(getErrorStyle(m.theme).Render(fmt.Sprintf("We had some trouble: %v", m.err)))
	}

	switch m.state {
	case StateLoading:
		return centerTextStyle(width).Render(lipgloss.JoinVertical(
			lipgloss.Center, m.loadingSpinner.View(), confirmQuitLabelForCompact,
		))
	case StatePlaying:
		durationInfo := ""
		if m.metadata.Duration != "" {
			elapsedSecs := min(int(m.elapsedSeconds), m.metadata.DurationSeconds)
			elapsed := utils.FormatDuration(elapsedSecs)
			total := utils.FormatDuration(m.metadata.DurationSeconds)
			durationInfo = elapsed + " / " + total
		}

		m.progress.SetWidth(width)
		// progressBarView := m.progress.View() // old bracket-less block-fill rendering
		progressBarView := renderProgressBar(m.progress, m.theme, width)

		titleLine := ""
		if m.minimal && m.metadata.Title != "" {
			titleLine = getLabelStyle(m.theme).Render(utils.TruncateTitle(m.metadata.Title, width))
		}

		pauseplay := getMutedLabelStyle(m.theme).Render("▶⏸")
		if m.pausePending {
			pauseplay = m.loadingSpinner.View()
		}
		stop := getMutedLabelStyle(m.theme).Render("⏹")
		shutdown := getMutedLabelStyle(m.theme).Render("⏻")
		playerBtns := lipgloss.JoinHorizontal(lipgloss.Left, stop, "  ", pauseplay, "  ", shutdown)

		return centerTextStyle(width).Render(lipgloss.JoinVertical(lipgloss.Center, progressBarView, titleLine, durationInfo, playerBtns, confirmQuitLabelForCompact))
	case StateStopped:
		stoppedLabel := ""
		if m.metadata.Title != "" {
			stoppedLabel = getMutedLabelStyle(m.theme).Render(labelLastPlayedPrefix + utils.TruncateTitle(m.metadata.Title, width))
		} else {
			stoppedLabel = getMutedLabelStyle(m.theme).Render(labelIdlePlaceholder)
		}
		return centerTextStyle(width).Render(lipgloss.JoinVertical(lipgloss.Center, stoppedLabel, confirmQuitLabelForCompact))
	case StateIdle:
		updateNotice := ""
		if m.latestVersion != "" {
			updateNotice = getMutedLabelStyle(m.theme).Render("Update available: " + m.latestVersion)
		}
		return centerTextStyle(width).Render(lipgloss.JoinVertical(lipgloss.Center,
			getMutedLabelStyle(m.theme).Render(labelIdlePlaceholder), updateNotice, confirmQuitLabelForCompact))
	default:
		return centerTextStyle(width).Render(lipgloss.JoinVertical(lipgloss.Center, getMutedLabelStyle(m.theme).Render(labelIdlePlaceholder), confirmQuitLabelForCompact))
	}
}

func getContentWidth(style lipgloss.Style) int {
	return style.GetWidth() - style.GetHorizontalFrameSize()
}

func (m model) renderTabStrip() string {
	labels := make([]string, len(m.tabs))
	for i, tab := range m.tabs {
		label := fmt.Sprintf("[%d:%s]", i+1, tab)
		style := getTabStyle(m.theme)
		if i == m.activeTab {
			style = getActiveTabStyle(m.theme)
		}
		labels[i] = style.Render(label)
	}
	return strings.Join(labels, " ")
}

func (m model) renderSidebar(playingPanelStartRow int) string {
	logoBox := getTitleStyle(m.theme).Render(strings.TrimSpace(logo))
	titleBox := getTitleStyle(m.theme).Render(appTitle)
	versionText := getLabelStyle(m.theme).Render(appVersion)
	top := lipgloss.JoinVertical(lipgloss.Center, logoBox, titleBox, versionText)

	sidebarStyle := sidebarStyleFor(m.termHeight, m.theme)
	body := top
	if m.state == StatePlaying && m.sidebarImage != "" {
		spacerLines := max(playingPanelStartRow-1-lipgloss.Height(top), 0) // -1: sidebar's own box content starts 1 row below screenthe topEdge label row
		maxSpacer := max(
			sidebarStyle.GetHeight()-sidebarStyle.GetVerticalFrameSize()-lipgloss.Height(top)-m.sidebarImageHeight-1, 0)
		spacerLines = min(spacerLines, maxSpacer)
		body = top + strings.Repeat("\n", spacerLines+1) + m.sidebarImage
	}

	dateAndTime := strings.ToLower(m.now.Format("Mon 2 Jan.")) + " " + m.now.Format("15:04") // from Phase 31
	topEdge := getBorderStyle(m.theme).Render(labeledTopEdge(lipgloss.RoundedBorder(), dateAndTime, sidebarWidth))
	box := sidebarStyle.Align(lipgloss.Center).Render(body)
	return lipgloss.JoinVertical(lipgloss.Left, topEdge, box)
}

func (m model) renderFrame(topContent, bottomExtra string) tea.View {

	if m.minimal {
		return renderMinimalTopPanel(m, topContent)
	}

	tabStrip := m.renderTabStrip()
	fullTop := lipgloss.JoinVertical(lipgloss.Left, tabStrip, topContent)
	playingRendered := playingPanelStyleFor(m.availableMainWidth(), m.theme).Render(m.renderPlayingPanel())

	panelStyle := topPanelStyleFor(0, m.availableMainWidth(), m.theme)

	naturalTopHeight := lipgloss.Height(fullTop) + lipgloss.Height(bottomExtra) + panelStyle.GetVerticalFrameSize()
	remaining := max(panelHeightFor(m.termHeight)-lipgloss.Height(playingRendered), naturalTopHeight)

	finalPanelStyle := topPanelStyleFor(remaining, m.availableMainWidth(), m.theme)
	innerHeight := remaining - finalPanelStyle.GetVerticalFrameSize()
	innerWidth := getContentWidth(finalPanelStyle)

	fullTopContent := fullTop
	if bottomExtra != "" {
		leftover := max(innerHeight-lipgloss.Height(fullTop), 0)
		placed := lipgloss.Place(innerWidth, leftover, lipgloss.Center, lipgloss.Bottom, bottomExtra)
		fullTopContent = lipgloss.JoinVertical(lipgloss.Left, fullTop, placed)
	}
	topRendered := finalPanelStyle.Render(fullTopContent)
	sideBar := m.renderSidebar(lipgloss.Height(topRendered))
	stackedColumn := lipgloss.JoinVertical(lipgloss.Left, topRendered, playingRendered)

	fullContent := lipgloss.JoinHorizontal(lipgloss.Top, sideBar, stackedColumn)
	v := tea.NewView(fullContent)
	v.AltScreen = true
	v.WindowTitle = m.windowTitle()
	return v
}

func renderMinimalTopPanel(m model, content string) tea.View {
	width := max(m.availableMainWidth(), 40)

	topLabel := appTitle
	if len(m.tabs) > 0 {
		topLabel = topLabel + ":" + m.tabs[m.activeTab]
	}

	bottomLabel := labelPlayingPanelMinimalTopBorder
	if m.err != nil {
		bottomLabel = labelErrorPanelMinimalTopBorder
	}
	bottomContent := m.renderPlayingPanel()

	frameSize := minimalBoxStyleFor(0, 0, m.theme).GetVerticalFrameSize()
	topBoxHeight := lipgloss.Height(content) + frameSize
	bottomBoxHeight := max(panelHeightFor(m.termHeight)-topBoxHeight-2, lipgloss.Height(bottomContent)+frameSize)

	topBox := renderLabeledBox(topLabel, width, topBoxHeight, m.theme, content)
	bottomBox := renderLabeledBox(bottomLabel, width, bottomBoxHeight, m.theme, bottomContent)

	stacked := lipgloss.JoinVertical(lipgloss.Left, topBox, bottomBox)
	v := tea.NewView(stacked)
	v.AltScreen = true
	v.WindowTitle = m.windowTitle()
	return v
}

func (m model) themePickerView() tea.View {
	m.themeList.Styles.Title = getTitleStyle(m.theme)
	content := lipgloss.JoinVertical(lipgloss.Top, m.themeList.View())
	v := tea.NewView(content)
	v.AltScreen = true
	v.WindowTitle = m.windowTitle()
	return v
}

func (m model) View() tea.View {

	if m.pickingTheme {
		return m.themePickerView()
	}

	if m.state == StateLoading {
		return m.loadingView()
	}

	if m.state == StatePlaying {
		return m.playingView()
	}

	if m.state == StateStopped {
		return m.stoppedView()
	}

	if m.err != nil {
		return m.errorView()
	}

	return m.idleView()

}

func (m model) idleView() tea.View {
	header := labelYtbPlayer
	label := getLabelStyle(m.theme).Render(labelEnterURL)
	footerStyled := m.topPanelFooter(idleKeys)
	title := getTitleStyle(m.theme).Render(header)
	textInputView := m.textInput.View()

	content := lipgloss.JoinVertical(lipgloss.Top, title, label, textInputView, footerStyled)
	panelStyle := topPanelStyleFor(panelHeightFor(m.termHeight), m.availableMainWidth(), m.theme)
	if m.minimal {
		panelStyle = minimalBoxStyleFor(m.availableMainWidth(), panelHeightFor(m.termHeight), m.theme)
	}
	c := m.setCursor(header, panelStyle, label)
	v := m.renderFrame(content, "")
	v.Cursor = c
	return v
}

func (m model) errorView() tea.View {
	header := labelYtbPlayer
	label := getLabelStyle(m.theme).Render(labelEnterURL)
	footerStyled := m.topPanelFooter(idleKeys)
	textInputView := m.textInput.View()
	title := getTitleStyle(m.theme).Render(header)

	content := lipgloss.JoinVertical(lipgloss.Top, title, label, textInputView, footerStyled)

	panelStyle := topPanelStyleFor(panelHeightFor(m.termHeight), m.availableMainWidth(), m.theme)
	if m.minimal {
		panelStyle = minimalBoxStyleFor(m.availableMainWidth(), panelHeightFor(m.termHeight), m.theme)
	}
	c := m.setCursor(header, panelStyle, label)
	v := m.renderFrame(content, "")
	v.Cursor = c
	return v
}

func (m model) stoppedView() tea.View {
	header := labelYtbPlayer
	playAnother := ""
	if m.err == nil {
		playAnother = getLabelStyle(m.theme).Render(labelPlayAnother)
	}
	title := getTitleStyle(m.theme).Render(header)
	footerStyled := m.topPanelFooter(idleKeys)
	textInputView := m.textInput.View()

	content := lipgloss.JoinVertical(lipgloss.Top, title, playAnother, textInputView, footerStyled)
	panelStyle := topPanelStyleFor(panelHeightFor(m.termHeight), m.availableMainWidth(), m.theme)
	if m.minimal {
		panelStyle = minimalBoxStyleFor(m.availableMainWidth(), panelHeightFor(m.termHeight), m.theme)
	}
	c := m.setCursor(header, panelStyle, playAnother)
	v := m.renderFrame(content, "")
	v.Cursor = c
	return v
}

func (m model) playingView() tea.View {
	statusView := "\n" + m.playingSpinner.View() + " Playing\n"
	upNextLine := ""
	if m.queueIndex+1 < len(m.queue) {
		upNextLine = "\n" + getMutedLabelStyle(m.theme).Render("Up next: "+m.queue[m.queueIndex+1].Title)
	}

	switch {
	case m.pausePending && m.player.IsPaused():
		statusView = "\n" + m.loadingSpinner.View() + " Resuming...\n"
	case m.pausePending:
		statusView = "\n" + m.loadingSpinner.View() + " Pausing...\n"
	case m.player.IsPaused():
		statusView = "\n⏸  Paused\n"
	}
	status := getStatusStyle(m.theme).Render(statusView) + upNextLine

	keys := playingKeys
	keys.Next.SetEnabled(m.queueIndex+1 < len(m.queue))
	keys.Prev.SetEnabled(m.queueIndex > 0)
	playingFooter := m.topPanelFooter(keys)

	if m.minimal {
		content := lipgloss.JoinVertical(lipgloss.Top, status, playingFooter)
		return m.renderFrame(content, "")
	}

	header := labelYtbPlayer
	audioInfo := ""
	if m.metadata.Title != "" {
		audioInfo = m.metadata.Title
	}
	title := getTitleStyle(m.theme).Render(header)
	topPanelWidth := getContentWidth(topPanelStyleFor(panelHeightFor(m.termHeight), m.availableMainWidth(), m.theme))
	audioTitle := getLabelStyle(m.theme).Render(utils.TruncateTitle(audioInfo, topPanelWidth))
	visualizer := "\n" + renderVisualizer(m.visualizerBars, m.availableMainWidth(), m.theme)

	content := lipgloss.JoinVertical(lipgloss.Top, title, status, audioTitle, playingFooter)
	return m.renderFrame(content, visualizer)
}

func (m model) loadingView() tea.View {
	header := labelYtbPlayer
	title := getTitleStyle(m.theme).Render(header)
	loadingLabel := getLabelStyle(m.theme).Render(labelLoading)
	content := lipgloss.JoinVertical(lipgloss.Top, title, loadingLabel)

	v := m.renderFrame(content, "")
	v.ProgressBar = tea.NewProgressBar(tea.ProgressBarIndeterminate, 0)
	return v
}

func (m model) inputRowOffset(panel lipgloss.Style, header string, others ...string) int {
	y := 0
	if !m.minimal {
		y = lipgloss.Height(m.renderTabStrip())
	}
	if m.minimal {
		y++ // for the top border
	}
	y += lipgloss.Height(header)
	for _, other := range others {
		y += lipgloss.Height(other)
	}
	y += panel.GetPaddingTop() + panel.GetBorderTopSize()
	return y
}

func (m model) inputColumnBounds(panel lipgloss.Style) (minX, maxX int) {
	minX = panel.GetPaddingLeft() + panel.GetBorderLeftSize()
	if !m.minimal {
		minX += lipgloss.Width(m.renderSidebar(0))
	}
	maxX = minX + m.textInput.Width()
	return minX, maxX
}

func (m model) setCursor(header string, panel lipgloss.Style, others ...string) *tea.Cursor {
	var c *tea.Cursor
	if !m.textInput.VirtualCursor() {
		c = m.textInput.Cursor()
		if c != nil {
			c.Y += m.inputRowOffset(panel, header, others...)
			minX, _ := m.inputColumnBounds(panel)
			c.X += minX
		}
	}
	return c
}
