package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

func assertBlockWidth(t *testing.T, block string, want int) {
	t.Helper()
	for lineNumber, line := range strings.Split(block, "\n") {
		if got := lipgloss.Width(stripANSI(line)); got != want {
			t.Fatalf("line %d width = %d, want %d\n%q", lineNumber+1, got, want, line)
		}
	}
}

func TestHeaderAlwaysFitsExactMockupWidth(t *testing.T) {
	for _, mode := range []string{"", "manual", "discography", "file", "account", "scrobbling", "history", "tools", "interface", "profiles"} {
		header := RenderHeader(140, stageInput, mode, "deathrashed", "", false)
		lines := strings.Split(header, "\n")
		for lineNumber, line := range lines {
			got := lipgloss.Width(stripANSI(line))
			if got > 67 {
				t.Fatalf("mode %q line %d width = %d, exceeds 67\n%q", mode, lineNumber+1, got, line)
			}
		}
		for _, lineNumber := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} {
			if got := lipgloss.Width(stripANSI(lines[lineNumber])); got != 67 {
				t.Fatalf("mode %q structural line %d width = %d, want 67", mode, lineNumber+1, got)
			}
		}
	}
}

func TestCompactHeadersUseTheFourLineSpec(t *testing.T) {
	modes := []string{"", "manual", "discography", "file", "account", "scrobbling", "history", "tools", "interface", "profiles", "info", "env", "profile", "connection", "diagnostics", "update"}
	for _, mode := range modes {
		header := RenderHeader(140, stageInput, mode, "deathrashed", "", true)
		lines := strings.Split(header, "\n")
		if len(lines) != compactHeaderLines {
			t.Fatalf("mode %q line count = %d, want %d", mode, len(lines), compactHeaderLines)
		}
		for lineNumber, line := range lines {
			if got := lipgloss.Width(stripANSI(line)); got != fullHeaderWidth {
				t.Fatalf("mode %q line %d width = %d, want %d\n%q", mode, lineNumber+1, got, fullHeaderWidth, line)
			}
		}
		if strings.Contains(header, "last.fm/user/") || strings.Contains(header, "┤") || strings.Contains(header, "├") {
			t.Fatalf("mode %q contains full-header or detached-badge content\n%s", mode, header)
		}
		spec := compactHeaderSpecFor(mode)
		assertCenteredHeaderText(t, lines[1], spec.Title)
		assertCenteredHeaderText(t, lines[2], spec.Subtitle)
		assertCenteredHeaderText(t, lines[3], spec.Icon)
	}
}

func TestCompactHeaderLongestSubtitleFits(t *testing.T) {
	for mode, spec := range compactHeaderSpecs {
		if got := lipgloss.Width(spec.Subtitle); got > fullHeaderWidth-2 {
			t.Fatalf("mode %q subtitle width = %d, exceeds %d", mode, got, fullHeaderWidth-2)
		}
	}
}

func TestCompactManualHeaderAddsArtistOnlyAfterResolution(t *testing.T) {
	unresolved := model{stage: stageSearch, modeChoice: "manual", cfg: config.Config{CompactHeader: true}}
	before := RenderHeaderWithHoverArtist(140, unresolved.stage, unresolved.modeChoice, "", "", true, false, unresolved.headerArtist())
	if lines := strings.Split(before, "\n"); len(lines) != compactHeaderLines {
		t.Fatalf("unresolved compact Manual header lines = %d, want %d", len(lines), compactHeaderLines)
	}
	if strings.Contains(stripANSI(before), "ARTIST ❯") {
		t.Fatalf("unresolved compact Manual header contains an artist row:\n%s", stripANSI(before))
	}

	resolved := model{
		stage:         stageResults,
		modeChoice:    "manual",
		cfg:           config.Config{CompactHeader: true},
		results:       []lastfm.Album{{Artist: "Enforced", Title: "War Remains"}},
		resultsCursor: 0,
	}
	header := RenderHeaderWithHoverArtist(140, resolved.stage, resolved.modeChoice, "", "", true, false, resolved.headerArtist())
	lines := strings.Split(header, "\n")
	if len(lines) != compactHeaderLines+1 {
		t.Fatalf("resolved compact Manual header lines = %d, want %d", len(lines), compactHeaderLines+1)
	}
	if !strings.Contains(stripANSI(lines[3]), "ARTIST ❯ ENFORCED") {
		t.Fatalf("resolved compact Manual header is missing artist row:\n%s", stripANSI(header))
	}
}

func TestCompactDiscographyHeaderAddsResolvedArtist(t *testing.T) {
	m := model{
		stage:             stageDiscographySelect,
		modeChoice:        "discography",
		cfg:               config.Config{CompactHeader: true},
		discographyArtist: "Oxygen Destroyer",
	}
	header := RenderHeaderWithHoverArtist(140, m.stage, m.modeChoice, "", "", true, false, m.headerArtist())
	lines := strings.Split(header, "\n")
	if len(lines) != compactHeaderLines+1 {
		t.Fatalf("compact Discography header lines = %d, want %d", len(lines), compactHeaderLines+1)
	}
	if !strings.Contains(stripANSI(lines[3]), "ARTIST ❯ OXYGEN DESTROYER") {
		t.Fatalf("compact Discography header is missing artist row:\n%s", stripANSI(header))
	}
}

