package setup

import (
	"runtime"
	"strings"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
)

type Page int

const (
	PageWelcome Page = iota
	PageSystem
	PageFont
	PageAccount
	PageStorage
	PageScrobbling
	PageInterface
	PageReview
	PageApply
	PageComplete
)

const NumberedPageCount = 8

type State struct {
	Page             Page
	Environment      Environment
	Terminal         Terminal
	Fonts            []FontChoice
	FontIndex        int
	InstallFont      bool
	TerminalDefault  bool
	Account          Account
	CredentialSource string
	Recommended      bool
	Loop             int
	Interval         time.Duration
	MouseEnabled     bool
	CompactHeader    bool
	Notify           bool
	Pending          config.Config
	Applying         bool
	ApplyResult      ApplyResult
	Field            int
}

type Account struct {
	Username  string
	Password  string
	APIKey    string
	APISecret string
}

type ApplyResult struct {
	FontStatus     string
	TerminalStatus string
	Configuration  string
	Credentials    string
	Connection     string
	Authentication string
	Error          error
}

func NewState(cfg config.Config) State {
	fonts := append([]FontChoice(nil), curatedFonts...)
	terminal := NewTerminalConfigurator().Detect()
	return State{
		Page:             PageWelcome,
		Environment:      DetectEnvironment(),
		Terminal:         terminal,
		Fonts:            fonts,
		FontIndex:        0,
		InstallFont:      true,
		TerminalDefault:  false,
		Account:          Account{Username: cfg.Username, Password: cfg.Password, APIKey: cfg.APIKey, APISecret: cfg.APISecret},
		CredentialSource: initialCredentialSource(cfg),
		Recommended:      true,
		Loop:             max(1, cfg.DefaultLoop),
		Interval:         maxDuration(cfg.DefaultInterval, 2*time.Second),
		MouseEnabled:     cfg.MouseEnabled,
		CompactHeader:    cfg.CompactHeader,
		Notify:           cfg.Notify,
		Pending:          cfg,
	}
}

func (s State) SelectedFont() FontChoice {
	if s.FontIndex < 0 || s.FontIndex >= len(s.Fonts) {
		return FontChoice{}
	}
	return s.Fonts[s.FontIndex]
}

func (s State) HasFont() bool { return s.SelectedFont().Asset != "" }

func (s State) ApplyConfig() config.Config {
	cfg := s.Pending
	cfg.Username = strings.TrimSpace(s.Account.Username)
	cfg.Password = s.Account.Password
	cfg.APIKey = strings.TrimSpace(s.Account.APIKey)
	cfg.APISecret = strings.TrimSpace(s.Account.APISecret)
	cfg.CredentialSource = s.CredentialSource
	cfg.DefaultLoop = max(1, s.Loop)
	cfg.DefaultInterval = maxDuration(s.Interval, time.Second)
	cfg.MouseEnabled = s.MouseEnabled
	cfg.CompactHeader = s.CompactHeader
	cfg.Notify = s.Notify
	if s.Recommended {
		cfg.DefaultLoop = 1
		cfg.DefaultInterval = 2 * time.Second
	}
	return cfg
}

func initialCredentialSource(cfg config.Config) string {
	if strings.EqualFold(cfg.CredentialSource, "keychain") && runtime.GOOS != "darwin" {
		return "file"
	}
	if cfg.CredentialSource == "" || cfg.CredentialSource == "auto" {
		return "file"
	}
	return cfg.CredentialSource
}

func max(value, fallback int) int {
	if value < fallback {
		return fallback
	}
	return value
}

func maxDuration(value, fallback time.Duration) time.Duration {
	if value < fallback {
		return fallback
	}
	return value
}
