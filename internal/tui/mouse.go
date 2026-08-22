package tui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/platform"
	"github.com/deathrashed/lastfm-scrobbler/internal/setup"
)

type mouseRegion struct {
	id                  string
	x, y, width, height int
	message             tea.KeyMsg
}

func (r mouseRegion) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.width && y >= r.y && y < r.y+r.height
}

type footerItem struct {
	id, key, label    string
	description       string
	descriptionAccent string
	group             string
	message           tea.KeyMsg
	interactive       bool
	tight             bool
}

func keyMessage(value string) tea.KeyMsg {
	switch value {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case "ctrl+c":
		return tea.KeyMsg{Type: tea.KeyCtrlC}
	case "space":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	}
	return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(value)}
}

func footerActionFor(id, key, label string, help ...string) footerItem {
	item := footerItem{id: id, key: key, label: label, message: keyMessage(key), interactive: true}
	if len(help) > 0 {
		item.description = help[0]
	}
	if len(help) > 1 {
		item.descriptionAccent = help[1]
	}
	return item
}
func footerStatic(key, label string) footerItem { return footerItem{key: key, label: label} }
func footerItemWidth(item footerItem) int       { return lipgloss.Width(item.key + item.label) }
func footerLineWidth(items []footerItem) int {
	width := 0
	for i, item := range items {
		if i > 0 && !item.tight {
			width += lipgloss.Width(" • ")
		}
		width += footerItemWidth(item)
	}
	return width
}

func footerGroupAction(id, key, group, description string, tight bool) footerItem {
	item := footerActionFor(id, key, "", description)
	item.group = group
	item.tight = tight
	return item
}

func footerGroupLabel(label, group string) footerItem {
	return footerItem{label: label, group: group, tight: true}
}

