package tui

import (
	"context"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPreviewEscapeCancelsPreparationAndIgnoresLateResult(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	m := model{stage: stagePreview, sessionCtx: ctx, sessionCancel: cancel, sessionID: 7}
	updated, _ := m.updatePreview(tea.KeyMsg{Type: tea.KeyEsc})
	afterEscape := updated.(model)
	if afterEscape.stage != stageTrackSelect {
		t.Fatalf("stage after Escape = %v, want track selection", afterEscape.stage)
	}
	if ctx.Err() != context.Canceled {
		t.Fatal("Escape did not cancel the preparation context")
	}

	late, _ := updateModel(afterEscape, scrobblePreparedMsg{sessionID: 7, queue: []queuedTrack{{Artist: "A", Album: "B", Title: "C"}}})
	if late.(model).stage == stageScrobbling {
		t.Fatal("late preparation result restarted scrobbling")
	}
}

func TestOldPreparationResultCannotStartNewSession(t *testing.T) {
	oldCtx, oldCancel := context.WithCancel(context.Background())
	m := model{stage: stagePreview, sessionCtx: oldCtx, sessionCancel: oldCancel, sessionID: 1}
	m.startScrobbleSession()
	if m.sessionID == 1 {
		t.Fatal("new session did not receive a new generation ID")
	}
	late, _ := updateModel(m, scrobblePreparedMsg{sessionID: 1, queue: []queuedTrack{{Artist: "A", Album: "B", Title: "C"}}})
	if late.(model).stage == stageScrobbling {
		t.Fatal("old preparation result started the new session")
	}
}
