package main

import (
	"charm.land/bubbles/v2/key"
)

type idleKeyMap struct {
	Enter, Paste, Clear, Quit, ThemePicker, QueueMode key.Binding
}

type playingKeyMap struct {
	PauseToggle, Stop, Quit, Next, Prev key.Binding
}

type themePickerKeyMap struct {
	Up, Down, Select, Cancel key.Binding
}

func (k idleKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Enter, k.Paste, k.Clear, k.Quit, k.ThemePicker, k.QueueMode}
}

func (k playingKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.PauseToggle, k.Stop, k.Quit, k.Next, k.Prev}
}

func (k themePickerKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Select, k.Cancel}
}

func (k idleKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func (k playingKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

func (k themePickerKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{k.ShortHelp()}
}

var idleKeys = idleKeyMap{
	Enter:       key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "play")),
	Paste:       key.NewBinding(key.WithKeys("ctrl+v"), key.WithHelp("ctrl+v", "paste")),
	Clear:       key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl+u", "clear")),
	ThemePicker: key.NewBinding(key.WithKeys("ctrl+t"), key.WithHelp("ctrl+t", "theme")),
	QueueMode:   key.NewBinding(key.WithKeys("ctrl+q"), key.WithHelp("ctrl+q", "queue mode")),
	Quit:        key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
}

var playingKeys = playingKeyMap{
	PauseToggle: key.NewBinding(key.WithKeys("space"), key.WithHelp("space", "pause/play")),
	Stop:        key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "stop")),
	Quit:        key.NewBinding(key.WithKeys("ctrl+c"), key.WithHelp("ctrl+c", "quit")),
	Next:        key.NewBinding(key.WithKeys("right", "l"), key.WithHelp("right/l", "next")),
	Prev:        key.NewBinding(key.WithKeys("left", "h"), key.WithHelp("left/h", "prev")),
}

var themePickerKeys = themePickerKeyMap{
	Select: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
	Cancel: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
}