func footerSpec(m model) [][]footerItem {
	a := footerActionFor
	s := footerStatic
	enter := func(label, description string) footerItem {
		return a("footer:enter", "enter", label, description)
	}
	esc := func(label, description string) footerItem {
		return a("footer:esc", "esc", label, description)
	}
	help := func() footerItem {
		return a("footer:help", "?", " help", "open keyboard and mouse help")
	}

	switch m.stage {
	case stageInput:
		return [][]footerItem{
			{
				enter(" select", "open the highlighted mode"),
				{key: "→ ↑"},
				{label: " navigate ", tight: true},
				{key: "↓ ←", tight: true},
				a("footer:s", "s", " settings", "open Settings"),
			},
			{
				a("footer:i", "i", " info", "open the information guide"),
				a("footer:h", "h", " history", "open completed session history"),
				{key: "m d"},
				{label: " quick ", tight: true},
				{key: "f q", tight: true},
				a("footer:r", "r", " rerun", "open the latest completed session"),
				help(),
			},
		}
	case stageSetup:
		if m.setup.Page == setup.PageWelcome {
			return [][]footerItem{{enter(" continue", "continue with setup")}, {esc(" skip", "skip setup without writing changes"), help()}}
		}
		if m.setup.Page == setup.PageComplete {
			return [][]footerItem{{enter(" continue", "open the dashboard")}, {help()}}
		}
		return [][]footerItem{
			{s("↑ ↓", " navigate"), enter(" continue", "continue with the current setup step")},
			{esc(" previous", "return to the previous setup step"), help()},
		}
	case stageImportSource:
		if m.filePathFocused {
			return [][]footerItem{
				{s("tab", " source"), enter(" import", "load albums from the current path")},
				{a("footer:o", "o", " picker", platform.PickerDescription()), esc(" menu", "return to the main menu"), help()},
			}
		}
		return [][]footerItem{
			{
				enter(" path", "move focus to the import path"),
				{key: "→ ↑", tight: false},
				{label: " source ", tight: true},
				{key: "↓ ←", tight: true},
				esc(" menu", "return to the main menu"),
			},
			{a("footer:o", "o", " picker", platform.PickerDescription()), help()},
		}
	case stageSearch:
		description := "search Last.fm with the current input"
		if m.modeChoice == "file" {
			description = "load albums from the current path"
		}
		line := []footerItem{enter(" continue", description)}
		if m.modeChoice == "file" {
			line = append(line, a("footer:o", "o", " picker", platform.PickerDescription()))
		}
		return [][]footerItem{append(line,
			esc(" back", "return to the previous screen"),
			a("footer:quit", "ctrl+c", " quit", "quit Last.fm Scrobbler"),
		)}
	case stageResults:
		similar := a("footer:s", "s", " similar", "load a list of similar albums to", footerAlbumTitle(m))
		return [][]footerItem{{
			s("↑ ↓", " navigate"),
			enter(" select", "load the highlighted album"),
			similar,
			esc(" back", "return to search"),
		}}
	case stageDiscographySelect:
		if m.discographyFiltering {
			return [][]footerItem{{
				s("type", " filter"),
				enter(" apply", "apply the current album filter"),
				esc(" cancel", "cancel filtering and return to the album list"),
			}}
		}
		return [][]footerItem{
			{
				a("footer:space", "space", " check", "select or deselect the highlighted album"),
				a("footer:a", "a", " all", "select or deselect all visible albums"),
				a("footer:c", "c", " clean", "hide or show obvious duplicate editions"),
				a("footer:filter", "/", " filter", "filter albums by title"),
				a("footer:s", "s", " sort", "cycle the album sort order"),
			},
			{s("↑ ↓", " navigate"), enter(" continue", "load tracks for the selected albums")},
		}
	case stageTrackSelect:
		similar := a("footer:s", "s", " similar", "load a list of similar albums to", footerAlbumTitle(m))
		intervalDown := footerGroupAction("footer:interval-down", "-", "interval", "decrease the time between each track scrobble", false)
		intervalUp := footerGroupAction("footer:interval-up", "+", "interval", "increase the time between each track scrobble", true)
		navUp := footerGroupAction("footer:nav-up", "↑", "navigate", "move to the previous track", false)
		navDown := footerGroupAction("footer:nav-down", "↓", "navigate", "move to the next track", true)
		loopDown := footerGroupAction("footer:loop-down", "-", "loop", "decrease the number of times to scrobble this album", false)
		loopUp := footerGroupAction("footer:loop-up", "+", "loop", "increase the number of times to scrobble this album", true)
		return [][]footerItem{
			{
				a("footer:space", "space", " check", "select or deselect tracks to scrobble"),
				similar,
				a("footer:a", "a", " all", "select or deselect all tracks"),
				enter(" continue", "continue with current selections"),
			},
			{
				intervalDown,
				footerGroupLabel(" interval ", "interval"),
				intervalUp,
				navUp,
				footerGroupLabel(" navigate ", "navigate"),
				navDown,
				loopDown,
				footerGroupLabel(" loop ", "loop"),
				loopUp,
			},
		}
	case stagePreview:
		similar := a("footer:s", "s", " similar", "load a list of similar albums to", footerAlbumTitle(m))
		return [][]footerItem{{
			enter(" start", "start scrobbling the current queue"),
			a("footer:e", "e", " export", "export the current queue"),
			similar,
			esc(" edit", "return to track selection"),
			help(),
		}}
	case stageConfig:
		return m.settingsFooterSpec()
	case stageEnvPath:
		return [][]footerItem{{
			enter(" save", "save the credentials path"),
			a("footer:o", "o", " picker", "choose a credentials file"),
			esc(" account", "return to Account settings"),
		}}
	case stageScrobbling:
		if m.scrobblePaused || m.authState == authInvalid {
			return [][]footerItem{{
				a("footer:a", "a", " re-authenticate", "repair the expired session and resume"),
				esc(" cancel", "cancel the active scrobble session"),
				a("footer:q", "q", " quit + resume later", "quit and keep the session available for recovery"),
			}}
		}
		return [][]footerItem{{
			esc(" cancel", "cancel the active scrobble session"),
			a("footer:q", "q", " quit + resume later", "quit and keep the session available for recovery"),
		}}
	case stageDone:
		similar := a("footer:s", "s", " similar", "load a list of similar albums to", footerAlbumTitle(m))
		return [][]footerItem{
			{
				enter(" another", "start another scrobble task"),
				a("footer:r", "r", " edit + re-run", "edit this completed queue before running it again"),
				a("footer:R", "R", " exact re-run", "run this completed queue again unchanged"),
				a("footer:e", "e", " export", "export this completed session"),
			},
			{
				similar,
				a("footer:h", "h", " history", "open completed session history"),
				esc(" menu", "return to the main menu"),
				a("footer:q", "q", " quit", "quit Last.fm Scrobbler"),
			},
		}
	case stageHistory:
		if m.settingsFocus == settingsFocusSections {
			return [][]footerItem{{
				s("↑ ↓ ← →", " section"),
				enter(" content", "move focus into History"),
				a("footer:tab", "tab", " content", "move focus into History"),
				esc(" back", "return to the main menu"),
			}}
		}
		return [][]footerItem{
			{
				s("↑ ↓", " navigate"),
				enter(" edit + re-run", "edit the highlighted session before running it again"),
				a("footer:R", "R", " exact re-run", "run the highlighted session again unchanged"),
			},
			{
				a("footer:e", "e", " export", "export the highlighted session"),
				a("footer:d", "d", " delete", "delete the highlighted history entry"),
				a("footer:tab", "tab", " sections", "move focus to the Settings sections"),
				esc(" menu", "return to the main menu"),
			},
		}
	case stageLastSession:
		return [][]footerItem{
			{enter(" rerun", "run the latest completed session unchanged"), a("footer:e", "e", " edit first", "edit the latest completed session before running it again")},
			{esc(" back", "return to the main menu"), help()},
		}
	case stageRecovery:
		return [][]footerItem{{
			enter(" resume", "resume the unfinished session"),
			a("footer:r", "r", " restart", "restart the unfinished session from the beginning"),
			a("footer:d", "d", " discard", "discard the unfinished session"),
			a("footer:q", "q", " quit", "quit Last.fm Scrobbler"),
		}}
	case stageSimilarSelect:
		return [][]footerItem{{
			s("↑ ↓", " navigate"),
			enter(" load", "load the highlighted similar album"),
			esc(" back", "return to the previous screen"),
		}}
	case stageProfiles:
		if m.settingsFocus == settingsFocusSections {
			return [][]footerItem{{
				s("↑ ↓ ← →", " section"),
				enter(" content", "move focus into Profiles"),
				a("footer:tab", "tab", " content", "move focus into Profiles"),
				esc(" back", "return to the main menu"),
			}}
		}
		return [][]footerItem{
			{
				s("↑ ↓", " navigate"),
				enter(" load", "load the highlighted profile"),
				a("footer:n", "n", " new", "create a new profile"),
				a("footer:s", "s", " save", "save the current configuration to this profile"),
			},
			{
				a("footer:d", "d", " delete", "delete the highlighted profile"),
				a("footer:tab", "tab", " sections", "move focus to the Settings sections"),
				esc(" menu", "return to the main menu"),
			},
		}
	case stageProfileName:
		return [][]footerItem{{
			enter(" create", "create the profile with this name"),
			esc(" profiles", "return to Profiles"),
		}}
	case stageInfo:
		return [][]footerItem{{
			s("← →", " section"),
			esc(" back", "return to the main menu"),
			a("footer:help", "?", " quick help", "open keyboard and mouse help"),
			a("footer:q", "q", " quit", "quit Last.fm Scrobbler"),
		}}
	case stageConnectionTest:
		return [][]footerItem{{
			a("footer:r", "r", " re-test", "run the Last.fm connection test again"),
			esc(" tools", "return to Tools"),
			a("footer:q", "q", " quit", "quit Last.fm Scrobbler"),
		}}
	case stageDiagnostics:
		return [][]footerItem{{
			enter(" export", "build and export a redacted diagnostics bundle"),
			a("footer:o", "o", " open folder", "open the diagnostics output folder"),
			esc(" tools", "return to Tools"),
			a("footer:q", "q", " quit", "quit Last.fm Scrobbler"),
		}}
	case stageUpdateCheck:
		return [][]footerItem{{
			a("footer:r", "r", " check again", "check the configured release source again"),
			a("footer:o", "o", " open release", "open the available release"),
			esc(" tools", "return to Tools"),
			a("footer:q", "q", " quit", "quit Last.fm Scrobbler"),
		}}
	case stageCompletions:
		return [][]footerItem{{
			s("↑ ↓", " navigate"),
			enter(" install", "install completion for the selected shell"),
			a("footer:r", "r", " refresh", "refresh completion status"),
		}, {
			esc(" tools", "return to Tools"),
			a("footer:q", "q", " quit", "quit Last.fm Scrobbler"),
		}}
	case stageAuth:
		switch m.authState {
		case authPending:
			return [][]footerItem{{
				a("footer:o", "o", " browser", "reopen the Last.fm authorization page"),
				enter(" get key", "exchange the authorized token for a session key"),
				esc(" back", "return to the previous screen"),
			}}
		case authExchanging:
			return [][]footerItem{{
				esc(" back", "return to the previous screen"),
			}}
		case authValid:
			label := " return"
			switch m.authReturnTarget() {
			case stageScrobbling:
				label = " scrobbling"
			case stageConfig:
				label = " settings"
			}
			return [][]footerItem{{
				enter(label, "leave the auth screen"),
				a("footer:q", "q", " quit", "quit Last.fm Scrobbler"),
			}}
		default: // authUnknown, authInvalid, authFailed
			if m.authState == authFailed || m.authState == authInvalid {
				return [][]footerItem{{
					enter(" start again", "request a fresh authorization token"),
					esc(" back", "return to the previous screen"),
				}}
			}
			return [][]footerItem{{
				enter(" continue", "request an authorization token"),
				esc(" back", "return to the previous screen"),
			}}
		}
	}
	return nil
}

