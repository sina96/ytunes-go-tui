package utils

import (
	"fmt"

	"github.com/charmbracelet/x/ansi"
)

func FormatDuration(seconds int) string {
	if seconds < 0 {
		return ""
	}

	if seconds >= 3600 {
		return fmt.Sprintf("%02d:%02d:%02d", seconds/3600, (seconds%3600)/60, seconds%60)
	}
	return fmt.Sprintf("%02d:%02d", seconds/60, seconds%60)
}

func TruncateTitle(s string, width int) string {
	return ansi.Truncate(s, width, "…")
}
