package tui

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/connection"
	"github.com/deathrashed/lastfm-scrobbler/internal/diagnostics"
	"github.com/deathrashed/lastfm-scrobbler/internal/exporter"
	"github.com/deathrashed/lastfm-scrobbler/internal/importer"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/logging"
	"github.com/deathrashed/lastfm-scrobbler/internal/platform"
	"github.com/deathrashed/lastfm-scrobbler/internal/scrobbler"
	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
	"github.com/deathrashed/lastfm-scrobbler/internal/updater"
	"github.com/deathrashed/lastfm-scrobbler/internal/version"
)

type scrobbleResultMsg struct {
	idx   int
	track queuedTrack
	err   error
}

type scrobbleCancelledMsg struct{}

type searchResultMsg struct {
	albums []lastfm.Album
	err    error
	direct bool
}

type loadedAlbumsMsg struct {
	albums []lastfm.Album
	err    error
}

type similarResultMsg struct {
	albums []lastfm.Album
	err    error
}

type scrobblePreparedMsg struct {
	queue   []queuedTrack
	skipped int
	err     error
}

type exportResultMsg struct {
	paths []string
	err   error
}

type connectionTestMsg struct {
	report connection.Report
}

type diagnosticsResultMsg struct {
	path string
	err  error
}

type updateCheckMsg struct {
	result updater.Result
	err    error
}

type headerURLMsg struct {
	err error
}

type filePickedMsg struct {
	path   string
	target string
	err    error
}

func isPlainTerminal() bool { return os.Getenv("NO_TUI") == "1" || os.Getenv("TERM") == "dumb" }

