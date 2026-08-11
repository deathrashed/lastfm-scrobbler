package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/spinner"

	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

func lastFMSpinner() spinner.Spinner {
	logo := theme.AccentTextStyle.Render(theme.IconDashboard)
	dot := theme.PrimaryTextStyle.Render("∙")

	return spinner.Spinner{
		Frames: []string{
			logo + " " + dot + " " + dot,
			dot + " " + logo + " " + dot,
			dot + " " + dot + " " + logo,
			dot + " " + logo + " " + dot,
		},
		FPS: time.Second / 7,
	}
}

func (m model) spinnerActive() bool {
	return m.searching ||
		m.connectionTesting ||
		m.diagnosticsBusy ||
		m.updateChecking ||
		m.setup.Applying ||
		m.stage == stageScrobbling
}
