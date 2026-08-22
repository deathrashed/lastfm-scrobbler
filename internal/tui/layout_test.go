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
	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
	setupstate "github.com/deathrashed/lastfm-scrobbler/internal/setup"
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
		header := RenderHeader(67, stageInput, mode, "deathrashed", "", false)
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
		header := RenderHeader(67, stageInput, mode, "deathrashed", "", true)
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
	before := RenderHeaderWithHoverArtist(67, unresolved.stage, unresolved.modeChoice, "", "", true, false, unresolved.headerArtist())
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
	header := RenderHeaderWithHoverArtist(67, resolved.stage, resolved.modeChoice, "", "", true, false, resolved.headerArtist())
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
	header := RenderHeaderWithHoverArtist(67, m.stage, m.modeChoice, "", "", true, false, m.headerArtist())
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

	header := RenderHeaderWithHoverArtist(67, stageResults, "manual", "", "", true, false, "Enforced")
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
	header := RenderHeaderWithHoverArtist(67, m.stage, m.modeChoice, "", "", true, false, m.headerArtist())
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

func TestResponsiveApplicationWidthAndCentering(t *testing.T) {
	for _, terminalWidth := range []int{67, 68, 79, 80, 95, 103, 104, 111, 127, 128, 160, 200} {
		width := appWidth(terminalWidth)
		if width < minAppWidth || width > maxAppWidth {
			t.Fatalf("terminal %d produced app width %d", terminalWidth, width)
		}
		offset := appOffset(terminalWidth)
		if terminalWidth > maxAppWidth && terminalWidth-width > 1 && offset == 0 {
			t.Fatalf("terminal %d did not center the capped application", terminalWidth)
		}
		if absInt(terminalWidth-width-2*offset) > 1 {
			t.Fatalf("terminal %d margins are not balanced: width=%d offset=%d", terminalWidth, width, offset)
		}
		header := RenderHeader(terminalWidth, stageInput, "", "user", "", false)
		for lineNumber, line := range strings.Split(header, "\n") {
			if got := displayWidth(stripANSI(line)); got != width {
				t.Fatalf("terminal %d header line %d width=%d want=%d", terminalWidth, lineNumber+1, got, width)
			}
		}
	}
}

func TestFullHeaderActivityAddsAttachedProfileAndStableActivityRow(t *testing.T) {
	without := model{width: 127, cfg: config.Config{NowPlaying: false}}
	with := without
	with.cfg.NowPlaying = true
	if got, want := without.headerHeight(), fullHeaderLines; got != want {
		t.Fatalf("activity-off header height=%d want=%d", got, want)
	}
	if got, want := with.headerHeight(), fullHeaderLines+2; got != want {
		t.Fatalf("activity-on header height=%d want=%d", got, want)
	}
	if got := len(strings.Split(with.renderHeader(), "\n")); got != with.headerHeight() {
		t.Fatalf("rendered activity header lines=%d want=%d", got, with.headerHeight())
	}
	with.cfg.CompactHeader = true
	if got := with.headerHeight(); got != compactHeaderLines {
		t.Fatalf("compact activity header height=%d want=%d", got, compactHeaderLines)
	}
}

func TestFullHeaderActivityStaysWithinResponsiveWidth(t *testing.T) {
	for _, width := range []int{67, 80, 104, 127, 160, 200} {
		m := model{
			width:         width,
			cfg:           config.Config{NowPlaying: true},
			activityState: activityCurrent,
			activityTrack: lastfm.RecentTrack{
				Artist: "A Very Long Artist Name That Should Truncate Safely",
				Title:  "A Very Long Track Title That Should Stay Inside The Header",
			},
		}
		for lineIndex, line := range strings.Split(stripANSI(m.renderHeader()), "\n") {
			if got := displayWidth(line); got > m.appWidth() {
				t.Fatalf("width %d activity line %d = %d, want <= %d: %q", width, lineIndex+1, got, m.appWidth(), line)
			}
		}
	}
}

