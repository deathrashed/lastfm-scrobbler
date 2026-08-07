package tui

import (
	"net/url"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

const (
	minHeaderWidth     = 67
	fullHeaderWidth    = 67
	fullHeaderLines    = 11
	compactHeaderLines = 4
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
		Subtitle: "enter artist - album manually",
		Icon:     theme.IconManual,
	},
	"discography": {
		Title:    "D I S C O G R A P H Y",
		Subtitle: "scrobble an artist's top albums",
		Icon:     theme.IconDiscography,
	},
	"file": {
		Title:    "F I L E",
		Subtitle: "load albums from a list, playlist, or music folder",
		Icon:     theme.IconFile,
	},
	"config": {
		Title:    "C O N F I G",
		Subtitle: "edit user settings",
		Icon:     theme.IconManual,
	},
	"advanced": {
		Title:    "A D V A N C E D",
		Subtitle: "edit advanced behaviour",
		Icon:     theme.IconSettings,
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
}

func RenderHeader(width int, stg stage, modeChoice, username, settingsLine string, compact bool) string {
	return RenderHeaderWithHover(width, stg, modeChoice, username, settingsLine, compact, false)
}

func RenderHeaderWithHover(width int, stg stage, modeChoice, username, settingsLine string, compact, urlHover bool) string {
	if compact {
		return renderCompactHeader(stg, modeChoice, username, settingsLine, urlHover)
	}
	return renderFullHeader(fullHeaderWidth, stg, modeChoice, username, settingsLine, urlHover)
}

func (m model) compactHeaderEnabled() bool {
	return m.cfg.CompactHeader
}

func (m model) headerHeight() int {
	if m.compactHeaderEnabled() {
		return compactHeaderLines
	}
	return fullHeaderLines
}

func renderCompactHeader(stg stage, modeChoice, username, settingsLine string, urlHover bool) string {
	_ = stg
	_ = username
	_ = urlHover
	spec := compactHeaderSpecFor(modeChoice)
	subtitle := theme.MutedStyle.Render(spec.Subtitle)
	if modeChoice == "" {
		subtitle = renderDashboardContext()
	}
	if settingsLine != "" {
		subtitle = renderSettingsContext(settingsLine)
	}
	return strings.Join([]string{
		renderCompactBorder('╭', '╮'),
		renderCompactContent(spec.Title, theme.HeaderTextStyle),
		renderCompactContentStyled(subtitle),
		renderCompactBottom(spec.Icon),
	}, "\n")
}

func compactHeaderSpecFor(modeChoice string) compactHeaderSpec {
	if spec, ok := compactHeaderSpecs[modeChoice]; ok {
		return spec
	}
	return compactHeaderSpecs[""]
}

func renderCompactBorder(left, right rune) string {
	return theme.BorderStyle.Render(string(left) + strings.Repeat("─", fullHeaderWidth-2) + string(right))
}

func renderCompactContent(value string, style lipgloss.Style) string {
	return renderCompactContentStyled(style.Render(value))
}

func renderCompactContentStyled(value string) string {
	return theme.BorderStyle.Render("│") + centerText(value, fullHeaderWidth-2) + theme.BorderStyle.Render("│")
}

func renderCompactBottom(icon string) string {
	innerWidth := fullHeaderWidth - 2
	iconWidth := lipgloss.Width(icon)
	dashes := maxInt(0, innerWidth-iconWidth)
	left := dashes / 2
	right := dashes - left
	return theme.BorderStyle.Render("╰"+strings.Repeat("─", left)) +
		theme.TitleIconStyle.Render(icon) +
		theme.BorderStyle.Render(strings.Repeat("─", right)+"╯")
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
	spec := compactHeaderSpecFor(modeChoice)
	return spec.Subtitle, spec.Title, spec.Icon
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
