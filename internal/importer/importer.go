package importer

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type Target struct {
	Artist string `json:"artist"`
	Album  string `json:"album"`
}

var audioExt = map[string]bool{
	".mp3": true, ".flac": true, ".m4a": true, ".alac": true,
	".wav": true, ".aiff": true, ".aif": true, ".ogg": true, ".opus": true,
}

func splitArtistAlbum(value string) (Target, bool) {
	for _, separator := range []string{" — ", " - ", "\t", "|"} {
		parts := strings.SplitN(value, separator, 2)
		if len(parts) == 2 {
			artist, album := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if artist != "" && album != "" {
				return Target{Artist: artist, Album: album}, true
			}
		}
	}
	return Target{}, false
}

func Load(path string) ([]Target, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	var targets []Target
	if info.IsDir() {
		targets, err = loadDirectory(path)
	} else {
		switch strings.ToLower(filepath.Ext(path)) {
		case ".csv", ".tsv":
			targets, err = loadCSV(path)
		case ".json":
			targets, err = loadJSON(path)
		case ".m3u", ".m3u8":
			targets, err = loadM3U(path)
		default:
			targets, err = loadText(path)
		}
	}
	if err != nil {
		return nil, err
	}
	targets = dedupe(targets, info.IsDir())
	if len(targets) == 0 {
		return nil, fmt.Errorf("no Artist - Album entries could be read from %s", path)
	}
	return targets, nil
}

func loadText(path string) ([]Target, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var out []Target
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if target, ok := splitArtistAlbum(line); ok {
			out = append(out, target)
		}
	}
	return out, scanner.Err()
}

func loadCSV(path string) ([]Target, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	reader := csv.NewReader(file)
	if strings.ToLower(filepath.Ext(path)) == ".tsv" {
		reader.Comma = '\t'
	}
	reader.FieldsPerRecord = -1
	var out []Target
	first := true
	artistCol, albumCol := 0, 1
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if first {
			first = false
			for i, cell := range record {
				switch strings.ToLower(strings.TrimSpace(cell)) {
				case "artist", "artist_name":
					artistCol = i
				case "album", "album_name", "title":
					albumCol = i
				}
			}
			if len(record) > max(artistCol, albumCol) && strings.EqualFold(strings.TrimSpace(record[artistCol]), "artist") {
				continue
			}
		}
		if len(record) <= max(artistCol, albumCol) {
			continue
		}
		artist, album := strings.TrimSpace(record[artistCol]), strings.TrimSpace(record[albumCol])
		if artist != "" && album != "" {
			out = append(out, Target{Artist: artist, Album: album})
		}
	}
	return out, nil
}

func loadJSON(path string) ([]Target, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var targets []Target
	if json.Unmarshal(data, &targets) == nil && len(targets) > 0 {
		return targets, nil
	}
	var stringsList []string
	if json.Unmarshal(data, &stringsList) == nil {
		for _, line := range stringsList {
			if target, ok := splitArtistAlbum(line); ok {
				targets = append(targets, target)
			}
		}
		return targets, nil
	}
	var wrapper struct {
		Albums []Target `json:"albums"`
	}
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return nil, err
	}
	return wrapper.Albums, nil
}

func loadM3U(path string) ([]Target, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	base := filepath.Dir(path)
	var out []Target
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		raw := strings.TrimSpace(scanner.Text())
		line := raw
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if isPathEntry(raw) {
			if !filepath.IsAbs(line) {
				line = filepath.Join(base, line)
			}
			if target, ok := targetFromAudioPath(line); ok {
				out = append(out, target)
			}
			continue
		}
		if target, ok := splitArtistAlbum(line); ok {
			out = append(out, target)
		}
	}
	return out, scanner.Err()
}

func loadDirectory(path string) ([]Target, error) {
	if hasAudioFiles(path) {
		return []Target{{Artist: filepath.Base(filepath.Dir(path)), Album: filepath.Base(path)}}, nil
	}
	entries, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	var out []Target
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		albumDir := filepath.Join(path, entry.Name())
		if hasAudioFiles(albumDir) {
			out = append(out, Target{Artist: filepath.Base(path), Album: entry.Name()})
			continue
		}
		artistEntries, _ := os.ReadDir(albumDir)
		for _, albumEntry := range artistEntries {
			if albumEntry.IsDir() && hasAudioFiles(filepath.Join(albumDir, albumEntry.Name())) {
				out = append(out, Target{Artist: entry.Name(), Album: albumEntry.Name()})
			}
		}
	}
	return out, nil
}

func hasAudioFiles(dir string) bool {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, entry := range entries {
		if !entry.IsDir() && audioExt[strings.ToLower(filepath.Ext(entry.Name()))] {
			return true
		}
	}
	return false
}

func isPathEntry(value string) bool {
	value = strings.TrimSpace(value)
	return filepath.IsAbs(value) || strings.ContainsAny(value, `/\\`) || strings.HasPrefix(value, "./") || strings.HasPrefix(value, "../")
}

func targetFromAudioPath(path string) (Target, bool) {
	albumDir := filepath.Dir(path)
	album := filepath.Base(albumDir)
	artist := filepath.Base(filepath.Dir(albumDir))
	if artist == "" || album == "" || artist == "." || album == "." {
		return Target{}, false
	}
	return Target{Artist: artist, Album: album}, true
}

func dedupe(in []Target, sortTargets bool) []Target {
	seen := map[string]bool{}
	var out []Target
	for _, target := range in {
		target.Artist = strings.TrimSpace(target.Artist)
		target.Album = strings.TrimSpace(target.Album)
		if target.Artist == "" || target.Album == "" {
			continue
		}
		key := strings.ToLower(target.Artist + "\x00" + target.Album)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, target)
	}
	if sortTargets {
		sort.SliceStable(out, func(i, j int) bool {
			if strings.EqualFold(out[i].Artist, out[j].Artist) {
				return strings.ToLower(out[i].Album) < strings.ToLower(out[j].Album)
			}
			return strings.ToLower(out[i].Artist) < strings.ToLower(out[j].Artist)
		})
	}
	return out
}