func TestCompactArtistHeaderUsesRequestedStyles(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(previousProfile) })

	header := RenderHeaderWithHoverArtist(140, stageResults, "manual", "", "", true, false, "Enforced")
	artistLine := strings.Split(header, "\n")[3]
	if !strings.Contains(artistLine, theme.AccentTextStyle.Render("ARTIST")) {
		t.Fatal("compact artist label is not Torch Red")
	}
	if !strings.Contains(artistLine, theme.PrimaryTextStyle.Render("❯")) {
		t.Fatal("compact artist arrow is not white")
	}
	if !strings.Contains(artistLine, theme.ArtistStyle.Render("ENFORCED")) {
		t.Fatal("compact artist value is not bold Torch Red")
	}
}

func TestCompactArtistHeaderTruncatesLongNamesWithoutChangingWidth(t *testing.T) {
	artist := "An Artist Name That Is Deliberately Much Longer Than The Compact Header Can Display"
	m := model{
		stage:      stageResults,
		modeChoice: "manual",
		cfg:        config.Config{CompactHeader: true},
		results:    []lastfm.Album{{Artist: artist, Title: "Long Name"}},
	}
	header := RenderHeaderWithHoverArtist(140, m.stage, m.modeChoice, "", "", true, false, m.headerArtist())
	lines := strings.Split(header, "\n")
	if len(lines) != compactHeaderLines+1 {
		t.Fatalf("long compact artist header lines = %d, want %d", len(lines), compactHeaderLines+1)
	}
	for lineNumber, line := range lines {
		if got := lipgloss.Width(stripANSI(line)); got != fullHeaderWidth {
			t.Fatalf("long compact artist line %d width = %d, want %d\n%q", lineNumber+1, got, fullHeaderWidth, line)
		}
	}
	plain := stripANSI(lines[3])
	if !strings.Contains(plain, "ARTIST ❯ ") || !strings.Contains(plain, "…") {
		t.Fatalf("long compact artist row does not preserve its prefix and ellipsis: %q", plain)
	}
	if strings.Contains(plain, strings.ToUpper(artist)) {
		t.Fatal("long compact artist value was not truncated")
	}
}

func TestHeaderHeightMatchesRenderedMode(t *testing.T) {
	full := model{width: 140}
	if got := full.headerHeight(); got != fullHeaderLines {
		t.Fatalf("full header height = %d, want %d", got, fullHeaderLines)
	}
	compact := model{width: 140, cfg: config.Config{CompactHeader: true}}
	if got := compact.headerHeight(); got != compactHeaderLines {
		t.Fatalf("compact header height = %d, want %d", got, compactHeaderLines)
	}
	small := model{width: 66}
	if got := small.headerHeight(); got != fullHeaderLines {
		t.Fatalf("small terminal header height = %d, want %d", got, fullHeaderLines)
	}
}

func TestViewRejectsTerminalsNarrowerThanHeader(t *testing.T) {
	for _, width := range []int{40, 66} {
		view := (model{width: width}).View()
		if !strings.Contains(stripANSI(view), "Terminal too narrow") {
			t.Fatalf("width %d did not show the narrow-terminal message: %q", width, view)
		}
	}
	for _, width := range []int{67, 120} {
		view := (model{width: width}).View()
		if strings.Contains(stripANSI(view), "Terminal too narrow") {
			t.Fatalf("width %d was rejected", width)
		}
	}
}

func assertCenteredHeaderText(t *testing.T, line, content string) {
	t.Helper()
	plain := stripANSI(line)
	inner := strings.TrimSuffix(strings.TrimPrefix(plain, "│"), "│")
	index := strings.Index(inner, content)
	if index < 0 {
		t.Fatalf("line does not contain %q: %q", content, plain)
	}
	left := lipgloss.Width(inner[:index])
	right := lipgloss.Width(inner[index+len(content):])
	if absInt(left-right) > 1 {
		t.Fatalf("content %q is not centered: left=%d right=%d line=%q", content, left, right, plain)
	}
}

func TestHeaderURLUsesOSC8AndConfiguredUsername(t *testing.T) {
	got := RenderHeaderWithHover(140, stageInput, "", "deathrashed", "", false, false)
	if !strings.Contains(got, "\x1b]8;;https://www.last.fm/user/deathrashed\x1b\\") {
		t.Fatal("header URL is not wrapped in an OSC 8 hyperlink")
	}
	if !strings.Contains(got, "last.fm/user/deathrashed") {
		t.Fatal("header URL does not use the compact visible form")
	}
	if theme.HeaderURLStyle.GetUnderline() || !theme.HeaderURLHoverStyle.GetUnderline() {
		t.Fatal("header URL hover state is not visually distinguished")
	}
	if theme.HeaderURLStyle.GetForeground() != theme.TorchRed || theme.HeaderURLHoverStyle.GetForeground() != theme.White {
		t.Fatal("header URL colors do not distinguish normal and hover states")
	}

	fallback := RenderHeaderWithHover(140, stageInput, "", "", "", false, false)
	if !strings.Contains(fallback, "https://www.last.fm") || strings.Contains(fallback, "/user/username") {
		t.Fatalf("unexpected fallback URL: %q", fallback)
	}
}

func TestHeaderURLMouseHoverAndClickBounds(t *testing.T) {
	m := model{cfg: config.Config{Username: "deathrashed", MouseEnabled: true}}
	left, top, width := headerURLBounds(m.cfg.Username)

	updated, cmd := m.updateMouse(tea.MouseMsg{X: left, Y: top, Action: tea.MouseActionMotion})
	if cmd != nil || !updated.(model).headerURLHover {
		t.Fatal("URL hover did not activate inside the header bounds")
	}

	updated, cmd = updated.(model).updateMouse(tea.MouseMsg{X: left + width, Y: top, Action: tea.MouseActionMotion})
	if cmd != nil || updated.(model).headerURLHover {
		t.Fatal("URL hover remained active outside the header bounds")
	}

	_, cmd = m.updateMouse(tea.MouseMsg{X: left, Y: top, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft, Ctrl: true})
	if cmd == nil {
		t.Fatal("URL click did not return a Bubble Tea command")
	}
}

