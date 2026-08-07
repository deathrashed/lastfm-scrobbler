package tui

import (
	"fmt"
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
	"github.com/deathrashed/lastfm-scrobbler/internal/version"
)

func joinThreeLineBoxes(boxes []string, middleSeparator string) string {
	rows := make([][]string, len(boxes))
	maxRows := 0
	for i, box := range boxes {
		rows[i] = strings.Split(box, "\n")
		if len(rows[i]) > maxRows {
			maxRows = len(rows[i])
		}
	}
	var output []string
	for row := 0; row < maxRows; row++ {
		separator := " "
		if row == maxRows/2 {
			separator = middleSeparator
		}
		parts := make([]string, 0, len(boxes)*2-1)
		for i := range rows {
			if i > 0 {
				parts = append(parts, separator)
			}
			if row < len(rows[i]) {
				parts = append(parts, rows[i][row])
			}
		}
		output = append(output, strings.Join(parts, ""))
	}
	return strings.Join(output, "\n")
}

func renderInputView(m model) string {
	labels := []string{"M A N U A L", "D I S C O G R A P H Y", "F I L E"}
	widths := []int{19, 25, 18}
	boxes := make([]string, len(labels))
	for i, label := range labels {
		boxes[i] = renderExactBox(label, widths[i], i == m.modeIndex)
	}
	row := joinThreeLineBoxes(boxes, theme.SepStyle.Render("•"))
	descriptions := []string{
		"enter artist - album manually",
		"scrobble an artists top albums",
		"load albums from a list, playlist, or music folder",
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		centerToHeader(row),
		centerToHeader(theme.MutedStyle.Render(descriptions[m.modeIndex])),
		"",
	)
}

func renderImportSourceView(m model) string {
	labels := []string{"L I S T   F I L E", "P L A Y L I S T", "A L B U M   F O L D E R", "A R T I S T   F O L D E R"}
	widths := []int{22, 19, 27, 29}
	firstRow := joinThreeLineBoxes([]string{
		renderExactBox(labels[0], widths[0], m.importSourceIndex == 0),
		renderExactBox(labels[1], widths[1], m.importSourceIndex == 1),
	}, theme.SepStyle.Render("•"))
	secondRow := joinThreeLineBoxes([]string{
		renderExactBox(labels[2], widths[2], m.importSourceIndex == 2),
		renderExactBox(labels[3], widths[3], m.importSourceIndex == 3),
	}, theme.SepStyle.Render("•"))
	descriptions := []string{
		"TXT, CSV, TSV, or JSON containing Artist - Album entries",
		"M3U or M3U8 playlist",
		"read one album folder containing audio files",
		"scan album folders inside one artist folder",
	}
	formats := renderInfoBox("IMPORTS", "TXT  •  CSV  •  TSV  •  JSON  •  M3U  •  M3U8  •  folders", "", 65, false)
	return lipgloss.JoinVertical(lipgloss.Left,
		centerToHeader(firstRow),
		centerToHeader(secondRow),
		centerToHeader(theme.MutedStyle.Render(descriptions[m.importSourceIndex])),
		"",
		centerToHeader(formats),
		"",
	)
}

