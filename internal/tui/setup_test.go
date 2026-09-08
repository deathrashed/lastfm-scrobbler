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
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

func setupTestModel(t *testing.T) model {
	t.Helper()
	cfg := config.Config{EnvPath: filepath.Join(t.TempDir(), ".env"), DefaultLoop: 1, DefaultInterval: 2 * time.Second, MouseEnabled: true}
	return NewSetup(cfg, nil).(model)
}

func TestSetupPanelsUseExactWidth(t *testing.T) {
	m := setupTestModel(t)
	pages := []setup.Page{setup.PageWelcome, setup.PageSystem, setup.PageAccount, setup.PageStorage, setup.PageScrobbling, setup.PageInterface, setup.PageReview, setup.PageComplete}
	for _, page := range pages {
		m.setup.Page = page
		lines := strings.Split(renderSetupBody(m), "\n")
		want := lipgloss.Width(stripANSI(lines[0]))
		for lineNumber, line := range lines {
			if got := lipgloss.Width(stripANSI(line)); got != want {
				t.Fatalf("page %d line %d width=%d want=%d", page, lineNumber+1, got, want)
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
	for _, page := range []setup.Page{setup.PageSystem, setup.PageReview, setup.PageComplete} {
		m.setup.Page = page
		var titleLine string
		for _, line := range strings.Split(stripANSI(renderSetupBody(m)), "\n") {
			if strings.Contains(line, "┤") && strings.Contains(line, "├") {
				titleLine = line
				break
			}
		}
		if titleLine == "" {
			t.Fatalf("page %d has no attached title line", page)
		}
		startByte := strings.Index(titleLine, "┤")
		endByte := strings.LastIndex(titleLine, "├")
		left := lipgloss.Width(titleLine[:startByte]) - 1 // exclude outer ╭
		rightSuffix := strings.TrimSuffix(titleLine[endByte+len("├"):], "╮")
		right := lipgloss.Width(rightSuffix)
		if left-right > 1 || right-left > 1 {
			t.Fatalf("page %d title segments differ: %d/%d", page, left, right)
		}
	}
}

func TestSetupProgressFitsAndCenters(t *testing.T) {
	for _, page := range []setup.Page{setup.PageSystem, setup.PageReview, setup.PageComplete} {
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

func TestSetupPanelsStartWithAttachedTitleOnly(t *testing.T) {
	m := setupTestModel(t)
	m.stage = stageSetup
	m.setup.Page = setup.PageInterface
	lines := strings.Split(stripANSI(renderSetupBody(m)), "\n")
	if len(lines) < 1 {
		t.Fatal("setup body missing panel")
	}
	if got := lipgloss.Width(lines[0]); got != setupNarrowPanelWidth(m, 37) {
		t.Fatalf("panel width=%d want=%d", got, setupNarrowPanelWidth(m, 37))
	}
	plain := strings.TrimSpace(lines[0])
	if !strings.HasPrefix(plain, "╭") || !strings.Contains(plain, "┤ I N T E R F A C E ├") || !strings.HasSuffix(plain, "╮") {
		t.Fatalf("malformed setup title line: %q", plain)
	}
}

func TestSetupNavigationSkipsAutomaticFontInstallPage(t *testing.T) {
	m := setupTestModel(t)
	m.stage = stageSetup
	m.setup.Page = setup.PageSystem
	updated, _ := m.setupContinue()
	got := updated.(model)
	if got.setup.Page != setup.PageAccount {
		t.Fatalf("system continue page=%d want account=%d", got.setup.Page, setup.PageAccount)
	}
	if got.setup.InstallFont || got.setup.TerminalDefault {
		t.Fatal("first-run setup unexpectedly enables font download/terminal font mutation")
	}
}

func TestSetupAccountPreviousSkipsLegacyFontPage(t *testing.T) {
	m := setupTestModel(t)
	m.stage = stageSetup
	m.setup.Page = setup.PageAccount
	updated, _ := m.setupPrevious()
	got := updated.(model)
	if got.setup.Page != setup.PageSystem {
		t.Fatalf("account previous page=%d want system=%d", got.setup.Page, setup.PageSystem)
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

func TestSetupWelcomeActionBoxIsCenteredAsAWhole(t *testing.T) {
	m := setupTestModel(t)
	m.setup.Page = setup.PageWelcome
	lines := strings.Split(stripANSI(renderSetupBody(m)), "\n")
	buttonLine := -1
	for index, line := range lines {
		if strings.Contains(line, "GET STARTED") {
			buttonLine = index
			break
		}
	}
	if buttonLine <= 0 || buttonLine+1 >= len(lines) {
		t.Fatalf("welcome action box not found: %q", strings.Join(lines, "\\n"))
	}
	for _, index := range []int{buttonLine - 1, buttonLine, buttonLine + 1} {
		line := lines[index]
		left := strings.IndexAny(line, "╭│╰")
		right := strings.LastIndexAny(line, "╮│╯")
		if left < 0 || right <= left {
			t.Fatalf("welcome button row %d malformed: %q", index, line)
		}
		// The inner button should have matching start/end columns on all rows.
		if index == buttonLine-1 {
			continue
		}
	}
	// Compare the actual inner-box left column on all three rows. The outer
	// setup panel border is column 0, so use the second structural glyph.
	positions := make([]int, 0, 3)
	for _, index := range []int{buttonLine - 1, buttonLine, buttonLine + 1} {
		line := lines[index]
		plain := []rune(line)
		found := -1
		for col := 1; col < len(plain); col++ {
			if plain[col] == '╭' || plain[col] == '│' || plain[col] == '╰' {
				found = col
				break
			}
		}
		positions = append(positions, found)
	}
	if positions[0] != positions[1] || positions[1] != positions[2] {
		t.Fatalf("welcome button rows drift horizontally: %v", positions)
	}
}

func TestSetupReviewContinueAppliesWithoutShowingApplyPage(t *testing.T) {
	m := setupTestModel(t)
	m.stage = stageSetup
	m.setup.Page = setup.PageReview
	updated, cmd := m.setupContinue()
	got := updated.(model)
	if got.setup.Page != setup.PageReview {
		t.Fatalf("review continue page=%d want review=%d", got.setup.Page, setup.PageReview)
	}
	if !got.setup.Applying {
		t.Fatal("review continue did not enter applying state")
	}
	if cmd == nil {
		t.Fatal("review continue did not return apply command")
	}
}

func TestSetupApplySuccessAutoRoutesForward(t *testing.T) {
	m := setupTestModel(t)
	m.stage = stageSetup
	m.setup.Page = setup.PageReview
	m.cfg.SessionKey = ""
	cfg := m.cfg
	updated, cmd := m.updateSetup(setupApplyMsg{result: setup.ApplyResult{Configuration: "saved"}, config: cfg})
	got := updated.(model)
	if got.stage != stageSetup || got.setup.Page != setup.PageComplete {
		t.Fatalf("apply success stage/page=%d/%d want setup/complete", got.stage, got.setup.Page)
	}
	if cmd != nil {
		t.Fatalf("apply success unexpectedly returned command: %v", cmd)
	}
}

func TestSetupCompleteRoutesToAuthWhenSessionMissing(t *testing.T) {
	m := setupTestModel(t)
	m.setup.Page = setup.PageComplete
	m.stage = stageSetup
	m.cfg.SessionKey = ""
	updated, cmd := m.setupContinue()
	got := updated.(model)
	if got.stage != stageAuth {
		t.Fatalf("complete setup stage=%d want auth", got.stage)
	}
	if got.authReturn != stageInput {
		t.Fatalf("auth return=%d want dashboard/input", got.authReturn)
	}
	if cmd == nil {
		t.Fatal("complete setup did not request a fresh auth token")
	}
}

func TestSetupCompleteOpensDashboardWhenSessionExists(t *testing.T) {
	m := setupTestModel(t)
	m.setup.Page = setup.PageComplete
	m.stage = stageSetup
	m.cfg.SessionKey = "session"
	updated, cmd := m.setupContinue()
	got := updated.(model)
	if got.stage != stageInput {
		t.Fatalf("complete setup stage=%d want dashboard/input", got.stage)
	}
	if cmd != nil {
		t.Fatalf("complete setup with session unexpectedly returned command: %v", cmd)
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