func footerAlbumTitle(m model) string {
	switch m.stage {
	case stageResults:
		if m.resultsCursor >= 0 && m.resultsCursor < len(m.results) {
			return m.results[m.resultsCursor].Title
		}
	case stageTrackSelect:
		refs := m.flattenedTracks()
		if len(refs) > 0 {
			return refs[minInt(maxInt(m.trackCursor, 0), len(refs)-1)].Album.Title
		}
		if len(m.selectedAlbums) > 0 {
			return m.selectedAlbums[minInt(maxInt(m.trackCursor, 0), len(m.selectedAlbums)-1)].Title
		}
		if strings.TrimSpace(m.selectedAlbum.Title) != "" {
			return m.selectedAlbum.Title
		}
	case stagePreview:
		albums := previewAlbums(m)
		if len(albums) > 0 {
			return albums[0].Title
		}
	case stageDone:
		if len(m.scrobbleQueue) > 0 {
			return m.scrobbleQueue[len(m.scrobbleQueue)-1].Album
		}
	}
	return ""
}

func (m model) bodyLineCount() int {
	body := m.renderBody()
	if body == "" {
		return 0
	}
	return strings.Count(body, "\n") + 1
}

func (m model) footerRegions() []mouseRegion {
	regions := []mouseRegion{}
	for lineIndex, items := range footerSpec(m) {
		x := (m.appWidth() - footerLineWidth(items)) / 2
		for index, item := range items {
			if index > 0 && !item.tight {
				x += lipgloss.Width(" • ")
			}
			if item.interactive {
				regions = append(regions, mouseRegion{id: item.id, x: x, y: m.headerHeight() + m.bodyLineCount() + lineIndex, width: footerItemWidth(item), height: 1, message: item.message})
			}
			x += footerItemWidth(item)
		}
	}
	return regions
}

