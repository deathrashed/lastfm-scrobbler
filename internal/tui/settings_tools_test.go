package tui

import (
	"strings"
	"testing"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
)

func TestToolsUsesGitHubReleasesWhenUpdateURLIsEmpty(t *testing.T) {
	m := model{cfg: config.Config{}, settingsSection: settingsTools}
	row := settingsRows(settingsTools)[1]
	if got := m.settingsOverviewValue(row); got != "GitHub Releases" {
		t.Fatalf("empty update source = %q, want GitHub Releases", got)
	}
	view := stripANSI(renderSettingsView(m))
	if strings.Contains(view, "not configured") {
		t.Fatalf("Tools view still presents an unconfigured normal update source: %q", view)
	}
}
