package exporter

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
)

var unsafeName = regexp.MustCompile(`[^a-zA-Z0-9._-]+`)

func safeName(value string) string {
	value = strings.TrimSpace(value)
	value = unsafeName.ReplaceAllString(value, "-")
	value = strings.Trim(value, "-._")
	if value == "" {
		return "session"
	}
	if len(value) > 80 {
		value = value[:80]
	}
	return value
}

func baseName(record sessionstore.Record) string {
	artist, album := "", ""
	if len(record.Queue) > 0 {
		artist, album = record.Queue[0].Artist, record.Queue[0].Album
	}
	stamp := record.StartedAt
	if stamp.IsZero() {
		stamp = time.Now()
	}
	return fmt.Sprintf("%s-%s-%s", stamp.Format("2006-01-02-150405"), safeName(artist), safeName(album))
}

// Export writes JSON, CSV, text and M3U8 files for a session.
func Export(record sessionstore.Record, dir string) ([]string, error) {
	if strings.TrimSpace(dir) == "" {
		dir = "."
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	base := filepath.Join(dir, baseName(record))
	var paths []string

	jsonPath := base + ".json"
	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(jsonPath, append(data, '\n'), 0644); err != nil {
		return nil, err
	}
	paths = append(paths, jsonPath)

	csvPath := base + ".csv"
	csvFile, err := os.Create(csvPath)
	if err != nil {
		return nil, err
	}
	writer := csv.NewWriter(csvFile)
	_ = writer.Write([]string{"artist", "album", "track", "album_index", "track_index", "loop", "failed", "error"})
	for _, track := range record.Queue {
		_ = writer.Write([]string{
			track.Artist, track.Album, track.Title,
			strconv.Itoa(track.AlbumIndex), strconv.Itoa(track.TrackIndex), strconv.Itoa(track.LoopIndex),
			strconv.FormatBool(track.Failed), track.Error,
		})
	}
	writer.Flush()
	closeErr := csvFile.Close()
	if err := writer.Error(); err != nil {
		return nil, err
	}
	if closeErr != nil {
		return nil, closeErr
	}
	paths = append(paths, csvPath)

	txtPath := base + ".txt"
	var text strings.Builder
	for _, track := range record.Queue {
		fmt.Fprintf(&text, "%s — %s — %s\n", track.Artist, track.Album, track.Title)
	}
	if err := os.WriteFile(txtPath, []byte(text.String()), 0644); err != nil {
		return nil, err
	}
	paths = append(paths, txtPath)

	m3uPath := base + ".m3u8"
	var m3u strings.Builder
	m3u.WriteString("#EXTM3U\n")
	for _, track := range record.Queue {
		fmt.Fprintf(&m3u, "#EXTINF:-1,%s - %s\nlastfm://%s/%s/%s\n", track.Artist, track.Title, urlPart(track.Artist), urlPart(track.Album), urlPart(track.Title))
	}
	if err := os.WriteFile(m3uPath, []byte(m3u.String()), 0644); err != nil {
		return nil, err
	}
	paths = append(paths, m3uPath)

	return paths, nil
}

func urlPart(value string) string {
	value = strings.ReplaceAll(strings.TrimSpace(value), " ", "+")
	return strings.ReplaceAll(value, "/", "_")
}
