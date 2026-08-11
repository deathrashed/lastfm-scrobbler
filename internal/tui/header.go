package tui

import (
	"net/url"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

const (
	minHeaderWidth          = minAppWidth
	fullHeaderWidth         = minAppWidth
	fullHeaderLines         = 11
	compactHeaderLines      = 4
	compactHeaderInnerWidth = fullHeaderWidth - 2
)

type compactHeaderSpec struct {
	Title    string
	Subtitle string
	Icon     string
}

var compactHeaderSpecs = map[string]compactHeaderSpec{
	"": {
		Title:    "D A S H B O A R D",
		Subtitle: "SEARCH  •  SELECT  •  SCROBBLE",
		Icon:     theme.IconDashboard,
	},
	"manual": {
		Title:    "M A N U A L",
		Subtitle: "search by artist, album, or both",
		Icon:     theme.IconManual,
	},
	"discography": {
		Title:    "D I S C O G R A P H Y",
		Subtitle: "scrobble multiple albums at once",
		Icon:     theme.IconDiscography,
	},
	"file": {
		Title:    "F I L E",
		Subtitle: "load albums from a list, playlist, or music folder",
		Icon:     theme.IconFile,
	},
	"history": {
		Title:    "H I S T O R Y",
		Subtitle: "review, export, or re-run sessions",
		Icon:     theme.IconHistory,
	},
	"profiles": {
		Title:    "P R O F I L E S",
		Subtitle: "switch Last.fm profiles",
		Icon:     theme.IconProfile,
	},
	"info": {
		Title:    "I N F O",
		Subtitle: "guide, formats, controls, and data locations",
		Icon:     theme.IconInfo,
	},
	"env": {
		Title:    "E N V",
		Subtitle: "choose a credentials file",
		Icon:     theme.IconManual,
	},
	"profile": {
		Title:    "P R O F I L E",
		Subtitle: "create a Last.fm profile",
		Icon:     theme.IconProfile,
	},
	"connection": {
		Title:    "C O N N E C T I O N",
		Subtitle: "test Last.fm lookup and authentication readiness",
		Icon:     theme.IconSettings,
	},
	"diagnostics": {
		Title:    "D I A G N O S T I C S",
		Subtitle: "export redacted logs and configuration",
		Icon:     theme.IconFile,
	},
	"update": {
		Title:    "U P D A T E",
		Subtitle: "check the configured release source",
		Icon:     theme.IconSettings,
	},
	"completions": {
		Title:    "C O M P L E T I O N S",
		Subtitle: "install shell completion for this user",
		Icon:     theme.IconSettings,
	},
	"setup": {
		Title:    "S E T U P",
		Subtitle: "first-time configuration wizard",
		Icon:     theme.IconSettings,
	},
}

func RenderHeader(width int, stg stage, modeChoice, username, settingsLine string, compact bool) string {
	return RenderHeaderWithHover(width, stg, modeChoice, username, settingsLine, compact, false)
}

func RenderHeaderWithHover(width int, stg stage, modeChoice, username, settingsLine string, compact, urlHover bool) string {
	return RenderHeaderWithHoverArtist(width, stg, modeChoice, username, settingsLine, compact, urlHover, "")
}

func RenderHeaderWithHoverArtist(width int, stg stage, modeChoice, username, settingsLine string, compact, urlHover bool, artist string) string {
	width = appWidth(width)
	if compact {
		return renderCompactHeader(width, stg, modeChoice, username, settingsLine, urlHover, compactHeaderArtistFor(modeChoice, artist))
	}
	return renderFullHeader(width, stg, modeChoice, username, settingsLine, urlHover, artist, "")
}

func (m model) renderHeader() string {
	if m.cfg.CompactHeader {
		return renderCompactHeader(m.appWidth(), m.stage, m.modeChoice, m.cfg.Username, m.headerSettingsLine(), m.headerURLHover, m.compactHeaderArtist())
	}
	return renderFullHeader(m.appWidth(), m.stage, m.modeChoice, m.cfg.Username, m.headerSettingsLine(), m.headerURLHover, m.headerArtist(), m.activityContent())
}

func (m model) compactHeaderEnabled() bool {
	return m.cfg.CompactHeader
}

func (m model) headerHeight() int {
	if m.compactHeaderEnabled() {
		if m.compactHeaderArtist() != "" {
			return compactHeaderLines + 1
		}
		return compactHeaderLines
	}
	height := fullHeaderLines
	if m.nowPlayingEnabled() {
		height++
	}
	if m.headerArtist() != "" {
		height += 2
	}
	return height
}

func renderCompactHeader(width int, stg stage, modeChoice, username, settingsLine string, urlHover bool, artist string) string {
	_ = stg
	_ = username
	_ = urlHover
	spec := compactHeaderSpecFor(modeChoice)
	subtitle := theme.MutedStyle.Render(spec.Subtitle)
	if modeChoice == "" {
		subtitle = renderDashboardContext(width)
	}
	if settingsLine != "" {
		subtitle = renderSettingsContext(settingsLine)
	}
	lines := []string{
		renderCompactBorder(width, '╭', '╮'),
		renderCompactContent(width, spec.Title, theme.HeaderTextStyle),
		renderCompactContentStyled(width, subtitle),
	}
	if artist != "" {
		lines = append(lines, renderCompactArtistRow(width, artist))
	}
	lines = append(lines, renderCompactBottom(width, spec.Icon))
	return strings.Join(lines, "\n")
}

func (m model) compactHeaderArtist() string {
	if !m.compactHeaderEnabled() {
		return ""
	}
	return compactHeaderArtistFor(m.modeChoice, m.headerArtist())
}

func compactHeaderArtistFor(modeChoice, artist string) string {
	if modeChoice != "manual" && modeChoice != "discography" {
		return ""
	}
	return strings.ToUpper(strings.TrimSpace(artist))
}

func renderCompactArtistRow(width int, artist string) string {
	const prefix = "ARTIST ❯ "
	artist = truncateToWidth(strings.ToUpper(strings.TrimSpace(artist)), maxInt(1, width-2-lipgloss.Width(prefix)))
	row := theme.AccentTextStyle.Render("ARTIST") + " " +
		theme.PrimaryTextStyle.Render("❯") + " " +
		theme.ArtistStyle.Render(artist)
	return renderCompactContentStyled(width, row)
}

func compactHeaderSpecFor(modeChoice string) compactHeaderSpec {
	if spec, ok := compactHeaderSpecs[modeChoice]; ok {
		return spec
	}
	return compactHeaderSpecs[""]
}

func renderCompactBorder(width int, left, right rune) string {
	return theme.BorderStyle.Render(string(left) + strings.Repeat("─", maxInt(1, width-2)) + string(right))
}

func renderCompactContent(width int, value string, style lipgloss.Style) string {
	return renderCompactContentStyled(width, style.Render(value))
}

func renderCompactContentStyled(width int, value string) string {
	return theme.BorderStyle.Render("│") + centerText(value, maxInt(1, width-2)) + theme.BorderStyle.Render("│")
}

func renderCompactBottom(width int, icon string) string {
	innerWidth := maxInt(1, width-2)
	iconWidth := lipgloss.Width(icon)
	dashes := maxInt(0, innerWidth-iconWidth)
	left := dashes / 2
	right := dashes - left
	return theme.BorderStyle.Render("╰"+strings.Repeat("─", left)) +
		theme.TitleIconStyle.Render(icon) +
		theme.BorderStyle.Render(strings.Repeat("─", right)+"╯")
}

func renderFullHeader(width int, stg stage, modeChoice, username, settingsLine string, urlHover bool, artist, activity string) string {
	outer := theme.BorderStyle
	inner := theme.BorderStyle
	wordmark := theme.TitleIconStyle
	contextLine, badgeText, badgeIcon := headerBadge(stg, modeChoice)
	badgeText = responsiveHeaderTitle(badgeText, width)
	isDashboard := modeChoice == ""
	wordmarkLines := renderWordmarkLinesFor(width)
	innerWidth := lipgloss.Width(wordmarkLines[0])
	frameWidth := innerWidth + 2
	leftPadding := maxInt(0, (width-2-frameWidth)/2)
	rightPadding := maxInt(0, width-2-frameWidth-leftPadding)

	lines := []string{outer.Render("╭" + strings.Repeat("─", width-2) + "╮")}
	urlText := renderHeaderURL(username, urlHover, width)
	lines = append(lines,
		outer.Render("│")+centerText(urlText, width-2)+outer.Render("│"),
	)
	if activity != "" {
		lines = append(lines, outer.Render("│")+centerText(activity, width-2)+outer.Render("│"))
	}
	lines = append(lines,
		outer.Render("│")+strings.Repeat(" ", leftPadding)+inner.Render("╭"+strings.Repeat("─", innerWidth)+"╮")+strings.Repeat(" ", rightPadding)+outer.Render("│"),
	)
	for _, line := range wordmarkLines {
		lines = append(lines, outer.Render("│")+strings.Repeat(" ", leftPadding)+inner.Render("│")+wordmark.Render(padRight(line, innerWidth))+inner.Render("│")+strings.Repeat(" ", rightPadding)+outer.Render("│"))
	}
	lines = append(lines, outer.Render("│")+strings.Repeat(" ", leftPadding)+inner.Render("╰"+strings.Repeat("─", innerWidth)+"╯")+strings.Repeat(" ", rightPadding)+outer.Render("│"))

	var context string
	switch {
	case settingsLine != "" && (stg == stageTrackSelect || stg == stagePreview || stg == stageScrobbling || stg == stageDone):
		context = renderSettingsContext(settingsLine)
	case isDashboard:
		context = renderDashboardContext(width)
	default:
		context = theme.SecondaryTextStyle.Render(contextLine)
	}
	lines = append(lines, outer.Render("│")+centerTextLeftHeavy(context, width-2)+outer.Render("│"))
	badgeTop, bottomLine, badgeUnderline := renderBadgeBottom(width, badgeText, badgeIcon)
	lines = append(lines, badgeTop, bottomLine)
	if strings.TrimSpace(artist) == "" {
		lines = append(lines, badgeUnderline)
	} else {
		lines = append(lines, renderArtistBadgeExtension(width, badgeText, badgeIcon, artist)...)
	}
	return strings.Join(lines, "\n")
}

func (m model) headerArtist() string {
	if m.modeChoice != "manual" && m.modeChoice != "discography" {
		return ""
	}
	switch m.stage {
	case stageResults:
		if m.modeChoice == "manual" && len(m.results) > 0 {
			index := minInt(maxInt(m.resultsCursor, 0), len(m.results)-1)
			return strings.TrimSpace(m.results[index].Artist)
		}
		return ""
	case stageDiscographySelect:
		artist := strings.TrimSpace(m.discographyArtist)
		if artist == "" && len(m.discography) > 0 {
			artist = strings.TrimSpace(m.discography[0].Artist)
		}
		return artist
	case stageTrackSelect, stagePreview, stageScrobbling, stageDone:
		return strings.TrimSpace(m.currentArtist())
	case stageSimilarSelect:
		return strings.TrimSpace(m.similarArtist)
	default:
		return ""
	}
}

func renderArtistBadgeExtension(width int, badgeText, badgeIcon, artist string) []string {
	border := theme.BadgeBorderStyle
	icon := theme.BadgeStyle
	artistStyle := theme.ArtistStyle

	sectionWidth := lipgloss.Width(badgeText) + 4
	artistText := spacedArtistNameForWidth(artist, width)
	maxArtistBoxWidth := minInt(101, maxInt(57, width-10))
	artistText = truncateToWidth(artistText, maxArtistBoxWidth-4)
	artistBoxWidth := maxInt(7, lipgloss.Width(artistText)+4)
	if artistBoxWidth > maxArtistBoxWidth {
		artistBoxWidth = maxArtistBoxWidth
	}
	// A one-cell shoulder on either side reads visually like an accidental
	// notch rather than a deliberate size transition. Snap near-equal artist
	// badges to the section width so names such as SLAYER form a clean vertical
	// continuation, while genuinely shorter/longer names still taper naturally.
	if artistBoxWidth >= sectionWidth-2 && artistBoxWidth <= sectionWidth+2 {
		artistBoxWidth = sectionWidth
	}

	var connector string
	switch {
	case artistBoxWidth < sectionWidth:
		left := maxInt(1, (sectionWidth-artistBoxWidth)/2)
		right := maxInt(1, sectionWidth-artistBoxWidth-left)
		middleWidth := maxInt(3, artistBoxWidth-2)
		connector = border.Render("╰"+strings.Repeat("─", left-1)+"╮") +
			renderIconRibbon(middleWidth, badgeIcon, border, icon) +
			border.Render("╭"+strings.Repeat("─", right-1)+"╯")
	case artistBoxWidth == sectionWidth:
		connector = border.Render("│") + renderIconRibbon(sectionWidth-2, badgeIcon, border, icon) + border.Render("│")
	default:
		left := maxInt(1, (artistBoxWidth-sectionWidth)/2)
		right := maxInt(1, artistBoxWidth-sectionWidth-left)
		middleWidth := maxInt(3, sectionWidth-2)
		connector = border.Render("╭"+strings.Repeat("─", left-1)+"╯") +
			renderIconRibbon(middleWidth, badgeIcon, border, icon) +
			border.Render("╰"+strings.Repeat("─", right-1)+"╮")
	}

	artistInner := artistBoxWidth - 2
	artistLine := border.Render("│") + centerText(artistStyle.Render(artistText), artistInner) + border.Render("│")
	artistBottom := border.Render("╰" + strings.Repeat("─", artistInner) + "╯")
	return []string{
		centerText(connector, width),
		centerText(artistLine, width),
		centerText(artistBottom, width),
	}
}

func renderIconRibbon(width int, badgeIcon string, border, icon lipgloss.Style) string {
	if width <= 0 {
		return ""
	}
	iconWidth := lipgloss.Width(badgeIcon)
	if iconWidth >= width {
		return icon.Render(truncateToWidth(badgeIcon, width))
	}
	usable := width - iconWidth
	left := usable / 2
	right := usable - left
	if width >= iconWidth+4 {
		left = maxInt(1, left-1)
		right = maxInt(1, right-1)
		return " " + border.Render(strings.Repeat("─", left)) + icon.Render(badgeIcon) + border.Render(strings.Repeat("─", right)) + " "
	}
	return border.Render(strings.Repeat("─", left)) + icon.Render(badgeIcon) + border.Render(strings.Repeat("─", right))
}

func spacedArtistName(value string) string {
	return spacedArtistNameWithSeparator(value, " ", "   ")
}

func spacedArtistNameForWidth(value string, width int) string {
	if densityFor(width) == densityWide {
		return spacedArtistNameWithSeparator(value, "  ", "    ")
	}
	return spacedArtistName(value)
}

func spacedArtistNameWithSeparator(value, letterSeparator, wordSeparator string) string {
	words := strings.Fields(strings.ToUpper(strings.TrimSpace(value)))
	spaced := make([]string, 0, len(words))
	for _, word := range words {
		runes := []rune(word)
		parts := make([]string, 0, len(runes))
		for _, r := range runes {
			parts = append(parts, string(r))
		}
		spaced = append(spaced, strings.Join(parts, letterSeparator))
	}
	return strings.Join(spaced, wordSeparator)
}

func lastfmURL(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return "https://www.last.fm"
	}
	return "https://www.last.fm/user/" + url.PathEscape(username)
}

