package tui

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
)

func testLastSessionRecord() sessionstore.Record {
	return sessionstore.Record{
		Mode: "manual", Status: "complete", StartedAt: time.Unix(1700000000, 0),
		Queue: []sessionstore.Track{{Artist: "Terror", Album: "Keepers Of The Faith", Title: "Track", TrackIndex: 1}},
		Loop:  1, Interval: 2 * time.Second,
	}
}

func TestDashboardRerunOpensLastSessionWithoutStarting(t *testing.T) {
	m := model{stage: stageInput, cfg: config.Config{MouseEnabled: true}, history: []sessionstore.Record{testLastSessionRecord()}}
	updated, cmd := updateModel(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("r")})
	got := updated.(model)
	if cmd != nil || got.stage != stageLastSession {
		t.Fatalf("r stage=%d cmd=%v, want Last Session and no command", got.stage, cmd)
	}
	if got.modeChoice != "" {
		t.Fatalf("Last Session confirmation changed Dashboard header mode to %q", got.modeChoice)
	}
}

func TestLastSessionEnterReusesExactRerunPath(t *testing.T) {
	record := testLastSessionRecord()
	m := model{stage: stageLastSession, cfg: config.Config{DefaultLoop: 1}, history: []sessionstore.Record{record}}
	updated, _ := m.updateLastSession(keyMessage("enter"))
	got := updated.(model)
	if got.stage != stagePreview || got.modeChoice != "manual" || len(got.scrobbleQueue) != 1 {
		t.Fatalf("enter stage=%d mode=%q queue=%d", got.stage, got.modeChoice, len(got.scrobbleQueue))
	}
}

func TestLastSessionFooterMouseReusesRerunAction(t *testing.T) {
	m := model{
		width: 120, stage: stageLastSession, cfg: config.Config{MouseEnabled: true},
		history: []sessionstore.Record{testLastSessionRecord()},
	}
	updated, cmd := clickRegion(t, m, "footer:enter")
	if cmd != nil || updated.stage != stagePreview || len(updated.scrobbleQueue) != 1 {
		t.Fatalf("mouse rerun stage=%d queue=%d cmd=%v", updated.stage, len(updated.scrobbleQueue), cmd)
	}
}

func TestLastSessionRendersNoSecretsAndResponsivePanel(t *testing.T) {
	record := testLastSessionRecord()
	m := model{width: 127, stage: stageLastSession, cfg: config.Config{Username: "user", Password: "secret", APISecret: "private"}, history: []sessionstore.Record{record}}
	view := stripANSI(renderLastSessionView(m))
	if !strings.Contains(view, "LAST") && !strings.Contains(view, "Terror") {
		t.Fatalf("Last Session view missing session content: %q", view)
	}
	if strings.Contains(view, "secret") || strings.Contains(view, "private") {
		t.Fatal("Last Session view exposed credential material")
	}
	for _, line := range strings.Split(view, "\n") {
		if width := displayWidth(line); width > m.appWidth() {
			t.Fatalf("Last Session line width=%d exceeds %d: %q", width, m.appWidth(), line)
		}
	}
}
