package tui

import (
	"math"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/platform"
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
		boxes[i] = renderDashboardBox(label, widths[i], i == m.modeIndex)
		if m.hoverRegion == "dashboard:"+strconv.Itoa(i) {
			boxes[i] = renderExactBoxWithMnemonicHover(label, widths[i], i == m.modeIndex, true, true)
		}
	}
	row := joinThreeLineBoxes(boxes, theme.SepStyle.Render("•"))
	descriptions := []string{
		"search by artist, album, or both",
		"scrobble multiple albums at once",
		"load albums from a list, playlist, or music folder",
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		centerToHeader(row),
		centerToHeader(theme.MutedStyle.Render(descriptions[m.modeIndex])),
		"",
	)
}

func renderImportSourceView(m model) string {
	labels := make([]string, len(fileSourceSpecs))
	widths := make([]int, len(fileSourceSpecs))
	for index, spec := range fileSourceSpecs {
		labels[index], widths[index] = spec.label, spec.width
	}
	firstRow := joinThreeLineBoxes([]string{
		renderChoiceBox(labels[0], widths[0], m.importSourceIndex == 0, m.hoverRegion == "import:0"),
		renderChoiceBox(labels[1], widths[1], m.importSourceIndex == 1, m.hoverRegion == "import:1"),
	}, theme.SepStyle.Render("•"))
	secondRow := joinThreeLineBoxes([]string{
		renderChoiceBox(labels[2], widths[2], m.importSourceIndex == 2, m.hoverRegion == "import:2"),
		renderChoiceBox(labels[3], widths[3], m.importSourceIndex == 3, m.hoverRegion == "import:3"),
	}, theme.SepStyle.Render("•"))
	spec := fileSourceSpecFor(m.importSourceIndex)
	lines := []string{
		centerToHeader(firstRow),
		centerToHeader(secondRow),
		centerToHeader(theme.MutedStyle.Render(spec.description)),
		"",
		centerToHeader(renderFilePathView(m)),
	}
	if m.searching {
		lines = append(lines, "", centerToHeader(m.spinner.View()+" "+theme.MutedStyle.Render("Loading imports…")))
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		lines...,
	)
}

