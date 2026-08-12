package utils

import (
	"net/url"
	"strings"
)

func isValidURL(s string) bool {
	u, err := url.Parse(s)
	return err == nil && u.Scheme != "" && u.Host != ""
}

func addScheme(s string) string {
	if !strings.HasPrefix(s, "http://") && !strings.HasPrefix(s, "https://") {
		return "https://" + s
	}
	return s
}

func IsYouTubeURL(s string) bool {
	return isValidURL(addScheme(s)) && (strings.Contains(s, "youtube.com") || strings.Contains(s, "youtu.be"))
}