func updateModel(m model, msg tea.Msg) (tea.Model, tea.Cmd) {
	if isPlainTerminal() {
		return m.updatePlainTerminal(msg)
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		if m.helpVisible {
			switch keyMsg.String() {
			case "?", "esc", "enter":
				m.helpVisible = false
				return m, nil
			case "q", "ctrl+c":
				return m, tea.Quit
			default:
				return m, nil
			}
		}
		if keyMsg.String() == "?" && m.helpAllowed() {
			m.helpVisible = true
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.stage {
		case stageInput:
			return m.updateInput(msg)
		case stageImportSource:
			return m.updateImportSource(msg)
		case stageSearch:
			return m.updateSearch(msg)
		case stageResults:
			return m.updateResults(msg)
		case stageDiscographySelect:
			return m.updateDiscographySelect(msg)
		case stageTrackSelect:
			return m.updateTrackSelect(msg)
		case stagePreview:
			return m.updatePreview(msg)
		case stageConfig:
			return m.updateConfig(msg)
		case stageAdvancedConfig:
			return m.updateAdvancedConfig(msg)
		case stageEnvPath:
			return m.updateEnvPath(msg)
		case stageScrobbling:
			return m.updateScrobbling(msg)
		case stageDone:
			return m.updateDone(msg)
		case stageHistory:
			return m.updateHistory(msg)
		case stageRecovery:
			return m.updateRecovery(msg)
		case stageSimilarSelect:
			return m.updateSimilarSelect(msg)
		case stageProfiles:
			return m.updateProfiles(msg)
		case stageProfileName:
			return m.updateProfileName(msg)
		case stageInfo:
			return m.updateInfo(msg)
		case stageConnectionTest:
			return m.updateConnectionTest(msg)
		case stageDiagnostics:
			return m.updateDiagnostics(msg)
		case stageUpdateCheck:
			return m.updateUpdateCheck(msg)
		}
	case tea.MouseMsg:
		if !m.cfg.MouseEnabled {
			return m, nil
		}
		return m.updateMouse(msg)
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.headerURLHover = false
		return m, nil
	case spinner.TickMsg:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	case searchResultMsg:
		m.searching = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		if m.modeChoice == "discography" {
			m.discography = msg.albums
			m.discographyCursor = 0
			m.discographySelected = map[int]bool{}
			m.discographyFilter = ""
			m.stage = stageDiscographySelect
			return m, nil
		}
		if msg.direct {
			m.selectedAlbums = msg.albums
			if len(msg.albums) > 0 {
				m.selectedAlbum = msg.albums[0]
			}
			m.initialiseTrackSelection()
			m.stage = stageTrackSelect
			return m, nil
		}
		m.results = msg.albums
		m.resultsCursor = 0
		m.stage = stageResults
		return m, nil
	case loadedAlbumsMsg:
		m.searching = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.selectedAlbums = msg.albums
		if len(msg.albums) > 0 {
			m.selectedAlbum = msg.albums[0]
		}
		m.initialiseTrackSelection()
		m.stage = stageTrackSelect
		return m, nil
	case similarResultMsg:
		m.searching = false
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		m.similar = msg.albums
		m.similarCursor = 0
		m.stage = stageSimilarSelect
		return m, nil
	case scrobblePreparedMsg:
		if m.sessionCtx != nil && m.sessionCtx.Err() != nil {
			return m, nil
		}
		if msg.err != nil {
			m.err = msg.err
			m.stage = stagePreview
			return m, nil
		}
		m.scrobbleQueue = msg.queue
		m.skippedDuplicates = msg.skipped
		m.scrobbleIdx = 0
		m.failures = nil
		m.scrobbleStarted = time.Now()
		m.stage = stageScrobbling
		m.err = nil
		_ = m.store.SavePending(m.queueRecord("pending"))
		return m, m.scrobbleNext()
	case scrobbleResultMsg:
		if m.stage != stageScrobbling {
			return m, nil
		}
		return m.updateScrobbling(msg)
	case scrobbleCancelledMsg:
		if m.stage == stageScrobbling {
			m.cancelScrobbleSession()
		}
		return m, nil
	case exportResultMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		if len(msg.paths) > 0 {
			m.exportStatus = "Exported to " + filepath.Dir(msg.paths[0])
			m.historyStatus = m.exportStatus
			m.previewStatus = m.exportStatus
		}
		return m, nil
	case connectionTestMsg:
		m.connectionTesting = false
		m.connectionReport = msg.report
		return m, nil
	case diagnosticsResultMsg:
		m.diagnosticsBusy = false
		if msg.err != nil {
			m.err = msg.err
			logging.Printf("diagnostics export failed: %v", msg.err)
		} else {
			m.err = nil
			m.diagnosticsPath = msg.path
		}
		return m, nil
	case updateCheckMsg:
		m.updateChecking = false
		if msg.err != nil {
			m.err = msg.err
			logging.Printf("update check failed: %v", msg.err)
		} else {
			m.err = nil
			m.updateResult = msg.result
		}
		return m, nil
	case headerURLMsg:
		m.err = msg.err
		if msg.err != nil {
			logging.Printf("header URL open failed: %v", msg.err)
		}
		return m, nil
	case filePickedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.err = nil
		switch msg.target {
		case "search":
			m.searchInput.SetValue(msg.path)
			m.searchInput.CursorEnd()
			return m, m.searchInput.Focus()
		case "import":
			m.stage = stageSearch
			m.searchInput.SetValue(msg.path)
			m.searchInput.CursorEnd()
			return m, m.searchInput.Focus()
		case "env":
			m.envInput.SetValue(msg.path)
			m.envInput.CursorEnd()
			return m, m.envInput.Focus()
		case "export":
			m.cfg.ExportDir = msg.path
			m.loadAdvancedField()
			return m, m.configInput.Focus()
		}
	}
	return m, nil
}

func (m model) helpAllowed() bool {
	if m.discographyFiltering || m.advancedEditing {
		return false
	}
	switch m.stage {
	case stageSearch, stageConfig, stageEnvPath, stageProfileName:
		return false
	default:
		return true
	}
}

func (m model) updatePlainTerminal(msg tea.Msg) (tea.Model, tea.Cmd) {
	if k, ok := msg.(tea.KeyMsg); ok && (k.String() == "q" || k.String() == "ctrl+c") {
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "up", "k":
		m.modeIndex = (m.modeIndex + 2) % 3
	case "right", "down", "j":
		m.modeIndex = (m.modeIndex + 1) % 3
	case "enter":
		return m.activateMode(m.modeIndex)
	case "m":
		return m.activateMode(0)
	case "d":
		return m.activateMode(1)
	case "f":
		return m.activateMode(2)
	case "c":
		return m.openConfig()
	case "h":
		m.stage = stageHistory
		m.modeChoice = "history"
		m.historyCursor = 0
		m.historyStatus = ""
	case "p":
		m.stage = stageProfiles
		m.modeChoice = "profiles"
		m.profileCursor = indexOf(m.profiles, m.cfg.Profile)
	case "i":
		m.stage = stageInfo
		m.modeChoice = "info"
		m.infoIndex = 0
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func modeName(i int) string {
	switch i {
	case 1:
		return "discography"
	case 2:
		return "file"
	default:
		return "manual"
	}
}

func modeIndex(name string) int {
	switch name {
	case "discography":
		return 1
	case "file":
		return 2
	default:
		return 0
	}
}

func (m model) activateMode(i int) (tea.Model, tea.Cmd) {
	m.clearSessionSelection(false)
	m.modeIndex = i
	m.modeChoice = modeName(i)
	m.err = nil
	m.searchInput.SetValue("")
	if m.modeChoice == "file" {
		m.stage = stageImportSource
		m.importSourceIndex = 0
	} else {
		m.stage = stageSearch
	}
	switch m.modeChoice {
	case "file":
		m.searchInput.Placeholder = "/path/to/list, playlist, or music folder"
	case "discography":
		m.searchInput.Placeholder = "Artist name..."
	default:
		m.searchInput.Placeholder = "Artist - Album..."
	}
	if m.stage == stageSearch {
		return m, m.searchInput.Focus()
	}
	return m, nil
}

func (m model) updateImportSource(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "up", "k":
		m.importSourceIndex = (m.importSourceIndex + 3) % 4
	case "right", "down", "j":
		m.importSourceIndex = (m.importSourceIndex + 1) % 4
	case "enter":
		m.stage = stageSearch
		placeholders := []string{
			"/path/to/TXT, CSV, TSV, or JSON list",
			"/path/to/M3U or M3U8 playlist",
			"/path/to/Artist/Album folder",
			"/path/to/Artist folder",
		}
		m.searchInput.Placeholder = placeholders[m.importSourceIndex]
		m.searchInput.SetValue("")
		return m, m.searchInput.Focus()
	case "o":
		return m, pickPathCmd("Choose a list, playlist, album folder, or artist folder", "import")
	case "esc":
		m.stage = stageInput
		m.modeChoice = ""
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateSearch(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		q := strings.TrimSpace(m.searchInput.Value())
		if q == "" {
			return m, nil
		}
		if m.modeChoice == "discography" {
			m.discographyArtist = q
		}
		m.searching = true
		m.err = nil
		return m, tea.Batch(m.runModeQuery(q), m.spinner.Tick)
	case "esc":
		if m.modeChoice == "file" {
			m.stage = stageImportSource
		} else {
			m.stage = stageInput
			m.modeChoice = ""
		}
		m.searchInput.Blur()
		m.err = nil
		return m, nil
	case "o":
		if m.modeChoice == "file" {
			return m, pickPathCmd("Choose an album-list file, playlist, or music folder", "search")
		}
	case "ctrl+c":
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m model) runModeQuery(q string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		switch m.modeChoice {
		case "discography":
			albums, err := m.client.GetDiscography(ctx, q)
			return searchResultMsg{albums: albums, err: err}
		case "file":
			albums, err := m.loadAlbumFile(ctx, q)
			return searchResultMsg{albums: albums, err: err, direct: true}
		default:
			first, second, ok := splitArtistAlbum(q)
			if ok {
				album, err := m.client.GetAlbumTracks(ctx, first, second)
				if err == nil {
					return searchResultMsg{albums: []lastfm.Album{album}, direct: true}
				}
				reversed, reverseErr := m.client.GetAlbumTracks(ctx, second, first)
				if reverseErr == nil {
					return searchResultMsg{albums: []lastfm.Album{reversed}, direct: true}
				}
				return searchResultMsg{err: err, direct: true}
			}
			albums, err := m.client.SearchAlbums(ctx, q)
			return searchResultMsg{albums: albums, err: err}
		}
	}
}

func splitArtistAlbum(value string) (string, string, bool) {
	for _, separator := range []string{" — ", " - ", "\t", "|"} {
		parts := strings.SplitN(value, separator, 2)
		if len(parts) == 2 {
			a, b := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if a != "" && b != "" {
				return a, b, true
			}
		}
	}
	return "", "", false
}

func (m model) loadAlbumFile(ctx context.Context, path string) ([]lastfm.Album, error) {
	path = config.ExpandPath(path)
	targets, err := importer.Load(path)
	if err != nil {
		return nil, err
	}
	out := make([]lastfm.Album, 0, len(targets))
	for i, target := range targets {
		album, fetchErr := m.client.GetAlbumTracks(ctx, target.Artist, target.Album)
		if fetchErr != nil {
			return nil, fmt.Errorf("entry %d (%s — %s): %w", i+1, target.Artist, target.Album, fetchErr)
		}
		out = append(out, album)
	}
	return out, nil
}

func (m model) updateResults(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := len(m.results)
	switch msg.String() {
	case "up", "k":
		if n > 0 {
			m.resultsCursor = (m.resultsCursor + n - 1) % n
		}
	case "down", "j":
		if n > 0 {
			m.resultsCursor = (m.resultsCursor + 1) % n
		}
	case "enter":
		if n > 0 {
			m.searching = true
			return m, tea.Batch(m.loadAlbums([]lastfm.Album{m.results[m.resultsCursor]}), m.spinner.Tick)
		}
	case "s":
		if n > 0 {
			return m.startSimilar(m.results[m.resultsCursor].Artist)
		}
	case "esc":
		m.stage = stageSearch
		return m, m.searchInput.Focus()
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateDiscographySelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.discographyFiltering {
		switch msg.String() {
		case "enter":
			m.discographyFilter = strings.TrimSpace(m.filterInput.Value())
			m.discographyFiltering = false
			m.filterInput.Blur()
			m.discographyCursor = 0
			return m, nil
		case "esc":
			m.discographyFiltering = false
			m.filterInput.Blur()
			return m, nil
		case "ctrl+c":
			return m, tea.Quit
		}
		var cmd tea.Cmd
		m.filterInput, cmd = m.filterInput.Update(msg)
		m.discographyFilter = m.filterInput.Value()
		m.discographyCursor = 0
		return m, cmd
	}

	visible := m.discographyVisibleIndexes()
	n := len(visible)
	switch msg.String() {
	case "up", "k":
		if n > 0 {
			m.discographyCursor = (m.discographyCursor + n - 1) % n
		}
	case "down", "j":
		if n > 0 {
			m.discographyCursor = (m.discographyCursor + 1) % n
		}
	case " ":
		if n > 0 {
			original := visible[m.discographyCursor]
			if m.discographySelected[original] {
				delete(m.discographySelected, original)
			} else {
				m.discographySelected[original] = true
			}
		}
	case "a":
		allVisibleSelected := n > 0
		for _, original := range visible {
			if !m.discographySelected[original] {
				allVisibleSelected = false
				break
			}
		}
		for _, original := range visible {
			if allVisibleSelected {
				delete(m.discographySelected, original)
			} else {
				m.discographySelected[original] = true
			}
		}
	case "c":
		m.discographyClean = !m.discographyClean
		m.discographyCursor = 0
	case "s":
		m.discographySort = (m.discographySort + 1) % 3
		m.discographyCursor = 0
	case "/", "f":
		m.discographyFiltering = true
		m.filterInput.SetValue(m.discographyFilter)
		m.filterInput.CursorEnd()
		return m, m.filterInput.Focus()
	case "+", "=":
		m.loopCount++
	case "-":
		if m.loopCount > 1 {
			m.loopCount--
		}
	case "enter":
		if n == 0 {
			return m, nil
		}
		if len(m.discographySelected) == 0 {
			m.discographySelected[visible[m.discographyCursor]] = true
		}
		chosen := make([]lastfm.Album, 0, len(m.discographySelected))
		for i, album := range m.discography {
			if m.discographySelected[i] {
				chosen = append(chosen, album)
			}
		}
		m.searching = true
		return m, tea.Batch(m.loadAlbums(chosen), m.spinner.Tick)
	case "esc":
		m.stage = stageSearch
		return m, m.searchInput.Focus()
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) discographyVisibleIndexes() []int {
	filter := strings.ToLower(strings.TrimSpace(m.discographyFilter))
	seenClean := map[string]bool{}
	indexes := make([]int, 0, len(m.discography))
	for i, album := range m.discography {
		title := strings.TrimSpace(album.Title)
		if filter != "" && !strings.Contains(strings.ToLower(title), filter) {
			continue
		}
		if m.discographyClean {
			if noisyEdition(title) {
				continue
			}
			key := cleanAlbumKey(title)
			if seenClean[key] {
				continue
			}
			seenClean[key] = true
		}
		indexes = append(indexes, i)
	}
	switch m.discographySort {
	case 1:
		sort.SliceStable(indexes, func(i, j int) bool {
			return strings.ToLower(m.discography[indexes[i]].Title) < strings.ToLower(m.discography[indexes[j]].Title)
		})
	case 2:
		sort.SliceStable(indexes, func(i, j int) bool {
			return strings.ToLower(m.discography[indexes[i]].Title) > strings.ToLower(m.discography[indexes[j]].Title)
		})
	}
	return indexes
}

func noisyEdition(title string) bool {
	lower := strings.ToLower(title)
	for _, marker := range []string{"reissue", "remaster", "bonus", "deluxe", "anniversary", "demo", "live", "anthology", "compilation", "disc 1", "disc 2", "disc one", "disc two"} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func cleanAlbumKey(title string) string {
	lower := strings.ToLower(title)
	for _, marker := range []string{"(remastered", "[remastered", "(reissue", "[reissue", "(bonus", "[bonus", "(deluxe", "[deluxe"} {
		if index := strings.Index(lower, marker); index >= 0 {
			lower = lower[:index]
		}
	}
	replacer := strings.NewReplacer("'", "", "\"", "", "-", " ", "—", " ", ":", " ", "(", " ", ")", " ", "[", " ", "]", " ")
	return strings.Join(strings.Fields(replacer.Replace(lower)), " ")
}

func (m model) loadAlbums(albums []lastfm.Album) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
		defer cancel()
		out := make([]lastfm.Album, 0, len(albums))
		for _, album := range albums {
			full, err := m.client.GetAlbumTracks(ctx, album.Artist, album.Title)
			if err != nil {
				return loadedAlbumsMsg{err: fmt.Errorf("%s — %s: %w", album.Artist, album.Title, err)}
			}
			out = append(out, full)
		}
		return loadedAlbumsMsg{albums: out}
	}
}

func (m model) updateTrackSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	refs := m.flattenedTracks()
	n := len(refs)
	switch msg.String() {
	case "up", "k":
		if n > 0 {
			m.trackCursor = (m.trackCursor + n - 1) % n
		}
	case "down", "j":
		if n > 0 {
			m.trackCursor = (m.trackCursor + 1) % n
		}
	case " ":
		if n > 0 {
			index := refs[m.trackCursor].GlobalIndex
			if m.trackSelected[index] {
				delete(m.trackSelected, index)
			} else {
				m.trackSelected[index] = true
			}
		}
	case "a":
		m.selectAllTracks(m.selectedTrackCount() != n)
	case "+", "=":
		m.loopCount++
		for album := range m.albumLoops {
			m.albumLoops[album] = m.loopCount
		}
	case "-":
		if m.loopCount > 1 {
			m.loopCount--
			for album := range m.albumLoops {
				m.albumLoops[album] = m.loopCount
			}
		}
	case "]":
		if n > 0 {
			album := refs[m.trackCursor].AlbumIndex
			m.albumLoops[album] = m.loopForAlbum(album) + 1
		}
	case "[":
		if n > 0 {
			album := refs[m.trackCursor].AlbumIndex
			if m.loopForAlbum(album) > 1 {
				m.albumLoops[album]--
			}
		}
	case "enter":
		if m.selectedTrackCount() == 0 {
			m.err = fmt.Errorf("select at least one track before continuing")
			return m, nil
		}
		m.buildScrobbleQueue()
		if len(m.scrobbleQueue) == 0 {
			m.err = fmt.Errorf("no tracks were added to the scrobble queue")
			return m, nil
		}
		m.err = nil
		m.previewStatus = ""
		m.stage = stagePreview
	case "s":
		artist := m.currentArtist()
		if artist != "" {
			return m.startSimilar(artist)
		}
	case "e":
		m.buildScrobbleQueue()
		return m, m.exportRecordCmd(m.queueRecord("preview"))
	case "esc":
		if m.modeChoice == "discography" {
			m.stage = stageDiscographySelect
		} else {
			m.stage = stageSearch
			return m, m.searchInput.Focus()
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	m.err = nil
	return m, nil
}

func (m *model) buildScrobbleQueue() {
	m.scrobbleQueue = nil
	m.albumsScrobbled = 0
	albums := m.selectedAlbums
	if len(albums) == 0 && m.selectedAlbum.Title != "" {
		albums = []lastfm.Album{m.selectedAlbum}
	}
	type albumSelection struct {
		album      lastfm.Album
		albumIndex int
		tracks     []lastfm.Track
	}
	var selections []albumSelection
	globalIndex := 0
	for albumIndex, album := range albums {
		var selected []lastfm.Track
		for _, track := range album.Tracks {
			if m.trackSelected[globalIndex] {
				selected = append(selected, track)
			}
			globalIndex++
		}
		if len(selected) > 0 {
			selections = append(selections, albumSelection{album: album, albumIndex: albumIndex, tracks: selected})
		}
	}
	m.albumsScrobbled = len(selections)
	for albumPosition, selection := range selections {
		loops := m.loopForAlbum(selection.albumIndex)
		for loop := 0; loop < loops; loop++ {
			for trackIndex, track := range selection.tracks {
				m.scrobbleQueue = append(m.scrobbleQueue, queuedTrack{
					Artist: selection.album.Artist, Title: track.Title, Album: selection.album.Title,
					AlbumIndex: albumPosition + 1, AlbumTotal: len(selections),
					TrackIndex: trackIndex + 1, TrackTotal: len(selection.tracks),
					LoopIndex: loop + 1, LoopTotal: loops,
				})
			}
		}
	}
}

func (m model) updatePreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if len(m.scrobbleQueue) == 0 {
			m.buildScrobbleQueue()
		}
		m.previewStatus = "Preparing session…"
		m.searching = true
		m.startScrobbleSession()
		return m, tea.Batch(m.prepareScrobble(), m.spinner.Tick)
	case "e":
		return m, m.exportRecordCmd(m.queueRecord("preview"))
	case "s":
		if artist := m.currentArtist(); artist != "" {
			return m.startSimilar(artist)
		}
	case "esc":
		m.stage = stageTrackSelect
	case "q", "ctrl+c":
		m.cancelActiveSession()
		return m, tea.Quit
	}
	return m, nil
}

func (m model) prepareScrobble() tea.Cmd {
	queue := append([]queuedTrack(nil), m.scrobbleQueue...)
	ctx := m.sessionContext()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 90*time.Second)
		defer cancel()
		if err := m.client.Authenticate(ctx); err != nil {
			return scrobblePreparedMsg{err: err}
		}
		if sessionClient, ok := m.client.(interface{ SessionKey() string }); ok {
			_ = config.PersistSessionKey(m.cfg, sessionClient.SessionKey())
		}
		skipped := 0
		if m.cfg.DuplicateGuard > 0 && strings.TrimSpace(m.cfg.Username) != "" {
			recent, err := m.client.GetRecentTracks(ctx, m.cfg.Username, time.Now().Add(-m.cfg.DuplicateGuard))
			if err == nil {
				seen := map[string]bool{}
				for _, track := range recent {
					seen[recentKey(track.Artist, track.Title, track.Album)] = true
				}
				filtered := queue[:0]
				for _, item := range queue {
					if seen[recentKey(item.Artist, item.Title, item.Album)] || seen[recentKey(item.Artist, item.Title, "")] {
						skipped++
						continue
					}
					filtered = append(filtered, item)
				}
				queue = filtered
			}
		}
		if len(queue) == 0 {
			return scrobblePreparedMsg{err: fmt.Errorf("all selected tracks were skipped by duplicate protection")}
		}
		return scrobblePreparedMsg{queue: queue, skipped: skipped}
	}
}

func recentKey(artist, title, album string) string {
	return strings.ToLower(strings.TrimSpace(artist) + "\x00" + strings.TrimSpace(title) + "\x00" + strings.TrimSpace(album))
}

func (m model) authenticateThenContinue() tea.Cmd {
	ctx := m.sessionContext()
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()
		if err := m.client.Authenticate(ctx); err != nil {
			if errors.Is(ctx.Err(), context.Canceled) {
				return scrobbleCancelledMsg{}
			}
			return scrobbleResultMsg{err: err}
		}
		return m.scrobbleNext()()
	}
}