func renderHeaderURL(username string, hover bool, width int) string {
	value := lastfmURL(username)
	display := truncateToWidth(headerURLDisplay(username), width-2)
	style := theme.HeaderURLStyle
	if hover {
		style = theme.HeaderURLHoverStyle
	}
	return renderOSC8(value, style.Render(display))
}

func renderOSC8(target, value string) string {
	return "\x1b]8;;" + target + "\x1b\\" + value + "\x1b]8;;\x1b\\"
}

func headerURLBounds(username string) (left, top, width int) {
	return headerURLBoundsAtWidth(username, fullHeaderWidth)
}

func headerURLBoundsAtWidth(username string, headerWidth int) (left, top, width int) {
	displayWidth := lipgloss.Width(truncateToWidth(headerURLDisplay(username), headerWidth-2))
	left = 1 + (headerWidth-2-displayWidth)/2
	return left, 1, displayWidth
}

func headerURLDisplay(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return "last.fm"
	}
	return "last.fm/user/" + url.PathEscape(username)
}

func renderDashboardContext(widths ...int) string {
	width := fullHeaderWidth
	if len(widths) > 0 {
		width = widths[0]
	}
	separator := "  •  "
	if densityFor(width) == densityWide {
		separator = "   •   "
	}
	return theme.MutedStyle.Render("SEARCH" + separator + "SELECT" + separator + "SCROBBLE")
}

