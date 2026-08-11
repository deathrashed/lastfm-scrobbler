package tui

import (
	"context"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/connection"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
	setupstate "github.com/deathrashed/lastfm-scrobbler/internal/setup"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
	"github.com/deathrashed/lastfm-scrobbler/internal/updater"
)

type stage int

const (
	stageInput stage = iota
	stageImportSource
	stageSearch
	stageResults
	stageDiscographySelect
	stageTrackSelect
	stagePreview
	stageConfig
	stageEnvPath
	stageScrobbling
	stageDone
	stageHistory
	stageRecovery
	stageSimilarSelect
	stageProfiles
	stageProfileName
	stageInfo
	stageConnectionTest
	stageDiagnostics
	stageUpdateCheck
	stageSetup
)

type model struct {
	cfg            config.Config
	client         lastfm.Client
	store          sessionstore.Store
	stage          stage
	returnStage    stage
	width, height  int
	helpVisible    bool
	headerURLHover bool
	hoverRegion    string
	sessionCtx     context.Context
	sessionCancel  context.CancelFunc
	sessionID      uint64

	modeChoice        string
	modeIndex         int
	importSourceIndex int
	infoIndex         int

	connectionReport  connection.Report
	connectionTesting bool
	diagnosticsPath   string
	diagnosticsBusy   bool
	updateResult      updater.Result
	updateChecking    bool

	searchInput  textinput.Model
	filterInput  textinput.Model
	profileInput textinput.Model
	spinner      spinner.Model
	searching    bool

	results       []lastfm.Album
	resultsCursor int

	discography          []lastfm.Album
	discographyArtist    string
	discographyCursor    int // position in the visible index slice
	discographySelected  map[int]bool
	discographyFilter    string
	discographyFiltering bool
	discographyClean     bool
	discographySort      int

	similar       []lastfm.Album
	similarCursor int
	similarArtist string

	selectedAlbums []lastfm.Album
	selectedAlbum  lastfm.Album

	trackLimit    int
	loopCount     int
	interval      time.Duration
	trackCursor   int
	trackSelected map[int]bool
	albumLoops    map[int]int

	settingsSection settingsSection
	settingsFocus   settingsFocus
	settingsRow     int

	configInput  textinput.Model
	configStatus string
	envInput     textinput.Model
	envStatus    string

	profiles      []string
	profileCursor int
	profileStatus string

	scrobbleQueue     []queuedTrack
	scrobbleIdx       int
	failures          []queuedTrack
	albumsScrobbled   int
	skippedDuplicates int
	scrobbleStarted   time.Time
	previewStatus     string
	exportStatus      string
	lastRecord        sessionstore.Record
	pending           *sessionstore.Record
	history           []sessionstore.Record
	historyCursor     int
	historyStatus     string
	err               error
	setup             setupstate.State
	setupInputs       [4]textinput.Model
	setupOriginal     config.Config
}

type queuedTrack struct {
	Artist, Title, Album string
	AlbumIndex           int
	AlbumTotal           int
	TrackIndex           int
	TrackTotal           int
	LoopIndex            int
	LoopTotal            int
	Attempts             int
	Failed               bool
	ErrMsg               string
}

func New(cfg config.Config, client lastfm.Client) tea.Model {
	store := sessionstore.New(config.DataDir())
	history, _ := store.LoadHistory()
	pending, _ := store.LoadPending()
	profiles, _ := config.ListProfiles()

	m := model{
		cfg: cfg, client: client, store: store, stage: stageInput,
		modeIndex: 0, trackLimit: cfg.DefaultLimit, loopCount: cfg.DefaultLoop,
		interval: cfg.DefaultInterval, discographySelected: map[int]bool{},
		trackSelected: map[int]bool{}, albumLoops: map[int]int{},
		discographyClean: cfg.CleanDiscography,
		settingsSection:  settingsScrobbling,
		settingsFocus:    settingsFocusContent,
		history:          history, pending: pending, profiles: profiles,
	}
	if m.loopCount < 1 {
		m.loopCount = 1
	}
	if len(m.profiles) == 0 {
		m.profiles = []string{"default"}
	}

	m.searchInput = newTextInput(512, 48)
	m.filterInput = newTextInput(128, 44)
	m.profileInput = newTextInput(64, 40)
	m.configInput = newTextInput(1024, 44)
	m.envInput = newTextInput(1024, 48)
	m.spinner = spinner.New()
	m.spinner.Spinner = lastFMSpinner()

	if pending != nil && len(pending.Queue) > 0 {
		m.stage = stageRecovery
		m.modeChoice = pending.Mode
		m.scrobbleQueue = queueFromRecord(*pending)
		m.scrobbleIdx = minInt(pending.Completed, len(m.scrobbleQueue))
		m.loopCount = maxInt(1, pending.Loop)
		m.interval = pending.Interval
	}
	return m
}

func NewSetup(cfg config.Config, client lastfm.Client) tea.Model {
	m := New(cfg, client).(model)
	m.stage = stageSetup
	m.modeChoice = "setup"
	m.setup = setupstate.NewState(cfg)
	m.setupOriginal = cfg
	m.cfg.MouseEnabled = true
	for index := range m.setupInputs {
		m.setupInputs[index] = newTextInput(1024, 48)
	}
	m.setupInputs[1].EchoMode = textinput.EchoPassword
	m.setupInputs[1].EchoCharacter = '•'
	m.setupInputs[3].EchoMode = textinput.EchoPassword
	m.setupInputs[3].EchoCharacter = '•'
	m.syncSetupInputs()
	return m
}

func (m *model) syncSetupInputs() {
	values := []string{m.setup.Account.Username, m.setup.Account.Password, m.setup.Account.APIKey, m.setup.Account.APISecret}
	for index, value := range values {
		m.setupInputs[index].SetValue(value)
	}
}

func newTextInput(limit, width int) textinput.Model {
	input := textinput.New()
	input.Prompt = ""
	input.CharLimit = limit
	input.Width = width
	input.TextStyle = theme.PrimaryTextStyle
	input.PlaceholderStyle = theme.SecondaryTextStyle
	input.Cursor.Style = theme.AccentTextStyle
	return input
}

func (m model) Init() tea.Cmd                           { return nil }
func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) { return updateModel(m, msg) }
