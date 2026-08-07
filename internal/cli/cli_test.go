package cli

import (
	"reflect"
	"strings"
	"testing"

	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
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

func TestCompletionsMatchCommandFlags(t *testing.T) {
	zsh, _ := Completion("zsh")
	bash, _ := Completion("bash")
	fish, _ := Completion("fish")
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

func sampleAlbums() []lastfm.Album {
	return []lastfm.Album{
		{Artist: "Artist", Title: "Album", Tracks: []lastfm.Track{{Title: "Track 1"}, {Title: "Track 2"}}},
		{Artist: "Other", Title: "Second", Tracks: []lastfm.Track{{Title: "Other 1"}, {Title: "Other 2"}}},
	}
}