func (m model) infoTabRegions() []mouseRegion {
	widths := []int{13, 23, 11, 19, 19}
	rows := [][]int{{0, 1, 2}, {3, 4}}
	regions := []mouseRegion{}
	for row, indexes := range rows {
		rowWidths := make([]int, len(indexes))
		for index, section := range indexes {
			rowWidths[index] = widths[section]
		}
		positions, _ := responsiveCardPositions(rowWidths, m.infoPanelWidth())
		for index, section := range indexes {
			regions = append(regions, mouseRegion{id: "info:tab:" + strconv.Itoa(section), x: m.infoX() + positions[index], y: m.headerHeight() + row*3, width: widths[section], height: 3})
		}
	}
	return regions
}

func (m model) helpCloseRegion() (mouseRegion, bool) {
	lines := strings.Split(renderHelpView(m), "\n")
	for lineIndex, line := range lines {
		plain := stripANSI(line)
		needle := "press ? or esc to close"
		at := strings.Index(plain, needle)
		if at < 0 {
			continue
		}
		return mouseRegion{
			id:     "help:close",
			x:      lipgloss.Width(plain[:at]),
			y:      m.headerHeight() + lineIndex,
			width:  lipgloss.Width(needle),
			height: 1,
		}, true
	}
	return mouseRegion{}, false
}