func renderSearchView(m model) string {
	label, placeholder := "SEARCH", "Artist, Album, or Both..."
	switch m.modeChoice {
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

const (
	discographySortTabWidth        = 15
	discographyFilterTabWidth      = 29
	discographyCleanTabWidth       = 15
	discographyExpandedFilterWidth = 61
	discographyFilterContentWidth  = 57
)

func renderDiscographySelectView(m model) string {
	visible := m.discographyVisibleIndexes()
	chooser := renderDiscographyChooser(m, visible)
	return lipgloss.JoinVertical(lipgloss.Left,
		centerToHeader(chooser),
		"",
	)
}

func renderDiscographyChooser(m model, visibleIndexes []int) string {
	border := theme.BorderStyle
	sortName := []string{"LFM", "A-Z", "Z-A"}[m.discographySort]
	clean := "OFF"
	if m.discographyClean {
		clean = "ON"
	}
	filterValue := strings.TrimSpace(m.discographyFilter)
	expanded := discographyFilterExpanded(m)
	compactFilter := filterValue
	if expanded {
		compactFilter = ""
	}

	sortTab := renderDiscographyControlTab("SORT", sortName, discographySortTabWidth, false, m.hoverRegion == "discography:sort", false)
	filterHovered := m.hoverRegion == "discography:filter" || m.hoverRegion == "discography:filter-input"
	filterTab := renderDiscographyControlTab("FILTER", compactFilter, discographyFilterTabWidth, m.discographyFiltering, filterHovered, expanded)
	cleanTab := renderDiscographyControlTab("CLEAN", clean, discographyCleanTabWidth, false, m.hoverRegion == "discography:clean", false)

	lines := []string{
		"  " + sortTab[0] + " " + filterTab[0] + " " + cleanTab[0] + "  ",
		border.Render("╭─") + sortTab[1] + border.Render("─") + filterTab[1] + border.Render("─") + cleanTab[1] + border.Render("─╮"),
		border.Render("│ ") + sortTab[2] + " " + filterTab[2] + " " + cleanTab[2] + border.Render(" │"),
	}
	if expanded {
		lines = append(lines, renderDiscographyExpandedFilter(m)...)
	}

	maxRows := discographyChooserMaxRows(m)
	rows := discographyListRows(m, visibleIndexes, maxRows)
	if len(rows) == 0 {
		rows = []string{theme.MutedStyle.Render("No albums match the current view.")}
	}
	for _, row := range rows {
		lines = append(lines, border.Render("│")+" "+padRight(fitStyled(row, 61), 61)+" "+border.Render("│"))
	}

	lines = append(lines, renderDiscographyCountAttachment(len(visibleIndexes), len(m.discographySelected))...)
	return strings.Join(lines, "\n")
}

func renderDiscographyControlTab(label, value string, totalWidth int, active, hovered, expanded bool) [3]string {
	border := theme.BorderStyle
	innerWidth := maxInt(1, totalWidth-2)
	labelStyle := theme.SummaryLabelStyle
	arrowStyle := theme.SummaryArrowStyle
	if active || hovered {
		labelStyle = theme.AccentTextStyle
		arrowStyle = theme.FocusedRowArrowStyle
	}
	content := labelStyle.Render(label+" ") + arrowStyle.Render("❯")
	if strings.TrimSpace(value) != "" {
		content += " " + theme.PrimaryTextStyle.Render(value)
	}
	content = centerText(fitStyled(content, innerWidth), innerWidth)
	bottom := border.Render("╰" + strings.Repeat("─", innerWidth) + "╯")
	if expanded {
		left := innerWidth / 2
		right := innerWidth - left - 1
		bottom = border.Render("╰" + strings.Repeat("─", left) + "┬" + strings.Repeat("─", right) + "╯")
	}
	return [3]string{
		border.Render("╭" + strings.Repeat("─", innerWidth) + "╮"),
		border.Render("┤") + content + border.Render("├"),
		bottom,
	}
}

func discographyCompactFilterValueWidth() int {
	return discographyFilterTabWidth - 2 - lipgloss.Width("FILTER ❯ ")
}

func discographyFilterExpanded(m model) bool {
	if m.discographyFiltering {
		return true
	}
	return lipgloss.Width(strings.TrimSpace(m.discographyFilter)) > discographyCompactFilterValueWidth()
}

func renderDiscographyExpandedFilter(m model) []string {
	border := theme.BorderStyle
	innerTop := border.Render("╭" + strings.Repeat("─", 29) + "┴" + strings.Repeat("─", 29) + "╮")
	lines := []string{border.Render("│") + " " + innerTop + " " + border.Render("│")}

	contentLines := discographyExpandedFilterContent(m)
	for _, content := range contentLines {
		content = fitStyled(content, discographyFilterContentWidth)
		lines = append(lines,
			border.Render("│")+" "+border.Render("│")+" "+padRight(content, discographyFilterContentWidth)+" "+border.Render("│")+" "+border.Render("│"),
		)
	}
	innerBottom := border.Render("╰" + strings.Repeat("─", discographyExpandedFilterWidth-2) + "╯")
	lines = append(lines, border.Render("│")+" "+innerBottom+" "+border.Render("│"))
	return lines
}

func discographyExpandedFilterContent(m model) []string {
	if m.discographyFiltering {
		input := m.filterInput
		input.Width = discographyFilterContentWidth
		shown := input.View()
		if strings.TrimSpace(stripANSI(shown)) == "" {
			shown = theme.MutedStyle.Render("type album title…")
		}
		return []string{shown}
	}

	query := strings.TrimSpace(m.discographyFilter)
	if query == "" {
		return []string{theme.MutedStyle.Render("type album title…")}
	}
	wrapped := wrapWords(query, discographyFilterContentWidth)
	if len(wrapped) == 0 {
		wrapped = []string{query}
	}
	if len(wrapped) > 2 {
		wrapped = wrapped[:2]
		wrapped[1] = truncateToWidth(wrapped[1]+"…", discographyFilterContentWidth)
	}
	for index := range wrapped {
		wrapped[index] = theme.PrimaryTextStyle.Render(wrapped[index])
	}
	return wrapped
}

func discographyExpandedFilterLineCount(m model) int {
	if !discographyFilterExpanded(m) {
		return 0
	}
	return len(discographyExpandedFilterContent(m)) + 2
}

func discographyChooserListRowOffset(m model) int {
	return 3 + discographyExpandedFilterLineCount(m)
}

func discographyChooserMaxRows(m model) int {
	rows := discographyMaxRows(m)
	if extra := discographyExpandedFilterLineCount(m); extra > 0 {
		rows = maxInt(5, rows-extra)
	}
	return rows
}

func renderDiscographyCountAttachment(results, selected int) []string {
	border := theme.BorderStyle
	const (
		totalWidth = 65
		badgeWidth = 18
		gap        = 1
	)
	innerWidth := totalWidth - 2
	groupWidth := badgeWidth*2 + gap
	leftDash := (innerWidth - groupWidth) / 2
	rightDash := innerWidth - groupWidth - leftDash

	leftContent := renderCountContent("RESULTS", results)
	rightContent := renderCountContent("SELECTED", selected)
	leftTop := border.Render("╭" + strings.Repeat("─", badgeWidth-2) + "╮")
	rightTop := border.Render("╭" + strings.Repeat("─", badgeWidth-2) + "╮")
	topGroup := leftTop + " " + rightTop
	line1 := border.Render("│") + strings.Repeat(" ", leftDash) + topGroup + strings.Repeat(" ", rightDash) + border.Render("│")

	leftMiddle := border.Render("┤") + centerText(leftContent, badgeWidth-2) + border.Render("├")
	rightMiddle := border.Render("┤") + centerText(rightContent, badgeWidth-2) + border.Render("├")
	line2 := border.Render("╰"+strings.Repeat("─", leftDash)) + leftMiddle + border.Render("─") + rightMiddle + border.Render(strings.Repeat("─", rightDash)+"╯")

	leftBottom := border.Render("╰" + strings.Repeat("─", badgeWidth-2) + "╯")
	rightBottom := border.Render("╰" + strings.Repeat("─", badgeWidth-2) + "╯")
	line3 := strings.Repeat(" ", leftDash+1) + leftBottom + " " + rightBottom
	line3 += strings.Repeat(" ", maxInt(0, totalWidth-lipgloss.Width(line3)))
	return []string{line1, line2, line3}
}

func discographyListRows(m model, visibleIndexes []int, maxRows int) []string {
	const contentWidth = 61
	total := len(visibleIndexes)
	if total == 0 {
		return nil
	}
	cursor := minInt(m.discographyCursor, total-1)
	start := visibleStart(cursor, total, maxRows)
	end := minInt(total, start+maxRows)
	visible := end - start
	thumb := scrollbarThumb(start, total, visible)
	rows := make([]string, 0, visible)
	for row, position := 0, start; position < end; position, row = position+1, row+1 {
		original := visibleIndexes[position]
		focused := position == cursor
		hovered := m.hoverRegion == "discography:"+strconv.Itoa(position)
		cursorMark := "  "
		if focused {
			cursorMark = theme.PromptStyle.Render("❯ ")
		}
		check := theme.SecondaryTextStyle.Render("○")
		if m.discographySelected[original] {
			check = theme.AccentTextStyle.Render("●")
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
		titleStyle := theme.PrimaryTextStyle
		if focused {
			titleStyle = theme.FocusedRowLabelStyle
		} else if hovered {
			titleStyle = theme.HoverRowLabelStyle
		}
		left := cursorMark + check + " " + titleStyle.Render(title)
		gap := contentWidth - lipgloss.Width(left) - 1
		if gap < 1 {
			gap = 1
		}
		rows = append(rows, fitStyled(left+strings.Repeat(" ", gap)+scroll, contentWidth))
	}
	return rows
}

func renderDiscographyList(m model, visibleIndexes []int) string {
	const totalWidth = 65
	rows := discographyListRows(m, visibleIndexes, discographyMaxRows(m))
	if len(rows) == 0 {
		return renderPanelBox([]string{theme.MutedStyle.Render("No albums match the current view.")}, totalWidth, theme.BorderStyle)
	}
	return renderPanelBox(rows, totalWidth, theme.BorderStyle)
}

func discographyMaxRows(m model) int {
	maxRows := 13
	if m.height > 0 {
		maxRows = maxInt(6, minInt(16, m.height-20))
	}
	return maxRows
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

func maskValue(value string) string {
	if value == "" {
		return ""
	}
	return strings.Repeat("•", minInt(16, len([]rune(value))))
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
	border := theme.BorderStyle
	if m.hoverRegion == "connection:action" {
		border = theme.InnerBorderStyle
	}
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(renderPanelBox(rows, 65, border)), "")
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
	border := theme.BorderStyle
	if m.diagnosticsBusy || m.hoverRegion == "diagnostics:action" {
		border = theme.InnerBorderStyle
	}
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(renderPanelBox(rows, 65, border)), "")
}

func renderUpdateCheckView(m model) string {
	if m.updateChecking {
		box := renderInfoBox("UPDATE", "Checking configured release source", m.spinner.View(), 65, true)
		return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(box), "")
	}
	if m.updateResult.Latest == "" {
		box := renderInfoBox("UPDATE", "No result yet", versionText(), 65, m.hoverRegion == "update:action")
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
	border := theme.BorderStyle
	if m.hoverRegion == "update:action" {
		border = theme.InnerBorderStyle
	}
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(renderPanelBox(rows, 65, border)), "")
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
	list := renderSimpleAlbumList(m.similar, m.similarCursor, 13, m.hoverRegion, "similar")
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(artistBox), centerToHeader(list), "")
}

