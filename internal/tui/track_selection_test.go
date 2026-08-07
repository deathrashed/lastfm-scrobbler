package tui

import (
	"testing"

	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
)

func testAlbum() lastfm.Album {
	return lastfm.Album{
		Artist: "Slayer",
		Title:  "Hell Awaits",
		Tracks: []lastfm.Track{
			{Title: "Hell Awaits"},
			{Title: "Kill Again"},
			{Title: "At Dawn They Sleep"},
		},
	}
}

func TestInitialiseTrackSelectionSelectsAll(t *testing.T) {
	m := model{selectedAlbums: []lastfm.Album{testAlbum()}, trackSelected: map[int]bool{}}
	m.initialiseTrackSelection()
	if got := m.selectedTrackCount(); got != 3 {
		t.Fatalf("selected %d tracks, want 3", got)
	}
}

func TestInitialiseTrackSelectionRespectsLimit(t *testing.T) {
	m := model{selectedAlbums: []lastfm.Album{testAlbum()}, trackLimit: 2, trackSelected: map[int]bool{}}
	m.initialiseTrackSelection()
	if got := m.selectedTrackCount(); got != 2 {
		t.Fatalf("selected %d tracks, want 2", got)
	}
}

func TestBuildScrobbleQueueUsesSelectionAndLoops(t *testing.T) {
	m := model{
		selectedAlbums: []lastfm.Album{testAlbum()},
		trackSelected:  map[int]bool{0: true, 2: true},
		loopCount:      2,
	}
	m.buildScrobbleQueue()

	if got := len(m.scrobbleQueue); got != 4 {
		t.Fatalf("queue contains %d tracks, want 4", got)
	}
	want := []string{"Hell Awaits", "At Dawn They Sleep", "Hell Awaits", "At Dawn They Sleep"}
	for i, title := range want {
		if m.scrobbleQueue[i].Title != title {
			t.Fatalf("queue[%d] = %q, want %q", i, m.scrobbleQueue[i].Title, title)
		}
	}
	if m.albumsScrobbled != 1 {
		t.Fatalf("albumsScrobbled = %d, want 1", m.albumsScrobbled)
	}
}

func TestBuildScrobbleQueueAddsAlbumAndTrackMetadata(t *testing.T) {
	second := testAlbum()
	second.Title = "Reign in Blood"
	second.Tracks = second.Tracks[:2]
	m := model{
		selectedAlbums: []lastfm.Album{testAlbum(), second},
		trackSelected:  map[int]bool{0: true, 1: true, 2: true, 3: true, 4: true},
		loopCount:      1,
	}
	m.buildScrobbleQueue()
	if got := len(m.scrobbleQueue); got != 5 {
		t.Fatalf("queue contains %d tracks, want 5", got)
	}
	first := m.scrobbleQueue[0]
	last := m.scrobbleQueue[len(m.scrobbleQueue)-1]
	if first.AlbumIndex != 1 || first.AlbumTotal != 2 || first.TrackIndex != 1 || first.TrackTotal != 3 {
		t.Fatalf("first metadata = %#v", first)
	}
	if last.AlbumIndex != 2 || last.AlbumTotal != 2 || last.TrackIndex != 2 || last.TrackTotal != 2 {
		t.Fatalf("last metadata = %#v", last)
	}
}

func TestProgressMarkersMarkAlbumBoundaries(t *testing.T) {
	queue := []queuedTrack{
		{AlbumIndex: 1}, {AlbumIndex: 1}, {AlbumIndex: 2}, {AlbumIndex: 2}, {AlbumIndex: 3},
	}
	markers := progressMarkers(queue, 20)
	if len(markers) != 2 {
		t.Fatalf("markers = %#v, want 2 boundaries", markers)
	}
}
