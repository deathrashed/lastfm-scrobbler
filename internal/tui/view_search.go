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
	row := joinResponsiveBoxes(boxes, widths, m.panelWidth(), theme.SepStyle.Render("•"))
	descriptions := []string{
		"search by artist, album, or both",
		"scrobble multiple albums at once",
		"load albums from a list, playlist, or music folder",
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		m.centerToApp(row),
		m.centerToApp(theme.MutedStyle.Render(descriptions[m.modeIndex])),
		"",
	)
}

func renderImportSourceView(m model) string {
	labels := make([]string, len(fileSourceSpecs))
	widths := make([]int, len(fileSourceSpecs))
	for index, spec := range fileSourceSpecs {
		labels[index], widths[index] = spec.label, spec.width
	}
	firstRow := joinResponsiveBoxes([]string{
		renderChoiceBox(labels[0], widths[0], m.importSourceIndex == 0, m.hoverRegion == "import:0"),
		renderChoiceBox(labels[1], widths[1], m.importSourceIndex == 1, m.hoverRegion == "import:1"),
	}, []int{widths[0], widths[1]}, m.panelWidth(), theme.SepStyle.Render("•"))
	secondRow := joinResponsiveBoxes([]string{
		renderChoiceBox(labels[2], widths[2], m.importSourceIndex == 2, m.hoverRegion == "import:2"),
		renderChoiceBox(labels[3], widths[3], m.importSourceIndex == 3, m.hoverRegion == "import:3"),
	}, []int{widths[2], widths[3]}, m.panelWidth(), theme.SepStyle.Render("•"))
	spec := fileSourceSpecFor(m.importSourceIndex)
	lines := []string{
		m.centerToApp(firstRow),
		m.centerToApp(secondRow),
		m.centerToApp(theme.MutedStyle.Render(spec.description)),
		"",
		m.centerToApp(renderFilePathView(m)),
	}
	if m.searching {
		lines = append(lines, "", m.centerToApp(m.spinner.View()+" "+theme.MutedStyle.Render("Loading imports…")))
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
	box := renderTextBox(label, m.searchInput.View(), placeholder, m.panelWidth(), m.searchInput.Focused())
	lines := []string{m.centerToApp(box), ""}
	if m.searching {
		lines = append(lines, m.centerToApp(m.spinner.View()+" "+theme.MutedStyle.Render("Loading from Last.fm…")), "")
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

func discographyFilterTabWidthFor(m model) int {
	extra := maxInt(0, m.panelWidth()-65)
	return minInt(maxInt(discographyFilterTabWidth, discographyFilterTabWidth+extra), maxInt(discographyFilterTabWidth, m.panelWidth()-discographySortTabWidth-discographyCleanTabWidth-2))
}

func discographyFilterContentWidthFor(m model) int {
	return maxInt(1, discographyFilterTabWidthFor(m)-4)
}

func discographyExpandedFilterWidthFor(m model) int {
	return maxInt(discographyExpandedFilterWidth, m.panelWidth()-4)
}

func discographyExpandedFilterContentWidthFor(m model) int {
	return maxInt(1, discographyExpandedFilterWidthFor(m)-4)
}

func discographySortTabX(m model) int {
	groupWidth := discographySortTabWidth + discographyFilterTabWidthFor(m) + discographyCleanTabWidth + 2
	return 1 + maxInt(0, (m.panelWidth()-groupWidth)/2)
}

func discographyFilterTabX(m model) int {
	return discographySortTabX(m) + discographySortTabWidth + 1
}

func discographyCleanTabX(m model) int {
	return discographyFilterTabX(m) + discographyFilterTabWidthFor(m) + 1
}

func discographyExpandedFilterX(m model) int {
	return 1 + maxInt(0, (m.panelWidth()-discographyExpandedFilterWidthFor(m))/2)
}

func renderDiscographySelectView(m model) string {
	visible := m.discographyVisibleIndexes()
	chooser := renderDiscographyChooser(m, visible)
	return lipgloss.JoinVertical(lipgloss.Left,
		m.centerToApp(chooser),
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
	filterWidth := discographyFilterTabWidthFor(m)
	filterTab := renderDiscographyControlTab("FILTER", compactFilter, filterWidth, m.discographyFiltering, filterHovered, expanded)
	cleanTab := renderDiscographyControlTab("CLEAN", clean, discographyCleanTabWidth, false, m.hoverRegion == "discography:clean", false)

	groupWidth := discographySortTabWidth + filterWidth + discographyCleanTabWidth + 2
	left := maxInt(0, (m.panelWidth()-groupWidth)/2)
	right := maxInt(0, m.panelWidth()-groupWidth-left)
	frameWidth := groupWidth + 4
	frameLeft := maxInt(0, (m.panelWidth()-frameWidth)/2)
	frameRight := maxInt(0, m.panelWidth()-frameWidth-frameLeft)
	lines := []string{
		border.Render(strings.Repeat("─", left)) + sortTab[0] + " " + filterTab[0] + " " + cleanTab[0] + border.Render(strings.Repeat("─", right)),
		border.Render(strings.Repeat("─", frameLeft)+"╭─") + sortTab[1] + border.Render("─") + filterTab[1] + border.Render("─") + cleanTab[1] + border.Render("─╮"+strings.Repeat("─", frameRight)),
		border.Render(strings.Repeat("─", frameLeft)+"│ ") + sortTab[2] + " " + filterTab[2] + " " + cleanTab[2] + border.Render(" │"+strings.Repeat("─", frameRight)),
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
		rowWidth := maxInt(1, m.panelWidth()-4)
		lines = append(lines, border.Render("│")+" "+padRight(fitStyled(row, rowWidth), rowWidth)+" "+border.Render("│"))
	}

	lines = append(lines, renderDiscographyCountAttachment(len(visibleIndexes), len(m.discographySelected), m.panelWidth())...)
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
	return lipgloss.Width(strings.TrimSpace(m.discographyFilter)) > discographyFilterContentWidthFor(m)
}

func renderDiscographyExpandedFilter(m model) []string {
	border := theme.BorderStyle
	width := discographyExpandedFilterWidthFor(m)
	left := (width - 1) / 2
	right := width - 2 - left
	innerTop := border.Render("╭" + strings.Repeat("─", left) + "┴" + strings.Repeat("─", right) + "╮")
	lines := []string{fitStyled(border.Render("│")+" "+innerTop+" "+border.Render("│"), m.panelWidth())}

	contentLines := discographyExpandedFilterContent(m)
	for _, content := range contentLines {
		content = fitStyled(content, discographyExpandedFilterContentWidthFor(m))
		lines = append(lines,
			border.Render("│")+" "+border.Render("│")+" "+padRight(content, discographyExpandedFilterContentWidthFor(m))+" "+border.Render("│")+" "+border.Render("│"),
		)
	}
	innerBottom := border.Render("╰" + strings.Repeat("─", discographyExpandedFilterWidthFor(m)-2) + "╯")
	lines = append(lines, border.Render("│")+" "+innerBottom+" "+border.Render("│"))
	return lines
}

func discographyExpandedFilterContent(m model) []string {
	if m.discographyFiltering {
		input := m.filterInput
		input.Width = discographyExpandedFilterContentWidthFor(m)
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
	wrapped := wrapWords(query, discographyExpandedFilterContentWidthFor(m))
	if len(wrapped) == 0 {
		wrapped = []string{query}
	}
	if len(wrapped) > 2 {
		wrapped = wrapped[:2]
		wrapped[1] = truncateToWidth(wrapped[1]+"…", discographyExpandedFilterContentWidthFor(m))
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

func renderDiscographyCountAttachment(results, selected, totalWidth int) []string {
	border := theme.BorderStyle
	const (
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
	countGroup := leftBottom + " " + rightBottom
	countGroupWidth := lipgloss.Width(countGroup)
	groupLeft := maxInt(0, (totalWidth-countGroupWidth)/2)
	groupRight := maxInt(0, totalWidth-countGroupWidth-groupLeft)
	line3 := strings.Repeat(" ", groupLeft) + countGroup + strings.Repeat(" ", groupRight)
	line3 = fitStyled(line3, totalWidth)
	return []string{line1, line2, line3}
}

func discographyListRows(m model, visibleIndexes []int, maxRows int) []string {
	contentWidth := maxInt(1, m.panelWidth()-4)
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
	totalWidth := m.panelWidth()
	rows := discographyListRows(m, visibleIndexes, discographyMaxRows(m))
	if len(rows) == 0 {
		return renderPanelBox([]string{theme.MutedStyle.Render("No albums match the current view.")}, totalWidth, theme.BorderStyle)
	}
	return renderPanelBox(rows, totalWidth, theme.BorderStyle)
}

func discographyMaxRows(m model) int {
	return m.visibleRows(20, 6, 32)
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
		box := renderInfoBox("CONNECTION", "Testing Last.fm lookup and authentication readiness", m.spinner.View(), m.panelWidth(), true)
		return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(box), "")
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
		detail := truncateToWidth(item.Detail, maxInt(1, m.panelWidth()-27))
		right := status + "  " + theme.MutedStyle.Render(detail)
		gap := m.panelWidth() - 6 - lipgloss.Width(left) - lipgloss.Width(right)
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
	return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(renderPanelBox(rows, m.panelWidth(), border)), "")
}

func renderDiagnosticsView(m model) string {
	var rows []string
	contentWidth := m.panelWidth() - 4
	switch {
	case m.diagnosticsBusy:
		rows = []string{theme.KeyStyle.Render("EXPORTING ❯ ") + m.spinner.View() + " " + theme.MutedStyle.Render("building redacted support bundle")}
	case m.diagnosticsPath != "":
		rows = []string{
			statLine("STATUS", "complete", contentWidth),
			statLine("PATH", truncateToWidth(m.diagnosticsPath, contentWidth-15), contentWidth),
			statLine("SECRETS", "redacted", contentWidth),
		}
	default:
		rows = []string{
			statLine("CONTENTS", "logs + redacted config + history summary", contentWidth),
			statLine("OUTPUT", truncateToWidth(m.cfg.ExportDir, contentWidth-16), contentWidth),
			statLine("SECRETS", "passwords, API secret and session key excluded", contentWidth),
		}
	}
	border := theme.BorderStyle
	if m.diagnosticsBusy || m.hoverRegion == "diagnostics:action" {
		border = theme.InnerBorderStyle
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(renderPanelBox(rows, m.panelWidth(), border)), "")
}

func renderUpdateCheckView(m model) string {
	if m.updateChecking {
		box := renderInfoBox("UPDATE", "Checking configured release source", m.spinner.View(), m.panelWidth(), true)
		return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(box), "")
	}
	if m.updateResult.Latest == "" {
		box := renderInfoBox("UPDATE", "No result yet", versionText(), m.panelWidth(), m.hoverRegion == "update:action")
		return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(box), "")
	}
	status := "UP TO DATE"
	if m.updateResult.Available {
		status = "AVAILABLE"
	}
	contentWidth := m.panelWidth() - 4
	rows := []string{
		statLine("CURRENT", m.updateResult.Current, contentWidth),
		statLine("LATEST", m.updateResult.Latest, contentWidth),
		statLine("STATUS", status, contentWidth),
	}
	if m.updateResult.URL != "" {
		rows = append(rows, statLine("RELEASE", truncateToWidth(m.updateResult.URL, contentWidth-16), contentWidth))
	}
	border := theme.BorderStyle
	if m.hoverRegion == "update:action" {
		border = theme.InnerBorderStyle
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(renderPanelBox(rows, m.panelWidth(), border)), "")
}

func versionText() string { return version.Version }

func renderEnvPathView(m model) string {
	box := renderTextBox("CREDENTIALS PATH", m.envInput.View(), "~/.config/lastfm-scrobbler/.env", m.panelWidth(), m.envInput.Focused())
	lines := []string{m.centerToApp(box), m.centerToApp(theme.MutedStyle.Render("enter path to credentials file")), ""}
	if m.envStatus != "" {
		lines = append(lines, m.centerToApp(theme.SuccessStyle.Render(m.envStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderSimilarSelectView(m model) string {
	artistBox := renderInfoBox("BASED ON", m.similarArtist, "", m.panelWidth(), false)
	list := renderSimpleAlbumList(m.similar, m.similarCursor, m.visibleRows(8, 6, 32), m.hoverRegion, "similar", m.panelWidth())
	return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(artistBox), m.centerToApp(list), "")
}

func renderInfoView(m model) string {
	labels := []string{"M O D E S", "A U T O M A T I O N", "D A T A", "C U R A T I O N", "I M P O R T S"}
	widths := []int{13, 23, 11, 19, 19}
	boxes := make([]string, len(labels))
	for index, label := range labels {
		boxes[index] = renderChoiceBox(label, widths[index], index == m.infoIndex, m.hoverRegion == "info:tab:"+strconv.Itoa(index))
	}
	firstRow := joinResponsiveBoxes(boxes[:3], widths[:3], m.panelWidth(), theme.SepStyle.Render("•"))
	secondRow := joinResponsiveBoxes(boxes[3:], widths[3:], m.panelWidth(), theme.SepStyle.Render("•"))
	infoRow := func(label, description string) string { return helpRow(label, description, m.panelWidth()-4) }
	sections := [][]string{
		{
			infoRow("MANUAL", "search by artist, album, or both"),
			infoRow("DISCOGRAPHY", "browse, filter, clean, sort, and select albums"),
			infoRow("FILE", "lists, playlists, album folders, artist folders"),
			infoRow("SIMILAR", "press S from results, preview, or completion"),
		},
		{
			infoRow("HEADLESS CLI", "manual, file, and discography commands"),
			infoRow("SETTINGS", "S opens the six-section Settings area"),
			infoRow("CONNECTION", "test API lookup and authentication from Tools"),
			infoRow("COMPLETIONS", "install zsh, bash, fish, or PowerShell completion"),
			infoRow("MOUSE", "click sections, rows, actions, tabs, and footer hints"),
		},
		{
			infoRow("HISTORY", "re-run, export, or delete saved sessions"),
			infoRow("RECOVERY", "resume an interrupted queue at startup"),
			infoRow("PROFILES", "switch named account configurations"),
			infoRow("ACCOUNT", "credentials, source, and credential path"),
			infoRow("STORAGE", "~/.config/lastfm-scrobbler"),
		},
		{
			infoRow("SPACE / A", "select tracks or all/none"),
			infoRow("- / +", "change the current album loop while selecting tracks"),
			infoRow("MOUSE FOOTER", "adjust interval, navigate, and change loop directly"),
			infoRow("RERUN", "edit a saved queue before starting it again"),
			infoRow("E", "export JSON, CSV, TXT, and M3U8"),
		},
		{
			infoRow("LIST FILE", "TXT, CSV, TSV, and JSON album lists"),
			infoRow("PLAYLIST", "M3U and M3U8 playlists"),
			infoRow("ALBUM FOLDER", "infer Artist/Album from one folder"),
			infoRow("ARTIST FOLDER", "scan album folders beneath an artist"),
			infoRow("O", platform.PickerDescription()),
		},
	}
	box := renderPanelBox(sections[m.infoIndex], m.panelWidth(), theme.BorderStyle)
	return lipgloss.JoinVertical(lipgloss.Left,
		m.centerToApp(firstRow),
		m.centerToApp(secondRow),
		m.centerToApp(box),
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
	box := renderPanelBox(rows, m.panelWidth(), theme.BorderStyle)
	lines := []string{m.centerToApp(box), ""}
	if m.profileStatus != "" {
		lines = append(lines, m.centerToApp(theme.SuccessStyle.Render(m.profileStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func renderProfileNameView(m model) string {
	box := renderTextBox("PROFILE NAME", m.profileInput.View(), "personal", m.panelWidth(), m.profileInput.Focused())
	return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(box), m.centerToApp(theme.MutedStyle.Render("letters, numbers, hyphens and underscores")), "")
}
