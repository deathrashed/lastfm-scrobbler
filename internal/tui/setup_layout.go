package tui

import (
	"fmt"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/setup"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

func setupPanel(title string, lines []string, widths ...int) string {
	totalWidth := setupPanelWidth
	if len(widths) > 0 {
		// An explicit width is authoritative. This lets compact setup cards
		// (Scrobbling, Interface, Completed) stay genuinely narrow while the
		// information-heavy pages continue to use the normal panel width.
		totalWidth = maxInt(31, widths[0])
	}
	titleCore := "┤ " + theme.HeaderTextStyle.Render(title) + " ├"
	remaining := totalWidth - 2 - lipgloss.Width(titleCore)
	left := remaining / 2
	right := remaining - left
	border := theme.BorderStyle
	out := []string{border.Render("╭" + strings.Repeat("─", left) + titleCore + strings.Repeat("─", right) + "╮")}
	closure := "╰" + strings.Repeat("─", lipgloss.Width(titleCore)-2) + "╯"
	line := strings.Repeat(" ", left) + closure + strings.Repeat(" ", right)
	out = append(out, border.Render("│")+fitSetupValue(line, totalWidth-2)+border.Render("│"))
	for _, value := range lines {
		for _, lineValue := range strings.Split(value, "\n") {
			out = append(out, border.Render("│")+" "+fitSetupValue(lineValue, totalWidth-4)+" "+border.Render("│"))
		}
	}
	out = append(out, border.Render("╰"+strings.Repeat("─", totalWidth-2)+"╯"))
	return strings.Join(out, "\n")
}

func setupNarrowPanelWidth(m model, preferred int) int {
	maxWidth := maxInt(31, m.panelWidth())
	return minInt(maxWidth, maxInt(31, preferred))
}

func setupPanelForModel(m model, title string, lines []string) string {
	return setupPanel(title, lines, m.panelWidth())
}

// setupBridgeLine visually continues the SETUP stage badge into the current
// wizard page. The bridge is generated from display widths instead of copied
// dash counts, so long titles such as CREDENTIAL STORAGE stay centered.
func setupBridgeLine(title string, width int) string {
	width = maxInt(setupPanelWidth, width)
	icon := theme.IconSettings
	ribbon := "────" + icon + "────"
	// The bridge widens only as much as the child title grows beyond SETUP.
	// That keeps short pages compact (╭───╯ ... ╰───╮) while long titles such
	// as CREDENTIAL STORAGE naturally receive longer shoulders.
	titleDelta := maxInt(0, lipgloss.Width(title)-lipgloss.Width("S E T U P"))
	leftShoulder := maxInt(3, titleDelta/2)
	rightShoulder := maxInt(3, titleDelta-leftShoulder)
	maxShoulders := maxInt(3, (width-lipgloss.Width(ribbon)-7)/2)
	leftShoulder = minInt(leftShoulder, maxShoulders)
	rightShoulder = minInt(rightShoulder, maxShoulders)
	bridge := theme.BadgeBorderStyle.Render("╭"+strings.Repeat("─", leftShoulder)+"╯ ") +
		theme.BadgeBorderStyle.Render(strings.Repeat("─", 4)) +
		theme.BadgeStyle.Render(icon) +
		theme.BadgeBorderStyle.Render(strings.Repeat("─", 4)+" ╰"+strings.Repeat("─", rightShoulder)+"╮")
	return centerText(bridge, width)
}

func setupRow(label, value string, focused bool, widths ...int) string {
	totalWidth := setupPanelWidth
	if len(widths) > 0 {
		totalWidth = maxInt(31, widths[0])
	}
	labelStyle, arrowStyle, valueStyle := theme.RowLabelStyle, theme.RowArrowStyle, theme.RowValueStyle
	if focused {
		labelStyle, arrowStyle, valueStyle = theme.FocusedRowLabelStyle, theme.FocusedRowArrowStyle, theme.FocusedRowValueStyle
	}
	labelText := padRight(labelStyle.Render(label), 22)
	row := labelText + " " + arrowStyle.Render("❯") + " " + valueStyle.Render(value)
	return fitSetupValue(row, totalWidth-4)
}

func setupInputRow(label, value string, focused bool, widths ...int) string {
	totalWidth := setupPanelWidth
	if len(widths) > 0 {
		totalWidth = maxInt(31, widths[0])
	}
	labelStyle, arrowStyle := theme.RowLabelStyle, theme.RowArrowStyle
	if focused {
		labelStyle, arrowStyle = theme.FocusedRowLabelStyle, theme.FocusedRowArrowStyle
	}
	return fitSetupValue(padRight(labelStyle.Render(label), 22)+" "+arrowStyle.Render("❯")+" "+value, totalWidth-4)
}

func setupStatusRow(label, status string, applying bool, widths ...int) string {
	totalWidth := setupPanelWidth
	if len(widths) > 0 {
		totalWidth = maxInt(31, widths[0])
	}
	if status == "" && applying {
		status = "waiting"
	}
	glyph := theme.SecondaryTextStyle.Render("○")
	statusStyle := theme.RowValueStyle
	if status == "installed" || status == "configured" || status == "saved" || status == "stored securely" || status == "connected" || status == "ready" {
		glyph = theme.CompleteStyle.Render(theme.IconSuccess)
		statusStyle = theme.CompleteStyle
	} else if status == "failed" {
		glyph = theme.ErrorStyle.Render(theme.IconError)
		statusStyle = theme.ErrorStyle
	}
	labelWidth := minInt(22, maxInt(14, totalWidth/4))
	row := glyph + "  " + padRight(theme.RowLabelStyle.Render(label), labelWidth) + "  " + statusStyle.Render(status)
	return fitSetupValue(row, totalWidth-4)
}

func fitSetupValue(value string, width int) string { return padRight(fitStyled(value, width), width) }

func setupOnOff(value bool) string {
	if value {
		return "ON"
	}
	return "OFF"
}

func setupProgress(page setup.Page) string {
	step := 0
	switch page {
	case setup.PageWelcome:
		step = 0
	case setup.PageSystem:
		step = 1
	case setup.PageAccount:
		step = 2
	case setup.PageStorage:
		step = 3
	case setup.PageScrobbling:
		step = 4
	case setup.PageInterface:
		step = 5
	case setup.PageReview:
		step = 6
	case setup.PageComplete:
		step = 7
	default:
		step = 6
	}
	dots := make([]string, 8)
	for index := range dots {
		style := theme.MutedStyle
		if index < step {
			style = theme.AccentTextStyle
		} else if index == step {
			style = theme.PrimaryTextStyle
		}
		if page == setup.PageComplete {
			style = theme.AccentTextStyle
		}
		dots[index] = style.Render("●")
	}
	return strings.Join(dots, "  ")
}

func setupScreenRegions(m model, bodyY int) []mouseRegion {
	lines := strings.Split(stripANSI(renderSetupBody(m)), "\n")
	regions := []mouseRegion{}
	add := func(id, needle string, message tea.KeyMsg) {
		for index, line := range lines {
			if strings.Contains(line, needle) {
				regions = append(regions, mouseRegion{id: id, x: m.workX(), y: bodyY + index, width: m.panelWidth(), height: 1, message: message})
				return
			}
		}
	}
	switch m.setup.Page {
	case setup.PageWelcome:
		add("setup:continue", "GET STARTED", keyMessage("enter"))
	case setup.PageAccount:
		for index, label := range []string{"USERNAME", "PASSWORD", "API KEY", "SECRET"} {
			add(fmt.Sprintf("setup:account:%d", index), label, keyMessage("tab"))
		}
	case setup.PageStorage:
		for index, choice := range credentialChoices() {
			add(fmt.Sprintf("setup:storage:%d", index), choice.label, keyMessage("enter"))
		}
	case setup.PageScrobbling:
		for index, label := range []string{"USE RECOMMENDED", "LOOP", "INTERVAL"} {
			add(fmt.Sprintf("setup:toggle:%d", index), label, keyMessage("space"))
		}
	case setup.PageInterface:
		for index, label := range []string{"MOUSE SUPPORT", "COMPACT HEADER", "NOTIFICATIONS"} {
			add(fmt.Sprintf("setup:toggle:%d", index), label, keyMessage("space"))
		}
	case setup.PageReview:
		add("setup:continue", "ACCOUNT", keyMessage("enter"))
	case setup.PageComplete:
		if strings.TrimSpace(m.cfg.SessionKey) == "" {
			add("setup:continue", "E N J O Y", keyMessage("enter"))
		} else {
			add("setup:continue", "E N J O Y", keyMessage("enter"))
		}
	}
	return regions
}

type credentialChoice struct{ label, value, note string }

func credentialChoices() []credentialChoice {
	choices := []credentialChoice{{"credentials file", "file", "recommended"}, {"environment variables", "environment", ""}}
	if runtime.GOOS == "darwin" {
		choices = append([]credentialChoice{{"macOS Keychain", "keychain", "recommended"}}, choices...)
	}
	return choices
}