func renderSettingsContext(settingsLine string) string {
	parts := strings.Split(settingsLine, "|")
	styled := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		label, value, found := strings.Cut(part, " ")
		if !found {
			styled = append(styled, theme.SecondaryTextStyle.Render(part))
			continue
		}
		styled = append(styled,
			theme.SecondaryTextStyle.Render(label+" ")+theme.PrimaryTextStyle.Render(strings.TrimSpace(value)),
		)
	}
	return strings.Join(styled, theme.SepStyle.Render("   •   "))
}

func renderBadgeBottom(width int, badgeText, badgeIcon string) (badgeTop, bottomLine, badgeUnderline string) {
	border := theme.BadgeBorderStyle
	label := theme.BadgeStyle
	icon := theme.BadgeStyle
	badgeInnerWidth := lipgloss.Width(" " + badgeText + " ")
	badgeTotalWidth := badgeInnerWidth + 2
	topBox := border.Render("╭" + strings.Repeat("─", badgeInnerWidth) + "╮")
	topLeft := (width - 2 - badgeTotalWidth) / 2
	topRight := width - 2 - badgeTotalWidth - topLeft
	badgeTop = border.Render("│") + strings.Repeat(" ", topLeft) + topBox + strings.Repeat(" ", topRight) + border.Render("│")
	totalDash := width - 2 - badgeTotalWidth
	leftDash := maxInt(0, totalDash/2)
	rightDash := maxInt(0, totalDash-leftDash)
	bottomLine = border.Render("╰"+strings.Repeat("─", leftDash)+"┤") + label.Render(" "+badgeText+" ") + border.Render("├"+strings.Repeat("─", rightDash)+"╯")
	iconWidth := lipgloss.Width(badgeIcon)
	dashes := maxInt(0, badgeInnerWidth-iconWidth)
	leftIconDash := dashes / 2
	rightIconDash := dashes - leftIconDash
	underline := border.Render("╰"+strings.Repeat("─", leftIconDash)) + icon.Render(badgeIcon) + border.Render(strings.Repeat("─", rightIconDash)+"╯")
	badgeUnderline = centerText(underline, width)
	return badgeTop, bottomLine, badgeUnderline
}