func (m model) scrobbleNext() tea.Cmd {
	ctx := m.sessionContext()
	return func() tea.Msg {
		if err := ctx.Err(); err != nil {
			return scrobbleCancelledMsg{}
		}
		if m.scrobbleIdx >= len(m.scrobbleQueue) {
			return scrobbleResultMsg{idx: m.scrobbleIdx}
		}
		item := m.scrobbleQueue[m.scrobbleIdx]
		attempts, err := scrobbler.RunOne(ctx, m.client, scrobbler.Track{Artist: item.Artist, Title: item.Title, Album: item.Album}, scrobbler.Options{Retries: m.cfg.RetryCount, RetryDelay: m.cfg.RetryDelay})
		if errors.Is(err, context.Canceled) {
			return scrobbleCancelledMsg{}
		}
		item.Attempts = attempts
		return scrobbleResultMsg{idx: m.scrobbleIdx + 1, track: item, err: err}
	}
}

func (m model) updateScrobbling(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q":
			m.cancelScrobbleSession()
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.cancelScrobbleSession()
			return m, nil
		}
	case scrobbleResultMsg:
		if errors.Is(msg.err, context.Canceled) {
			m.cancelScrobbleSession()
			return m, nil
		}
		if msg.idx == 0 && msg.err != nil {
			m.err = msg.err
			m.stage = stagePreview
			return m, nil
		}
		if msg.idx > 0 && msg.idx <= len(m.scrobbleQueue) {
			item := msg.track
			if msg.err != nil {
				item.Failed = true
				item.ErrMsg = msg.err.Error()
				m.failures = append(m.failures, item)
				m.err = msg.err
			} else {
				m.err = nil
			}
			m.scrobbleQueue[msg.idx-1] = item
			m.scrobbleIdx = msg.idx
			_ = m.store.SavePending(m.queueRecord("pending"))
		}
		if m.scrobbleIdx >= len(m.scrobbleQueue) {
			return m.finishSession()
		}
		return m, tea.Sequence(waitCmd(m.sessionContext(), m.interval), m.scrobbleNext())
	}
	return m, nil
}

