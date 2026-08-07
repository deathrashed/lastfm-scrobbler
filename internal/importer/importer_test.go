package importer

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadTextCSVJSONAndPlaylist(t *testing.T) {
	dir := t.TempDir()
	tests := []struct {
		name    string
		content string
	}{
		{"albums.txt", "Slayer - Hell Awaits\nDemolition Hammer — Epidemic of Violence\n"},
		{"albums.csv", "artist,album\nSlayer,Hell Awaits\nDemolition Hammer,Epidemic of Violence\n"},
		{"albums.json", `[{"artist":"Slayer","album":"Hell Awaits"},{"artist":"Demolition Hammer","album":"Epidemic of Violence"}]`},
		{"albums.m3u8", "/Music/Slayer/Hell Awaits/01 Hell Awaits.flac\n/Music/Demolition Hammer/Epidemic of Violence/01 Skull.flac\n"},
	}
	for _, test := range tests {
		path := filepath.Join(dir, test.name)
		if err := os.WriteFile(path, []byte(test.content), 0600); err != nil {
			t.Fatal(err)
		}
		targets, err := Load(path)
		if err != nil {
			t.Fatalf("%s: %v", test.name, err)
		}
		if len(targets) != 2 {
			t.Fatalf("%s: got %#v", test.name, targets)
		}
	}
}

func TestLoadDirectoryInfersArtistAndAlbum(t *testing.T) {
	root := t.TempDir()
	albumDir := filepath.Join(root, "Slayer", "Hell Awaits")
	if err := os.MkdirAll(albumDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(albumDir, "01 Hell Awaits.flac"), []byte("audio"), 0600); err != nil {
		t.Fatal(err)
	}
	targets, err := Load(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 || targets[0].Artist != "Slayer" || targets[0].Album != "Hell Awaits" {
		t.Fatalf("targets = %#v", targets)
	}
}

func TestLoadTextPreservesInputOrder(t *testing.T) {
	path := filepath.Join(t.TempDir(), "albums.txt")
	content := "Slayer - Hell Awaits\nMetallica - Ride the Lightning\nMegadeth - Peace Sells\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	targets, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if targets[0].Artist != "Slayer" || targets[1].Artist != "Metallica" || targets[2].Artist != "Megadeth" {
		t.Fatalf("targets = %#v", targets)
	}
}

func TestLoadM3UChecksPathsBeforeArtistAlbumParsing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "albums.m3u8")
	content := "/Music/Artist - Live/Album - 2024/01 Track.flac\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	targets, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 || targets[0].Artist != "Artist - Live" || targets[0].Album != "Album - 2024" {
		t.Fatalf("targets = %#v", targets)
	}
}