func renderSearchView(m model) string {
	label, placeholder := "SEARCH", "Artist - Album..."
	switch m.modeChoice {
	case "file":
		label, placeholder = "PATH", "/path/to/list, playlist, or music folder"
	case "discography":
		label, placeholder = "ARTIST", "Artist name..."
	}
	box := renderTextBox(label, m.searchInput.View(), placeholder, 65, m.searchInput.Focused())
	lines := []string{centerToHeader(box), ""}
	if m.searching {
		lines = append(lines, centerToHeader(m.spinner.View()+" "+theme.MutedStyle.Render("Loading from Last.fm…")), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderDiscographySelectView(m model) string {
	artist := strings.TrimSpace(m.discographyArtist)
	if artist == "" && len(m.discography) > 0 {
		artist = m.discography[0].Artist
	}
	visible := m.discographyVisibleIndexes()
	artistBox := renderInfoBox("ARTIST", artist, "", 65, false)
	selectedBox := renderSelectedBadge(len(m.discographySelected), len(visible))
	listBox := renderDiscographyList(m, visible)
	var filterLine string
	if m.discographyFiltering {
		filterLine = renderTextBox("FILTER", m.filterInput.View(), "type album title…", 65, true)
	} else {
		clean := "off"
		if m.discographyClean {
			clean = "on"
		}
		sortName := []string{"Last.fm", "A–Z", "Z–A"}[m.discographySort]
		filterValue := m.discographyFilter
		if filterValue == "" {
			filterValue = "none"
		}
		filterLine = renderInfoBox("VIEW", fmt.Sprintf("filter %s   •   clean %s   •   sort %s", filterValue, clean, sortName), fmt.Sprintf("%d / %d", len(visible), len(m.discography)), 65, false)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		centerToHeader(artistBox),
		centerToHeader(selectedBox),
		centerToHeader(filterLine),
		centerToHeader(listBox),
		"",
	)
}

func renderDiscographyList(m model, visibleIndexes []int) string {
	const totalWidth = 65
	maxRows := 13
	if m.height > 0 {
		maxRows = maxInt(6, minInt(16, m.height-20))
	}
	innerWidth := totalWidth - 2
	contentWidth := innerWidth - 2
	total := len(visibleIndexes)
	if total == 0 {
		return renderPanelBox([]string{theme.MutedStyle.Render("No albums match the current view.")}, totalWidth, theme.BorderStyle)
	}
	cursor := minInt(m.discographyCursor, total-1)
	start := visibleStart(cursor, total, maxRows)
	end := minInt(total, start+maxRows)
	visible := end - start
	thumb := scrollbarThumb(start, total, visible)
	rows := make([]string, 0, visible)
	for row, position := 0, start; position < end; position, row = position+1, row+1 {
		original := visibleIndexes[position]
		cursorMark := "  "
		if position == cursor {
			cursorMark = theme.PromptStyle.Render("▸ ")
		}
		check := theme.BorderStyle.Render("☐")
		if m.discographySelected[original] {
			check = theme.KeyStyle.Render("☑")
		}
		title := strings.TrimSpace(m.discography[original].Title)
		if title == "" {
			title = "(untitled album)"
		}
		title = truncateToWidth(title, contentWidth-7)
		scroll := " "
		if total > visible {
			scroll = theme.MutedStyle.Render("│")
			if row == thumb {
				scroll = theme.KeyStyle.Render("█")
			}
		}
		left := cursorMark + check + " " + theme.AlbumStyle.Render(title)
		gap := contentWidth - lipgloss.Width(left) - 1
		if gap < 1 {
			gap = 1
		}
		rows = append(rows, left+strings.Repeat(" ", gap)+scroll)
	}
	return renderPanelBox(rows, totalWidth, theme.BorderStyle)
}

func visibleStart(cursor, total, maxRows int) int {
	if total <= maxRows {
		return 0
	}
	start := cursor - maxRows/2
	if start < 0 {
		start = 0
	}
	if start > total-maxRows {
		start = total - maxRows
	}
	return start
}

func scrollbarThumb(start, total, visible int) int {
	if total <= visible || visible <= 1 {
		return 0
	}
	maxStart := total - visible
	return int(math.Round(float64(start) / float64(maxStart) * float64(visible-1)))
}

func renderConfigView(m model) string {
	firstLabels := []string{"L O O P", "I N T E R V A L", "U S E R N A M E", "A P I"}
	firstWidths := []int{11, 19, 19, 9}
	firstBoxes := make([]string, len(firstLabels))
	for i, label := range firstLabels {
		firstBoxes[i] = renderExactBox(label, firstWidths[i], i == m.configIndex)
	}
	secondLabels := []string{"A D V A N C E D", "H I S T O R Y", "P R O F I L E S"}
	secondWidths := []int{19, 17, 19}
	secondBoxes := make([]string, len(secondLabels))
	for i, label := range secondLabels {
		secondBoxes[i] = renderExactBox(label, secondWidths[i], i+4 == m.configIndex)
	}
	firstRow := joinThreeLineBoxes(firstBoxes, theme.SepStyle.Render("•"))
	secondRow := joinThreeLineBoxes(secondBoxes, theme.SepStyle.Render("•"))
	fieldBox := renderConfigFields(m)
	hints := []string{
		"how many times to scrobble each album",
		"interval between scrobbles (seconds)",
		"Last.fm account details; Tab changes field",
		"Last.fm application credentials; Tab changes field",
		"additional settings",
		"review, re-run, delete, and export completed sessions",
		"save and switch named Last.fm configurations",
	}
	lines := []string{
		centerToHeader(firstRow),
		centerToHeader(secondRow),
		centerToHeader(fieldBox),
		centerToHeader(theme.MutedStyle.Render(hints[m.configIndex])),
		"",
	}
	if m.configStatus != "" {
		lines = append(lines, centerToHeader(theme.SuccessStyle.Render(m.configStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderConfigFields(m model) string {
	switch m.configIndex {
	case 0:
		return renderTextBox("LOOP", renderConfigEditValue(m, false), "1", 65, true)
	case 1:
		return renderTextBox("INTERVAL", renderConfigEditValue(m, false), "2", 65, true)
	case 2:
		username := m.cfg.Username
		password := maskValue(m.cfg.Password)
		if m.configFieldIndex == 0 {
			username = renderConfigEditValue(m, false)
		} else {
			password = renderConfigEditValue(m, true)
		}
		return renderMultiFieldBox([]configRenderField{{"LASTFM USERNAME", username, m.configFieldIndex == 0}, {"LASTFM PASSWORD", password, m.configFieldIndex == 1}}, 65)
	case 3:
		apiKey := m.cfg.APIKey
		apiSecret := maskValue(m.cfg.APISecret)
		if m.configFieldIndex == 0 {
			apiKey = renderConfigEditValue(m, false)
		} else {
			apiSecret = renderConfigEditValue(m, true)
		}
		return renderMultiFieldBox([]configRenderField{{"API KEY", apiKey, m.configFieldIndex == 0}, {"API SECRET", apiSecret, m.configFieldIndex == 1}}, 65)
	case 4:
		return renderInfoBox("ADVANCED", "additional settings", "ENTER", 65, true)
	case 5:
		return renderInfoBox("HISTORY", fmt.Sprintf("%d saved session(s)", len(m.history)), "ENTER", 65, true)
	case 6:
		return renderInfoBox("PROFILE", fmt.Sprintf("%s   •   %d available", m.cfg.Profile, len(m.profiles)), "ENTER", 65, true)
	}
	return ""
}

func renderConfigEditValue(m model, secret bool) string {
	value := m.configInput.Value()
	if secret {
		value = strings.Repeat("•", len([]rune(value)))
	}
	// Config panels use the raw value rather than textinput.View(). The latter
	// pads to its viewport width and was the source of the detached right
	// borders shown with long usernames, passwords, and API values.
	if m.configInput.Focused() {
		value += theme.KeyStyle.Render("▏")
	}
	return value
}

type configRenderField struct {
	Label, Value string
	Active       bool
}

func renderMultiFieldBox(fields []configRenderField, totalWidth int) string {
	contentWidth := totalWidth - 4
	lines := make([]string, 0, len(fields))
	active := false
	for _, field := range fields {
		prefix := theme.KeyStyle.Render(field.Label + " ❯ ")
		available := maxInt(1, contentWidth-lipgloss.Width(prefix))
		value := fitStyled(field.Value, available)
		line := fitStyled(prefix+value, contentWidth)
		if field.Active {
			active = true
			line = theme.ActiveRowStyle.Render(line)
		}
		lines = append(lines, line)
	}
	border := theme.BorderStyle
	if active {
		border = theme.InnerBorderStyle
	}
	return renderPanelBox(lines, totalWidth, border)
}

func maskValue(value string) string {
	if value == "" {
		return ""
	}
	return strings.Repeat("•", minInt(16, len([]rune(value))))
}

func renderAdvancedConfigView(m model) string {
	maxRows := 11
	if m.height > 0 {
		maxRows = maxInt(7, minInt(13, m.height-19))
	}
	cursor := minInt(maxInt(m.advancedIndex, 0), len(advancedLabels)-1)
	start := visibleStart(cursor, len(advancedLabels), maxRows)
	end := minInt(len(advancedLabels), start+maxRows)
	visible := end - start
	thumb := scrollbarThumb(start, len(advancedLabels), visible)
	rows := make([]string, 0, visible)
	for row, index := 0, start; index < end; index, row = index+1, row+1 {
		label := advancedLabels[index]
		cursorMark := "  "
		if index == cursor {
			cursorMark = theme.PromptStyle.Render("▸ ")
		}
		value := advancedValue(m, index)
		left := cursorMark + theme.KeyStyle.Render(label+" ❯ ")
		right := theme.AlbumStyle.Render(value)
		scroll := " "
		if len(advancedLabels) > visible {
			scroll = theme.MutedStyle.Render("│")
			if row == thumb {
				scroll = theme.KeyStyle.Render("█")
			}
		}
		gap := 58 - lipgloss.Width(left) - lipgloss.Width(right)
		if gap < 1 {
			gap = 1
		}
		rows = append(rows, left+strings.Repeat(" ", gap)+right+scroll)
	}
	list := renderPanelBox(rows, 65, theme.BorderStyle)
	descriptions := []string{
		"automatic retries after a failed Last.fm request",
		"pause before each retry",
		"skip matching recent scrobbles; 0 disables it",
		"send a macOS notification when a session finishes",
		"use a shorter header on small screens",
		"hide obvious reissues, demos and duplicate editions",
		"folder used for JSON, CSV, TXT, M3U8, and diagnostics",
		"auto, file, environment, or keychain",
		"enable clickable controls and mouse-wheel navigation",
		"GitHub release API or custom JSON update endpoint",
		"verify API lookup and authentication readiness",
		"export logs and redacted configuration for troubleshooting",
		"compare this build with the configured release endpoint",
	}
	var action string
	if advancedEditable(m.advancedIndex) {
		action = renderTextBox(advancedLabels[m.advancedIndex], m.configInput.View(), "", 65, true)
	} else {
		action = renderInfoBox(advancedLabels[m.advancedIndex], descriptions[m.advancedIndex], "ENTER", 65, true)
	}
	lines := []string{centerToHeader(list), centerToHeader(action), centerToHeader(theme.MutedStyle.Render(descriptions[m.advancedIndex])), ""}
	if m.configStatus != "" {
		lines = append(lines, centerToHeader(theme.SuccessStyle.Render(m.configStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func advancedValue(m model, index int) string {
	switch index {
	case 0:
		return fmt.Sprintf("%d", m.cfg.RetryCount)
	case 1:
		return m.cfg.RetryDelay.String()
	case 2:
		return m.cfg.DuplicateGuard.String()
	case 3:
		return boolWord(m.cfg.Notify)
	case 4:
		return boolWord(m.cfg.CompactHeader)
	case 5:
		return boolWord(m.cfg.CleanDiscography)
	case 6:
		return truncateToWidth(m.cfg.ExportDir, 26)
	case 7:
		return m.cfg.CredentialSource
	case 8:
		return boolWord(m.cfg.MouseEnabled)
	case 9:
		if strings.TrimSpace(m.cfg.UpdateURL) == "" {
			return "not configured"
		}
		return truncateToWidth(m.cfg.UpdateURL, 24)
	case 10, 11, 12:
		return "ENTER"
	}
	return ""
}

func renderConnectionTestView(m model) string {
	if m.connectionTesting {
		box := renderInfoBox("CONNECTION", "Testing Last.fm lookup and authentication readiness", m.spinner.View(), 65, true)
		return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(box), "")
	}
	rows := make([]string, 0, len(m.connectionReport.Items))
	for _, item := range m.connectionReport.Items {
		status := theme.ErrorStyle.Render("FAIL")
		if item.Skipped {
			status = theme.MutedStyle.Render("SKIP")
		} else if item.OK {
			status = theme.SuccessStyle.Render("OK")
		}
		left := theme.KeyStyle.Render(item.Label + " ❯")
		detail := truncateToWidth(item.Detail, 38)
		right := status + "  " + theme.MutedStyle.Render(detail)
		gap := 59 - lipgloss.Width(left) - lipgloss.Width(right)
		if gap < 1 {
			gap = 1
		}
		rows = append(rows, left+strings.Repeat(" ", gap)+right)
	}
	if len(rows) == 0 {
		rows = []string{theme.MutedStyle.Render("No test results yet.")}
	}
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(renderPanelBox(rows, 65, theme.BorderStyle)), "")
}

func renderDiagnosticsView(m model) string {
	var rows []string
	switch {
	case m.diagnosticsBusy:
		rows = []string{theme.KeyStyle.Render("EXPORTING ❯ ") + m.spinner.View() + " " + theme.MutedStyle.Render("building redacted support bundle")}
	case m.diagnosticsPath != "":
		rows = []string{
			statLine("STATUS", "complete", 59),
			statLine("PATH", truncateToWidth(m.diagnosticsPath, 44), 59),
			statLine("SECRETS", "redacted", 59),
		}
	default:
		rows = []string{
			statLine("CONTENTS", "logs + redacted config + history summary", 59),
			statLine("OUTPUT", truncateToWidth(m.cfg.ExportDir, 43), 59),
			statLine("SECRETS", "passwords, API secret and session key excluded", 59),
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(renderPanelBox(rows, 65, theme.BorderStyle)), "")
}

func renderUpdateCheckView(m model) string {
	if m.updateChecking {
		box := renderInfoBox("UPDATE", "Checking configured release source", m.spinner.View(), 65, true)
		return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(box), "")
	}
	if m.updateResult.Latest == "" {
		box := renderInfoBox("UPDATE", "No result yet", versionText(), 65, false)
		return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(box), "")
	}
	status := "UP TO DATE"
	if m.updateResult.Available {
		status = "AVAILABLE"
	}
	rows := []string{
		statLine("CURRENT", m.updateResult.Current, 59),
		statLine("LATEST", m.updateResult.Latest, 59),
		statLine("STATUS", status, 59),
	}
	if m.updateResult.URL != "" {
		rows = append(rows, statLine("RELEASE", truncateToWidth(m.updateResult.URL, 43), 59))
	}
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(renderPanelBox(rows, 65, theme.BorderStyle)), "")
}

func versionText() string { return version.Version }

func renderEnvPathView(m model) string {
	box := renderTextBox("CREDENTIALS PATH", m.envInput.View(), "~/.config/lastfm-scrobbler/.env", 65, m.envInput.Focused())
	lines := []string{centerToHeader(box), centerToHeader(theme.MutedStyle.Render("enter path to credentials file")), ""}
	if m.envStatus != "" {
		lines = append(lines, centerToHeader(theme.SuccessStyle.Render(m.envStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderSimilarSelectView(m model) string {
	artistBox := renderInfoBox("BASED ON", m.similarArtist, "", 65, false)
	list := renderSimpleAlbumList(m.similar, m.similarCursor, 13)
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(artistBox), centerToHeader(list), "")
}

func renderInfoView(m model) string {
	labels := []string{"M O D E S", "A U T O M A T I O N", "D A T A", "C U R A T I O N", "I M P O R T S"}
	widths := []int{13, 23, 11, 19, 19}
	boxes := make([]string, len(labels))
	for index, label := range labels {
		boxes[index] = renderExactBox(label, widths[index], index == m.infoIndex)
	}
	firstRow := joinThreeLineBoxes(boxes[:3], theme.SepStyle.Render("•"))
	secondRow := joinThreeLineBoxes(boxes[3:], theme.SepStyle.Render("•"))
	sections := [][]string{
		{
			helpRow("MANUAL", "enter Artist - Album or search by name"),
			helpRow("DISCOGRAPHY", "filter, clean, sort, and select albums"),
			helpRow("FILE", "lists, playlists, album folders, artist folders"),
			helpRow("SIMILAR", "press S from results, preview, or completion"),
		},
		{
			helpRow("HEADLESS CLI", "manual, file, and discography commands"),
			helpRow("CONNECTION", "test API lookup and authentication"),
			helpRow("COMPLETION", "zsh, bash, and fish shell completion"),
			helpRow("MOUSE", "click tabs; use the wheel in long lists"),
			helpRow("UPDATES", "custom or GitHub release endpoint"),
		},
		{
			helpRow("HISTORY", "re-run, export, or delete sessions"),
			helpRow("RECOVERY", "resume an interrupted queue at startup"),
			helpRow("PROFILES", "switch named account/settings profiles"),
			helpRow("DIAGNOSTICS", "redacted logs and configuration bundle"),
			helpRow("STORAGE", "~/.config/lastfm-scrobbler"),
		},
		{
			helpRow("SPACE / A", "select tracks or all/none"),
			helpRow("- / +", "change the global loop count"),
			helpRow("[ / ]", "change the current album loop"),
			helpRow("RERUN", "edit a saved queue before starting it again"),
			helpRow("E", "export JSON, CSV, TXT, and M3U8"),
		},
		{
			helpRow("LIST FILE", "TXT, CSV, TSV, and JSON album lists"),
			helpRow("PLAYLIST", "M3U and M3U8 playlists"),
			helpRow("ALBUM FOLDER", "infer Artist/Album from one folder"),
			helpRow("ARTIST FOLDER", "scan album folders beneath an artist"),
			helpRow("O", "open the native macOS file/folder picker"),
		},
	}
	box := renderPanelBox(sections[m.infoIndex], 65, theme.BorderStyle)
	return lipgloss.JoinVertical(lipgloss.Left,
		centerToHeader(firstRow),
		centerToHeader(secondRow),
		centerToHeader(box),
		"",
	)
}

func renderProfilesView(m model) string {
	rows := make([]string, 0, len(m.profiles))
	for index, profile := range m.profiles {
		cursor := "  "
		if index == m.profileCursor {
			cursor = theme.PromptStyle.Render("▸ ")
		}
		active := ""
		if strings.EqualFold(profile, m.cfg.Profile) {
			active = theme.KeyStyle.Render("  ACTIVE")
		}
		line := cursor + theme.AlbumStyle.Render(profile) + active
		rows = append(rows, line)
	}
	if len(rows) == 0 {
		rows = []string{theme.MutedStyle.Render("No profiles saved.")}
	}
	box := renderPanelBox(rows, 65, theme.BorderStyle)
	lines := []string{centerToHeader(box), ""}
	if m.profileStatus != "" {
		lines = append(lines, centerToHeader(theme.SuccessStyle.Render(m.profileStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderProfileNameView(m model) string {
	box := renderTextBox("PROFILE NAME", m.profileInput.View(), "personal", 65, m.profileInput.Focused())
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(box), centerToHeader(theme.MutedStyle.Render("letters, numbers, hyphens and underscores")), "")
}