func waitCmd(ctx context.Context, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		if duration <= 0 {
			return nil
		}
		timer := time.NewTimer(duration)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return scrobbleCancelledMsg{}
		case <-timer.C:
			return nil
		}
	}
}

func (m model) finishSession() (tea.Model, tea.Cmd) {
	m.cancelActiveSession()
	m.stage = stageDone
	record := m.queueRecord("complete")
	m.lastRecord = record
	_ = m.store.Append(record)
	_ = m.store.ClearPending()
	m.history, _ = m.store.LoadHistory()
	m.pending = nil
	message := fmt.Sprintf("Scrobbled %d track(s)", len(m.scrobbleQueue)-len(m.failures))
	if len(m.failures) > 0 {
		message += fmt.Sprintf("; %d failed", len(m.failures))
	}
	if m.cfg.Notify {
		return m, func() tea.Msg {
			_ = platform.Notify("Last.fm Scrobbler", message)
			return nil
		}
	}
	return m, nil
}

func (m *model) startScrobbleSession() {
	m.cancelActiveSession()
	m.sessionCtx, m.sessionCancel = context.WithCancel(context.Background())
}

func (m model) sessionContext() context.Context {
	if m.sessionCtx != nil {
		return m.sessionCtx
	}
	return context.Background()
}

func (m *model) cancelActiveSession() {
	if m.sessionCancel != nil {
		m.sessionCancel()
	}
	m.sessionCtx = nil
	m.sessionCancel = nil
}

func (m *model) cancelScrobbleSession() {
	m.cancelActiveSession()
	record := m.queueRecord("cancelled")
	_ = m.store.Append(record)
	_ = m.store.SavePending(record)
	m.pending = &record
	m.history, _ = m.store.LoadHistory()
	m.stage = stageTrackSelect
	m.err = nil
}

func (m model) openConfig() (tea.Model, tea.Cmd) {
	m.stage = stageConfig
	m.modeChoice = "config"
	m.returnStage = stageInput
	m.configIndex = 0
	m.configFieldIndex = 0
	m.configStatus = ""
	m.loadConfigField()
	return m, m.configInput.Focus()
}

func (m *model) configFieldCount() int {
	switch m.configIndex {
	case 0, 1:
		return 1
	case 2, 3:
		return 2
	default:
		return 0
	}
}

func (m model) configEditable() bool { return m.configIndex >= 0 && m.configIndex < 4 }

func configLabel(index, field int) string {
	switch index {
	case 0:
		return "LOOP"
	case 1:
		return "INTERVAL"
	case 2:
		if field == 0 {
			return "LASTFM USERNAME"
		}
		return "LASTFM PASSWORD"
	case 3:
		if field == 0 {
			return "API KEY"
		}
		return "API SECRET"
	}
	return "VALUE"
}

func (m *model) loadConfigField() {
	if !m.configEditable() {
		m.configInput.Blur()
		return
	}
	m.configInput.EchoMode = textinput.EchoNormal
	m.configInput.EchoCharacter = '•'
	label := configLabel(m.configIndex, m.configFieldIndex)
	// The view renders the cursor itself. Keeping the Bubbles input narrow
	// prevents its padded viewport from ever escaping the fixed panel width.
	m.configInput.Width = maxInt(8, 54-len([]rune(label)))
	switch m.configIndex {
	case 0:
		m.configInput.SetValue(strconv.Itoa(m.cfg.DefaultLoop))
	case 1:
		m.configInput.SetValue(strings.TrimSuffix(m.cfg.DefaultInterval.String(), "s"))
	case 2:
		if m.configFieldIndex == 0 {
			m.configInput.SetValue(m.cfg.Username)
		} else {
			m.configInput.SetValue(m.cfg.Password)
			m.configInput.EchoMode = textinput.EchoPassword
		}
	case 3:
		if m.configFieldIndex == 0 {
			m.configInput.SetValue(m.cfg.APIKey)
		} else {
			m.configInput.SetValue(m.cfg.APISecret)
			m.configInput.EchoMode = textinput.EchoPassword
		}
	}
	m.configInput.CursorEnd()
}