func renderInfoView(m model) string {
	labels := []string{"M O D E S", "A U T O M A T I O N", "D A T A", "C U R A T I O N", "I M P O R T S"}
	widths := []int{13, 23, 11, 19, 19}
	boxes := make([]string, len(labels))
	for index, label := range labels {
		boxes[index] = renderChoiceBox(label, widths[index], index == m.infoIndex, m.hoverRegion == "info:tab:"+strconv.Itoa(index))
	}
	firstRow := joinThreeLineBoxes(boxes[:3], theme.SepStyle.Render("•"))
	secondRow := joinThreeLineBoxes(boxes[3:], theme.SepStyle.Render("•"))
	sections := [][]string{
		{
			helpRow("MANUAL", "search by artist, album, or both"),
			helpRow("DISCOGRAPHY", "browse, filter, clean, sort, and select albums"),
			helpRow("FILE", "lists, playlists, album folders, artist folders"),
			helpRow("SIMILAR", "press S from results, preview, or completion"),
		},
		{
			helpRow("HEADLESS CLI", "manual, file, and discography commands"),
			helpRow("SETTINGS", "S opens the six-section Settings area"),
			helpRow("CONNECTION", "test API lookup and authentication from Tools"),
			helpRow("COMPLETIONS", "install zsh, bash, fish, or PowerShell completion"),
			helpRow("MOUSE", "click sections, rows, actions, tabs, and footer hints"),
		},
		{
			helpRow("HISTORY", "re-run, export, or delete saved sessions"),
			helpRow("RECOVERY", "resume an interrupted queue at startup"),
			helpRow("PROFILES", "switch named account configurations"),
			helpRow("ACCOUNT", "credentials, source, and credential path"),
			helpRow("STORAGE", "~/.config/lastfm-scrobbler"),
		},
		{
			helpRow("SPACE / A", "select tracks or all/none"),
			helpRow("- / +", "change the current album loop while selecting tracks"),
			helpRow("MOUSE FOOTER", "adjust interval, navigate, and change loop directly"),
			helpRow("RERUN", "edit a saved queue before starting it again"),
			helpRow("E", "export JSON, CSV, TXT, and M3U8"),
		},
		{
			helpRow("LIST FILE", "TXT, CSV, TSV, and JSON album lists"),
			helpRow("PLAYLIST", "M3U and M3U8 playlists"),
			helpRow("ALBUM FOLDER", "infer Artist/Album from one folder"),
			helpRow("ARTIST FOLDER", "scan album folders beneath an artist"),
			helpRow("O", platform.PickerDescription()),
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
		focused := m.settingsFocus == settingsFocusContent && index == m.profileCursor
		hovered := m.hoverRegion == "profiles:"+strconv.Itoa(index)
		cursor := "  "
		if focused {
			cursor = theme.PromptStyle.Render("❯ ")
		}
		nameStyle := theme.PrimaryTextStyle
		stateStyle := theme.SecondaryTextStyle
		if focused {
			nameStyle = theme.FocusedRowLabelStyle
			stateStyle = theme.FocusedRowValueStyle
		} else if hovered {
			nameStyle = theme.HoverRowLabelStyle
			stateStyle = theme.HoverRowValueStyle
		}
		active := ""
		if strings.EqualFold(profile, m.cfg.Profile) {
			active = stateStyle.Render("  ACTIVE")
		}
		line := cursor + nameStyle.Render(profile) + active
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