func TestResponsiveReferenceViewsDoNotOverflow(t *testing.T) {
	for _, width := range []int{67, 80, 104, 127, 160, 200} {
		base := model{
			width: width, height: 40,
			cfg:         config.Config{MouseEnabled: true},
			modeChoice:  "discography",
			discography: makeAlbums(8), discographyArtist: "Oxygen Destroyer",
			discographySelected: map[int]bool{},
			searchInput:         newTextInput(512, 48), filterInput: newTextInput(128, 44),
			configInput: newTextInput(1024, 44), envInput: newTextInput(1024, 48), profileInput: newTextInput(64, 40),
			profiles: []string{"default"}, history: []sessionstore.Record{testLastSessionRecord()},
			setup: setupstate.NewState(config.Config{}),
		}
		views := []string{
			base.renderHeader(), renderInputView(base), renderSearchView(base),
			renderDiscographySelectView(base), renderImportSourceView(base),
			renderSettingsView(base), renderInfoView(base), renderCompletionsView(base),
			renderHistoryView(base), renderLastSessionView(base), renderSetupView(base),
		}
		for viewIndex, view := range views {
			for lineIndex, line := range strings.Split(stripANSI(view), "\n") {
				if got := displayWidth(line); got > base.appWidth() {
					t.Fatalf("width %d view %d line %d width=%d app=%d: %q", width, viewIndex, lineIndex+1, got, base.appWidth(), line)
				}
			}
		}
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
	dashboard := dashboardCardPositions(compact.panelWidth())
	updated, _ := compact.updateMouse(tea.MouseMsg{X: compact.appX() + compact.workX() + dashboard[1], Y: compact.headerHeight(), Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	if got := updated.(model).modeChoice; got != "discography" {
		t.Fatalf("compact dashboard click selected %q, want discography", got)
	}

	file := model{width: 140, cfg: config.Config{CompactHeader: true, MouseEnabled: true}, stage: stageImportSource, modeChoice: "file"}
	fileCards := fileSourceCardPositions(file.panelWidth())
	updated, _ = file.updateMouse(tea.MouseMsg{X: file.appX() + file.workX() + fileCards[0][0], Y: file.headerHeight() + fileCards[0][1], Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
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
	updated, _ = settingsModel.updateMouse(tea.MouseMsg{X: settingsModel.appX() + region.x, Y: region.y, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
	gotSettings := updated.(model)
	if gotSettings.currentSettingsSection() != settingsHistory || gotSettings.stage != stageHistory {
		t.Fatalf("compact Settings click selected section=%d stage=%d", gotSettings.currentSettingsSection(), gotSettings.stage)
	}

	full := model{width: 140, cfg: config.Config{MouseEnabled: true}, stage: stageInput, searchInput: newTextInput(512, 48)}
	fullDashboard := dashboardCardPositions(full.panelWidth())
	updated, _ = full.updateMouse(tea.MouseMsg{X: full.appX() + full.workX() + fullDashboard[1], Y: full.headerHeight(), Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
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
			header := RenderHeader(67, stageConfig, spec.Mode, "deathrashed", "", compact)
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
		width: 67,
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
	for _, required := range []string{"enter select", "→ ↑ navigate ↓ ←", "s settings", "i info", "h history", "m d quick f q", "r rerun", "? help"} {
		if !strings.Contains(plain, required) {
			t.Fatalf("Dashboard footer missing %q: %q", required, plain)
		}
	}
	if strings.Contains(plain, "p profiles") {
		t.Fatalf("Dashboard footer still advertises Profiles shortcut: %q", plain)
	}
}

func TestDashboardFooterMatchesApprovedOrder(t *testing.T) {
	lines := strings.Split(stripANSI(renderFooter(model{stage: stageInput})), "\n")
	want := []string{
		"enter select • → ↑ navigate ↓ ← • s settings",
		"i info • h history • m d quick f q • r rerun • ? help",
	}
	if len(lines) < len(want) {
		t.Fatalf("Dashboard footer lines=%d want at least %d: %q", len(lines), len(want), lines)
	}
	for index := range want {
		if lines[index] != want[index] {
			t.Fatalf("Dashboard footer line %d=%q want=%q", index+1, lines[index], want[index])
		}
	}
}

func TestSettingsSectionsContainEachExpectedRowExactlyOnce(t *testing.T) {
	want := map[settingsSection][]string{
		settingsAccount:    {"username", "password", "api-key", "api-secret", "credential-source", "credential-path", "auth-status", "reauthenticate"},
		settingsScrobbling: {"loop", "interval", "retry-count", "retry-delay", "duplicate-guard", "clean-top-albums"},
		settingsTools:      {"export-dir", "update-url", "connection-test", "diagnostics", "completions", "check-updates"},
		settingsInterface:  {"notifications", "now-playing", "compact-header", "mouse-support"},
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
			67,
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

func TestDiscographyExpandedFilterJointAlignsWithFilterTabAtResponsiveWidths(t *testing.T) {
	for _, width := range []int{67, 80, 81, 82, 95, 104, 111, 127} {
		m := model{
			width:                width,
			stage:                stageDiscographySelect,
			modeChoice:           "discography",
			discography:          makeAlbums(3),
			discographySelected:  map[int]bool{},
			discographyFiltering: true,
			filterInput:          newTextInput(128, 44),
		}
		plain := stripANSI(renderDiscographyChooser(m, m.discographyVisibleIndexes()))
		lines := strings.Split(plain, "\n")
		filterJointX := -1
		expandedJointX := -1
		for _, line := range lines {
			if filterJointX < 0 {
				if index := strings.Index(line, "┬"); index >= 0 {
					filterJointX = lipgloss.Width(line[:index])
				}
			}
			if expandedJointX < 0 {
				if index := strings.Index(line, "┴"); index >= 0 {
					expandedJointX = lipgloss.Width(line[:index])
				}
			}
		}
		if filterJointX < 0 || expandedJointX < 0 {
			t.Fatalf("width %d missing filter connector joints:\n%s", width, plain)
		}
		if filterJointX != expandedJointX {
			t.Fatalf("width %d FILTER connector is offset: tab joint x=%d expanded joint x=%d:\n%s", width, filterJointX, expandedJointX, plain)
		}
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

func TestResponsiveWorkAreaCapsAndCenters(t *testing.T) {
	tests := []struct {
		terminal int
		app      int
		panel    int
		workX    int
	}{
		{67, 67, 65, 1},
		{80, 80, 78, 1},
		{104, 104, 102, 1},
		{127, 127, 103, 12},
		{160, 127, 103, 12},
	}
	for _, tt := range tests {
		m := model{width: tt.terminal}
		if got := m.appWidth(); got != tt.app {
			t.Fatalf("terminal %d app width=%d want=%d", tt.terminal, got, tt.app)
		}
		if got := m.panelWidth(); got != tt.panel {
			t.Fatalf("terminal %d work width=%d want=%d", tt.terminal, got, tt.panel)
		}
		if got := m.workX(); got != tt.workX {
			t.Fatalf("terminal %d work x=%d want=%d", tt.terminal, got, tt.workX)
		}
		left := m.workX()
		right := m.appWidth() - m.panelWidth() - left
		if absInt(left-right) > 1 {
			t.Fatalf("terminal %d work margins left=%d right=%d", tt.terminal, left, right)
		}
	}
}

func TestHelpUsesFullResponsiveInfoWidth(t *testing.T) {
	for _, tt := range []struct {
		name string
		m    model
		want int
	}{
		{"compact", model{width: 67}, 65},
		{"wide", model{width: 127}, 125},
	} {
		lines := strings.Split(stripANSI(renderHelpView(tt.m)), "\n")
		if len(lines) == 0 {
			t.Fatalf("%s help rendered no lines", tt.name)
		}
		if got := lipgloss.Width(strings.TrimSpace(lines[0])); got != tt.want {
			t.Fatalf("%s help panel width=%d want=%d", tt.name, got, tt.want)
		}
	}
}

func TestResponsiveCardGapsAreBoundedInsteadOfGreedy(t *testing.T) {
	tests := []struct {
		panel int
		want  int
	}{
		{65, 1},
		{78, 3},
		{102, 5},
		{103, 5},
	}
	for _, tt := range tests {
		positions, gap := responsiveCardPositions([]int{19, 25, 18}, tt.panel)
		if gap != tt.want {
			t.Fatalf("panel %d gap=%d want=%d", tt.panel, gap, tt.want)
		}
		groupStart := positions[0]
		groupEnd := positions[len(positions)-1] + 18
		if absInt(groupStart-(tt.panel-groupEnd)) > 1 {
			t.Fatalf("panel %d card group not centered: left=%d right=%d", tt.panel, groupStart, tt.panel-groupEnd)
		}
	}
}

func assertCardSeparatorsOnlyOnMiddleRows(t *testing.T, name, block string, groups int, separatorsPerMiddle ...int) {
	t.Helper()
	if len(separatorsPerMiddle) == 0 {
		t.Fatal("separator expectations are missing")
	}
	lines := strings.Split(stripANSI(block), "\n")
	need := groups * 3
	if len(lines) < need {
		t.Fatalf("%s lines=%d want at least %d\n%s", name, len(lines), need, block)
	}
	for group := 0; group < groups; group++ {
		for row := 0; row < 3; row++ {
			got := strings.Count(lines[group*3+row], "•")
			want := 0
			if row == 1 {
				want = separatorsPerMiddle[minInt(group, len(separatorsPerMiddle)-1)]
			}
			if got != want {
				t.Fatalf("%s group %d row %d bullet count=%d want=%d: %q", name, group, row, got, want, lines[group*3+row])
			}
		}
	}
}

func TestWideNavigationGroupsDoNotCreateVerticalBulletColumns(t *testing.T) {
	m := model{width: 127, modeIndex: 1, infoIndex: 0, importSourceIndex: 0, searchInput: newTextInput(512, 48)}
	assertCardSeparatorsOnlyOnMiddleRows(t, "dashboard", renderInputView(m), 1, 2)
	assertCardSeparatorsOnlyOnMiddleRows(t, "settings", renderSettingsGrid(m), 2, 2)
	assertCardSeparatorsOnlyOnMiddleRows(t, "file", renderImportSourceView(m), 2, 1)
	assertCardSeparatorsOnlyOnMiddleRows(t, "info", renderInfoView(m), 2, 2, 1)
}

func TestWideCompactHeaderRemainsFixedSizeAndCentered(t *testing.T) {
	header := RenderHeader(127, stageInput, "manual", "deathrashed", "", true)
	lines := strings.Split(stripANSI(header), "\n")
	if len(lines) != compactHeaderLines {
		t.Fatalf("compact header lines=%d want=%d", len(lines), compactHeaderLines)
	}
	for index, line := range lines {
		if got := displayWidth(line); got != 127 {
			t.Fatalf("wide compact line %d width=%d want=127", index+1, got)
		}
		trimmed := strings.TrimSpace(line)
		if got := displayWidth(trimmed); got != fullHeaderWidth {
			t.Fatalf("wide compact inner line %d width=%d want=%d: %q", index+1, got, fullHeaderWidth, trimmed)
		}
		left := len(line) - len(strings.TrimLeft(line, " "))
		right := len(line) - len(strings.TrimRight(line, " "))
		if absInt(left-right) > 1 {
			t.Fatalf("wide compact line %d margins left=%d right=%d", index+1, left, right)
		}
	}
}

func TestWideDiscographyUsesCenteredBoundedWorkGeometry(t *testing.T) {
	m := model{
		width:               127,
		stage:               stageDiscographySelect,
		modeChoice:          "discography",
		discography:         makeAlbums(5),
		discographySelected: map[int]bool{},
		discographyArtist:   "Oxygen Destroyer",
		filterInput:         newTextInput(128, 44),
	}
	chooser := renderDiscographyChooser(m, m.discographyVisibleIndexes())
	assertBlockWidth(t, chooser, maxWorkWidth)
	lines := strings.Split(stripANSI(chooser), "\n")
	if len(lines) < 3 {
		t.Fatalf("wide chooser has only %d lines", len(lines))
	}
	if strings.HasPrefix(lines[0], "─") || strings.HasPrefix(lines[0], "╭─") {
		t.Fatalf("wide chooser has detached/floating top-border stub: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "╭") || !strings.HasSuffix(lines[1], "╮") {
		t.Fatalf("wide chooser outer control border malformed: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "│") || !strings.HasSuffix(lines[2], "│") {
		t.Fatalf("wide chooser control content border malformed: %q", lines[2])
	}
	if got := discographyFilterTabWidthFor(m); got > discographyFilterTabMaxWidth {
		t.Fatalf("wide FILTER width=%d exceeds cap=%d", got, discographyFilterTabMaxWidth)
	}
	if got := discographyExpandedFilterWidthFor(m); got > discographyExpandedFilterMax {
		t.Fatalf("wide expanded FILTER width=%d exceeds cap=%d", got, discographyExpandedFilterMax)
	}
	left := discographySortTabX(m) - m.workX()
	right := m.panelWidth() - ((discographyCleanTabX(m) - m.workX()) + discographyCleanTabWidth)
	if absInt(left-right) > 1 {
		t.Fatalf("Discography control group margins left=%d right=%d", left, right)
	}
}

func TestLastSessionUsesAttachedSessionPanel(t *testing.T) {
	m := model{width: 127, stage: stageLastSession, history: []sessionstore.Record{testLastSessionRecord()}}
	view := stripANSI(renderLastSessionView(m))
	if !strings.Contains(view, "L A S T   S E S S I O N") {
		t.Fatalf("Last Session attached title missing:\n%s", view)
	}
	for _, want := range []string{"ARTIST      ❯ Terror", "ALBUM       ❯ Keepers Of The Faith", "TRACKS      ❯ 1", "LOOP        ❯ 1", "INTERVAL    ❯ 2s"} {
		if !strings.Contains(view, want) {
			t.Fatalf("Last Session missing %q:\n%s", want, view)
		}
	}
}

func TestResponsiveListBudgetsKeepFullViewWithinTerminal(t *testing.T) {
	for _, height := range []int{28, 30, 32, 36, 40, 44, 50, 56, 64} {
		manual := model{
			width: 127, height: height, stage: stageResults, modeChoice: "manual",
			cfg: config.Config{NowPlaying: true}, results: makeAlbums(50),
		}
		if got := len(strings.Split(manual.View(), "\n")); got > height {
			t.Fatalf("manual height %d rendered %d lines", height, got)
		}
	}

	for _, height := range []int{40, 44, 50} {
		history := make([]sessionstore.Record, 40)
		for index := range history {
			history[index] = testLastSessionRecord()
			history[index].ID = fmt.Sprintf("history-%d", index)
		}
		historyView := model{
			width: 127, height: height, stage: stageHistory, modeChoice: "history",
			settingsSection: settingsHistory, settingsFocus: settingsFocusContent,
			cfg: config.Config{NowPlaying: true}, history: history,
		}
		if got := len(strings.Split(historyView.View(), "\n")); got > height {
			t.Fatalf("history height %d rendered %d lines", height, got)
		}
	}
}

func TestNowPlayingPromotesProfileURLIntoAttachedTopBadge(t *testing.T) {
	m := model{width: 67, cfg: config.Config{Username: "deathrashed", NowPlaying: true}, activityState: activityRecent, activityTrack: lastfm.RecentTrack{Artist: "Hypocrisy", Title: "Fire in the Sky"}}
	lines := strings.Split(stripANSI(m.renderHeader()), "\n")
	if len(lines) != fullHeaderLines+2 {
		t.Fatalf("activity header lines=%d want=%d", len(lines), fullHeaderLines+2)
	}
	if !strings.Contains(lines[1], "┤ last.fm/user/deathrashed ├") {
		t.Fatalf("profile URL is not attached to the top border: %q", lines[1])
	}
	if !strings.Contains(lines[2], "╰") || !strings.Contains(lines[2], "╯") {
		t.Fatalf("attached profile badge has no closure row: %q", lines[2])
	}
	if !strings.Contains(lines[3], "Hypocrisy - Fire in the Sky") {
		t.Fatalf("activity row is not directly below attached profile badge: %q", lines[3])
	}
	for index, line := range lines {
		if got := displayWidth(line); got != 67 {
			t.Fatalf("attached profile line %d width=%d want=67: %q", index+1, got, line)
		}
	}
}

func TestHeaderLayoutAdaptsToTerminalHeight(t *testing.T) {
	tests := []struct {
		name   string
		height int
		want   headerLayout
	}{
		{name: "short", height: 20, want: headerCompact},
		{name: "medium", height: 27, want: headerClassic},
		{name: "tall", height: 40, want: headerHero},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := model{width: 100, height: test.height}
			if got := m.headerLayout(); got != test.want {
				t.Fatalf("height %d header layout=%d want=%d", test.height, got, test.want)
			}
			if got := len(strings.Split(m.renderHeader(), "\n")); got != m.headerHeight() {
				t.Fatalf("height %d rendered header lines=%d headerHeight=%d", test.height, got, m.headerHeight())
			}
		})
	}

	forced := model{width: 100, height: 60, cfg: config.Config{CompactHeader: true}}
	if got := forced.headerLayout(); got != headerCompact {
		t.Fatalf("forced compact layout=%d want=%d", got, headerCompact)
	}
}

func TestHeroHeaderRestoresOriginalFrameStructure(t *testing.T) {
	m := model{width: 100, height: 40, stage: stageInput, modeChoice: "", cfg: config.Config{Username: "deathrashed"}}
	lines := strings.Split(stripANSI(m.renderHeader()), "\n")
	if got, want := len(lines), heroHeaderLines; got != want {
		t.Fatalf("hero header lines=%d want=%d\n%s", got, want, stripANSI(m.renderHeader()))
	}
	// Top: one outer frame whose border is interrupted by the centered URL
	// capsule, exactly like the original Scrobbler header.
	if !strings.HasPrefix(lines[1], "╭") || !strings.HasSuffix(lines[1], "╮") ||
		!strings.Contains(lines[1], "┤") || !strings.Contains(lines[1], "├") {
		t.Fatalf("top border is not a single frame interrupted by the URL capsule: %q", lines[1])
	}
	if !strings.Contains(lines[1], "last.fm/user/deathrashed") {
		t.Fatalf("URL capsule does not carry the dynamic profile URL: %q", lines[1])
	}
	if !strings.HasPrefix(lines[2], "│") || !strings.HasSuffix(lines[2], "│") ||
		!strings.Contains(lines[2], "╰") || !strings.Contains(lines[2], "╯") {
		t.Fatalf("URL capsule has no closure row inside the frame: %q", lines[2])
	}
	// The floating design caption above the frame is gone.
	if strings.Contains(stripANSI(m.renderHeader()), "•  SCROBBLER  •") {
		t.Fatal("hero header still carries the floating scrobbler caption")
	}
	// Wordmark sits inside the outer frame.
	for row := 3; row < 3+len(heroWordmarkLines()); row++ {
		if !strings.HasPrefix(lines[row], "│") || !strings.HasSuffix(lines[row], "│") {
			t.Fatalf("wordmark row %d is not inside the outer frame: %q", row, lines[row])
		}
	}
	if !strings.Contains(lines[3], "███████╗") {
		t.Fatalf("hero wordmark is missing the block brand: %q", lines[3])
	}
	// A single subheader line sits inside the frame, directly below the
	// wordmark, with the static fallback when nothing is playing.
	subheaderRow := 3 + len(heroWordmarkLines())
	if !strings.HasPrefix(lines[subheaderRow], "│") || !strings.HasSuffix(lines[subheaderRow], "│") {
		t.Fatalf("subheader row is not inside the outer frame: %q", lines[subheaderRow])
	}
	if !strings.Contains(lines[subheaderRow], "SEARCH  •  SELECT  •  SCROBBLE") {
		t.Fatalf("subheader row does not show the static SEARCH • SELECT • SCROBBLE line: %q", lines[subheaderRow])
	}
	// Bottom: stage badge interrupts the lower border; the small icon badge
	// sits directly beneath it.
	badgeTopRow := subheaderRow + 1
	bottomRow := subheaderRow + 2
	iconRow := subheaderRow + 3
	if !strings.HasPrefix(lines[badgeTopRow], "│") || !strings.Contains(lines[badgeTopRow], "╭") {
		t.Fatalf("stage badge top is not attached inside the frame: %q", lines[badgeTopRow])
	}
	if !strings.HasPrefix(lines[bottomRow], "╰") || !strings.Contains(lines[bottomRow], "┤ D A S H B O A R D ├") {
		t.Fatalf("stage badge does not interrupt the bottom border: %q", lines[bottomRow])
	}
	if !strings.Contains(lines[iconRow], "╰") || !strings.Contains(lines[iconRow], theme.IconDashboard) || !strings.Contains(lines[iconRow], "╯") {
		t.Fatalf("icon badge is not directly beneath the stage badge: %q", lines[iconRow])
	}
	for index, line := range lines {
		if got := displayWidth(line); got != m.appWidth() {
			t.Fatalf("hero line %d width=%d want=%d: %q", index+1, got, m.appWidth(), line)
		}
	}
}

func TestHeroHeaderShowsNowPlayingInSubheaderRow(t *testing.T) {
	m := model{
		width:         100,
		height:        40,
		stage:         stageInput,
		cfg:           config.Config{Username: "deathrashed", NowPlaying: true},
		activityState: activityCurrent,
		activityTrack: lastfm.RecentTrack{Artist: "Overpower", Title: "They Came From Beyond", NowPlaying: true},
		activityFrame: 1,
	}
	lines := strings.Split(stripANSI(m.renderHeader()), "\n")
	if got, want := len(lines), heroHeaderLines; got != want {
		t.Fatalf("hero activity header lines=%d want=%d", got, want)
	}
	subheaderRow := 1 + len(heroWordmarkLines()) + 2
	subheader := lines[subheaderRow]
	// Now Playing is a plain centered row inside the frame — never a ┤caption├
	// treatment on the frame border.
	if !strings.HasPrefix(subheader, "│") || !strings.HasSuffix(subheader, "│") || strings.Contains(subheader, "┤") {
		t.Fatalf("Now Playing is not a plain row inside the frame: %q", subheader)
	}
	if !strings.Contains(subheader, "Overpower - They Came From Beyond") {
		t.Fatalf("Now Playing is missing from the subheader row: %q", subheader)
	}
	if !strings.Contains(subheader, activityVolumeFrames[1]) {
		t.Fatalf("current activity does not render its volume icon: %q", subheader)
	}

	// Most-recently-played keeps the static history icon and no animation.
	recent := m
	recent.activityState = activityRecent
	recent.activityFrame = 0
	recentLines := strings.Split(stripANSI(recent.renderHeader()), "\n")
	recentSubheader := recentLines[subheaderRow]
	if !strings.Contains(recentSubheader, "Overpower - They Came From Beyond") || !strings.Contains(recentSubheader, theme.IconHistory) {
		t.Fatalf("recent activity lost its static history icon: %q", recentSubheader)
	}
	if strings.Contains(recentSubheader, activityVolumeFrames[1]) {
		t.Fatalf("recent activity incorrectly animates the volume icon: %q", recentSubheader)
	}
	for index, line := range lines {
		if got := displayWidth(line); got != m.appWidth() {
			t.Fatalf("hero activity line %d width=%d want=%d: %q", index+1, got, m.appWidth(), line)
		}
	}
}

func TestHeroHeaderFallsBackToStaticSubheader(t *testing.T) {
	cases := []struct {
		name       string
		state      activityState
		nowPlaying bool
	}{
		{name: "disabled", state: activityLoading, nowPlaying: false},
		{name: "loading", state: activityLoading, nowPlaying: true},
		{name: "no tracks", state: activityNoTracks, nowPlaying: true},
		{name: "unavailable", state: activityUnavailable, nowPlaying: true},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			m := model{
				width:         100,
				height:        40,
				stage:         stageInput,
				cfg:           config.Config{Username: "deathrashed", NowPlaying: test.nowPlaying},
				activityState: test.state,
			}
			lines := strings.Split(stripANSI(m.renderHeader()), "\n")
			subheader := lines[1+len(heroWordmarkLines())+2]
			if !strings.Contains(subheader, "SEARCH  •  SELECT  •  SCROBBLE") {
				t.Fatalf("subheader=%q want static SEARCH • SELECT • SCROBBLE", subheader)
			}
			for _, banned := range []string{"loading", "unavailable", "no recent", "error"} {
				if strings.Contains(strings.ToLower(subheader), banned) {
					t.Fatalf("subheader leaks transient state %q: %q", banned, subheader)
				}
			}
		})
	}
}

func TestHeroHeaderTruncatesLongNowPlayingWithinFrame(t *testing.T) {
	artist := strings.Repeat("A Very Long Artist ", 8)
	title := strings.Repeat("A Very Long Track ", 8)
	for _, width := range []int{67, 104, 127} {
		m := model{
			width:         width,
			height:        40,
			stage:         stageInput,
			cfg:           config.Config{Username: "deathrashed", NowPlaying: true},
			activityState: activityCurrent,
			activityTrack: lastfm.RecentTrack{Artist: artist, Title: title, NowPlaying: true},
			activityFrame: 1,
		}
		lines := strings.Split(stripANSI(m.renderHeader()), "\n")
		if got, want := len(lines), heroHeaderLines; got != want {
			t.Fatalf("width %d hero lines=%d want=%d", width, got, want)
		}
		for index, line := range lines {
			if got := displayWidth(line); got != m.appWidth() {
				t.Fatalf("width %d line %d width=%d want=%d", width, index+1, got, m.appWidth())
			}
		}
		if !strings.Contains(lines[1+len(heroWordmarkLines())+2], "…") {
			t.Fatalf("width %d long Now Playing text was not ellipsis-truncated", width)
		}
	}
}

func TestHeroHeaderFitsResponsiveWidths(t *testing.T) {
	for _, width := range []int{67, 80, 104, 127, 160} {
		m := model{width: width, height: 40, stage: stageInput, cfg: config.Config{Username: "deathrashed"}}
		for lineIndex, line := range strings.Split(stripANSI(m.renderHeader()), "\n") {
			if got := displayWidth(line); got > m.appWidth() {
				t.Fatalf("width %d hero line %d=%d want <=%d: %q", width, lineIndex+1, got, m.appWidth(), line)
			}
		}
	}
}

func TestFooterDetailRowsAdaptToTerminalHeight(t *testing.T) {
	if got := footerDetailRows(model{height: 20}); got != 0 {
		t.Fatalf("short footer detail rows=%d want=0", got)
	}
	if got := footerDetailRows(model{height: 28}); got != 1 {
		t.Fatalf("medium footer detail rows=%d want=1", got)
	}
	if got := footerDetailRows(model{height: 40}); got != 2 {
		t.Fatalf("tall footer detail rows=%d want=2", got)
	}
}

func TestHeroHeaderURLMouseBoundsMatchRenderedCapsule(t *testing.T) {
	m := model{width: 100, height: 40, stage: stageInput, cfg: config.Config{Username: "deathrashed", MouseEnabled: true}}
	lines := strings.Split(stripANSI(m.renderHeader()), "\n")
	borderLine := lines[1]

	// The interrupted top border is the Unicode regression fixture: the
	// capsule is built from box-drawing characters, so column math must be
	// display-width aware, never raw byte offsets.
	if !strings.HasPrefix(borderLine, "╭") || !strings.HasSuffix(borderLine, "╮") ||
		!strings.Contains(borderLine, "┤") || !strings.Contains(borderLine, "├") {
		t.Fatalf("hero top border is not interrupted by the URL capsule: %q", borderLine)
	}

	profile := headerURLDisplay(m.cfg.Username)
	index := strings.Index(borderLine, profile)
	if index < 0 {
		t.Fatalf("URL is not embedded in the interrupted top border: %q", borderLine)
	}
	left := displayWidth(borderLine[:index])
	right := displayWidth(borderLine[index+len(profile):])
	if absInt(left-right) > 1 {
		t.Fatalf("URL capsule is not centered on the top border: left=%d right=%d", left, right)
	}

	top := 1
	if !m.headerURLContains(left, top) {
		t.Fatalf("URL hover did not activate at its rendered capsule columns: left=%d top=%d", left, top)
	}
	for _, probe := range []struct{ x, y int }{
		{left - 1, top},                     // left border ┤ edge
		{left + displayWidth(profile), top}, // right border ├ edge
		{left, top - 1},                     // capsule cap row
		{left, top + 1},                     // capsule closure row
	} {
		if m.headerURLContains(probe.x, probe.y) {
			t.Fatalf("hero URL hitbox leaked outside the capsule at (%d,%d)", probe.x, probe.y)
		}
	}
}

func TestHeroHeaderShowsSearchedArtistContext(t *testing.T) {
	m := model{
		width:         100,
		height:        40,
		stage:         stageResults,
		modeChoice:    "manual",
		results:       []lastfm.Album{{Artist: "Death", Title: "Scream Bloody Gore"}},
		resultsCursor: 0,
		cfg:           config.Config{Username: "deathrashed"},
	}
	lines := strings.Split(stripANSI(m.renderHeader()), "\n")
	// Active workflow context (a searched artist) adds the contextual badge
	// beneath the stage badge, exactly like the original Scrobbler header.
	if got, want := len(lines), heroHeaderLines+2; got != want {
		t.Fatalf("searched-artist hero lines=%d want=%d", got, want)
	}
	joined := strings.ToUpper(strings.ReplaceAll(strings.Join(lines[len(lines)-3:], ""), " ", ""))
	if !strings.Contains(joined, "DEATH") {
		t.Fatalf("searched artist context not shown in badge: %q", strings.Join(lines, "\n"))
	}
	// The subheader keeps the static tagline; Now Playing is a dashboard-only
	// enhancement and must not overwrite the active context.
	subheader := lines[1+len(heroWordmarkLines())+2]
	if !strings.Contains(subheader, "SEARCH  •  SELECT  •  SCROBBLE") {
		t.Fatalf("searched-artist subheader lost the static tagline: %q", subheader)
	}
	for index, line := range lines {
		if got := displayWidth(line); got != m.appWidth() {
			t.Fatalf("searched-artist hero line %d width=%d want=%d: %q", index+1, got, m.appWidth(), line)
		}
	}
}

func TestHeroHeaderShowsDiscographyArtistContext(t *testing.T) {
	m := model{
		width:             100,
		height:            40,
		stage:             stageDiscographySelect,
		modeChoice:        "discography",
		discography:       []lastfm.Album{{Artist: "Oxygen Destroyer", Title: "Best Logic"}},
		discographyArtist: "Oxygen Destroyer",
		cfg:               config.Config{Username: "deathrashed"},
	}
	lines := strings.Split(stripANSI(m.renderHeader()), "\n")
	if got, want := len(lines), heroHeaderLines+2; got != want {
		t.Fatalf("discography hero lines=%d want=%d", got, want)
	}
	joined := strings.ToUpper(strings.ReplaceAll(strings.Join(lines[len(lines)-3:], ""), " ", ""))
	if !strings.Contains(joined, "OXYGENDESTROYER") {
		t.Fatalf("discography artist context not shown in badge: %q", strings.Join(lines, "\n"))
	}
	subheader := lines[1+len(heroWordmarkLines())+2]
	if !strings.Contains(subheader, "SEARCH  •  SELECT  •  SCROBBLE") {
		t.Fatalf("discography subheader lost the static tagline: %q", subheader)
	}
}

func TestHeroHeaderContextualArtistSuppressesNowPlaying(t *testing.T) {
	withActivity := model{
		width:         100,
		height:        40,
		stage:         stageResults,
		modeChoice:    "manual",
		results:       []lastfm.Album{{Artist: "Death", Title: "Scream Bloody Gore"}},
		resultsCursor: 0,
		cfg:           config.Config{Username: "deathrashed", NowPlaying: true},
		activityState: activityCurrent,
		activityTrack: lastfm.RecentTrack{Artist: "Hellwitch", Title: "Nosferatu", NowPlaying: true},
		activityFrame: 1,
	}
	lines := strings.Split(stripANSI(withActivity.renderHeader()), "\n")
	subheader := lines[1+len(heroWordmarkLines())+2]
	// Now Playing must NOT overwrite the active workflow context.
	if strings.Contains(subheader, "Hellwitch") || strings.Contains(subheader, "Nosferatu") {
		t.Fatalf("Now Playing overwrote searched artist context: %q", subheader)
	}
	joined := strings.ToUpper(strings.ReplaceAll(strings.Join(lines[len(lines)-3:], ""), " ", ""))
	if !strings.Contains(joined, "DEATH") {
		t.Fatalf("searched artist lost under Now Playing: %q", strings.Join(lines, "\n"))
	}
	// Returning to the dashboard restores Now Playing in the subheader.
	dash := withActivity
	dash.stage = stageInput
	dash.modeChoice = ""
	dash.results = nil
	dash.resultsCursor = 0
	dashLines := strings.Split(stripANSI(dash.renderHeader()), "\n")
	if got, want := len(dashLines), heroHeaderLines; got != want {
		t.Fatalf("dashboard-after-context hero lines=%d want=%d", got, want)
	}
	dashSubheader := dashLines[1+len(heroWordmarkLines())+2]
	if !strings.Contains(dashSubheader, "Hellwitch - Nosferatu") {
		t.Fatalf("returning to dashboard did not restore Now Playing: %q", dashSubheader)
	}
}

func TestHeroWordmarkTwoTonePreservesVisibleASCII(t *testing.T) {
	lines := heroWordmarkLines()
	if len(lines) != 6 {
		t.Fatalf("block wordmark rows=%d want=6", len(lines))
	}
	for _, line := range lines {
		styled := theme.HeroWordmarkLine(line)
		if got, want := stripANSI(styled), line; got != want {
			t.Fatalf("two-tone styling changed the visible ASCII:\n got %q\nwant %q", got, want)
		}
		if got, want := displayWidth(styled), displayWidth(line); got != want {
			t.Fatalf("two-tone styling changed display width: %d -> %d for %q", want, got, line)
		}
	}
	// The hero header is width-stable once its wordmark is painted.
	m := model{width: 100, height: 40, stage: stageInput, cfg: config.Config{Username: "deathrashed"}}
	for index, line := range strings.Split(stripANSI(m.renderHeader()), "\n") {
		if got := displayWidth(line); got != m.appWidth() {
			t.Fatalf("painted hero line %d width=%d want=%d", index+1, got, m.appWidth())
		}
	}
}

func TestHeroWordmarkTwoToneColorsByGlyph(t *testing.T) {
	previousProfile := lipgloss.ColorProfile()
	lipgloss.SetColorProfile(termenv.TrueColor)
	t.Cleanup(func() { lipgloss.SetColorProfile(previousProfile) })

	outlineGlyphs := []rune{'╔', '╗', '╚', '╝', '═', '║'}
	for _, line := range heroWordmarkLines() {
		styled := theme.HeroWordmarkLine(line)
		// Every full block cell is painted with the Last.fm red style.
		if got, want := strings.Count(styled, theme.HeroWordmarkRed.Render("█")), strings.Count(line, "█"); got != want {
			t.Fatalf("row %q has %d red blocks, want %d", line, got, want)
		}
		// Every outline/shadow cell is painted with the dim hint tone.
		for _, glyph := range outlineGlyphs {
			got := strings.Count(styled, theme.HeroWordmarkShadow.Render(string(glyph)))
			want := strings.Count(line, string(glyph))
			if got != want {
				t.Fatalf("row %q has %d dim-toned %q, want %d", line, got, string(glyph), want)
			}
		}
		// Spaces and the literal row are otherwise untouched.
		if strings.Contains(styled, theme.HeroWordmarkRed.Render(" ")) || strings.Contains(styled, theme.HeroWordmarkShadow.Render(" ")) {
			t.Fatalf("row %q paints spaces: %q", line, styled)
		}
		// The removed gradient palette must not survive anywhere.
		for _, stale := range []string{"#ff5964", "#ff9aa3", "#ffe8eb"} {
			if strings.Contains(styled, stale) {
				t.Fatalf("row %q still carries gradient color %s", line, stale)
			}
		}
	}
}
