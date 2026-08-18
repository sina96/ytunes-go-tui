package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	tea "charm.land/bubbletea/v2"
)

const timoutSec = 3 * time.Second

type updateCheckMsg struct {
	tag string // empty if the request/parse failed — Update doesn't need to know why
}

type githubRelease struct {
	TagName string `json:"tag_name"`
}

func checkForUpdate() tea.Cmd {
	return func() tea.Msg {
		req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, releaseAPIURL, nil)
		if err != nil {
			return updateCheckMsg{}
		}
		req.Header.Set("User-Agent", "ytunes")

		client := http.Client{Timeout: timoutSec}
		resp, err := client.Do(req)
		if err != nil {
			return updateCheckMsg{}
		}
		defer func() { _ = resp.Body.Close() }()

		if resp.StatusCode != http.StatusOK {
			return updateCheckMsg{}
		}

		var release githubRelease
		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return updateCheckMsg{}
		}

		return updateCheckMsg{tag: release.TagName}
	}
}
