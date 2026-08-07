package tui

import (
	"net/url"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

const (
	minHeaderWidth  = 67
	fullHeaderWidth = 67
)

func RenderHeader(width int, stg stage, modeChoice, username, settingsLine string, compact bool) string {
	return RenderHeaderWithHover(width, stg, modeChoice, username, settingsLine, compact, false)
}

func RenderHeaderWithHover(width int, stg stage, modeChoice, username, settingsLine string, compact, urlHover bool) string {
	if width > 0 && width < minHeaderWidth {
		return renderCompactHeader(stg, modeChoice, username, settingsLine, urlHover)
	}
	if compact {
		return renderCompactHeader(stg, modeChoice, username, settingsLine, urlHover)
	}
	return renderFullHeader(fullHeaderWidth, stg, modeChoice, username, settingsLine, urlHover)
}

func renderCompactHeader(stg stage, modeChoice, username, settingsLine string, urlHover bool) string {
	width := fullHeaderWidth
	outer := theme.BorderStyle
	contextLine, badgeText, badgeIcon := headerBadge(stg, modeChoice)
	context := theme.HeaderTextStyle.Render(contextLine)
	if settingsLine != "" {
		context = renderSettingsContext(settingsLine)
	}
	lines := []string{
		outer.Render("╭" + strings.Repeat("─", width-2) + "╮"),
		outer.Render("│") + centerText(renderHeaderURL(username, urlHover), width-2) + outer.Render("│"),
		outer.Render("│") + centerText(context, width-2) + outer.Render("│"),
	}
	_, bottom, underline := renderBadgeBottom(width, badgeText, badgeIcon)
	lines = append(lines, bottom, underline)
	return strings.Join(lines, "\n")
}

func renderFullHeader(width int, stg stage, modeChoice, username, settingsLine string, urlHover bool) string {
	outer := theme.BorderStyle
	inner := theme.BorderStyle
	wordmark := theme.TitleIconStyle
	contextLine, badgeText, badgeIcon := headerBadge(stg, modeChoice)
	isDashboard := modeChoice == ""
	innerWidth := width - 12

	lines := []string{outer.Render("╭" + strings.Repeat("─", width-2) + "╮")}
	urlText := renderHeaderURL(username, urlHover)
	lines = append(lines,
		outer.Render("│")+centerText(urlText, width-2)+outer.Render("│"),
		outer.Render("│    ")+inner.Render("╭"+strings.Repeat("─", innerWidth)+"╮")+outer.Render("    │"),
	)
	for _, line := range renderWordmarkLines() {
		lines = append(lines, outer.Render("│    ")+inner.Render("│")+wordmark.Render(line)+inner.Render("│")+outer.Render("    │"))
	}
	lines = append(lines, outer.Render("│    ")+inner.Render("╰"+strings.Repeat("─", innerWidth)+"╯")+outer.Render("    │"))

	var context string
	switch {
	case settingsLine != "" && (stg == stageTrackSelect || stg == stagePreview || stg == stageScrobbling || stg == stageDone):
		context = renderSettingsContext(settingsLine)
	case isDashboard:
		context = renderDashboardContext()
	case modeChoice == "info" || modeChoice == "file":
		context = theme.MutedStyle.Render(contextLine)
	default:
		context = theme.HeaderTextStyle.Render(contextLine)
	}
	lines = append(lines, outer.Render("│")+centerTextLeftHeavy(context, width-2)+outer.Render("│"))
	badgeTop, bottomLine, badgeUnderline := renderBadgeBottom(width, badgeText, badgeIcon)
	lines = append(lines, badgeTop, bottomLine, badgeUnderline)
	return strings.Join(lines, "\n")
}

func lastfmURL(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return "https://www.last.fm"
	}
	return "https://www.last.fm/user/" + url.PathEscape(username)
}

func renderHeaderURL(username string, hover bool) string {
	value := lastfmURL(username)
	display := truncateToWidth(headerURLDisplay(username), fullHeaderWidth-2)
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
	displayWidth := lipgloss.Width(truncateToWidth(headerURLDisplay(username), fullHeaderWidth-2))
	left = 1 + (fullHeaderWidth-2-displayWidth)/2
	return left, 1, displayWidth
}

func headerURLDisplay(username string) string {
	username = strings.TrimSpace(username)
	if username == "" {
		return "last.fm"
	}
	return "last.fm/user/" + url.PathEscape(username)
}

func renderDashboardContext() string {
	return theme.HeaderMetaStyle.Render("SEARCH") +
		theme.SepStyle.Render("  •  ") +
		theme.HeaderMetaStyle.Render("SELECT") +
		theme.SepStyle.Render("  •  ") +
		theme.HeaderMetaStyle.Render("SCROBBLE")
}

func renderSettingsContext(settingsLine string) string {
	parts := strings.Split(settingsLine, "|")
	styled := make([]string, 0, len(parts))
	for _, part := range parts {
		styled = append(styled, theme.HeaderTextStyle.Render(strings.TrimSpace(part)))
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
	badgeUnderline = strings.Repeat(" ", maxInt(0, (width-lipgloss.Width(underline))/2)) + underline
	return badgeTop, bottomLine, badgeUnderline
}

func headerBadge(stg stage, modeChoice string) (contextLine, badgeText, badgeIcon string) {
	_ = stg
	switch modeChoice {
	case "manual":
		return "enter artist - album manually", "M A N U A L", theme.IconManual
	case "file":
		return "load albums from a list, playlist, or music folder", "F I L E", theme.IconFile
	case "config":
		return "edit user settings", "C O N F I G", theme.IconManual
	case "advanced":
		return "edit advanced behaviour", "A D V A N C E D", theme.IconSettings
	case "env":
		return "choose a credentials file", "E N V", theme.IconManual
	case "discography":
		return "scrobble an artists top albums", "D I S C O G R A P H Y", theme.IconDiscography
	case "history":
		return "review, export, or re-run sessions", "H I S T O R Y", theme.IconHistory
	case "profiles":
		return "switch Last.fm profiles", "P R O F I L E S", theme.IconProfile
	case "profile":
		return "create a Last.fm profile", "P R O F I L E", theme.IconProfile
	case "info":
		return "guide, formats, controls, and data locations", "I N F O", theme.IconInfo
	case "connection":
		return "test Last.fm lookup and authentication readiness", "C O N N E C T I O N", theme.IconSettings
	case "diagnostics":
		return "export redacted logs and configuration", "D I A G N O S T I C S", theme.IconFile
	case "update":
		return "check the configured release source", "U P D A T E", theme.IconSettings
	default:
		return "SEARCH  •  SELECT  •  SCROBBLE", "D A S H B O A R D", theme.IconDashboard
	}
}

func renderWordmarkLines() []string {
	return []string{
		"   ╦  ╔═╗╔═╗╔╦╗ ╔═╗╔╦╗  ╔═╗╔═╗╦═╗╔═╗╔╗ ╔╗ ╦  ╔═╗╦═╗    ",
		"   ║  ╠═╣╚═╗ ║  ╠╣ ║║║  ╚═╗║  ╠╦╝║ ║╠╩╗╠╩╗║  ║╣ ╠╦╝    ",
		"   ╩═╝╩ ╩╚═╝ ╩ o╚  ╩ ╩  ╚═╝╚═╝╩╚═╚═╝╚═╝╚═╝╩═╝╚═╝╩╚═    ",
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