func TestCompactHeaderDisablesURLHitTesting(t *testing.T) {
	m := model{width: 140, cfg: config.Config{Username: "deathrashed", CompactHeader: true, MouseEnabled: true}}
	left, top, width := headerURLBounds(m.cfg.Username)
	if m.headerURLContains(left, top) {
		t.Fatal("compact header exposed a URL hit area")
	}
	updated, cmd := m.updateMouse(tea.MouseMsg{X: left, Y: top, Action: tea.MouseActionMotion})
	if cmd != nil || updated.(model).headerURLHover {
		t.Fatal("compact header activated URL hover")
	}
	if got := width; got == 0 {
		t.Fatal("URL test fixture unexpectedly has no width")
	}
}

func TestMouseCoordinatesFollowCompactAndFullHeaderHeights(t *testing.T) {
	compact := model{width: 140, cfg: config.Config{CompactHeader: true, MouseEnabled: true}, stage: stageInput, searchInput: newTextInput(512, 48)}
	updated, _ := compact.updateMouse(tea.MouseMsg{X: 30, Y: compact.headerHeight(), Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if got := updated.(model).modeChoice; got != "discography" {
		t.Fatalf("compact dashboard click selected %q, want discography", got)
	}

	file := model{width: 140, cfg: config.Config{CompactHeader: true, MouseEnabled: true}, stage: stageImportSource, modeChoice: "file"}
	updated, _ = file.updateMouse(tea.MouseMsg{X: 20, Y: file.headerHeight(), Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if got := updated.(model).importSourceIndex; got != 0 {
		t.Fatalf("compact file click selected source %d, want 0", got)
	}

	settingsModel := model{
		width:           140,
		stage:           stageConfig,
		modeChoice:      "scrobbling",
		settingsSection: settingsScrobbling,
		settingsFocus:   settingsFocusContent,
		settingsRow:     0,
		cfg: config.Config{
			CompactHeader:   true,
			MouseEnabled:    true,
			DefaultLoop:     1,
			DefaultInterval: 2 * time.Second,
		},
		configInput: newTextInput(1024, 44),
	}
	region := settingsModel.settingsGridRegion(settingsHistory)
	updated, _ = settingsModel.updateMouse(tea.MouseMsg{X: region.x, Y: region.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	gotSettings := updated.(model)
	if gotSettings.currentSettingsSection() != settingsHistory || gotSettings.stage != stageHistory {
		t.Fatalf("compact Settings click selected section=%d stage=%d", gotSettings.currentSettingsSection(), gotSettings.stage)
	}

	full := model{width: 140, cfg: config.Config{MouseEnabled: true}, stage: stageInput, searchInput: newTextInput(512, 48)}
	updated, _ = full.updateMouse(tea.MouseMsg{X: 30, Y: full.headerHeight(), Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if got := updated.(model).modeChoice; got != "discography" {
		t.Fatalf("full dashboard click selected %q, want discography", got)
	}
}

func absInt(value int) int {
	if value < 0 {
		return -value
	}
	return value
}

func TestDashboardAndSettingsGridFitHeader(t *testing.T) {
	dashboard := joinThreeLineBoxes([]string{
		renderExactBox("M A N U A L", 19, true),
		renderExactBox("D I S C O G R A P H Y", 25, false),
		renderExactBox("F I L E", 18, false),
	}, "•")
	for _, line := range strings.Split(dashboard, "\n") {
		if got := lipgloss.Width(stripANSI(line)); got > headerContentWidth {
			t.Fatalf("dashboard width = %d, exceeds %d: %q", got, headerContentWidth, line)
		}
	}

	m := model{stage: stageConfig, modeChoice: "scrobbling", settingsSection: settingsScrobbling, settingsFocus: settingsFocusContent}
	grid := renderSettingsGrid(m)
	lines := strings.Split(grid, "\n")
	if len(lines) != 6 {
		t.Fatalf("Settings grid lines = %d, want 6", len(lines))
	}
	for lineNumber, line := range lines {
		if got := lipgloss.Width(stripANSI(line)); got != settingsGridWidth {
			t.Fatalf("Settings grid line %d width = %d, want %d: %q", lineNumber+1, got, settingsGridWidth, line)
		}
	}

	wantNames := []string{"ACCOUNT", "SCROBBLING", "HISTORY", "TOOLS", "INTERFACE", "PROFILES"}
	if got := settingsSectionNames(); strings.Join(got, ",") != strings.Join(wantNames, ",") {
		t.Fatalf("Settings section order = %v, want %v", got, wantNames)
	}

	for row := 0; row < 2; row++ {
		for col, want := range []struct {
			x     int
			width int
		}{{1, settingsSideWidth}, {21, settingsCenterWidth}, {47, settingsSideWidth}} {
			section := settingsSection(row*3 + col)
			region := m.settingsGridRegion(section)
			if region.x != want.x || region.width != want.width || region.y != m.headerHeight()+row*3 {
				t.Fatalf("section %d region = x%d y%d w%d, want x%d y%d w%d", section, region.x, region.y, region.width, want.x, m.headerHeight()+row*3, want.width)
			}
		}
	}
}

func TestSettingsHeaderTitlesAndSubtitles(t *testing.T) {
	for section := settingsAccount; section <= settingsProfiles; section++ {
		spec := settingsSpec(section)
		headerSpec := compactHeaderSpecFor(spec.Mode)
		if headerSpec.Title != spec.Label || headerSpec.Subtitle != spec.Subtitle || headerSpec.Icon != spec.Icon {
			t.Fatalf("section %d header = %#v, want %#v", section, headerSpec, spec)
		}
		for _, compact := range []bool{false, true} {
			header := RenderHeader(140, stageConfig, spec.Mode, "deathrashed", "", compact)
			plain := stripANSI(header)
			if !strings.Contains(plain, spec.Label) || !strings.Contains(plain, spec.Subtitle) {
				t.Fatalf("section %d compact=%t missing title/subtitle\n%s", section, compact, plain)
			}
		}
	}
}

func TestFullHeaderUsesMutedDescriptiveSubtitles(t *testing.T) {
	for _, mode := range []string{"manual", "discography", "file", "account", "scrobbling", "history", "tools", "interface", "profiles"} {
		spec := compactHeaderSpecFor(mode)
		header := RenderHeader(140, stageInput, mode, "deathrashed", "", false)
		if !strings.Contains(header, theme.SecondaryTextStyle.Render(spec.Subtitle)) {
			t.Fatalf("mode %q subtitle is not rendered with the muted semantic style", mode)
		}
	}
}

func TestPreviewSettingsContextUsesMutedLabelsAndWhiteValues(t *testing.T) {
	context := renderSettingsContext("loop 1|interval 2s|limit all")
	for _, label := range []string{"loop ", "interval ", "limit "} {
		if !strings.Contains(context, theme.SecondaryTextStyle.Render(label)) {
			t.Fatalf("settings context missing muted label %q: %q", label, context)
		}
	}
	for _, value := range []string{"1", "2s", "all"} {
		if !strings.Contains(context, theme.PrimaryTextStyle.Render(value)) {
			t.Fatalf("settings context missing primary value %q: %q", value, context)
		}
	}
}

func TestPreviewSummaryCardsUseCentered57CellGroup(t *testing.T) {
	group := renderPreviewSummaryCards("1", "10", "2s", "10", "18s", "1")
	for lineNumber, line := range strings.Split(group, "\n") {
		if got := lipgloss.Width(stripANSI(line)); got != 57 {
			t.Fatalf("summary group line %d width = %d, want 57: %q", lineNumber+1, got, line)
		}
	}
	centered := centerToHeader(group)
	for lineNumber, line := range strings.Split(centered, "\n") {
		plain := stripANSI(line)
		if got := lipgloss.Width(plain); got != 67 {
			t.Fatalf("centered summary line %d width = %d, want 67", lineNumber+1, got)
		}
		left := len(plain) - len(strings.TrimLeft(plain, " "))
		right := len(plain) - len(strings.TrimRight(plain, " "))
		if left != 5 || right != 5 {
			t.Fatalf("centered summary line %d padding = left %d right %d, want 5/5 including the outer 67-cell frame alignment", lineNumber+1, left, right)
		}
	}
}

func TestSelectedBadgeUsesRedNumbersMutedSlashAndRedArrow(t *testing.T) {
	badge := renderSelectedBadge(1, 200)
	if !strings.Contains(badge, theme.SummaryLabelStyle.Render("SELECTED ")) {
		t.Fatal("selected badge label is not primary white")
	}
	if !strings.Contains(badge, theme.SummaryArrowStyle.Render("❯")) {
		t.Fatal("selected badge arrow is not accent red")
	}
	if !strings.Contains(badge, theme.AccentTextStyle.Render("1")) ||
		!strings.Contains(badge, theme.AccentTextStyle.Render("200")) {
		t.Fatal("selected badge numbers are not accent red")
	}
	if !strings.Contains(badge, theme.MutedStyle.Render("/")) {
		t.Fatal("selected badge slash is not muted")
	}
}

func TestLastFMSpinnerUsesRedLogoAndWhiteDots(t *testing.T) {
	spin := lastFMSpinner()
	if len(spin.Frames) != 4 {
		t.Fatalf("spinner frame count = %d, want 4", len(spin.Frames))
	}
	for index, frame := range spin.Frames {
		if got := lipgloss.Width(stripANSI(frame)); got != 5 {
			t.Fatalf("spinner frame %d width = %d, want 5: %q", index, got, frame)
		}
		if !strings.Contains(frame, theme.AccentTextStyle.Render(theme.IconDashboard)) {
			t.Fatalf("spinner frame %d does not contain the red Last.fm logo", index)
		}
		if !strings.Contains(frame, theme.PrimaryTextStyle.Render("∙")) {
			t.Fatalf("spinner frame %d does not contain white dots", index)
		}
	}
}

func TestProgressBoxUsesSingleLeftInsetAndDoneState(t *testing.T) {
	spin := spinner.New()
	spin.Spinner = lastFMSpinner()
	m := model{spinner: spin}

	active := renderProgressBox(m, 0.5, false)
	activeLines := strings.Split(stripANSI(active), "\n")
	if len(activeLines) < 2 || !strings.HasPrefix(activeLines[1], "│ "+theme.IconDashboard+" ∙ ∙ ") {
		t.Fatalf("active progress row does not start one cell inside the border: %q", activeLines)
	}

	done := renderProgressBox(m, 1, true)
	doneLines := strings.Split(stripANSI(done), "\n")
	if len(doneLines) < 2 || !strings.HasPrefix(doneLines[1], "│ "+theme.IconSuccess+"  "+theme.IconDashboard+"  ") {
		t.Fatalf("done progress row has unexpected prefix: %q", doneLines)
	}
	if !strings.Contains(done, theme.CompleteStyle.Render(theme.IconSuccess)) {
		t.Fatal("done progress row does not render the completion tick in green")
	}
	if !strings.Contains(done, theme.PrimaryTextStyle.Render(theme.IconDashboard)) {
		t.Fatal("done progress row does not render the Last.fm logo in white")
	}
	if !strings.Contains(done, theme.CompleteStyle.Render("DONE")) {
		t.Fatal("done progress row does not render DONE with the completion style")
	}
	if strings.Contains(doneLines[1], "∙") {
		t.Fatalf("done progress row still renders loading dots: %q", doneLines[1])
	}
	for index, line := range doneLines {
		if got := lipgloss.Width(line); got != 65 {
			t.Fatalf("done progress line %d width = %d, want 65: %q", index, got, line)
		}
	}
}

func TestSpinnerStopsSchedulingAfterDone(t *testing.T) {
	m := model{stage: stageDone}
	if m.spinnerActive() {
		t.Fatal("spinner remains active after the session reaches stageDone")
	}
	m.stage = stageScrobbling
	if !m.spinnerActive() {
		t.Fatal("spinner is not active during stageScrobbling")
	}
}

func TestExpandedSettingsAndFootersFitHeader(t *testing.T) {
	base := model{
		width: 140,
		cfg: config.Config{
			Username:         strings.Repeat("username", 20),
			Password:         strings.Repeat("password", 20),
			APIKey:           strings.Repeat("a", 120),
			APISecret:        strings.Repeat("s", 120),
			DefaultLoop:      12,
			DefaultInterval:  2500 * time.Millisecond,
			RetryCount:       12,
			RetryDelay:       15 * time.Second,
			DuplicateGuard:   2 * time.Hour,
			ExportDir:        "/Users/example/" + strings.Repeat("very-long-export-folder/", 5),
			EnvPath:          "/Users/example/.config/lastfm-scrobbler/.env",
			CredentialSource: "environment",
			UpdateURL:        "https://example.invalid/" + strings.Repeat("release/", 10),
			MouseEnabled:     true,
			CleanDiscography: true,
			Notify:           true,
			Profile:          "default",
		},
		profiles:    []string{"default", "archive"},
		configInput: newTextInput(1024, 44),
	}

	for _, section := range []settingsSection{settingsAccount, settingsScrobbling, settingsTools, settingsInterface} {
		updated, _ := base.openSettingsSection(section, settingsFocusContent)
		m := updated.(model)
		rows := settingsRows(section)
		for index := range rows {
			m.settingsRow = index
			m.loadSettingsField()
			for lineNumber, line := range strings.Split(renderSettingsView(m), "\n") {
				if got := lipgloss.Width(stripANSI(line)); got > headerContentWidth {
					t.Fatalf("section %d row %d line %d width = %d, exceeds %d\n%q", section, index, lineNumber+1, got, headerContentWidth, line)
				}
			}
		}
	}

	for _, current := range []struct {
		stage   stage
		section settingsSection
	}{
		{stageInput, settingsScrobbling},
		{stageConfig, settingsScrobbling},
		{stageHistory, settingsHistory},
		{stageProfiles, settingsProfiles},
		{stageInfo, settingsScrobbling},
	} {
		m := base
		m.stage = current.stage
		m.settingsSection = current.section
		m.settingsFocus = settingsFocusContent
		if current.stage == stageConfig {
			m.modeChoice = settingsSpec(current.section).Mode
			m.settingsRow = 0
			m.loadSettingsField()
		}
		for _, line := range strings.Split(renderFooter(m), "\n") {
			if got := lipgloss.Width(stripANSI(line)); got > headerContentWidth {
				t.Fatalf("footer stage %d width = %d, exceeds %d: %q", current.stage, got, headerContentWidth, line)
			}
		}
	}
}

func TestDashboardAdvertisesSettingsShortcutNotConfig(t *testing.T) {
	m := model{stage: stageInput}
	plain := stripANSI(renderFooter(m))
	if !strings.Contains(plain, "s settings") {
		t.Fatalf("Dashboard footer does not advertise s settings: %q", plain)
	}
	if strings.Contains(strings.ToLower(plain), "c config") {
		t.Fatalf("Dashboard footer still advertises c config: %q", plain)
	}
}

func TestSettingsSectionsContainEachExpectedRowExactlyOnce(t *testing.T) {
	want := map[settingsSection][]string{
		settingsAccount:    {"username", "password", "api-key", "api-secret", "credential-source", "credential-path"},
		settingsScrobbling: {"loop", "interval", "retry-count", "retry-delay", "duplicate-guard", "clean-top-albums"},
		settingsTools:      {"export-dir", "update-url", "connection-test", "diagnostics", "completions", "check-updates"},
		settingsInterface:  {"notifications", "compact-header", "mouse-support"},
	}
	seen := map[string]settingsSection{}
	for section, wantIDs := range want {
		rows := settingsRows(section)
		if len(rows) != len(wantIDs) {
			t.Fatalf("section %d row count = %d, want %d", section, len(rows), len(wantIDs))
		}
		for index, wantID := range wantIDs {
			if rows[index].ID != wantID {
				t.Fatalf("section %d row %d = %q, want %q", section, index, rows[index].ID, wantID)
			}
			if previous, exists := seen[rows[index].ID]; exists {
				t.Fatalf("row %q appears in sections %d and %d", rows[index].ID, previous, section)
			}
			seen[rows[index].ID] = section
		}
	}
}

func TestChoiceBoxUsesMutedHoverAndSelectedCardStates(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(profile)

	idle := renderChoiceBox("A C C O U N T", settingsSideWidth, false, false)
	cardTop := "╭" + strings.Repeat("─", settingsSideWidth-2) + "╮"
	if !strings.Contains(idle, theme.SecondaryTextStyle.Render("A C C O U N T")) {
		t.Fatal("idle choice card label is not muted")
	}

	hovered := renderChoiceBox("A C C O U N T", settingsSideWidth, false, true)
	if !strings.Contains(hovered, theme.AccentTextStyle.Render("A C C O U N T")) {
		t.Fatal("hovered choice card label is not Torch Red")
	}
	if !strings.Contains(hovered, theme.BorderStyle.Render(cardTop)) {
		t.Fatal("hovered unselected choice card should keep the white border")
	}

	selected := renderChoiceBox("A C C O U N T", settingsSideWidth, true, false)
	if !strings.Contains(selected, theme.SelectedModeStyle.Render("A C C O U N T")) {
		t.Fatal("selected choice card label is not bold white")
	}
	if !strings.Contains(selected, theme.InnerBorderStyle.Render(cardTop)) {
		t.Fatal("selected choice card border is not Torch Red")
	}
}

func TestFocusedTextBoxKeepsStructuralBorderWhite(t *testing.T) {
	profile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	defer lipgloss.SetColorProfile(profile)

	box := renderTextBox("SEARCH", "Slayer", "Artist, Album, or Both...", 65, true)
	if !strings.Contains(box, theme.BorderStyle.Render("╭"+strings.Repeat("─", 63)+"╮")) {
		t.Fatal("focused text box did not keep white structural border")
	}
	if strings.Contains(box, theme.InnerBorderStyle.Render("╭"+strings.Repeat("─", 63)+"╮")) {
		t.Fatal("focused text box incorrectly uses Torch Red border")
	}
	if !strings.Contains(box, theme.FocusedRowLabelStyle.Render("SEARCH ")) {
		t.Fatal("focused text box label is not Torch Red")
	}
	if !strings.Contains(box, theme.FocusedRowArrowStyle.Render("❯ ")) {
		t.Fatal("focused text box arrow is not white")
	}
}

func TestDynamicArtistHeaderBadgeScalesAndStaysWithinFrame(t *testing.T) {
	for _, artist := range []string{"Jet", "Slayer", "Oxygen Destroyer", "An Artist Name That Is Deliberately Much Longer Than The Header Badge"} {
		header := RenderHeaderWithHoverArtist(
			140,
			stageTrackSelect,
			"manual",
			"deathrashed",
			"loop 1|interval 2s|limit all",
			false,
			false,
			artist,
		)
		lines := strings.Split(header, "\n")
		if len(lines) != fullHeaderLines+2 {
			t.Fatalf("artist %q header lines = %d, want %d", artist, len(lines), fullHeaderLines+2)
		}
		for lineNumber, line := range lines {
			if got := lipgloss.Width(stripANSI(line)); got != fullHeaderWidth {
				t.Fatalf("artist %q line %d width = %d, want %d\n%q", artist, lineNumber+1, got, fullHeaderWidth, line)
			}
		}
		if !strings.Contains(header, theme.ArtistStyle.Render(truncateToWidth(spacedArtistName(artist), 53))) {
			t.Fatalf("artist %q is not rendered in the dynamic artist badge", artist)
		}
	}
}

func TestHeaderHeightIncludesArtistExtensionOnlyForFullHeader(t *testing.T) {
	full := model{
		stage:             stageTrackSelect,
		modeChoice:        "manual",
		discographyArtist: "Slayer",
	}
	if got := full.headerHeight(); got != fullHeaderLines+2 {
		t.Fatalf("full artist header height = %d, want %d", got, fullHeaderLines+2)
	}
	full.cfg.CompactHeader = true
	if got := full.headerHeight(); got != compactHeaderLines+1 {
		t.Fatalf("compact artist header height = %d, want %d", got, compactHeaderLines+1)
	}
}

func TestCompactArtistHeaderKeepsMouseBodyOffsetsAligned(t *testing.T) {
	manual := model{
		width:      140,
		stage:      stageResults,
		modeChoice: "manual",
		cfg:        config.Config{CompactHeader: true, MouseEnabled: true},
		results:    []lastfm.Album{{Artist: "Enforced", Title: "War Remains"}},
	}
	if got := manual.headerHeight(); got != compactHeaderLines+1 {
		t.Fatalf("compact Manual header height = %d, want %d", got, compactHeaderLines+1)
	}
	var manualRow mouseRegion
	for _, region := range manual.screenRegions() {
		if region.id == "results:0" {
			manualRow = region
			break
		}
	}
	if manualRow.y != manual.headerHeight()+4 {
		t.Fatalf("compact Manual result row y = %d, want %d", manualRow.y, manual.headerHeight()+4)
	}
	updated, _ := manual.updateMouse(tea.MouseMsg{X: manualRow.x, Y: manualRow.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if got := updated.(model).resultsCursor; got != 0 {
		t.Fatalf("compact Manual result click moved cursor to %d, want 0", got)
	}

	discography := model{
		width:               140,
		stage:               stageDiscographySelect,
		modeChoice:          "discography",
		cfg:                 config.Config{CompactHeader: true, MouseEnabled: true},
		discographyArtist:   "Oxygen Destroyer",
		discography:         makeAlbums(3),
		discographySelected: map[int]bool{},
	}
	if got := discography.headerHeight(); got != compactHeaderLines+1 {
		t.Fatalf("compact Discography header height = %d, want %d", got, compactHeaderLines+1)
	}
	var discographyRow mouseRegion
	for _, region := range discography.screenRegions() {
		if region.id == "discography:0" {
			discographyRow = region
			break
		}
	}
	if discographyRow.y != discography.headerHeight()+discographyChooserListRowOffset(discography) {
		t.Fatalf("compact Discography row y = %d, want %d", discographyRow.y, discography.headerHeight()+discographyChooserListRowOffset(discography))
	}
	updated, _ = discography.updateMouse(tea.MouseMsg{X: discographyRow.x, Y: discographyRow.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if got := updated.(model).discographyCursor; got != 0 {
		t.Fatalf("compact Discography click moved cursor to %d, want 0", got)
	}

	full := manual
	full.cfg.CompactHeader = false
	if got := full.headerHeight(); got != fullHeaderLines+2 {
		t.Fatalf("full artist header height = %d, want %d", got, fullHeaderLines+2)
	}
	for _, region := range full.screenRegions() {
		if region.id == "results:0" && region.y != full.headerHeight()+4 {
			t.Fatalf("full Manual result row y = %d, want %d", region.y, full.headerHeight()+4)
		}
	}
}

func TestTrackFooterUsesInteractiveIntervalNavigationAndLoopControls(t *testing.T) {
	m := model{stage: stageTrackSelect, modeChoice: "manual", interval: 2 * time.Second}
	plain := stripANSI(renderFooter(m))
	lines := strings.Split(plain, "\n")
	if len(lines) < 2 {
		t.Fatalf("track footer lines = %d, want at least 2: %q", len(lines), plain)
	}
	if lines[0] != "space check • s similar • a all • enter continue" {
		t.Fatalf("track footer line 1 = %q", lines[0])
	}
	if lines[1] != "- interval + • ↑ navigate ↓ • - loop +" {
		t.Fatalf("track footer line 2 = %q", lines[1])
	}
	for _, id := range []string{"footer:interval-down", "footer:interval-up", "footer:nav-up", "footer:nav-down", "footer:loop-down", "footer:loop-up"} {
		found := false
		for _, region := range m.footerRegions() {
			if region.id == id {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("missing mouse region %q", id)
		}
	}
}

func TestFooterHoverDescriptionsUseWhiteWithRedDynamicAlbum(t *testing.T) {
	m := model{
		stage:             stageTrackSelect,
		modeChoice:        "manual",
		discographyArtist: "Slayer",
		hoverRegion:       "footer:s",
		selectedAlbum:     lastfm.Album{Artist: "Slayer", Title: "Hell Awaits"},
	}
	footer := renderFooter(m)
	if !strings.Contains(footer, theme.PrimaryTextStyle.Render("load a list of similar albums to")) {
		t.Fatal("similar hover description is not primary white")
	}
	if !strings.Contains(footer, theme.AccentTextStyle.Render("Hell Awaits")) {
		t.Fatal("dynamic album name is not accent red")
	}

	m.hoverRegion = "footer:interval-down"
	footer = renderFooter(m)
	if !strings.Contains(footer, theme.SelectedModeStyle.Render("-")) {
		t.Fatal("hovered interval minus is not white")
	}
	if !strings.Contains(footer, theme.AccentTextStyle.Render(" interval ")) {
		t.Fatal("interval group label is not red on hover")
	}
	if !strings.Contains(footer, theme.PrimaryTextStyle.Render("decrease the time between each track scrobble")) {
		t.Fatal("interval hover description is not white")
	}
}

func TestSelectedBadgeCanAttachToTrackPanelWithoutChangingWidth(t *testing.T) {
	panel := renderPanelBoxWithSelectedAttachment([]string{"●  1  Hell Awaits", "●  2  Kill Again"}, 65, 2, 2, theme.BorderStyle)
	for lineNumber, line := range strings.Split(panel, "\n") {
		if got := lipgloss.Width(stripANSI(line)); got != 65 {
			t.Fatalf("attached badge line %d width = %d, want 65: %q", lineNumber+1, got, line)
		}
	}
	plain := stripANSI(panel)
	if !strings.Contains(plain, "SELECTED ❯   2 / 2") {
		t.Fatalf("attached SELECTED badge missing: %q", plain)
	}
	if !strings.Contains(plain, "┤") || !strings.Contains(plain, "├") {
		t.Fatalf("attached SELECTED badge is not connected to the panel border: %q", plain)
	}
}

func TestPreviewQueueShowsQueuedAlbumsInOrderAndCapsLongLists(t *testing.T) {
	m := model{modeChoice: "discography"}
	for i := 1; i <= 6; i++ {
		m.scrobbleQueue = append(m.scrobbleQueue, queuedTrack{Artist: "Slayer", Album: fmt.Sprintf("Album %d", i)})
	}
	box := stripANSI(renderPreviewQueueBox(m))
	for i := 1; i <= 5; i++ {
		if !strings.Contains(box, fmt.Sprintf("Album %d", i)) {
			t.Fatalf("queue preview missing Album %d: %q", i, box)
		}
	}
	if strings.Contains(box, "Album 6") {
		t.Fatalf("queue preview should cap the visible list at five albums: %q", box)
	}
	if !strings.Contains(box, "… 1 more") {
		t.Fatalf("queue preview does not summarize hidden albums: %q", box)
	}
}

func TestArtistHeaderRemovesDuplicateArtistFromScrobbleStatus(t *testing.T) {
	for _, mode := range []string{"manual", "discography"} {
		album := lastfm.Album{Artist: "Slayer", Title: "Hell Awaits"}
		m := model{
			stage:          stageScrobbling,
			modeChoice:     mode,
			selectedAlbum:  album,
			selectedAlbums: []lastfm.Album{album},
			scrobbleQueue: []queuedTrack{{
				Artist:     "Slayer",
				Album:      "Hell Awaits",
				Title:      "Kill Again",
				AlbumIndex: 1,
				AlbumTotal: 1,
				TrackIndex: 2,
				TrackTotal: 7,
			}},
			interval: time.Second,
		}
		plain := stripANSI(renderScrobbleStatus(m, false))
		if strings.Contains(plain, "ARTIST ❯") {
			t.Fatalf("%s status duplicates the artist already shown in the header: %q", mode, plain)
		}
		if strings.Contains(plain, "ALBUM ❯ Slayer — Hell Awaits") {
			t.Fatalf("%s album row repeats the header artist: %q", mode, plain)
		}
		if !strings.Contains(plain, "ALBUM ❯ Hell Awaits") {
			t.Fatalf("%s status does not show the album beneath the artist header: %q", mode, plain)
		}
	}
}

func TestFooterHoverHelpKeepsActionRowsStable(t *testing.T) {
	m := model{
		stage:         stageTrackSelect,
		modeChoice:    "manual",
		selectedAlbum: lastfm.Album{Artist: "Slayer", Title: "Hell Awaits"},
	}
	idle := strings.Split(stripANSI(renderFooter(m)), "\n")
	m.hoverRegion = "footer:s"
	hovered := strings.Split(stripANSI(renderFooter(m)), "\n")
	if len(idle) != 4 || len(hovered) != 4 {
		t.Fatalf("footer line count changed on hover: idle=%d hovered=%d", len(idle), len(hovered))
	}
	for index := 0; index < 2; index++ {
		if idle[index] != hovered[index] {
			t.Fatalf("action row %d changed on hover: idle=%q hovered=%q", index+1, idle[index], hovered[index])
		}
	}
	if hovered[2] != "load a list of similar albums to" || hovered[3] != "Hell Awaits" {
		t.Fatalf("unexpected reserved hover help rows: %q", hovered[2:])
	}
}

func TestManualResultsAttachResultsCountWithoutSelectedBadge(t *testing.T) {
	m := model{
		stage:         stageResults,
		modeChoice:    "manual",
		results:       []lastfm.Album{{Artist: "Enforced", Title: "War Remains"}, {Artist: "Enforced", Title: "Kill Grid"}},
		resultsCursor: 0,
	}
	view := renderResultsView(m)
	plain := stripANSI(view)
	if !strings.Contains(plain, "MATCH ❯ Enforced — War Remains") {
		t.Fatalf("manual result view missing MATCH panel: %q", plain)
	}
	if !strings.Contains(plain, "RESULTS ❯  2") {
		t.Fatalf("manual result view missing attached results count: %q", plain)
	}
	if strings.Contains(plain, "SELECTED") {
		t.Fatalf("manual result view still exposes a multiselect SELECTED badge: %q", plain)
	}
	for lineNumber, line := range strings.Split(view, "\n") {
		if got := lipgloss.Width(stripANSI(line)); got > headerContentWidth {
			t.Fatalf("manual result line %d width=%d exceeds %d: %q", lineNumber+1, got, headerContentWidth, line)
		}
	}
}

func TestDiscographyChooserIntegratesControlsListAndCounts(t *testing.T) {
	m := model{
		stage:               stageDiscographySelect,
		modeChoice:          "discography",
		discography:         makeAlbums(3),
		discographySelected: map[int]bool{1: true},
		discographyCursor:   0,
	}
	chooser := renderDiscographyChooser(m, m.discographyVisibleIndexes())
	assertBlockWidth(t, chooser, 65)
	plain := stripANSI(chooser)
	for _, want := range []string{"SORT ❯ LFM", "FILTER ❯", "CLEAN ❯ OFF", "RESULTS ❯  3", "SELECTED ❯  1"} {
		if !strings.Contains(plain, want) {
			t.Fatalf("discography chooser missing %q:\n%s", want, plain)
		}
	}
	if strings.Contains(plain, "VIEW ❯") {
		t.Fatalf("discography chooser still contains the old detached VIEW panel:\n%s", plain)
	}
}

func TestDiscographyLongFilterExpandsBelowCompactControl(t *testing.T) {
	query := "this is for even longer filter queries that do not fit in the provided box parameters"
	m := model{
		stage:               stageDiscographySelect,
		modeChoice:          "discography",
		discography:         makeAlbums(3),
		discographySelected: map[int]bool{},
		discographyFilter:   query,
	}
	if !discographyFilterExpanded(m) {
		t.Fatal("long Discography filter did not expand")
	}
	chooser := renderDiscographyChooser(m, m.discographyVisibleIndexes())
	assertBlockWidth(t, chooser, 65)
	plain := stripANSI(chooser)
	if !strings.Contains(plain, "┴") {
		t.Fatalf("expanded filter is not visually connected to the FILTER tab:\n%s", plain)
	}
	if !strings.Contains(plain, "this is for even longer filter queries") {
		t.Fatalf("expanded filter does not show the long query:\n%s", plain)
	}
}

func TestManualResultsExposeResolvedArtistInHeader(t *testing.T) {
	m := model{
		stage:         stageResults,
		modeChoice:    "manual",
		results:       []lastfm.Album{{Artist: "Enforced", Title: "War Remains"}},
		resultsCursor: 0,
	}
	if got := m.headerArtist(); got != "Enforced" {
		t.Fatalf("manual results header artist=%q, want Enforced", got)
	}
}