func (m *model) commitConfigField() {
	if !m.configEditable() {
		return
	}
	value := strings.TrimSpace(m.configInput.Value())
	switch m.configIndex {
	case 0:
		if number, err := strconv.Atoi(value); err == nil && number > 0 {
			m.cfg.DefaultLoop = number
			m.loopCount = number
		}
	case 1:
		if duration, err := time.ParseDuration(value); err == nil {
			m.cfg.DefaultInterval = duration
			m.interval = duration
		} else if seconds, err := strconv.ParseFloat(value, 64); err == nil {
			m.cfg.DefaultInterval = time.Duration(seconds * float64(time.Second))
			m.interval = m.cfg.DefaultInterval
		}
	case 2:
		if m.configFieldIndex == 0 {
			m.cfg.MarkCredentialEdited("username")
			m.cfg.Username = value
		} else {
			m.cfg.MarkCredentialEdited("password")
			m.cfg.Password = value
		}
	case 3:
		if m.configFieldIndex == 0 {
			m.cfg.MarkCredentialEdited("api_key")
			m.cfg.APIKey = value
		} else {
			m.cfg.MarkCredentialEdited("api_secret")
			m.cfg.APISecret = value
		}
	}
}

func configHorizontal(index, delta int) int {
	rowStart, rowLength := 0, 4
	if index >= 4 {
		rowStart, rowLength = 4, 3
	}
	column := index - rowStart
	column = (column + delta + rowLength) % rowLength
	return rowStart + column
}

func configVertical(index int) int {
	if index < 4 {
		return 4 + minInt(index, 2)
	}
	return index - 4
}

func (m model) selectConfigTab(index int) (tea.Model, tea.Cmd) {
	m.commitConfigField()
	m.configIndex = index
	m.configFieldIndex = 0
	m.configStatus = ""
	m.loadConfigField()
	if m.configEditable() {
		return m, m.configInput.Focus()
	}
	m.configInput.Blur()
	return m, nil
}

func (m model) openConfigUtility() (tea.Model, tea.Cmd) {
	switch m.configIndex {
	case 4:
		return m.openAdvancedConfig()
	case 5:
		m.stage = stageHistory
		m.modeChoice = "history"
		m.returnStage = stageConfig
		m.historyCursor = 0
		m.historyStatus = ""
		return m, nil
	case 6:
		m.stage = stageProfiles
		m.modeChoice = "profiles"
		m.returnStage = stageConfig
		m.profileCursor = indexOf(m.profiles, m.cfg.Profile)
		return m, nil
	}
	return m, nil
}

func (m model) updateConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left":
		return m.selectConfigTab(configHorizontal(m.configIndex, -1))
	case "right":
		return m.selectConfigTab(configHorizontal(m.configIndex, 1))
	case "up", "down":
		return m.selectConfigTab(configVertical(m.configIndex))
	case "tab", "shift+tab":
		if m.configFieldCount() > 1 {
			m.commitConfigField()
			delta := 1
			if msg.String() == "shift+tab" {
				delta = -1
			}
			m.configFieldIndex = (m.configFieldIndex + delta + m.configFieldCount()) % m.configFieldCount()
			m.loadConfigField()
			return m, m.configInput.Focus()
		}
	case "enter":
		if m.configEditable() {
			m.commitConfigField()
			return m.saveConfig()
		}
		return m.openConfigUtility()
	case "ctrl+p":
		m.commitConfigField()
		return m.openEnvPath()
	case "ctrl+g":
		m.commitConfigField()
		return m.openAdvancedConfig()
	case "ctrl+o":
		m.commitConfigField()
		m.stage = stageInfo
		m.modeChoice = "info"
		m.returnStage = stageConfig
		m.infoIndex = 0
		m.configInput.Blur()
		return m, nil
	case "esc":
		m.commitConfigField()
		m.stage = stageInput
		m.modeChoice = ""
		m.returnStage = stageInput
		m.configInput.Blur()
		return m, nil
	case "ctrl+c":
		return m, tea.Quit
	}
	if !m.configEditable() {
		return m, nil
	}
	var cmd tea.Cmd
	m.configInput, cmd = m.configInput.Update(msg)
	return m, cmd
}

func (m model) saveConfig() (tea.Model, tea.Cmd) {
	if err := config.Save(m.cfg); err != nil {
		m.err = err
		m.configStatus = ""
	} else {
		_ = config.RememberEnvPath(m.cfg.EnvPath)
		m.err = nil
		m.configStatus = "Saved to " + m.cfg.EnvPath
		m.client = lastfm.New(m.cfg.APIKey, m.cfg.APISecret, m.cfg.Username, m.cfg.Password, m.cfg.SessionKey)
	}
	return m, nil
}

func (m model) openAdvancedConfig() (tea.Model, tea.Cmd) {
	m.stage = stageAdvancedConfig
	m.modeChoice = "advanced"
	m.advancedIndex = 0
	m.advancedEditing = false
	m.configStatus = ""
	m.loadAdvancedField()
	return m, m.configInput.Focus()
}

var advancedLabels = []string{
	"RETRY COUNT", "RETRY DELAY", "DUPLICATE GUARD", "NOTIFICATIONS",
	"COMPACT HEADER", "CLEAN DISCOGRAPHY", "EXPORT DIR", "CREDENTIAL SOURCE",
	"MOUSE SUPPORT", "UPDATE URL", "CONNECTION TEST", "DIAGNOSTICS BUNDLE", "CHECK FOR UPDATES",
}

func advancedEditable(index int) bool { return index >= 0 && index < 10 }
func advancedAction(index int) bool   { return index >= 10 && index < len(advancedLabels) }

func (m *model) loadAdvancedField() {
	if !advancedEditable(m.advancedIndex) {
		m.configInput.SetValue("")
		m.configInput.Blur()
		return
	}
	m.configInput.EchoMode = textinput.EchoNormal
	m.configInput.Width = maxInt(12, 57-len([]rune(advancedLabels[m.advancedIndex])))
	switch m.advancedIndex {
	case 0:
		m.configInput.SetValue(strconv.Itoa(m.cfg.RetryCount))
	case 1:
		m.configInput.SetValue(m.cfg.RetryDelay.String())
	case 2:
		m.configInput.SetValue(m.cfg.DuplicateGuard.String())
	case 3:
		m.configInput.SetValue(boolWord(m.cfg.Notify))
	case 4:
		m.configInput.SetValue(boolWord(m.cfg.CompactHeader))
	case 5:
		m.configInput.SetValue(boolWord(m.cfg.CleanDiscography))
	case 6:
		m.configInput.SetValue(m.cfg.ExportDir)
	case 7:
		m.configInput.SetValue(m.cfg.CredentialSource)
	case 8:
		m.configInput.SetValue(boolWord(m.cfg.MouseEnabled))
	case 9:
		m.configInput.SetValue(m.cfg.UpdateURL)
	}
	m.configInput.CursorEnd()
}

func boolWord(value bool) string {
	if value {
		return "on"
	}
	return "off"
}

func (m *model) commitAdvancedField() {
	if !advancedEditable(m.advancedIndex) {
		return
	}
	value := strings.TrimSpace(m.configInput.Value())
	switch m.advancedIndex {
	case 0:
		if number, err := strconv.Atoi(value); err == nil && number >= 0 {
			m.cfg.RetryCount = number
		}
	case 1:
		if duration, err := time.ParseDuration(value); err == nil && duration >= 0 {
			m.cfg.RetryDelay = duration
		}
	case 2:
		if value == "0" || value == "off" {
			m.cfg.DuplicateGuard = 0
		} else if duration, err := time.ParseDuration(value); err == nil && duration >= 0 {
			m.cfg.DuplicateGuard = duration
		}
	case 3:
		m.cfg.Notify = parseToggle(value, m.cfg.Notify)
	case 4:
		m.cfg.CompactHeader = parseToggle(value, m.cfg.CompactHeader)
	case 5:
		m.cfg.CleanDiscography = parseToggle(value, m.cfg.CleanDiscography)
		m.discographyClean = m.cfg.CleanDiscography
	case 6:
		if value != "" {
			m.cfg.ExportDir = config.ExpandPath(value)
		}
	case 7:
		source := strings.ToLower(value)
		if source == "auto" || source == "file" || source == "environment" || source == "keychain" {
			m.cfg.CredentialSource = source
		}
	case 8:
		m.cfg.MouseEnabled = parseToggle(value, m.cfg.MouseEnabled)
	case 9:
		m.cfg.UpdateURL = value
	}
}

