package cli

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
)

func TestNormalizeArgsAllowsFlagsAfterPositionals(t *testing.T) {
	got := normalizeArgs(
		[]string{"Demolition Hammer", "--first", "3", "--dry-run", "--interval=500ms"},
		map[string]bool{"first": true, "interval": true},
	)
	want := []string{"--first", "3", "--dry-run", "--interval=500ms", "Demolition Hammer"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeArgs() = %#v, want %#v", got, want)
	}
}

func TestSetupIsRecognizedAsInteractiveCommand(t *testing.T) {
	if !IsCommand([]string{"setup"}) {
		t.Fatal("setup command is not recognized")
	}
}

func TestCompletionsMatchCommandFlags(t *testing.T) {
	zsh, _ := Completion("zsh")
	bash, _ := Completion("bash")
	fish, _ := Completion("fish")
	powershell, _ := Completion("powershell")
	for _, output := range []string{zsh, bash} {
		if !strings.Contains(output, "--artist") || !strings.Contains(output, "--album") {
			t.Fatalf("manual credential flags missing from completion:\n%s", output)
		}
		if !strings.Contains(output, "--all") || !strings.Contains(output, "--clean") {
			t.Fatalf("discography flags missing from completion:\n%s", output)
		}
	}
	if !strings.Contains(fish, "-l artist") || !strings.Contains(fish, "-l album") || !strings.Contains(fish, "-l all") || !strings.Contains(fish, "-l clean") {
		t.Fatal("fish command flags are incomplete")
	}
	if !strings.Contains(powershell, "Register-ArgumentCompleter") || !strings.Contains(powershell, "powershell") {
		t.Fatal("PowerShell completion is incomplete")
	}
	if strings.Contains(bash, `file) COMPREPLY=( $(compgen -W "--loop --limit --interval --dry-run --json --artist`) {
		t.Fatal("bash file completion advertises manual-only flags")
	}
	if strings.Contains(zsh, "file) _arguments '--artist") {
		t.Fatal("zsh file completion advertises manual-only flags")
	}
	if strings.Contains(fish, "__fish_seen_subcommand_from file' -l artist") {
		t.Fatal("fish file completion advertises manual-only flags")
	}
}

func TestNormalizeArgsPreservesDoubleDashPositionals(t *testing.T) {
	got := normalizeArgs(
		[]string{"--loop", "2", "--", "Artist", "- Album"},
		map[string]bool{"loop": true},
	)
	want := []string{"--loop", "2", "Artist", "- Album"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizeArgs() = %#v, want %#v", got, want)
	}
}

func TestBuildQueueAppliesLoopAndLimit(t *testing.T) {
	// Kept here because buildQueue is the common path for Manual, File, and
	// Discography headless commands.
	albums := sampleAlbums()
	queue := buildQueue(albums, 2, 1)
	if len(queue) != 4 {
		t.Fatalf("len(queue) = %d, want 4", len(queue))
	}
	if queue[0].Title != "Track 1" || queue[1].Title != "Track 1" || queue[2].Title != "Other 1" {
		t.Fatalf("unexpected queue ordering: %#v", queue)
	}
}

func TestCLIQueueRecordsPreserveAlbumTrackAndLoopMetadata(t *testing.T) {
	albums := sampleAlbums()
	queue := buildQueue(albums, 2, 0)
	record := recordFromQueue("manual", structConfig(), queue, commonOptions{loop: 2, interval: time.Second})
	if len(record.Queue) != 8 {
		t.Fatalf("queue length = %d, want 8", len(record.Queue))
	}
	want := []sessionstore.Track{
		{AlbumIndex: 1, AlbumTotal: 2, TrackIndex: 1, TrackTotal: 2, LoopIndex: 1, LoopTotal: 2},
		{AlbumIndex: 1, AlbumTotal: 2, TrackIndex: 2, TrackTotal: 2, LoopIndex: 1, LoopTotal: 2},
		{AlbumIndex: 1, AlbumTotal: 2, TrackIndex: 1, TrackTotal: 2, LoopIndex: 2, LoopTotal: 2},
		{AlbumIndex: 1, AlbumTotal: 2, TrackIndex: 2, TrackTotal: 2, LoopIndex: 2, LoopTotal: 2},
		{AlbumIndex: 2, AlbumTotal: 2, TrackIndex: 1, TrackTotal: 2, LoopIndex: 1, LoopTotal: 2},
	}
	for index, expected := range want {
		got := record.Queue[index]
		if got.AlbumIndex != expected.AlbumIndex || got.AlbumTotal != expected.AlbumTotal || got.TrackIndex != expected.TrackIndex || got.TrackTotal != expected.TrackTotal || got.LoopIndex != expected.LoopIndex || got.LoopTotal != expected.LoopTotal {
			t.Fatalf("queue[%d] metadata = %#v, want %#v", index, got, expected)
		}
	}
}

func structConfig() config.Config { return config.Config{Profile: "default"} }

func sampleAlbums() []lastfm.Album {
	return []lastfm.Album{
		{Artist: "Artist", Title: "Album", Tracks: []lastfm.Track{{Title: "Track 1"}, {Title: "Track 2"}}},
		{Artist: "Other", Title: "Second", Tracks: []lastfm.Track{{Title: "Other 1"}, {Title: "Other 2"}}},
	}
}
