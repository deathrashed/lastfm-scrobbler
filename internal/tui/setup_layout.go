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

func setupPanel(title string, lines []string) string {
	titleCore := "┤ " + theme.HeaderTextStyle.Render(title) + " ├"
	remaining := setupPanelWidth - 2 - lipgloss.Width(titleCore)
	left := remaining / 2
	right := remaining - left
	border := theme.BorderStyle
	out := []string{border.Render("╭" + strings.Repeat("─", left) + titleCore + strings.Repeat("─", right) + "╮")}
	closure := "╰" + strings.Repeat("─", lipgloss.Width(titleCore)-2) + "╯"
	line := strings.Repeat(" ", left) + closure + strings.Repeat(" ", right)
	out = append(out, border.Render("│")+fitSetupValue(line, setupPanelWidth-2)+border.Render("│"))
	for _, value := range lines {
		for _, lineValue := range strings.Split(value, "\n") {
			out = append(out, border.Render("│")+" "+fitSetupValue(lineValue, setupPanelWidth-4)+" "+border.Render("│"))
		}
	}
	out = append(out, border.Render("╰"+strings.Repeat("─", setupPanelWidth-2)+"╯"))
	return strings.Join(out, "\n")
}

func setupRow(label, value string, focused bool) string {
	labelStyle, arrowStyle, valueStyle := theme.RowLabelStyle, theme.RowArrowStyle, theme.RowValueStyle
	if focused {
		labelStyle, arrowStyle, valueStyle = theme.FocusedRowLabelStyle, theme.FocusedRowArrowStyle, theme.FocusedRowValueStyle
	}
	labelText := padRight(labelStyle.Render(label), 22)
	row := labelText + " " + arrowStyle.Render("❯") + " " + valueStyle.Render(value)
	return fitSetupValue(row, setupPanelWidth-4)
}

func setupInputRow(label, value string, focused bool) string {
	labelStyle, arrowStyle := theme.RowLabelStyle, theme.RowArrowStyle
	if focused {
		labelStyle, arrowStyle = theme.FocusedRowLabelStyle, theme.FocusedRowArrowStyle
	}
	return fitSetupValue(padRight(labelStyle.Render(label), 22)+" "+arrowStyle.Render("❯")+" "+value, setupPanelWidth-4)
}

func setupStatusRow(label, status string, applying bool) string {
	if status == "" && applying {
		status = "waiting"
	}
	if status == "installed" || status == "configured" || status == "saved" || status == "stored securely" || status == "connected" || status == "ready" {
		return theme.CompleteStyle.Render(theme.IconSuccess) + "  " + theme.RowLabelStyle.Render(label) + "  " + theme.CompleteStyle.Render(status)
	}
	if status == "failed" {
		return theme.ErrorStyle.Render(theme.IconError) + "  " + theme.RowLabelStyle.Render(label) + "  " + theme.ErrorStyle.Render(status)
	}
	return theme.SecondaryTextStyle.Render("○") + "  " + theme.RowLabelStyle.Render(label) + "  " + theme.RowValueStyle.Render(status)
}

func fitSetupValue(value string, width int) string { return padRight(fitStyled(value, width), width) }

func setupOnOff(value bool) string {
	if value {
		return "ON"
	}
	return "OFF"
}

func setupFontDetection(m model) string {
	if m.setup.HasFont() {
		return "selected"
	}
	return "not detected"
}

func setupTerminalChoice(m model) string {
	if !m.setup.HasFont() {
		return "NO"
	}
	if m.setup.TerminalDefault && m.setup.Terminal.Supported {
		return "YES"
	}
	if m.setup.TerminalDefault {
		return "MANUAL"
	}
	return "NO"
}

func setupProgress(page setup.Page) string {
	step := -1
	switch page {
	case setup.PageSystem:
		step = 0
	case setup.PageFont:
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
	case setup.PageApply:
		step = 7
	case setup.PageComplete:
		step = setup.NumberedPageCount
	}
	dots := make([]string, setup.NumberedPageCount)
	for index := range dots {
		style := theme.MutedStyle
		if index < step || step == setup.NumberedPageCount {
			style = theme.AccentTextStyle
		} else if index == step {
			style = theme.PrimaryTextStyle
		}
		dots[index] = style.Render("●")
	}
	if page == setup.PageComplete {
		dots = append(dots, theme.CompleteStyle.Render(theme.IconSuccess))
	}
	return strings.Join(dots, "  ")
}

func setupScreenRegions(m model, bodyY int) []mouseRegion {
	lines := strings.Split(stripANSI(renderSetupBody(m)), "\n")
	regions := []mouseRegion{}
	add := func(id, needle string, message tea.KeyMsg) {
		for index, line := range lines {
			if strings.Contains(line, needle) {
				regions = append(regions, mouseRegion{id: id, x: 1, y: bodyY + index, width: setupPanelWidth, height: 1, message: message})
				return
			}
		}
	}
	switch m.setup.Page {
	case setup.PageWelcome:
		add("setup:continue", "GET STARTED", keyMessage("enter"))
	case setup.PageFont:
		for index, font := range m.setup.Fonts {
			add(fmt.Sprintf("setup:font:%d", index), font.Name, keyMessage("enter"))
		}
		add("setup:toggle:install", "INSTALL FONT", keyMessage("space"))
		add("setup:toggle:terminal", "TERMINAL DEFAULT", keyMessage("space"))
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
	case setup.PageApply:
		add("setup:continue", "C O N N E C T I O N", keyMessage("enter"))
	case setup.PageComplete:
		add("setup:continue", "S E T U P", keyMessage("enter"))
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
