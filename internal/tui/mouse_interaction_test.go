package tui

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

func makeAlbums(count int) []lastfm.Album {
	albums := make([]lastfm.Album, count)
	for index := range albums {
		albums[index] = lastfm.Album{Artist: "Artist", Title: "Album " + strconv.Itoa(index)}
	}
	return albums
}

func settingsTestModel(t *testing.T, section settingsSection, compact bool) model {
	t.Helper()
	m := model{
		width: 120,
		cfg: config.Config{
			CompactHeader:    compact,
			MouseEnabled:     true,
			DefaultLoop:      1,
			DefaultInterval:  2 * time.Second,
			RetryCount:       2,
			CredentialSource: "auto",
			ExportDir:        "/tmp",
		},
		configInput:  newTextInput(1024, 44),
		envInput:     newTextInput(1024, 48),
		profileInput: newTextInput(64, 40),
		profiles:     []string{"default", "archive"},
	}
	updated, _ := m.openSettingsSection(section, settingsFocusContent)
	return updated.(model)
}

func clickRegion(t *testing.T, m model, id string) (model, tea.Cmd) {
	t.Helper()
	for _, region := range m.screenRegions() {
		if region.id == id {
			updated, cmd := m.updateMouse(tea.MouseMsg{X: region.x, Y: region.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
			return updated.(model), cmd
		}
	}
	t.Fatalf("region %q not found", id)
	return m, nil
}

func TestInfoTabsAreClickableInBothHeaderModes(t *testing.T) {
	for _, compact := range []bool{false, true} {
		m := model{width: 120, stage: stageInfo, cfg: config.Config{CompactHeader: compact, MouseEnabled: true}}
		for index := 0; index < 5; index++ {
			m.infoIndex = 0
			updated, _ := clickRegion(t, m, "info:tab:"+strconv.Itoa(index))
			if updated.infoIndex != index {
				t.Fatalf("compact=%t tab=%d selected %d", compact, index, updated.infoIndex)
			}
		}
	}
}

func TestInfoWheelMovesSections(t *testing.T) {
	m := model{stage: stageInfo, infoIndex: 0}
	updated, _ := m.mouseMove(1)
	if updated.(model).infoIndex != 1 {
		t.Fatal("wheel down did not move to next Info section")
	}
	updated, _ = updated.(model).mouseMove(-1)
	if updated.(model).infoIndex != 0 {
		t.Fatal("wheel up did not move to previous Info section")
	}
}

func TestDashboardSettingsShortcutAndFooterOpenScrobbling(t *testing.T) {
	m := model{stage: stageInput, cfg: config.Config{MouseEnabled: true}, configInput: newTextInput(1024, 44)}
	updated, _ := m.updateInput(keyMessage("s"))
	got := updated.(model)
	if got.stage != stageConfig || got.currentSettingsSection() != settingsScrobbling || got.settingsFocus != settingsFocusContent {
		t.Fatalf("s settings opened %s", got.settingsDebugString())
	}

	m = model{width: 120, stage: stageInput, cfg: config.Config{MouseEnabled: true}, configInput: newTextInput(1024, 44)}
	got, _ = clickRegion(t, m, "footer:s")
	if got.currentSettingsSection() != settingsScrobbling || got.modeChoice != "scrobbling" {
		t.Fatalf("footer settings opened %s", got.settingsDebugString())
	}
}

func TestDashboardHistoryAndProfilesOpenInsideSettingsShell(t *testing.T) {
	for _, test := range []struct {
		key     string
		stage   stage
		section settingsSection
	}{
		{"h", stageHistory, settingsHistory},
		{"p", stageProfiles, settingsProfiles},
	} {
		m := model{stage: stageInput, configInput: newTextInput(1024, 44), profiles: []string{"default"}}
		updated, _ := m.updateInput(keyMessage(test.key))
		got := updated.(model)
		if got.stage != test.stage || got.currentSettingsSection() != test.section || got.settingsFocus != settingsFocusContent {
			t.Fatalf("%s opened stage=%d section=%d focus=%d", test.key, got.stage, got.currentSettingsSection(), got.settingsFocus)
		}
	}
}

func TestDashboardNoLongerUsesCForSettings(t *testing.T) {
	m := model{stage: stageInput, modeIndex: 0}
	updated, _ := m.updateInput(keyMessage("c"))
	got := updated.(model)
	if got.stage != stageInput || got.modeChoice != "" {
		t.Fatalf("legacy c shortcut changed Dashboard state: stage=%d mode=%q", got.stage, got.modeChoice)
	}
}

func TestSettingsGridIsClickableInBothHeaderModes(t *testing.T) {
	for _, compact := range []bool{false, true} {
		for section := settingsAccount; section <= settingsProfiles; section++ {
			m := settingsTestModel(t, settingsScrobbling, compact)
			updated, _ := clickRegion(t, m, "settings:section:"+strconv.Itoa(int(section)))
			if updated.currentSettingsSection() != section {
				t.Fatalf("compact=%t section=%d selected=%d", compact, section, updated.currentSettingsSection())
			}
			if updated.settingsFocus != settingsFocusSections {
				t.Fatalf("compact=%t section click did not focus grid", compact)
			}
		}
	}
}

func TestSettingsKeyboardFocusMovesBetweenGridAndContent(t *testing.T) {
	m := settingsTestModel(t, settingsScrobbling, false)
	updated, _ := m.updateSettings(keyMessage("tab"))
	got := updated.(model)
	if got.settingsFocus != settingsFocusSections {
		t.Fatal("Tab did not focus Settings sections")
	}

	updated, _ = got.updateSettingsGrid(keyMessage("right"))
	got = updated.(model)
	if got.currentSettingsSection() != settingsHistory || got.stage != stageHistory {
		t.Fatalf("right from Scrobbling selected %s", got.settingsDebugString())
	}

	updated, _ = got.updateHistory(keyMessage("enter"))
	got = updated.(model)
	if got.settingsFocus != settingsFocusContent {
		t.Fatal("Enter on section grid did not focus section content")
	}

	m = settingsTestModel(t, settingsScrobbling, false)
	m.settingsRow = 0
	updated, _ = m.updateSettings(keyMessage("up"))
	if updated.(model).settingsFocus != settingsFocusSections {
		t.Fatal("Up from first Settings row did not return focus to section grid")
	}
}

func TestSettingsOverviewColorHierarchy(t *testing.T) {
	m := settingsTestModel(t, settingsScrobbling, false)
	row := settingsRows(settingsScrobbling)[1]

	m.settingsFocus = settingsFocusSections
	idle := renderSettingsOverviewRow(m, row, 1)
	if !strings.Contains(idle, theme.RowLabelStyle.Render("INTERVAL ")) ||
		!strings.Contains(idle, theme.RowArrowStyle.Render("❯")) ||
		!strings.Contains(idle, theme.RowValueStyle.Render("2s")) {
		t.Fatalf("idle row does not use white label/red arrow/muted value: %q", idle)
	}

	m.settingsFocus = settingsFocusContent
	m.settingsRow = 1
	focused := renderSettingsOverviewRow(m, row, 1)
	if !strings.Contains(focused, theme.FocusedRowLabelStyle.Render("INTERVAL ")) ||
		!strings.Contains(focused, theme.FocusedRowArrowStyle.Render("❯")) ||
		!strings.Contains(focused, theme.FocusedRowValueStyle.Render("2s")) {
		t.Fatalf("focused row does not use red label/white arrow/value: %q", focused)
	}

	m.settingsFocus = settingsFocusSections
	m.hoverRegion = "settings:row:1"
	hovered := renderSettingsOverviewRow(m, row, 1)
	if !strings.Contains(hovered, theme.HoverRowLabelStyle.Render("INTERVAL ")) ||
		!strings.Contains(hovered, theme.HoverRowArrowStyle.Render("❯")) ||
		!strings.Contains(hovered, theme.HoverRowValueStyle.Render("2s")) {
		t.Fatalf("hovered row does not use red label/white arrow/value: %q", hovered)
	}
}

func TestFileSourceCardsHoverWithoutMovingKeyboardSelection(t *testing.T) {
	m := model{width: 120, stage: stageImportSource, modeChoice: "file", importSourceIndex: 0, cfg: config.Config{MouseEnabled: true}}
	var target mouseRegion
	for _, region := range m.screenRegions() {
		if region.id == "import:1" {
			target = region
			break
		}
	}
	if target.width == 0 {
		t.Fatal("playlist hover region missing")
	}
	updated, _ := m.updateMouse(tea.MouseMsg{X: target.x, Y: target.y, Action: tea.MouseActionMotion})
	got := updated.(model)
	if got.importSourceIndex != 0 {
		t.Fatalf("hover changed keyboard source from 0 to %d", got.importSourceIndex)
	}
	if got.hoverRegion != "import:1" {
		t.Fatalf("hover region = %q, want import:1", got.hoverRegion)
	}
	rendered := renderImportSourceView(got)
	if !strings.Contains(rendered, theme.AccentTextStyle.Render("P L A Y L I S T")) {
		t.Fatal("hovered File source is not Torch Red")
	}
	if !strings.Contains(rendered, theme.SelectedModeStyle.Render("L I S T   F I L E")) {
		t.Fatal("selected File source is not bold white while another card is hovered")
	}
	if !strings.Contains(rendered, theme.InnerBorderStyle.Render("╭")) {
		t.Fatal("selected File source lost its Torch Red border")
	}
}

func TestHistoryHoverDoesNotMoveKeyboardCursor(t *testing.T) {
	history := []sessionstore.Record{
		{Status: "complete", Queue: []sessionstore.Track{{Artist: "First", Album: "One"}}},
		{Status: "complete", Queue: []sessionstore.Track{{Artist: "Second", Album: "Two"}}},
	}
	m := model{width: 120, stage: stageHistory, modeChoice: "history", settingsSection: settingsHistory, settingsFocus: settingsFocusContent, history: history, historyCursor: 0, cfg: config.Config{MouseEnabled: true}}
	var target mouseRegion
	for _, region := range m.screenRegions() {
		if region.id == "history:1" {
			target = region
			break
		}
	}
	if target.width == 0 {
		t.Fatal("history hover region missing")
	}
	updated, _ := m.updateMouse(tea.MouseMsg{X: target.x, Y: target.y, Action: tea.MouseActionMotion})
	got := updated.(model)
	if got.historyCursor != 0 {
		t.Fatalf("hover moved History cursor to %d", got.historyCursor)
	}
	rendered := renderHistoryView(got)
	if !strings.Contains(rendered, theme.FocusedRowLabelStyle.Render("First — One")) {
		t.Fatal("keyboard-focused History entry is not red")
	}
	if !strings.Contains(rendered, theme.HoverRowLabelStyle.Render("Second — Two")) {
		t.Fatal("hovered History entry is not red")
	}
}

func TestDiscographyHoverDoesNotMoveKeyboardCursor(t *testing.T) {
	m := model{
		width:               120,
		stage:               stageDiscographySelect,
		modeChoice:          "discography",
		discography:         makeAlbums(3),
		discographySelected: map[int]bool{},
		discographyCursor:   0,
		cfg:                 config.Config{MouseEnabled: true},
	}
	var target mouseRegion
	for _, region := range m.screenRegions() {
		if region.id == "discography:1" {
			target = region
			break
		}
	}
	if target.width == 0 {
		t.Fatal("discography hover region missing")
	}
	updated, _ := m.updateMouse(tea.MouseMsg{X: target.x, Y: target.y, Action: tea.MouseActionMotion})
	got := updated.(model)
	if got.discographyCursor != 0 {
		t.Fatalf("hover moved Discography cursor to %d", got.discographyCursor)
	}
	rendered := renderDiscographyList(got, got.discographyVisibleIndexes())
	if !strings.Contains(rendered, theme.FocusedRowLabelStyle.Render("Album 0")) {
		t.Fatal("focused Discography album is not red")
	}
	if !strings.Contains(rendered, theme.HoverRowLabelStyle.Render("Album 1")) {
		t.Fatal("hovered Discography album is not red")
	}
}

func TestTallDiscographyRowsKeepMouseRegionsInSync(t *testing.T) {
	m := model{
		width:               120,
		height:              40,
		stage:               stageDiscographySelect,
		modeChoice:          "discography",
		discography:         makeAlbums(20),
		discographySelected: map[int]bool{},
		discographyCursor:   0,
		cfg:                 config.Config{MouseEnabled: true},
	}
	if got := discographyMaxRows(m); got != 16 {
		t.Fatalf("discography max rows = %d, want 16", got)
	}
	found := false
	for _, region := range m.screenRegions() {
		if region.id == "discography:15" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("16th visible Discography row has no mouse region")
	}
}

func TestSettingsRowsAndDetailAreMouseTargets(t *testing.T) {
	m := settingsTestModel(t, settingsInterface, false)
	updated, _ := clickRegion(t, m, "settings:row:2")
	if updated.settingsRow != 2 || updated.settingsFocus != settingsFocusContent {
		t.Fatalf("row click selected row=%d focus=%d", updated.settingsRow, updated.settingsFocus)
	}
	beforeMouse := updated.cfg.MouseEnabled
	updated, _ = clickRegion(t, updated, "settings:detail")
	if updated.cfg.MouseEnabled == beforeMouse {
		t.Fatal("toggle detail click did not toggle Mouse Support")
	}

	m = settingsTestModel(t, settingsTools, false)
	m.settingsRow = 2
	m.loadSettingsField()
	updated, _ = clickRegion(t, m, "settings:detail")
	if updated.stage != stageConnectionTest || !updated.connectionTesting {
		t.Fatal("Connection Test detail did not invoke action")
	}

	m = settingsTestModel(t, settingsAccount, false)
	m.settingsRow = 5
	m.loadSettingsField()
	updated, _ = clickRegion(t, m, "settings:detail")
	if updated.stage != stageEnvPath {
		t.Fatal("Credential Path detail did not open path editor")
	}
}

func TestSettingsCommitPreservesUnchangedEnvironmentCredential(t *testing.T) {
	path := filepath.Join(t.TempDir(), "credentials.env")
	if err := os.WriteFile(path, []byte("LASTFM_CREDENTIAL_SOURCE=auto\nLASTFM_USERNAME=file-user\nLASTFM_PASSWORD=file-pass\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		"API_KEY", "API_SECRET", "LASTFM_API_KEY", "LASTFM_API_SECRET", "LASTFM_SHARED_SECRET",
		"LASTFM_USER", "LASTFM_PASSWORD", "LASTFM_SESSION_KEY", "SESSION_KEY",
		"LASTFM_CREDENTIAL_SOURCE", "SCROBBLE_CREDENTIAL_SOURCE", "LASTFM_PROFILE", "SCROBBLE_PROFILE",
	} {
		t.Setenv(key, "")
	}
	t.Setenv("LASTFM_USERNAME", "environment-user")

	cfg, err := config.LoadFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	m := model{
		stage:           stageConfig,
		settingsSection: settingsAccount,
		settingsFocus:   settingsFocusContent,
		configInput:     newTextInput(1024, 44),
		cfg:             cfg,
	}
	m.loadSettingsField()
	m.selectSettingsRow(1)
	if err := config.Save(m.cfg); err != nil {
		t.Fatal(err)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(contents), "LASTFM_USERNAME=file-user") {
		t.Fatal("moving through Settings persisted an unchanged environment credential")
	}
}

func TestHelpKeyboardCloseStillWorks(t *testing.T) {
	for _, key := range []string{"?", "esc", "enter"} {
		m := model{stage: stageInput, helpVisible: true}
		updated, _ := updateModel(m, keyMessage(key))
		if updated.(model).helpVisible {
			t.Fatalf("Help remained open after %s", key)
		}
	}
}

func TestHelpHasMouseCloseInBothHeaderModes(t *testing.T) {
	for _, compact := range []bool{false, true} {
		m := model{width: 120, stage: stageInput, helpVisible: true, cfg: config.Config{CompactHeader: compact, MouseEnabled: true}}
		var region mouseRegion
		for _, candidate := range m.screenRegions() {
			if candidate.id == "help:close" {
				region = candidate
				break
			}
		}
		if region.width == 0 {
			t.Fatalf("compact=%t help close region missing", compact)
		}
		updated, _ := m.updateMouse(tea.MouseMsg{X: region.x, Y: region.y, Action: tea.MouseActionMotion})
		if updated.(model).hoverRegion != "help:close" || !strings.Contains(renderHelpView(updated.(model)), theme.SuccessStyle.Render("close")) {
			t.Fatalf("compact=%t help close hover not rendered", compact)
		}
		updated, _ = updated.(model).updateMouse(tea.MouseMsg{X: region.x, Y: region.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
		if updated.(model).helpVisible {
			t.Fatalf("compact=%t help mouse close did not close", compact)
		}
	}
}

func TestActionCardsUseTheirVisibleKeyboardActions(t *testing.T) {
	m := model{width: 120, stage: stageDiagnostics, cfg: config.Config{MouseEnabled: true}}
	updated, _ := clickRegion(t, m, "diagnostics:action")
	if !updated.diagnosticsBusy {
		t.Fatal("diagnostics action card did not invoke export")
	}

	m = model{width: 120, stage: stageConnectionTest, cfg: config.Config{MouseEnabled: true}}
	for _, region := range m.screenRegions() {
		if region.id == "connection:action" && region.message.String() != "r" {
			t.Fatalf("connection region synthesized %q, want r", region.message.String())
		}
	}
	updated, _ = clickRegion(t, m, "connection:action")
	if !updated.connectionTesting {
		t.Fatal("connection card did not invoke re-test")
	}

	m = model{width: 120, stage: stageUpdateCheck, cfg: config.Config{MouseEnabled: true}}
	updated, _ = clickRegion(t, m, "update:action")
	if !updated.updateChecking {
		t.Fatal("update card did not invoke check again")
	}
}

func TestActionCardHoverRegionsUseTorchRedBorders(t *testing.T) {
	for _, test := range []struct {
		stage stage
		id    string
	}{
		{stageDiagnostics, "diagnostics:action"},
		{stageConnectionTest, "connection:action"},
		{stageUpdateCheck, "update:action"},
	} {
		m := model{width: 120, stage: test.stage, cfg: config.Config{MouseEnabled: true}}
		var region mouseRegion
		for _, candidate := range m.screenRegions() {
			if candidate.id == test.id {
				region = candidate
				break
			}
		}
		if region.width == 0 {
			t.Fatalf("%s region not found", test.id)
		}
		updated, _ := m.updateMouse(tea.MouseMsg{X: region.x, Y: region.y, Action: tea.MouseActionMotion})
		if updated.(model).hoverRegion != test.id {
			t.Fatalf("hover region = %q, want %q", updated.(model).hoverRegion, test.id)
		}
		if !strings.Contains(renderActionCard(updated.(model), test.stage), theme.InnerBorderStyle.Render("╭")) {
			t.Fatalf("%s hover did not render Torch Red", test.id)
		}
	}
}

func renderActionCard(m model, current stage) string {
	switch current {
	case stageDiagnostics:
		return renderDiagnosticsView(m)
	case stageConnectionTest:
		return renderConnectionTestView(m)
	case stageUpdateCheck:
		return renderUpdateCheckView(m)
	}
	return ""
}

func TestEveryRenderedFooterActionHasARegion(t *testing.T) {
	stages := []stage{stageInput, stageImportSource, stageSearch, stageResults, stageDiscographySelect, stageTrackSelect, stagePreview, stageConfig, stageEnvPath, stageScrobbling, stageDone, stageHistory, stageRecovery, stageSimilarSelect, stageProfiles, stageProfileName, stageInfo, stageConnectionTest, stageDiagnostics, stageUpdateCheck, stageCompletions}
	for _, current := range stages {
		m := model{
			width:           120,
			stage:           current,
			settingsSection: settingsScrobbling,
			settingsFocus:   settingsFocusContent,
			cfg:             config.Config{MouseEnabled: true, DefaultLoop: 1, DefaultInterval: 2 * time.Second},
			configInput:     newTextInput(64, 30),
			envInput:        newTextInput(64, 30),
			profileInput:    newTextInput(64, 30),
		}
		regions := map[string]bool{}
		for _, region := range m.footerRegions() {
			regions[region.id] = true
		}
		for _, items := range footerSpec(m) {
			for _, item := range items {
				if item.interactive && !regions[item.id] {
					t.Fatalf("stage %v action %q has no mouse region", current, item.id)
				}
			}
		}
	}
}

func TestScrollableListRegionsTrackRenderedRowsAcrossHeaders(t *testing.T) {
	for _, compact := range []bool{false, true} {
		m := model{width: 120, stage: stageResults, cfg: config.Config{CompactHeader: compact}, results: makeAlbums(30), resultsCursor: 15}
		regions := m.screenRegions()
		rows := 0
		for _, region := range regions {
			if strings.HasPrefix(region.id, "results:") {
				rows++
			}
		}
		if rows != 13 {
			t.Fatalf("compact=%t visible result rows=%d, want 13", compact, rows)
		}
		first := visibleStart(m.resultsCursor, len(m.results), 13)
		for _, want := range []int{first, first + 6, first + 12} {
			id := "results:" + strconv.Itoa(want)
			found := false
			for _, region := range regions {
				if region.id == id {
					found = true
					if region.y != m.headerHeight()+4+(want-first) {
						t.Fatalf("compact=%t %s y=%d", compact, id, region.y)
					}
				}
			}
			if !found {
				t.Fatalf("compact=%t missing %s", compact, id)
			}
		}
	}
}

func TestHistoryAndProfilesRegionsAccountForSettingsGrid(t *testing.T) {
	for _, compact := range []bool{false, true} {
		m := model{
			width:           120,
			cfg:             config.Config{CompactHeader: compact, MouseEnabled: true},
			history:         make([]sessionstore.Record, 30),
			historyCursor:   15,
			profiles:        make([]string, 30),
			profileCursor:   15,
			settingsFocus:   settingsFocusContent,
			settingsSection: settingsHistory,
			stage:           stageHistory,
		}
		for _, prefix := range []string{"history", "profiles"} {
			m.stage = stageHistory
			m.settingsSection = settingsHistory
			if prefix == "profiles" {
				m.stage = stageProfiles
				m.settingsSection = settingsProfiles
			}
			regions := m.screenRegions()
			start := visibleStart(15, 30, 13)
			for _, index := range []int{start, start + 6, start + 12} {
				id := prefix + ":" + strconv.Itoa(index)
				found := false
				for _, region := range regions {
					if region.id == id {
						found = true
						wantY := m.headerHeight() + settingsSectionContentStartY() + 1 + (index - start)
						if region.y != wantY {
							t.Fatalf("compact=%t %s y=%d want=%d", compact, id, region.y, wantY)
						}
					}
				}
				if !found {
					t.Fatalf("compact=%t missing %s", compact, id)
				}
			}
		}
	}
}

func TestMouseDisabledDoesNotChangeHoverOrCloseHelp(t *testing.T) {
	m := model{width: 120, stage: stageInput, helpVisible: true, cfg: config.Config{MouseEnabled: false}, hoverRegion: "existing"}
	updated, cmd := updateModel(m, tea.MouseMsg{X: 10, Y: 5, Action: tea.MouseActionMotion})
	if cmd != nil || updated.(model).hoverRegion != "existing" || !updated.(model).helpVisible {
		t.Fatal("mouse-disabled model changed on motion")
	}
}

func TestURLColorsMatchInteractionContract(t *testing.T) {
	if theme.HeaderURLStyle.GetForeground() != theme.TorchRed || theme.HeaderURLHoverStyle.GetForeground() != theme.White {
		t.Fatalf("URL colors are idle=%v hover=%v", theme.HeaderURLStyle.GetForeground(), theme.HeaderURLHoverStyle.GetForeground())
	}
	if lipgloss.Width(stripANSI(RenderHeaderWithHover(120, stageInput, "", "user", "", false, false))) == 0 {
		t.Fatal("URL header did not render")
	}
}

func TestSettingsMouseDisabledIgnoresSectionClicks(t *testing.T) {
	m := settingsTestModel(t, settingsScrobbling, false)
	m.cfg.MouseEnabled = false
	region := m.settingsGridRegion(settingsAccount)
	updated, cmd := updateModel(m, tea.MouseMsg{X: region.x, Y: region.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	got := updated.(model)
	if cmd != nil || got.currentSettingsSection() != settingsScrobbling {
		t.Fatalf("mouse-disabled Settings click changed section=%d cmd=%v", got.currentSettingsSection(), cmd)
	}
}

func TestMultiSelectListsUseCircleSelectorsAndChevronCursor(t *testing.T) {
	m := model{
		width:               120,
		stage:               stageDiscographySelect,
		modeChoice:          "discography",
		discography:         makeAlbums(3),
		discographySelected: map[int]bool{1: true},
		discographyCursor:   0,
	}
	plain := stripANSI(renderDiscographyList(m, m.discographyVisibleIndexes()))
	if strings.ContainsAny(plain, "☐☑▸") {
		t.Fatalf("discography still contains legacy selector/cursor glyphs: %q", plain)
	}
	if !strings.Contains(plain, "❯ ○ Album 0") {
		t.Fatalf("focused unselected album does not use ❯ ○: %q", plain)
	}
	if !strings.Contains(plain, "● Album 1") {
		t.Fatalf("selected album does not use ●: %q", plain)
	}
}

func TestTrackFooterMouseControlsAdjustIntervalNavigationAndLoop(t *testing.T) {
	album := lastfm.Album{
		Artist: "Slayer",
		Title:  "Hell Awaits",
		Tracks: []lastfm.Track{{Title: "Hell Awaits"}, {Title: "Kill Again"}},
	}
	m := model{
		width:          120,
		stage:          stageTrackSelect,
		modeChoice:     "manual",
		cfg:            config.Config{MouseEnabled: true},
		selectedAlbum:  album,
		selectedAlbums: []lastfm.Album{album},
		trackSelected:  map[int]bool{0: true, 1: true},
		albumLoops:     map[int]int{0: 1},
		loopCount:      1,
		interval:       2 * time.Second,
	}

	got, _ := clickRegion(t, m, "footer:interval-up")
	if got.interval != 3*time.Second {
		t.Fatalf("interval after footer + = %s, want 3s", got.interval)
	}
	got, _ = clickRegion(t, got, "footer:interval-down")
	if got.interval != 2*time.Second {
		t.Fatalf("interval after footer - = %s, want 2s", got.interval)
	}
	got, _ = clickRegion(t, got, "footer:nav-down")
	if got.trackCursor != 1 {
		t.Fatalf("track cursor after footer ↓ = %d, want 1", got.trackCursor)
	}
	got, _ = clickRegion(t, got, "footer:nav-up")
	if got.trackCursor != 0 {
		t.Fatalf("track cursor after footer ↑ = %d, want 0", got.trackCursor)
	}
	got, _ = clickRegion(t, got, "footer:loop-up")
	if got.loopCount != 2 {
		t.Fatalf("manual loop after footer + = %d, want 2", got.loopCount)
	}
	got, _ = clickRegion(t, got, "footer:loop-down")
	if got.loopCount != 1 {
		t.Fatalf("manual loop after footer - = %d, want 1", got.loopCount)
	}
}

func TestDiscographyTrackFooterLoopAdjustsCurrentAlbum(t *testing.T) {
	albums := []lastfm.Album{
		{Artist: "Slayer", Title: "Hell Awaits", Tracks: []lastfm.Track{{Title: "Hell Awaits"}}},
		{Artist: "Slayer", Title: "South of Heaven", Tracks: []lastfm.Track{{Title: "South of Heaven"}}},
	}
	m := model{
		width:          120,
		stage:          stageTrackSelect,
		modeChoice:     "discography",
		cfg:            config.Config{MouseEnabled: true},
		selectedAlbums: albums,
		trackSelected:  map[int]bool{0: true, 1: true},
		albumLoops:     map[int]int{0: 1, 1: 1},
		loopCount:      1,
		trackCursor:    1,
		interval:       2 * time.Second,
	}
	got, _ := clickRegion(t, m, "footer:loop-up")
	if got.loopForAlbum(0) != 1 || got.loopForAlbum(1) != 2 {
		t.Fatalf("album loops after footer + = [%d %d], want [1 2]", got.loopForAlbum(0), got.loopForAlbum(1))
	}
	if got.loopCount != 1 {
		t.Fatalf("global loop changed to %d while adjusting a Discography album", got.loopCount)
	}
}

func TestDiscographyChooserControlsAreMouseTargets(t *testing.T) {
	m := model{
		width:               120,
		stage:               stageDiscographySelect,
		modeChoice:          "discography",
		discography:         makeAlbums(5),
		discographySelected: map[int]bool{},
		cfg:                 config.Config{MouseEnabled: true},
		filterInput:         newTextInput(128, 44),
	}

	got, _ := clickRegion(t, m, "discography:sort")
	if got.discographySort != 1 {
		t.Fatalf("sort control click produced sort=%d, want 1", got.discographySort)
	}
	got, _ = clickRegion(t, got, "discography:clean")
	if !got.discographyClean {
		t.Fatal("clean control click did not enable clean mode")
	}
	got, _ = clickRegion(t, got, "discography:filter")
	if !got.discographyFiltering || !got.filterInput.Focused() {
		t.Fatal("filter control click did not focus the integrated filter input")
	}
}

func TestExpandedDiscographyFilterKeepsListMouseGeometryInSync(t *testing.T) {
	albums := makeAlbums(20)
	albums[0].Title = strings.TrimSpace(strings.Repeat("Album ", 10))
	m := model{
		width:               120,
		height:              40,
		stage:               stageDiscographySelect,
		modeChoice:          "discography",
		discography:         albums,
		discographySelected: map[int]bool{},
		discographyCursor:   0,
		discographyFilter:   strings.TrimSpace(strings.Repeat("Album ", 10)),
		cfg:                 config.Config{MouseEnabled: true},
	}
	wantY := m.headerHeight() + discographyChooserListRowOffset(m)
	for _, region := range m.screenRegions() {
		if region.id == "discography:0" {
			if region.y != wantY {
				t.Fatalf("expanded filter first album y=%d, want %d", region.y, wantY)
			}
			return
		}
	}
	t.Fatal("expanded filter first album has no mouse region")
}
