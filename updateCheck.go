package main

import (
	"encoding/json"
	"net/http"
	"time"

	tea "charm.land/bubbletea/v2"
)

type updateCheckMsg struct {
	tag string // empty if the request/parse failed — Update doesn't need to know why
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func checkForUpdate() tea.Cmd {
	return func() tea.Msg {
		client := http.Client{Timeout: 3 * time.Second}
		resp, err := client.Get(releaseAPIURL)
		if err != nil {
			return updateCheckMsg{} // silent failure — empty tag means "show nothing"
		}
		defer resp.Body.Close()

		var release githubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return updateCheckMsg{}
		}

		return updateCheckMsg{tag: release.TagName}
	}
}