func (m model) screenRegions() []mouseRegion {
	if m.helpVisible {
		if region, ok := m.helpCloseRegion(); ok {
			return []mouseRegion{region}
		}
		return nil
	}
	regions := m.footerRegions()
	if m.inSettingsArea() {
		regions = append(regions, m.settingsSectionRegions()...)
	}
	if m.stage == stageInfo {
		regions = append(regions, m.infoTabRegions()...)
	}
	bodyY := m.headerHeight()
	workX := m.workX()
	switch m.stage {
	case stageInput:
		positions := dashboardCardPositions(m.panelWidth())
		for i, width := range []int{19, 25, 18} {
			regions = append(regions, mouseRegion{id: "dashboard:" + strconv.Itoa(i), x: workX + positions[i], y: bodyY, width: width, height: 3, message: keyMessage([]string{"m", "d", "f"}[i])})
		}
	case stageSetup:
		regions = append(regions, setupScreenRegions(m, bodyY)...)
	case stageCompletions:
		regions = append(regions, completionScreenRegions(m, bodyY)...)
	case stageAuth:
		// Exact regions matching the visible action buttons only. Buttons are
		// centered under the panel; compute their x from the group layout.
		bodyY := m.headerHeight()
		totalWidth := m.panelWidth()
		var boxes []struct {
			id    string
			width int
			key   string
		}
		switch m.authState {
		case authPending:
			boxes = append(boxes,
				struct {
					id    string
					width int
					key   string
				}{"auth:open", 20, "o"},
				struct {
					id    string
					width int
					key   string
				}{"auth:get-session", 22, "enter"},
			)
		case authFailed, authInvalid:
			boxes = append(boxes, struct {
				id    string
				width int
				key   string
			}{"auth:retry", 18, "enter"})
		case authValid:
			label, _ := m.authReturnAction()
			boxes = append(boxes, struct {
				id    string
				width int
				key   string
			}{"auth:return", maxInt(24, minInt(totalWidth-6, 30)), "enter"})
			_ = label
		}
		if len(boxes) > 0 {
			gap := 2
			groupWidth := 0
			for i, b := range boxes {
				if i > 0 {
					groupWidth += gap
				}
				groupWidth += b.width
			}
			left := maxInt(0, (totalWidth-groupWidth)/2)
			x := left + m.workX()
			y := bodyY // first button row is attached to the panel bottom border
			for _, b := range boxes {
				regions = append(regions, mouseRegion{id: b.id, x: x, y: y, width: b.width, height: 3, message: keyMessage(b.key)})
				x += b.width + gap
			}
		}
	case stageSearch:
		regions = append(regions, mouseRegion{id: "search:input", x: workX, y: bodyY, width: m.panelWidth(), height: 3})
	case stageImportSource:
		fileRegions := fileSourceRegions(bodyY, m.panelWidth())
		for index := range fileRegions {
			fileRegions[index].x += workX
		}
		regions = append(regions, fileRegions...)
	case stageConfig:
		regions = append(regions, m.settingsRowRegions()...)
	case stageResults:
		regions = append(regions, listRegion(workX, bodyY+3, len(m.results), m.resultsCursor, manualResultsMaxRows(m), m.panelWidth(), "results")...)
	case stageDiscographySelect:
		if !m.discographyFiltering {
			regions = append(regions,
				mouseRegion{id: "discography:sort", x: discographySortTabX(m), y: bodyY, width: discographySortTabWidth, height: 3, message: keyMessage("s")},
				mouseRegion{id: "discography:filter", x: discographyFilterTabX(m), y: bodyY, width: discographyFilterTabWidthFor(m), height: 3, message: keyMessage("/")},
				mouseRegion{id: "discography:clean", x: discographyCleanTabX(m), y: bodyY, width: discographyCleanTabWidth, height: 3, message: keyMessage("c")},
			)
		} else {
			regions = append(regions, mouseRegion{id: "discography:filter", x: discographyFilterTabX(m), y: bodyY, width: discographyFilterTabWidthFor(m), height: 3})
		}
		if discographyFilterExpanded(m) {
			regions = append(regions, mouseRegion{id: "discography:filter-input", x: discographyExpandedFilterX(m), y: bodyY + 3, width: discographyExpandedFilterWidthFor(m), height: discographyExpandedFilterLineCount(m), message: keyMessage("/")})
		}
		visible := m.discographyVisibleIndexes()
		maxRows := discographyChooserMaxRows(m)
		start := visibleStart(m.discographyCursor, len(visible), maxRows)
		end := minInt(len(visible), start+maxRows)
		firstRowY := bodyY + discographyChooserListRowOffset(m)
		for row, position := 0, start; position < end; position, row = position+1, row+1 {
			regions = append(regions, mouseRegion{id: "discography:" + strconv.Itoa(position), x: workX, y: firstRowY + row, width: m.panelWidth(), height: 1})
		}
	case stageTrackSelect:
		regions = append(regions, listRegion(workX, bodyY+trackSelectListTopOffset(m), len(m.flattenedTracks()), m.trackCursor, trackListMaxRows(m), m.panelWidth(), "tracks")...)
	case stageSimilarSelect:
		regions = append(regions, listRegion(workX, bodyY+3, len(m.similar), m.similarCursor, similarMaxRows(m), m.panelWidth(), "similar")...)
	case stageHistory:
		regions = append(regions, listRegion(workX, bodyY+settingsSectionContentStartY(), len(m.history), m.historyCursor, historyMaxRows(m), m.panelWidth(), "history")...)
	case stageProfiles:
		regions = append(regions, listRegion(workX, bodyY+settingsSectionContentStartY(), len(m.profiles), m.profileCursor, profilesMaxRows(m), m.panelWidth(), "profiles")...)
	case stageDiagnostics:
		regions = append(regions, mouseRegion{id: "diagnostics:action", x: workX, y: bodyY, width: m.panelWidth(), height: 6, message: keyMessage("enter")})
	case stageConnectionTest:
		if !m.connectionTesting {
			regions = append(regions, mouseRegion{id: "connection:action", x: workX, y: bodyY, width: m.panelWidth(), height: 6, message: keyMessage("r")})
		}
	case stageUpdateCheck:
		regions = append(regions, mouseRegion{id: "update:action", x: workX, y: bodyY, width: m.panelWidth(), height: 6, message: keyMessage("r")})
	case stageEnvPath:
		regions = append(regions, mouseRegion{id: "env:input", x: workX, y: bodyY, width: m.panelWidth(), height: 3})
	case stageProfileName:
		regions = append(regions, mouseRegion{id: "profile:input", x: workX, y: bodyY, width: m.panelWidth(), height: 3})
	}
	return regions
}

func listRegion(x, y, total, cursor, maxRows, width int, prefix string) []mouseRegion {
	if total == 0 {
		return nil
	}
	start := visibleStart(cursor, total, maxRows)
	end := minInt(total, start+maxRows)
	regions := []mouseRegion{}
	for row, index := 0, start; index < end; index, row = index+1, row+1 {
		regions = append(regions, mouseRegion{id: prefix + ":" + strconv.Itoa(index), x: x, y: y + row + 1, width: width, height: 1})
	}
	return regions
}
