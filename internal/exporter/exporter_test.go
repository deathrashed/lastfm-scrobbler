package exporter

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
)

func TestExportWritesAllFormats(t *testing.T) {
	dir := t.TempDir()
	record := sessionstore.Record{
		ID: "session", StartedAt: time.Unix(1700000000, 0), Mode: "manual",
		Queue: []sessionstore.Track{{Artist: "Slayer", Album: "Hell Awaits", Title: "Hell Awaits"}},
	}
	paths, err := Export(record, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 4 {
		t.Fatalf("paths = %#v", paths)
	}
	for _, path := range paths {
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("missing %s: %v", filepath.Base(path), err)
		}
	}
}
