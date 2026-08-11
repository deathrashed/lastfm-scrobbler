package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/setup"
)

func setupTestModel(t *testing.T) model {
	t.Helper()
	cfg := config.Config{EnvPath: filepath.Join(t.TempDir(), ".env"), DefaultLoop: 1, DefaultInterval: 2 * time.Second, MouseEnabled: true}
	return NewSetup(cfg, nil).(model)
}

func TestSetupPanelsUseExactWidth(t *testing.T) {
	m := setupTestModel(t)
	pages := []setup.Page{setup.PageWelcome, setup.PageSystem, setup.PageFont, setup.PageAccount, setup.PageStorage, setup.PageScrobbling, setup.PageInterface, setup.PageReview, setup.PageApply, setup.PageComplete}
	for _, page := range pages {
		m.setup.Page = page
		for lineNumber, line := range strings.Split(renderSetupBody(m), "\n") {
			if got := lipgloss.Width(stripANSI(line)); got != setupPanelWidth {
				t.Fatalf("page %d line %d width=%d want=%d", page, lineNumber+1, got, setupPanelWidth)
			}
		}
	}
}

func TestSetupPanelsTruncateLongUnicodeValues(t *testing.T) {
	m := setupTestModel(t)
	m.setup.Environment.Terminal = strings.Repeat("終", 100)
	m.setup.Environment.PackageManager = strings.Repeat("manager", 20)
	m.setup.Account.Username = strings.Repeat("artist", 20)
	for _, page := range []setup.Page{setup.PageSystem, setup.PageReview} {
		m.setup.Page = page
		for _, line := range strings.Split(renderSetupBody(m), "\n") {
			if lipgloss.Width(stripANSI(line)) != setupPanelWidth {
				t.Fatalf("page %d overflowed long values", page)
			}
		}
	}
}

func TestSetupAttachedTitlesStayCentered(t *testing.T) {
	m := setupTestModel(t)
	for _, page := range []setup.Page{setup.PageSystem, setup.PageFont, setup.PageReview, setup.PageComplete} {
		m.setup.Page = page
		line := stripANSI(strings.Split(renderSetupBody(m), "\n")[0])
		titleStart := strings.Index(line, "┤")
		titleEnd := strings.LastIndex(line, "├")
		left := strings.Count(line[:titleStart], "─")
		right := strings.Count(strings.TrimSuffix(line[titleEnd+len("├"):], "╮"), "─")
		if left-right > 1 || right-left > 1 {
			t.Fatalf("page %d title segments differ: %d/%d", page, left, right)
		}
	}
}

func TestSetupProgressFitsAndCenters(t *testing.T) {
	for _, page := range []setup.Page{setup.PageSystem, setup.PageFont, setup.PageReview, setup.PageApply, setup.PageComplete} {
		line := centerToHeader(setupProgress(page))
		if lipgloss.Width(stripANSI(line)) != headerContentWidth {
			t.Fatalf("page %d progress width mismatch", page)
		}
		plain := stripANSI(line)
		if !strings.Contains(plain, "●") {
			t.Fatalf("page %d has no progress dots", page)
		}
	}
}

func TestSetupWelcomeEscapesWithoutWriting(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	m := NewSetup(config.Config{EnvPath: path}, nil).(model)
	updated, _ := m.updateSetup(tea.KeyMsg{Type: tea.KeyEscape})
	got := updated.(model)
	if got.stage != stageInput {
		t.Fatalf("welcome skip did not return to dashboard stage: %d", got.stage)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("welcome skip wrote a credentials file")
	}
}

func TestSetupMouseRegionsFollowHeaderHeight(t *testing.T) {
	for _, compact := range []bool{false, true} {
		m := setupTestModel(t)
		m.cfg.CompactHeader = compact
		m.setup.Page = setup.PageInterface
		regions := m.screenRegions()
		found := false
		for _, region := range regions {
			if region.id == "setup:toggle:1" {
				found = true
				if region.y < m.headerHeight() {
					t.Fatalf("compact=%t setup region overlaps header", compact)
				}
				break
			}
		}
		if !found {
			t.Fatalf("compact=%t setup region missing", compact)
		}
	}
}

func TestSetupAccountInputIsStaged(t *testing.T) {
	m := setupTestModel(t)
	m.setup.Page = setup.PageAccount
	m.setupInputs[0].Focus()
	updated, _ := m.updateSetup(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("user")})
	got := updated.(model)
	if got.setup.Account.Username != "user" {
		t.Fatal("account input did not update staged setup state")
	}
	if _, err := os.Stat(got.setup.Pending.EnvPath); !os.IsNotExist(err) {
		t.Fatal("account input wrote persistent configuration")
	}
}