func parseToggle(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "on", "true", "yes", "1":
		return true
	case "off", "false", "no", "0":
		return false
	default:
		return fallback
	}
}

func (m model) updateAdvancedConfig(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		m.commitAdvancedField()
		m.advancedIndex = (m.advancedIndex + len(advancedLabels) - 1) % len(advancedLabels)
		m.loadAdvancedField()
		if advancedEditable(m.advancedIndex) {
			return m, m.configInput.Focus()
		}
		return m, nil
	case "down", "j":
		m.commitAdvancedField()
		m.advancedIndex = (m.advancedIndex + 1) % len(advancedLabels)
		m.loadAdvancedField()
		if advancedEditable(m.advancedIndex) {
			return m, m.configInput.Focus()
		}
		return m, nil
	case "left", "right", " ":
		if !advancedEditable(m.advancedIndex) {
			return m, nil
		}
		switch m.advancedIndex {
		case 3:
			m.cfg.Notify = !m.cfg.Notify
		case 4:
			m.cfg.CompactHeader = !m.cfg.CompactHeader
		case 5:
			m.cfg.CleanDiscography = !m.cfg.CleanDiscography
			m.discographyClean = m.cfg.CleanDiscography
		case 7:
			sources := []string{"auto", "file", "environment", "keychain"}
			index := indexOf(sources, m.cfg.CredentialSource)
			if msg.String() == "left" {
				index = (index + len(sources) - 1) % len(sources)
			} else {
				index = (index + 1) % len(sources)
			}
			m.cfg.CredentialSource = sources[index]
		case 8:
			m.cfg.MouseEnabled = !m.cfg.MouseEnabled
		}
		m.loadAdvancedField()
		return m, m.configInput.Focus()
	case "o":
		if m.advancedIndex == 6 {
			return m, pickFolderCmd("Choose an export folder", "export")
		}
	case "enter":
		if advancedAction(m.advancedIndex) {
			return m.openAdvancedAction()
		}
		m.commitAdvancedField()
		return m.saveConfig()
	case "esc":
		m.commitAdvancedField()
		m.stage = stageConfig
		m.modeChoice = "config"
		m.loadConfigField()
		return m, m.configInput.Focus()
	case "ctrl+c":
		return m, tea.Quit
	}
	if !advancedEditable(m.advancedIndex) {
		return m, nil
	}
	var cmd tea.Cmd
	m.configInput, cmd = m.configInput.Update(msg)
	return m, cmd
}

func (m model) openAdvancedAction() (tea.Model, tea.Cmd) {
	switch m.advancedIndex {
	case 10:
		return m.openConnectionTest()
	case 11:
		return m.openDiagnostics()
	case 12:
		return m.openUpdateCheck()
	}
	return m, nil
}

func (m model) openConnectionTest() (tea.Model, tea.Cmd) {
	m.stage = stageConnectionTest
	m.modeChoice = "connection"
	m.returnStage = stageAdvancedConfig
	m.connectionReport = connection.Report{}
	m.connectionTesting = true
	m.err = nil
	return m, tea.Batch(m.connectionTestCmd(), m.spinner.Tick)
}

func (m model) connectionTestCmd() tea.Cmd {
	cfg := m.cfg
	client := lastfm.New(cfg.APIKey, cfg.APISecret, cfg.Username, cfg.Password, cfg.SessionKey)
	return func() tea.Msg {
		return connectionTestMsg{report: connection.Test(context.Background(), cfg, client)}
	}
}

func (m model) updateConnectionTest(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r", "enter":
		m.connectionTesting = true
		m.connectionReport = connection.Report{}
		m.err = nil
		return m, tea.Batch(m.connectionTestCmd(), m.spinner.Tick)
	case "esc":
		m.stage = stageAdvancedConfig
		m.modeChoice = "advanced"
		m.advancedIndex = 10
		m.loadAdvancedField()
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) openDiagnostics() (tea.Model, tea.Cmd) {
	m.stage = stageDiagnostics
	m.modeChoice = "diagnostics"
	m.returnStage = stageAdvancedConfig
	m.diagnosticsPath = ""
	m.diagnosticsBusy = false
	m.err = nil
	return m, nil
}

func (m model) diagnosticsCmd() tea.Cmd {
	cfg := m.cfg
	history := append([]sessionstore.Record(nil), m.history...)
	return func() tea.Msg {
		path, err := diagnostics.Create(cfg, history, logging.Path(), version.Version, version.Commit)
		return diagnosticsResultMsg{path: path, err: err}
	}
}

func (m model) updateDiagnostics(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.diagnosticsBusy = true
		m.diagnosticsPath = ""
		m.err = nil
		return m, tea.Batch(m.diagnosticsCmd(), m.spinner.Tick)
	case "o":
		if m.diagnosticsPath != "" {
			return m, openFolderCmd(filepath.Dir(m.diagnosticsPath))
		}
	case "esc":
		m.stage = stageAdvancedConfig
		m.modeChoice = "advanced"
		m.advancedIndex = 11
		m.loadAdvancedField()
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) openUpdateCheck() (tea.Model, tea.Cmd) {
	m.stage = stageUpdateCheck
	m.modeChoice = "update"
	m.returnStage = stageAdvancedConfig
	m.updateResult = updater.Result{}
	m.updateChecking = true
	m.err = nil
	return m, tea.Batch(m.updateCheckCmd(), m.spinner.Tick)
}

func (m model) updateCheckCmd() tea.Cmd {
	url := m.cfg.UpdateURL
	return func() tea.Msg {
		result, err := (updater.Checker{}).Check(context.Background(), version.Version, url, version.Repository)
		return updateCheckMsg{result: result, err: err}
	}
}

func (m model) updateUpdateCheck(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r", "enter":
		m.updateChecking = true
		m.updateResult = updater.Result{}
		m.err = nil
		return m, tea.Batch(m.updateCheckCmd(), m.spinner.Tick)
	case "o":
		if m.updateResult.URL != "" {
			return m, openURLCmd(m.updateResult.URL)
		}
	case "esc":
		m.stage = stageAdvancedConfig
		m.modeChoice = "advanced"
		m.advancedIndex = 12
		m.loadAdvancedField()
		return m, nil
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) openEnvPath() (tea.Model, tea.Cmd) {
	m.stage = stageEnvPath
	m.modeChoice = "env"
	m.envStatus = ""
	m.envInput.SetValue(m.cfg.EnvPath)
	m.envInput.CursorEnd()
	return m, m.envInput.Focus()
}

func (m model) updateEnvPath(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		path := config.ExpandPath(m.envInput.Value())
		if path == "" {
			m.err = fmt.Errorf("credentials path is empty")
			return m, nil
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			loaded, loadErr := config.LoadFromPath(path)
			if loadErr != nil {
				m.err = loadErr
				return m, nil
			}
			m.cfg = loaded
		} else {
			m.cfg.EnvPath = path
			if saveErr := config.Save(m.cfg); saveErr != nil {
				m.err = saveErr
				return m, nil
			}
		}
		if err := config.RememberEnvPath(path); err != nil {
			m.err = err
			return m, nil
		}
		m.cfg.EnvPath = path
		m.applyConfig()
		m.err = nil
		m.envStatus = "Using " + path
		return m, nil
	case "o":
		return m, pickPathCmd("Choose a credentials file", "env")
	case "esc":
		m.stage = stageConfig
		m.modeChoice = "config"
		m.envInput.Blur()
		m.loadConfigField()
		return m, m.configInput.Focus()
	case "ctrl+c":
		return m, tea.Quit
	}
	var cmd tea.Cmd
	m.envInput, cmd = m.envInput.Update(msg)
	return m, cmd
}

