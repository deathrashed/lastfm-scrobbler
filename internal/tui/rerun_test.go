package tui

import (
	"testing"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
)

func TestRestoreRecordForEditCollapsesLoops(t *testing.T) {
	record := sessionstore.Record{Mode: "manual", Loop: 2, Interval: 2 * time.Second, Queue: []sessionstore.Track{
		{Artist: "Slayer", Album: "Hell Awaits", Title: "Hell Awaits", TrackIndex: 1, TrackTotal: 2, LoopIndex: 1, LoopTotal: 2},
		{Artist: "Slayer", Album: "Hell Awaits", Title: "Kill Again", TrackIndex: 2, TrackTotal: 2, LoopIndex: 1, LoopTotal: 2},
		{Artist: "Slayer", Album: "Hell Awaits", Title: "Hell Awaits", TrackIndex: 1, TrackTotal: 2, LoopIndex: 2, LoopTotal: 2},
		{Artist: "Slayer", Album: "Hell Awaits", Title: "Kill Again", TrackIndex: 2, TrackTotal: 2, LoopIndex: 2, LoopTotal: 2},
	}}
	m := model{}
	m.restoreRecordForEdit(record)
	if len(m.selectedAlbums) != 1 || len(m.selectedAlbums[0].Tracks) != 2 {
		t.Fatalf("unexpected albums: %#v", m.selectedAlbums)
	}
	if m.loopForAlbum(0) != 2 {
		t.Fatalf("loop=%d", m.loopForAlbum(0))
	}
	if m.selectedTrackCount() != 2 {
		t.Fatalf("selected=%d", m.selectedTrackCount())
	}
}
