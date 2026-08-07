package sessionstore

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Track struct {
	Artist     string `json:"artist"`
	Title      string `json:"title"`
	Album      string `json:"album"`
	AlbumIndex int    `json:"album_index"`
	AlbumTotal int    `json:"album_total"`
	TrackIndex int    `json:"track_index"`
	TrackTotal int    `json:"track_total"`
	LoopIndex  int    `json:"loop_index"`
	LoopTotal  int    `json:"loop_total"`
	Failed     bool   `json:"failed,omitempty"`
	Error      string `json:"error,omitempty"`
}

type Record struct {
	ID                string        `json:"id"`
	Mode              string        `json:"mode"`
	Profile           string        `json:"profile,omitempty"`
	StartedAt         time.Time     `json:"started_at"`
	CompletedAt       time.Time     `json:"completed_at,omitempty"`
	Status            string        `json:"status"` // pending, complete, cancelled, failed
	Queue             []Track       `json:"queue"`
	Completed         int           `json:"completed"`
	Failures          int           `json:"failures"`
	SkippedDuplicates int           `json:"skipped_duplicates,omitempty"`
	Loop              int           `json:"loop"`
	Interval          time.Duration `json:"interval"`
	Limit             string        `json:"limit"`
}

type Store struct{ dir string }

func New(dir string) Store { return Store{dir: dir} }

func (s Store) ensure() error       { return os.MkdirAll(s.dir, 0700) }
func (s Store) historyPath() string { return filepath.Join(s.dir, "history.json") }
func (s Store) pendingPath() string { return filepath.Join(s.dir, "pending-session.json") }

func readJSON(path string, value any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if len(data) == 0 {
		return errors.New("empty JSON file")
	}
	return json.Unmarshal(data, value)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, append(data, '\n'), 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (s Store) LoadHistory() ([]Record, error) {
	var records []Record
	if err := readJSON(s.historyPath(), &records); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	sort.SliceStable(records, func(i, j int) bool { return records[i].StartedAt.After(records[j].StartedAt) })
	return records, nil
}

func (s Store) SaveHistory(records []Record) error {
	if err := s.ensure(); err != nil {
		return err
	}
	return writeJSON(s.historyPath(), records)
}

func (s Store) Append(record Record) error {
	records, err := s.LoadHistory()
	if err != nil {
		return err
	}
	records = append([]Record{record}, records...)
	if len(records) > 500 {
		records = records[:500]
	}
	return s.SaveHistory(records)
}

func (s Store) Delete(id string) error {
	records, err := s.LoadHistory()
	if err != nil {
		return err
	}
	out := records[:0]
	for _, record := range records {
		if record.ID != id {
			out = append(out, record)
		}
	}
	return s.SaveHistory(out)
}

func (s Store) SavePending(record Record) error {
	if err := s.ensure(); err != nil {
		return err
	}
	record.Status = "pending"
	return writeJSON(s.pendingPath(), record)
}

func (s Store) LoadPending() (*Record, error) {
	var record Record
	if err := readJSON(s.pendingPath(), &record); err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	if len(record.Queue) == 0 {
		return nil, fmt.Errorf("pending session has an empty queue")
	}
	return &record, nil
}

func (s Store) ClearPending() error {
	err := os.Remove(s.pendingPath())
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func NewID(now time.Time) string { return now.UTC().Format("20060102T150405.000000000Z") }
