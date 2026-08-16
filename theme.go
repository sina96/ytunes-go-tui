package main

import (
	"fmt"
	"image/color"
	"io"
	"os"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

type themeItem struct {
	name  string
	theme Theme
}

func (i themeItem) Title() string       { return i.name }
func (i themeItem) Description() string { return "" }
func (i themeItem) FilterValue() string { return i.name }

type themeDelegate struct{}

func (d themeDelegate) Height() int                         { return 1 }
func (d themeDelegate) Spacing() int                        { return 0 }
func (d themeDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }
func (d themeDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(themeItem)
	if !ok {
		return
	}
	str := i.name
	if index == m.Index() {
		str = "> " + str
	}
	fmt.Fprint(w, lipgloss.NewStyle().Foreground(i.theme.Accent).Render(str))
}

func newThemeList(width, height int, title string) list.Model {
	items := make([]list.Item, len(Themes))
	for i, t := range Themes {
		items[i] = themeItem{name: t.Name, theme: t.Theme}
	}

	l := list.New(items, themeDelegate{}, width, height)
	l.Title = title
	l.AdditionalFullHelpKeys = themePickerKeys.ShortHelp
	l.SetShowHelp(true)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	return l
}

type Theme struct {
	Accent color.Color
	Muted  color.Color
	Error  color.Color
	Border color.Color
	Gray   color.Color
	Peak   color.Color // hottest stop of the visualizer's gradient — deliberately distinct from Error
}

// Default
var ThemeDefault = Theme{
	Accent: lipgloss.Color("205"),
	Muted:  lipgloss.Color("241"),
	Error:  lipgloss.Color("196"),
	Border: lipgloss.Color("62"),
	Gray:   lipgloss.Color("235"),
	Peak:   lipgloss.Color("214"), // orange
}

// Terminal-aligned — the 16 basic ANSI indices
var ThemeTerminal = Theme{
	Accent: lipgloss.Color("5"), // magenta
	Muted:  lipgloss.Color("8"), // bright black (gray)
	Error:  lipgloss.Color("1"), // red
	Border: lipgloss.Color("4"), // blue
	Gray:   lipgloss.Color("0"), // black
	Peak:   lipgloss.Color("3"), // yellow
}

// Catppuccin (Mocha) — https://catppuccin.com/palette
var ThemeCatppuccin = Theme{
	Accent: lipgloss.Color("#cba6f7"), // Mauve
	Muted:  lipgloss.Color("#a6adc8"), // Subtext0
	Error:  lipgloss.Color("#f38ba8"), // Red
	Border: lipgloss.Color("#6c7086"), // Overlay0
	Gray:   lipgloss.Color("#45475a"), // Surface1
	Peak:   lipgloss.Color("#fab387"), // Peach
}

// Gruvbox (dark) — https://github.com/morhetz/gruvbox
var ThemeGruvbox = Theme{
	Accent: lipgloss.Color("#fabd2f"), // yellow/orang
	Muted:  lipgloss.Color("#a89984"), // gray
	Error:  lipgloss.Color("#fb4934"), // red
	Border: lipgloss.Color("#665c54"), // subtle gray
	Gray:   lipgloss.Color("#282828"), // bg1
	Peak:   lipgloss.Color("#fe8019"), // orange
}

// Dracula — https://draculatheme.com/contribute
var ThemeDracula = Theme{
	Accent: lipgloss.Color("#bd93f9"), // Purple
	Muted:  lipgloss.Color("#6272a4"), // Comment
	Error:  lipgloss.Color("#ff5555"), // Red
	Border: lipgloss.Color("#6272a4"), // Current Line
	Gray:   lipgloss.Color("#282a36"),
	Peak:   lipgloss.Color("#ffb86c"), // Orange
}

// Themes is the picker's data source, in display order.
var Themes = []struct {
	Name  string
	Theme Theme
}{
	{"Default", ThemeDefault},
	{"Terminal-aligned", ThemeTerminal},
	{"Catppuccin", ThemeCatppuccin},
	{"Gruvbox", ThemeGruvbox},
	{"Dracula", ThemeDracula},
}

func themeConfigPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "ytunes", "theme"), nil
}

func loadSavedTheme() Theme {
	path, err := themeConfigPath()
	if err != nil {
		return ThemeDefault
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return ThemeDefault
	}

	name := strings.TrimSpace(string(data))
	for _, theme := range Themes {
		if theme.Name == name {
			return theme.Theme
		}
	}
	return ThemeDefault
}

func saveTheme(name string) error {
	path, err := themeConfigPath()
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	return os.WriteFile(path, []byte(name), 0644)
}