func (m *model) applyConfig() {
	m.loopCount = maxInt(1, m.cfg.DefaultLoop)
	m.interval = m.cfg.DefaultInterval
	m.trackLimit = m.cfg.DefaultLimit
	m.discographyClean = m.cfg.CleanDiscography
	m.client = lastfm.New(m.cfg.APIKey, m.cfg.APISecret, m.cfg.Username, m.cfg.Password, m.cfg.SessionKey)
}

func (m model) updateDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		mode := m.modeIndex
		m.clearSessionSelection(false)
		return m.activateMode(mode)
	case "r":
		record := m.lastRecord
		if len(record.Queue) == 0 {
			record = m.queueRecord("complete")
		}
		if len(record.Queue) == 0 {
			return m, nil
		}
		m.restoreRecordForEdit(record)
		m.stage = stageTrackSelect
	case "R":
		if len(m.scrobbleQueue) == 0 {
			m.scrobbleQueue = queueFromRecord(m.lastRecord)
		}
		if len(m.scrobbleQueue) == 0 {
			return m, nil
		}
		m.resetQueueForRun()
		m.stage = stagePreview
	case "e":
		record := m.lastRecord
		if len(record.Queue) == 0 {
			record = m.queueRecord("complete")
		}
		return m, m.exportRecordCmd(record)
	case "s":
		if artist := m.currentArtist(); artist != "" {
			return m.startSimilar(artist)
		}
	case "h":
		m.stage = stageHistory
		m.modeChoice = "history"
	case "esc":
		m.clearSessionSelection(true)
		m.stage = stageInput
		m.modeChoice = ""
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) clearSessionSelection(clearDiscography bool) {
	m.selectedAlbums = nil
	m.selectedAlbum = lastfm.Album{}
	m.trackSelected = map[int]bool{}
	m.albumLoops = map[int]int{}
	m.trackCursor = 0
	m.scrobbleQueue = nil
	m.scrobbleIdx = 0
	m.failures = nil
	m.skippedDuplicates = 0
	m.previewStatus = ""
	m.exportStatus = ""
	m.err = nil
	if clearDiscography {
		m.discography = nil
		m.discographySelected = map[int]bool{}
		m.discographyCursor = 0
		m.discographyArtist = ""
		m.discographyFilter = ""
	}
}

