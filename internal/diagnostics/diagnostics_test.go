package diagnostics

import (
	"archive/zip"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
)

func TestCreateRedactsSecrets(t *testing.T) {
	dir := t.TempDir()
	logPath := filepath.Join(dir, "scrobbler.log")
	if err := os.WriteFile(logPath, []byte("safe log\n"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := config.Config{APIKey: "1234567890abcdef", APISecret: "supersecret", Password: "pw-secret-XYZ", SessionKey: "session-secret-XYZ", Username: "user", ExportDir: dir}
	path, err := Create(cfg, nil, logPath, "v1.0.0", "test")
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	var combined strings.Builder
	for _, entry := range archive.File {
		f, _ := entry.Open()
		data := make([]byte, entry.UncompressedSize64)
		_, _ = f.Read(data)
		_ = f.Close()
		combined.Write(data)
	}
	text := combined.String()
	for _, secret := range []string{"supersecret", "pw-secret-XYZ", "session-secret-XYZ"} {
		if strings.Contains(text, secret) {
			t.Fatalf("bundle leaked %q", secret)
		}
	}
}
