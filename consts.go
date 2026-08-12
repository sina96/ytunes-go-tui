package main

const (

	// side bar
	appTitle   = "yTunes"
	appVersion = "v0.1.0"
	logo       = `
 __  ____     ____  __
 \ \/ / /____/_  / / /
  \  / __/ _ \/ /_/_/
  /_/\__/_//_/___(_)
	`

	sidebarMinHeight = 11 // logo + title + version + spacer + clock + padding + border

	// labels
	labelIdlePlaceholder  = "play that tubez!"         // was duplicated in view.go twice
	labelConfirmQuit      = "(press again to confirm)" // was duplicated three times
	labelEnterURL         = "Enter a url:"             // was duplicated in idleView/errorView
	labelLastPlayedPrefix = "LastPlayed: "
	labelPlayAnother      = "\nPlay Another url:"
	labelLoading          = "\nLoading..."
)
