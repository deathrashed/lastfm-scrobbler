package diagnostics

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
)

type Summary struct {
	GeneratedAt   time.Time `json:"generated_at"`
	Version       string    `json:"version"`
	Commit        string    `json:"commit"`
	GoVersion     string    `json:"go_version"`
	OS            string    `json:"os"`
	Architecture  string    `json:"architecture"`
	Profile       string    `json:"profile"`
	CredentialSrc string    `json:"credential_source"`
	EnvPath       string    `json:"env_path"`
	DataDir       string    `json:"data_dir"`
	HistoryCount  int       `json:"history_count"`
	LogIncluded   bool      `json:"log_included"`
}

// Create writes a redacted support bundle. Passwords, API secrets, session
// keys, and complete API keys are never included.
func Create(cfg config.Config, history []sessionstore.Record, logPath, version, commit string) (string, error) {
	dir := cfg.ExportDir
	if strings.TrimSpace(dir) == "" {
		dir = config.DataDir()
	}
	dir = config.ExpandPath(dir)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	stamp := time.Now().Format("20060102-150405")
	path := filepath.Join(dir, "lastfm-scrobbler-diagnostics-"+stamp+".zip")
	file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return "", err
	}
	archive := zip.NewWriter(file)
	closeWithError := func(current error) error {
		zipErr := archive.Close()
		fileErr := file.Close()
		if current != nil {
			return current
		}
		if zipErr != nil {
			return zipErr
		}
		return fileErr
	}

	summary := Summary{
		GeneratedAt: time.Now().UTC(), Version: version, Commit: commit,
		GoVersion: runtime.Version(), OS: runtime.GOOS, Architecture: runtime.GOARCH,
		Profile: cfg.Profile, CredentialSrc: cfg.CredentialSource,
		EnvPath: cfg.EnvPath, DataDir: config.DataDir(), HistoryCount: len(history),
	}
	if info, statErr := os.Stat(logPath); statErr == nil && !info.IsDir() {
		summary.LogIncluded = true
	}
	if err := addJSON(archive, "diagnostics.json", summary); err != nil {
		return "", closeWithError(err)
	}
	if err := addText(archive, "config.redacted.env", redactedConfig(cfg)); err != nil {
		return "", closeWithError(err)
	}
	if err := addJSON(archive, "history-summary.json", redactHistory(history)); err != nil {
		return "", closeWithError(err)
	}
	if summary.LogIncluded {
		if err := addFileTail(archive, "scrobbler.log", logPath, 512*1024); err != nil {
			return "", closeWithError(err)
		}
	}
	readme := `Last.fm Scrobbler diagnostics bundle

This archive is intended for troubleshooting. It contains:
- runtime and version information
- redacted configuration
- summary-only session history
- the tail of the application log, when available

It does NOT include the Last.fm password, API secret, session key, or full API key.
Review the files before sharing the archive.
`
	if err := addText(archive, "README.txt", readme); err != nil {
		return "", closeWithError(err)
	}
	if err := closeWithError(nil); err != nil {
		return "", err
	}
	return path, nil
}

func redactedConfig(cfg config.Config) string {
	lines := []string{
		"API_KEY=" + redactPublic(cfg.APIKey),
		"API_SECRET=<redacted>",
		"LASTFM_USERNAME=" + cfg.Username,
		"LASTFM_PASSWORD=<redacted>",
		"LASTFM_SESSION_KEY=<redacted>",
		"LASTFM_PROFILE=" + cfg.Profile,
		"LASTFM_CREDENTIAL_SOURCE=" + cfg.CredentialSource,
		"LASTFM_ENV_FILE=" + cfg.EnvPath,
		"SCROBBLE_INTERVAL=" + cfg.DefaultInterval.String(),
		fmt.Sprintf("SCROBBLE_LIMIT=%d", cfg.DefaultLimit),
		fmt.Sprintf("SCROBBLE_LOOP=%d", cfg.DefaultLoop),
		fmt.Sprintf("SCROBBLE_RETRIES=%d", cfg.RetryCount),
		"SCROBBLE_RETRY_DELAY=" + cfg.RetryDelay.String(),
		"SCROBBLE_DUPLICATE_GUARD=" + cfg.DuplicateGuard.String(),
		fmt.Sprintf("SCROBBLE_NOTIFY=%t", cfg.Notify),
		fmt.Sprintf("SCROBBLE_COMPACT_HEADER=%t", cfg.CompactHeader),
		fmt.Sprintf("SCROBBLE_CLEAN_DISCOGRAPHY=%t", cfg.CleanDiscography),
		"SCROBBLE_EXPORT_DIR=" + cfg.ExportDir,
		fmt.Sprintf("SCROBBLE_MOUSE=%t", cfg.MouseEnabled),
		"SCROBBLER_UPDATE_URL=" + cfg.UpdateURL,
	}
	return strings.Join(lines, "\n") + "\n"
}

func redactPublic(value string) string {
	value = strings.TrimSpace(value)
	if len(value) <= 8 {
		if value == "" {
			return "<missing>"
		}
		return "<configured>"
	}
	return value[:4] + "…" + value[len(value)-4:]
}

type historySummary struct {
	ID          string    `json:"id"`
	Mode        string    `json:"mode"`
	Profile     string    `json:"profile,omitempty"`
	StartedAt   time.Time `json:"started_at"`
	Status      string    `json:"status"`
	Albums      int       `json:"albums"`
	Tracks      int       `json:"tracks"`
	Completed   int       `json:"completed"`
	Failures    int       `json:"failures"`
	FirstArtist string    `json:"first_artist,omitempty"`
}

func redactHistory(records []sessionstore.Record) []historySummary {
	out := make([]historySummary, 0, len(records))
	for _, record := range records {
		seen := map[string]bool{}
		firstArtist := ""
		for _, track := range record.Queue {
			if firstArtist == "" {
				firstArtist = track.Artist
			}
			seen[strings.ToLower(track.Artist+"\x00"+track.Album)] = true
		}
		out = append(out, historySummary{
			ID: record.ID, Mode: record.Mode, Profile: record.Profile,
			StartedAt: record.StartedAt, Status: record.Status,
			Albums: len(seen), Tracks: len(record.Queue), Completed: record.Completed,
			Failures: record.Failures, FirstArtist: firstArtist,
		})
	}
	return out
}

func addJSON(archive *zip.Writer, name string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return addBytes(archive, name, append(data, '\n'))
}

func addText(archive *zip.Writer, name, value string) error {
	return addBytes(archive, name, []byte(value))
}

func addBytes(archive *zip.Writer, name string, data []byte) error {
	entry, err := archive.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate})
	if err != nil {
		return err
	}
	_, err = entry.Write(data)
	return err
}

func addFileTail(archive *zip.Writer, name, path string, maxBytes int64) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return err
	}
	start := info.Size() - maxBytes
	if start < 0 {
		start = 0
	}
	if _, err := file.Seek(start, io.SeekStart); err != nil {
		return err
	}
	entry, err := archive.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Deflate})
	if err != nil {
		return err
	}
	_, err = io.Copy(entry, file)
	return err
}
