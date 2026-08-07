package tui

import (
	"strings"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
)

type trackRef struct {
	GlobalIndex int
	AlbumIndex  int
	TrackIndex  int
	Album       lastfm.Album
	Track       lastfm.Track
}

func (m model) flattenedTracks() []trackRef {
	albums := m.selectedAlbums
	if len(albums) == 0 && m.selectedAlbum.Title != "" {
		albums = []lastfm.Album{m.selectedAlbum}
	}

	var refs []trackRef
	global := 0
	for albumIndex, album := range albums {
		for trackIndex, track := range album.Tracks {
			refs = append(refs, trackRef{
				GlobalIndex: global,
				AlbumIndex:  albumIndex,
				TrackIndex:  trackIndex,
				Album:       album,
				Track:       track,
			})
			global++
		}
	}
	return refs
}

func (m model) selectedTrackCount() int {
	count := 0
	for _, ref := range m.flattenedTracks() {
		if m.trackSelected[ref.GlobalIndex] {
			count++
		}
	}
	return count
}

func (m *model) initialiseTrackSelection() {
	m.trackCursor = 0
	m.trackSelected = map[int]bool{}
	m.albumLoops = map[int]int{}

	perAlbum := map[int]int{}
	for _, ref := range m.flattenedTracks() {
		if _, ok := m.albumLoops[ref.AlbumIndex]; !ok {
			m.albumLoops[ref.AlbumIndex] = maxInt(1, m.loopCount)
		}
		if m.trackLimit == 0 || perAlbum[ref.AlbumIndex] < m.trackLimit {
			m.trackSelected[ref.GlobalIndex] = true
			perAlbum[ref.AlbumIndex]++
		}
	}
}

func (m *model) selectAllTracks(selectAll bool) {
	m.trackSelected = map[int]bool{}
	if !selectAll {
		return
	}
	for _, ref := range m.flattenedTracks() {
		m.trackSelected[ref.GlobalIndex] = true
	}
}

func (m model) loopForAlbum(index int) int {
	if value := m.albumLoops[index]; value > 0 {
		return value
	}
	return maxInt(1, m.loopCount)
}

func (m model) mixedLoops() bool {
	if len(m.albumLoops) < 2 {
		return false
	}
	first := 0
	for _, loop := range m.albumLoops {
		if first == 0 {
			first = loop
		} else if loop != first {
			return true
		}
	}
	return false
}

func (m model) queueRecord(status string) sessionstore.Record {
	queue := make([]sessionstore.Track, 0, len(m.scrobbleQueue))
	for _, item := range m.scrobbleQueue {
		queue = append(queue, sessionstore.Track{
			Artist: item.Artist, Title: item.Title, Album: item.Album,
			AlbumIndex: item.AlbumIndex, AlbumTotal: item.AlbumTotal,
			TrackIndex: item.TrackIndex, TrackTotal: item.TrackTotal,
			LoopIndex: item.LoopIndex, LoopTotal: item.LoopTotal,
			Failed: item.Failed, Error: item.ErrMsg,
		})
	}
	started := m.scrobbleStarted
	if started.IsZero() {
		started = time.Now()
	}
	record := sessionstore.Record{
		ID: sessionstore.NewID(started), Mode: m.modeChoice, Profile: m.cfg.Profile,
		StartedAt: started, Status: status, Queue: queue,
		Completed: minInt(m.scrobbleIdx, len(queue)), Failures: len(m.failures),
		SkippedDuplicates: m.skippedDuplicates, Loop: m.loopCount,
		Interval: m.interval, Limit: m.limitLabel(),
	}
	if status == "complete" || status == "cancelled" || status == "failed" {
		record.CompletedAt = time.Now()
	}
	return record
}

func queueFromRecord(record sessionstore.Record) []queuedTrack {
	queue := make([]queuedTrack, 0, len(record.Queue))
	for _, item := range record.Queue {
		queue = append(queue, queuedTrack{
			Artist: item.Artist, Title: item.Title, Album: item.Album,
			AlbumIndex: item.AlbumIndex, AlbumTotal: item.AlbumTotal,
			TrackIndex: item.TrackIndex, TrackTotal: item.TrackTotal,
			LoopIndex: item.LoopIndex, LoopTotal: item.LoopTotal,
			Failed: item.Failed, ErrMsg: item.Error,
		})
	}
	return queue
}