func (m model) updateHistory(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := len(m.history)
	switch msg.String() {
	case "up", "k":
		if n > 0 {
			m.historyCursor = (m.historyCursor + n - 1) % n
		}
	case "down", "j":
		if n > 0 {
			m.historyCursor = (m.historyCursor + 1) % n
		}
	case "enter", "r":
		if n > 0 {
			record := m.history[m.historyCursor]
			m.restoreRecordForEdit(record)
			m.stage = stageTrackSelect
		}
	case "R":
		if n > 0 {
			record := m.history[m.historyCursor]
			m.restoreRecord(record)
			m.resetQueueForRun()
			m.stage = stagePreview
			m.modeChoice = record.Mode
			m.modeIndex = modeIndex(record.Mode)
		}
	case "e":
		if n > 0 {
			return m, m.exportRecordCmd(m.history[m.historyCursor])
		}
	case "d":
		if n > 0 {
			_ = m.store.Delete(m.history[m.historyCursor].ID)
			m.history, _ = m.store.LoadHistory()
			if m.historyCursor >= len(m.history) {
				m.historyCursor = maxInt(0, len(m.history)-1)
			}
			m.historyStatus = "History entry deleted"
		}
	case "esc":
		if m.returnStage == stageConfig {
			m.stage = stageConfig
			m.modeChoice = "config"
			m.configIndex = 5
			m.returnStage = stageInput
		} else {
			m.stage = stageInput
			m.modeChoice = ""
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) restoreRecord(record sessionstore.Record) {
	m.scrobbleQueue = queueFromRecord(record)
	m.scrobbleIdx = 0
	m.failures = nil
	m.loopCount = maxInt(1, record.Loop)
	m.interval = record.Interval
	m.scrobbleStarted = time.Time{}
	m.lastRecord = record
}

func (m model) updateRecovery(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.pending == nil {
		m.stage = stageInput
		m.modeChoice = ""
		return m, nil
	}
	switch msg.String() {
	case "enter":
		m.restoreRecord(*m.pending)
		m.startScrobbleSession()
		m.scrobbleIdx = minInt(m.pending.Completed, len(m.scrobbleQueue))
		m.scrobbleStarted = m.pending.StartedAt
		m.stage = stageScrobbling
		return m, m.authenticateThenContinue()
	case "r":
		m.restoreRecord(*m.pending)
		m.startScrobbleSession()
		m.resetQueueForRun()
		m.scrobbleStarted = time.Now()
		m.stage = stageScrobbling
		_ = m.store.SavePending(m.queueRecord("pending"))
		return m, m.authenticateThenContinue()
	case "d", "esc":
		_ = m.store.ClearPending()
		m.pending = nil
		m.stage = stageInput
		m.modeChoice = ""
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) startSimilar(artist string) (tea.Model, tea.Cmd) {
	m.returnStage = m.stage
	m.similarArtist = artist
	m.searching = true
	m.err = nil
	return m, tea.Batch(func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		albums, err := m.client.GetSimilarAlbums(ctx, artist, 15)
		return similarResultMsg{albums: albums, err: err}
	}, m.spinner.Tick)
}

func (m model) updateSimilarSelect(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := len(m.similar)
	switch msg.String() {
	case "up", "k":
		if n > 0 {
			m.similarCursor = (m.similarCursor + n - 1) % n
		}
	case "down", "j":
		if n > 0 {
			m.similarCursor = (m.similarCursor + 1) % n
		}
	case "enter":
		if n > 0 {
			m.modeChoice = "manual"
			m.modeIndex = 0
			m.searching = true
			return m, tea.Batch(m.loadAlbums([]lastfm.Album{m.similar[m.similarCursor]}), m.spinner.Tick)
		}
	case "esc":
		m.stage = m.returnStage
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateProfiles(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := len(m.profiles)
	switch msg.String() {
	case "up", "k":
		if n > 0 {
			m.profileCursor = (m.profileCursor + n - 1) % n
		}
	case "down", "j":
		if n > 0 {
			m.profileCursor = (m.profileCursor + 1) % n
		}
	case "enter":
		if n > 0 {
			name := m.profiles[m.profileCursor]
			if name == "default" {
				_ = config.RememberProfile(name)
				loaded, err := config.Load()
				if err != nil {
					m.err = err
					return m, nil
				}
				m.cfg = loaded
			} else {
				loaded, err := config.LoadProfile(name)
				if err != nil {
					m.err = err
					return m, nil
				}
				m.cfg = loaded
			}
			_ = config.RememberProfile(name)
			m.applyConfig()
			m.profileStatus = "Loaded profile " + name
		}
	case "n":
		m.stage = stageProfileName
		m.modeChoice = "profile"
		m.profileInput.SetValue("")
		return m, m.profileInput.Focus()
	case "s":
		if n > 0 {
			name := m.profiles[m.profileCursor]
			if err := config.SaveProfile(name, m.cfg); err != nil {
				m.err = err
			} else {
				m.profileStatus = "Saved profile " + name
			}
		}
	case "d":
		if n > 0 {
			name := m.profiles[m.profileCursor]
			if err := config.DeleteProfile(name); err != nil {
				m.err = err
			} else {
				m.profiles, _ = config.ListProfiles()
				m.profileCursor = minInt(m.profileCursor, maxInt(0, len(m.profiles)-1))
				m.profileStatus = "Deleted profile " + name
			}
		}
	case "esc":
		if m.returnStage == stageConfig {
			m.stage = stageConfig
			m.modeChoice = "config"
			m.configIndex = 6
			m.returnStage = stageInput
		} else {
			m.stage = stageInput
			m.modeChoice = ""
		}
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) updateProfileName(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		name := strings.TrimSpace(m.profileInput.Value())
		if name == "" {
			return m, nil
		}
		if err := config.SaveProfile(name, m.cfg); err != nil {
			m.err = err
			return m, nil
		}
		m.profiles, _ = config.ListProfiles()
		m.profileCursor = indexOf(m.profiles, name)
		m.profileStatus = "Created profile " + name
		m.stage = stageProfiles
		m.modeChoice = "profiles"
	case "esc":
		m.stage = stageProfiles
		m.modeChoice = "profiles"
	case "ctrl+c":
		return m, tea.Quit
	default:
		var cmd tea.Cmd
		m.profileInput, cmd = m.profileInput.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m model) exportRecordCmd(record sessionstore.Record) tea.Cmd {
	dir := config.ExpandPath(m.cfg.ExportDir)
	return func() tea.Msg {
		paths, err := exporter.Export(record, dir)
		return exportResultMsg{paths: paths, err: err}
	}
}

func (m model) updateInfo(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "up", "k":
		m.infoIndex = (m.infoIndex + 4) % 5
	case "right", "down", "j":
		m.infoIndex = (m.infoIndex + 1) % 5
	case "esc":
		if m.returnStage == stageConfig {
			m.stage = stageConfig
			m.modeChoice = "config"
		} else {
			m.stage = stageInput
			m.modeChoice = ""
		}
		m.returnStage = stageInput
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func pickPathCmd(prompt, target string) tea.Cmd {
	return func() tea.Msg {
		path, err := platform.PickFileOrFolder(prompt)
		return filePickedMsg{path: path, target: target, err: err}
	}
}

func pickFolderCmd(prompt, target string) tea.Cmd {
	return func() tea.Msg {
		path, err := platform.PickFolder(prompt)
		return filePickedMsg{path: path, target: target, err: err}
	}
}

func openFolderCmd(path string) tea.Cmd {
	return func() tea.Msg {
		if err := platform.OpenFolder(path); err != nil {
			return diagnosticsResultMsg{err: err}
		}
		return nil
	}
}

func openURLCmd(value string) tea.Cmd {
	return func() tea.Msg {
		if err := platform.OpenURL(value); err != nil {
			return updateCheckMsg{err: err}
		}
		return nil
	}
}

func openHeaderURLCmd(value string) tea.Cmd {
	return func() tea.Msg {
		return headerURLMsg{err: platform.OpenURL(value)}
	}
}

func (m model) headerURLContains(x, y int) bool {
	if m.compactHeaderEnabled() {
		return false
	}
	left, top, width := headerURLBounds(m.cfg.Username)
	return y == top && x >= left && x < left+width
}

func (m model) updateMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if msg.Action == tea.MouseActionMotion {
		m.headerURLHover = m.headerURLContains(msg.X, msg.Y)
		return m, nil
	}
	if msg.Button == tea.MouseButtonWheelUp {
		return m.mouseMove(-1)
	}
	if msg.Button == tea.MouseButtonWheelDown {
		return m.mouseMove(1)
	}
	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}
	x, y := msg.X, msg.Y
	if m.headerURLContains(x, y) {
		return m, openHeaderURLCmd(lastfmURL(m.cfg.Username))
	}
	bodyY := y - m.headerHeight()
	switch m.stage {
	case stageInput:
		if bodyY >= 0 && bodyY <= 2 {
			switch {
			case x >= 1 && x <= 19:
				return m.activateMode(0)
			case x >= 21 && x <= 45:
				return m.activateMode(1)
			case x >= 47 && x <= 65:
				return m.activateMode(2)
			}
		}
	case stageImportSource:
		switch {
		case bodyY >= 0 && bodyY <= 2 && x >= 13 && x <= 32:
			m.importSourceIndex = 0
		case bodyY >= 0 && bodyY <= 2 && x >= 34 && x <= 52:
			m.importSourceIndex = 1
		case bodyY >= 3 && bodyY <= 5 && x >= 2 && x <= 32:
			m.importSourceIndex = 2
		case bodyY >= 3 && bodyY <= 5 && x >= 34 && x <= 64:
			m.importSourceIndex = 3
		default:
			return m, nil
		}
	case stageConfig:
		index := -1
		if bodyY >= 0 && bodyY <= 2 {
			switch {
			case x >= 3 && x <= 13:
				index = 0
			case x >= 15 && x <= 33:
				index = 1
			case x >= 35 && x <= 53:
				index = 2
			case x >= 55 && x <= 63:
				index = 3
			}
		} else if bodyY >= 3 && bodyY <= 5 {
			switch {
			case x >= 5 && x <= 23:
				index = 4
			case x >= 25 && x <= 41:
				index = 5
			case x >= 43 && x <= 61:
				index = 6
			}
		}
		if index >= 0 {
			return m.selectConfigTab(index)
		}
	case stageAdvancedConfig:
		row := bodyY - 1
		maxRows := 11
		if m.height > 0 {
			maxRows = maxInt(7, minInt(13, m.height-19))
		}
		start := visibleStart(m.advancedIndex, len(advancedLabels), maxRows)
		index := start + row
		if row >= 0 && index >= 0 && index < len(advancedLabels) {
			m.commitAdvancedField()
			m.advancedIndex = index
			m.loadAdvancedField()
			if advancedEditable(index) {
				return m, m.configInput.Focus()
			}
			return m, nil
		}
	}
	return m, nil
}

func (m model) mouseMove(delta int) (tea.Model, tea.Cmd) {
	switch m.stage {
	case stageInput:
		m.modeIndex = (m.modeIndex + delta + 3) % 3
	case stageImportSource:
		m.importSourceIndex = (m.importSourceIndex + delta + 4) % 4
	case stageResults:
		if len(m.results) > 0 {
			m.resultsCursor = (m.resultsCursor + delta + len(m.results)) % len(m.results)
		}
	case stageDiscographySelect:
		visible := m.discographyVisibleIndexes()
		if len(visible) > 0 {
			m.discographyCursor = (m.discographyCursor + delta + len(visible)) % len(visible)
		}
	case stageTrackSelect:
		if n := len(m.flattenedTracks()); n > 0 {
			m.trackCursor = (m.trackCursor + delta + n) % n
		}
	case stageConfig:
		return m.selectConfigTab(configHorizontal(m.configIndex, delta))
	case stageAdvancedConfig:
		m.commitAdvancedField()
		m.advancedIndex = (m.advancedIndex + delta + len(advancedLabels)) % len(advancedLabels)
		m.loadAdvancedField()
		if advancedEditable(m.advancedIndex) {
			return m, m.configInput.Focus()
		}
	case stageHistory:
		if len(m.history) > 0 {
			m.historyCursor = (m.historyCursor + delta + len(m.history)) % len(m.history)
		}
	case stageSimilarSelect:
		if len(m.similar) > 0 {
			m.similarCursor = (m.similarCursor + delta + len(m.similar)) % len(m.similar)
		}
	case stageProfiles:
		if len(m.profiles) > 0 {
			m.profileCursor = (m.profileCursor + delta + len(m.profiles)) % len(m.profiles)
		}
	}
	return m, nil
}

func (m model) currentArtist() string {
	if len(m.selectedAlbums) > 0 {
		return m.selectedAlbums[0].Artist
	}
	if m.selectedAlbum.Artist != "" {
		return m.selectedAlbum.Artist
	}
	if len(m.scrobbleQueue) > 0 {
		return m.scrobbleQueue[0].Artist
	}
	if m.resultsCursor >= 0 && m.resultsCursor < len(m.results) {
		return m.results[m.resultsCursor].Artist
	}
	return m.discographyArtist
}

func indexOf(values []string, target string) int {
	for i, value := range values {
		if strings.EqualFold(value, target) {
			return i
		}
	}
	return 0
}