func headerBadge(stg stage, modeChoice string) (contextLine, badgeText, badgeIcon string) {
	_ = stg
	spec := compactHeaderSpecFor(modeChoice)
	return spec.Subtitle, spec.Title, spec.Icon
}

func responsiveHeaderTitle(value string, width int) string {
	if densityFor(width) != densityWide {
		return value
	}
	wide := strings.Join(strings.Fields(value), "  ")
	if lipgloss.Width(wide)+1 <= width-4 {
		return wide
	}
	return value
}

func renderWordmarkLines() []string {
	return []string{
		"   ╦  ╔═╗╔═╗╔╦╗ ╔═╗╔╦╗  ╔═╗╔═╗╦═╗╔═╗╔╗ ╔╗ ╦  ╔═╗╦═╗    ",
		"   ║  ╠═╣╚═╗ ║  ╠╣ ║║║  ╚═╗║  ╠╦╝║ ║╠╩╗╠╩╗║  ║╣ ╠╦╝    ",
		"   ╩═╝╩ ╩╚═╝ ╩ o╚  ╩ ╩  ╚═╝╚═╝╩╚═╚═╝╚═╝╚═╝╩═╝╚═╝╩╚═    ",
	}
}

func renderWordmarkLinesFor(width int) []string {
	if densityFor(width) != densityWide {
		return renderWordmarkLines()
	}
	return []string{
		"  ╦   ╔═╗ ╔═╗ ╔╦╗  ╔═╗ ╔╦╗    ╔═╗ ╔═╗ ╦═╗ ╔═╗ ╔╗  ╔╗  ╦   ╔═╗ ╦═╗  ",
		"  ║   ╠═╣ ╚═╗  ║   ╠╣  ║║║    ╚═╗ ║   ╠╦╝ ║ ║ ╠╩╗ ╠╩╗ ║   ║╣  ╠╦╝  ",
		"  ╩═╝ ╩ ╩ ╚═╝  ╩  o╚   ╩ ╩    ╚═╝ ╚═╝ ╩╚═ ╚═╝ ╚═╝ ╚═╝ ╩═╝ ╚═╝ ╩╚═  ",
	}
}

func padRight(value string, width int) string {
	current := lipgloss.Width(value)
	if current >= width {
		return value
	}
	return value + strings.Repeat(" ", width-current)
}

func centerTextLeftHeavy(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	right := (width - textWidth) / 2
	left := width - textWidth - right
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

func centerText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	left := (width - textWidth) / 2
	right := width - textWidth - left
	return strings.Repeat(" ", left) + text + strings.Repeat(" ", right)
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}