func (m model) limitLabel() string {
	total := len(m.flattenedTracks())
	selected := m.selectedTrackCount()
	if total == 0 && len(m.scrobbleQueue) > 0 {
		if strings.TrimSpace(m.lastRecord.Limit) != "" {
			return m.lastRecord.Limit
		}
		return "all"
	}
	if total > 0 && selected < total {
		return strings.TrimSpace(strings.Join([]string{itoa(selected), "selected"}, " "))
	}
	return "all"
}

func (m model) previewTrackCount() int {
	if selected := m.selectedTrackCount(); selected > 0 {
		return selected
	}
	seen := map[string]bool{}
	for _, item := range m.scrobbleQueue {
		key := strings.ToLower(item.Artist + "\x00" + item.Album + "\x00" + itoa(item.TrackIndex) + "\x00" + item.Title)
		seen[key] = true
	}
	return len(seen)
}

func (m *model) resetQueueForRun() {
	for index := range m.scrobbleQueue {
		m.scrobbleQueue[index].Attempts = 0
		m.scrobbleQueue[index].Failed = false
		m.scrobbleQueue[index].ErrMsg = ""
	}
	m.scrobbleIdx = 0
	m.failures = nil
	m.skippedDuplicates = 0
	m.err = nil
}

func itoa(value int) string {
	if value == 0 {
		return "0"
	}
	negative := value < 0
	if negative {
		value = -value
	}
	var digits [32]byte
	i := len(digits)
	for value > 0 {
		i--
		digits[i] = byte('0' + value%10)
		value /= 10
	}
	if negative {
		i--
		digits[i] = '-'
	}
	return string(digits[i:])
}

// restoreRecordForEdit reconstructs albums and unique tracks from a saved
// queue so the user can change track selection and per-album loops before
// running it again. Repeated loop entries are collapsed back to one track row.
func (m *model) restoreRecordForEdit(record sessionstore.Record) {
	type albumBuild struct {
		album lastfm.Album
		loops int
		seen  map[string]bool
	}
	var builds []*albumBuild
	byKey := map[string]*albumBuild{}
	for _, item := range record.Queue {
		key := strings.ToLower(strings.TrimSpace(item.Artist) + "\x00" + strings.TrimSpace(item.Album))
		build := byKey[key]
		if build == nil {
			build = &albumBuild{album: lastfm.Album{Artist: item.Artist, Title: item.Album}, loops: maxInt(1, item.LoopTotal), seen: map[string]bool{}}
			byKey[key] = build
			builds = append(builds, build)
		}
		if item.LoopTotal > build.loops {
			build.loops = item.LoopTotal
		}
		trackKey := strings.ToLower(itoa(item.TrackIndex) + "\x00" + item.Title)
		if !build.seen[trackKey] {
			build.seen[trackKey] = true
			build.album.Tracks = append(build.album.Tracks, lastfm.Track{Title: item.Title})
		}
	}
	m.selectedAlbums = make([]lastfm.Album, 0, len(builds))
	m.albumLoops = map[int]int{}
	for index, build := range builds {
		m.selectedAlbums = append(m.selectedAlbums, build.album)
		m.albumLoops[index] = maxInt(1, build.loops)
	}
	if len(m.selectedAlbums) > 0 {
		m.selectedAlbum = m.selectedAlbums[0]
	} else {
		m.selectedAlbum = lastfm.Album{}
	}
	m.trackSelected = map[int]bool{}
	for _, ref := range m.flattenedTracks() {
		m.trackSelected[ref.GlobalIndex] = true
	}
	m.trackCursor = 0
	m.loopCount = maxInt(1, record.Loop)
	m.interval = record.Interval
	m.modeChoice = record.Mode
	m.modeIndex = modeIndex(record.Mode)
	m.scrobbleQueue = nil
	m.scrobbleIdx = 0
	m.failures = nil
	m.lastRecord = record
	m.err = nil
	m.previewStatus = "Edit the queue, then press Enter to preview the re-run."
}
