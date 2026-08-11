package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

func TestActivityVolumeFramesAreSingleCell(t *testing.T) {
	for _, frame := range activityVolumeFrames {
		if got := lipgloss.Width(frame); got != 1 {
			t.Fatalf("activity frame %q width = %d, want 1", frame, got)
		}
	}
}

func TestActivityContentStates(t *testing.T) {
	m := model{width: 67, cfg: config.Config{NowPlaying: true}, activityState: activityCurrent, activityTrack: lastfm.RecentTrack{Artist: "Enforced", Title: "War Remains"}}
	current := m.activityContent()
	if got := stripANSI(current); !strings.Contains(got, "Enforced - War Remains") {
		t.Fatalf("current activity = %q", got)
	}
	if !strings.Contains(current, theme.AccentTextStyle.Render(activityVolumeFrames[0])) ||
		!strings.Contains(current, theme.AlbumStyle.Render("Enforced")) ||
		!strings.Contains(current, theme.PrimaryTextStyle.Render("War Remains")) {
		t.Fatalf("current activity styles are incorrect: %q", current)
	}
	m.activityState = activityRecent
	if got := stripANSI(m.activityContent()); !strings.Contains(got, "Enforced - War Remains") {
		t.Fatalf("recent activity = %q", got)
	}
	m.activityState = activityNoTracks
	if got := stripANSI(m.activityContent()); got != "no recent scrobbles" {
		t.Fatalf("no-track activity = %q", got)
	}
	m.activityState = activityUnavailable
	if got := stripANSI(m.activityContent()); got != "Last.fm activity unavailable" {
		t.Fatalf("unavailable activity = %q", got)
	}
	m.cfg.CompactHeader = true
	if got := m.activityContent(); got != "" {
		t.Fatalf("compact activity = %q, want empty", got)
	}
	m.cfg.CompactHeader = false
	m.activityState = activityLoading
	if got := stripANSI(m.activityContent()); got != "Last.fm activity unavailable" {
		t.Fatalf("activity without configured polling = %q", got)
	}
}

func TestActivityResultPreservesUsefulStateOnTransientError(t *testing.T) {
	m := model{width: 67, cfg: config.Config{Username: "user", NowPlaying: true}, activityState: activityCurrent, activitySeq: 4}
	updated, _ := updateModel(m, activityResultMsg{seq: 4, err: errActivityTest})
	got := updated.(model)
	if got.activityState != activityCurrent {
		t.Fatalf("activity state after transient error = %d, want current", got.activityState)
	}
}

func TestActivityTextPartsPreserveSeparatorAndWidth(t *testing.T) {
	artist, title := activityTextParts("A very long artist name", "A very long track title", 24)
	if got := displayWidth(artist) + 3 + displayWidth(title); got > 24 {
		t.Fatalf("activity text width = %d, want <= 24: %q - %q", got, artist, title)
	}
	if artist == "" || title == "" {
		t.Fatalf("activity text lost a side of the separator: %q - %q", artist, title)
	}
	artist, title = activityTextParts("Enforced", "War Remains", 64)
	if artist != "Enforced" || title != "War Remains" {
		t.Fatalf("short activity text changed: %q - %q", artist, title)
	}
}

var errActivityTest = activityTestError{}

type activityTestError struct{}

func (activityTestError) Error() string { return "test activity failure" }
