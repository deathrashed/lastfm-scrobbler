package tui

import (
	"context"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
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
	recent := m.activityContent()
	if got := stripANSI(recent); !strings.Contains(got, "Enforced - War Remains") {
		t.Fatalf("recent activity = %q", got)
	}
	if strings.Contains(recent, theme.AccentTextStyle.Render(activityVolumeFrames[0])) {
		t.Fatalf("recent activity incorrectly uses the live volume animation: %q", recent)
	}
	if !strings.Contains(recent, theme.MutedStyle.Render(theme.IconHistory)) {
		t.Fatalf("recent activity is missing the static recent icon: %q", recent)
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

func TestCurrentActivityResultStartsAnimation(t *testing.T) {
	m := model{
		width:       67,
		cfg:         config.Config{Username: "deathrashed", APIKey: "test-key", NowPlaying: true},
		client:      activityClientStub{},
		activitySeq: 9,
	}
	updated, cmd := updateModel(m, activityResultMsg{
		seq:    9,
		tracks: []lastfm.RecentTrack{{Artist: "Hypocrisy", Title: "Elastic Inverted Visions (live)", NowPlaying: true}},
	})
	got := updated.(model)
	if got.activityState != activityCurrent {
		t.Fatalf("activity state = %d, want current", got.activityState)
	}
	if cmd == nil {
		t.Fatal("current activity did not schedule refresh/animation commands")
	}
	if !got.activityShouldAnimate() {
		t.Fatal("current activity is not considered animation-eligible")
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

type activityClientStub struct{}

func (activityClientStub) Authenticate(context.Context) error           { return nil }
func (activityClientStub) GetAuthToken(context.Context) (string, error) { return "token", nil }
func (activityClientStub) GetSession(context.Context, string) (lastfm.Session, error) {
	return lastfm.Session{Name: "user", Key: "key"}, nil
}
func (activityClientStub) AuthURL(token string) string {
	return "https://www.last.fm/api/auth/?token=" + token
}
func (activityClientStub) SessionKey() string { return "key" }
func (activityClientStub) SearchAlbums(context.Context, string) ([]lastfm.Album, error) {
	return nil, nil
}
func (activityClientStub) GetAlbumTracks(context.Context, string, string) (lastfm.Album, error) {
	return lastfm.Album{}, nil
}
func (activityClientStub) GetDiscography(context.Context, string) ([]lastfm.Album, error) {
	return nil, nil
}
func (activityClientStub) GetSimilarAlbums(context.Context, string, int) ([]lastfm.Album, error) {
	return nil, nil
}
func (activityClientStub) GetRecentTracks(context.Context, string, time.Time) ([]lastfm.RecentTrack, error) {
	return nil, nil
}
func (activityClientStub) Scrobble(context.Context, string, string, string, int64) error { return nil }

func TestCurrentActivityAnimationAdvancesAndReschedules(t *testing.T) {
	m := model{
		width:         67,
		cfg:           config.Config{Username: "deathrashed", APIKey: "test-key", NowPlaying: true},
		client:        activityClientStub{},
		activitySeq:   11,
		activityState: activityCurrent,
		activityTrack: lastfm.RecentTrack{Artist: "Hypocrisy", Title: "Elastic Inverted Visions (live)", NowPlaying: true},
		activityFrame: 0,
	}
	updated, cmd := updateModel(m, activityAnimationMsg{seq: 11})
	got := updated.(model)
	if got.activityFrame != 1 {
		t.Fatalf("activity frame=%d want=1", got.activityFrame)
	}
	if cmd == nil {
		t.Fatal("current activity animation did not reschedule itself")
	}
	if plain := stripANSI(got.activityContent()); !strings.Contains(plain, activityVolumeFrames[1]) {
		t.Fatalf("activity content did not render advanced frame: %q", plain)
	}
}

func TestResizeFromCompactToHeroStartsActivityPolling(t *testing.T) {
	m := model{
		width:       100,
		height:      20,
		cfg:         config.Config{Username: "deathrashed", APIKey: "test-key", NowPlaying: true},
		client:      activityClientStub{},
		activitySeq: 3,
	}
	if m.activityPollingEnabled() {
		t.Fatal("short compact layout unexpectedly enabled activity polling")
	}
	updated, cmd := updateModel(m, tea.WindowSizeMsg{Width: 100, Height: 40})
	got := updated.(model)
	if got.headerLayout() != headerHero {
		t.Fatalf("resized header layout=%d want hero", got.headerLayout())
	}
	if !got.activityPollingEnabled() {
		t.Fatal("hero layout did not re-enable activity polling")
	}
	if cmd == nil {
		t.Fatal("resizing into hero layout did not schedule an activity fetch")
	}
}
