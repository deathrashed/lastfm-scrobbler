package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/setup"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

const setupPanelWidth = 65

func renderSetupView(m model) string {
	content := renderSetupBody(m)
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(content), centerToHeader(setupProgress(m.setup.Page)))
}

func renderSetupBody(m model) string {
	var content string
	switch m.setup.Page {
	case setup.PageWelcome:
		content = renderSetupWelcome()
	case setup.PageSystem:
		content = renderSetupSystem(m)
	case setup.PageFont:
		content = renderSetupFont(m)
	case setup.PageAccount:
		content = renderSetupAccount(m)
	case setup.PageStorage:
		content = renderSetupStorage(m)
	case setup.PageScrobbling:
		content = renderSetupScrobbling(m)
	case setup.PageInterface:
		content = renderSetupInterface(m)
	case setup.PageReview:
		content = renderSetupReview(m)
	case setup.PageApply:
		content = renderSetupApply(m)
	case setup.PageComplete:
		content = renderSetupComplete()
	}
	return content
}

func renderSetupWelcome() string {
	lines := append([]string{}, renderWordmarkLines()...)
	lines = append(lines, centerText(theme.MutedStyle.Render("first-time configuration wizard"), setupPanelWidth-4))
	lines = append(lines, "", centerText(renderExactBox("GET STARTED  ❯", 20, true), setupPanelWidth-4))
	return setupPanel("W E L C O M E", lines)
}

func renderSetupSystem(m model) string {
	env := m.setup.Environment
	rows := []string{
		setupRow("OPERATING SYSTEM", env.OperatingSystem, false),
	}
	if env.Distribution != "" && env.Distribution != "unknown" {
		rows = append(rows, setupRow("DISTRIBUTION", env.Distribution, false))
	}
	rows = append(rows,
		setupRow("ARCHITECTURE", env.Architecture, false),
		setupRow("TERMINAL", env.Terminal, false),
		setupRow("NERD FONT", setupFontDetection(m), false),
		setupRow("PACKAGE MANAGER", env.PackageManager, false),
	)
	return setupPanel("S Y S T E M", rows)
}

func renderSetupFont(m model) string {
	rows := make([]string, 0, len(m.setup.Fonts)+2)
	for index, font := range m.setup.Fonts {
		value := font.Name
		if index == m.setup.FontIndex {
			value = theme.AccentTextStyle.Render("❯") + " " + theme.PrimaryTextStyle.Render(font.Name)
		} else {
			value = theme.SecondaryTextStyle.Render("○") + " " + theme.RowLabelStyle.Render(font.Name)
		}
		rows = append(rows, fitSetupValue(value, setupPanelWidth-4))
	}
	rows = append(rows,
		setupRow("INSTALL FONT", setupOnOff(m.setup.InstallFont), m.setup.Field == len(m.setup.Fonts)),
		setupRow("TERMINAL DEFAULT", setupTerminalChoice(m), m.setup.Field == len(m.setup.Fonts)+1),
	)
	return setupPanel("N E R D  F O N T", rows)
}

func renderSetupAccount(m model) string {
	labels := []string{"USERNAME", "PASSWORD", "API KEY", "SECRET"}
	rows := make([]string, len(labels))
	for index, label := range labels {
		input := m.setupInputs[index].View()
		rows[index] = setupInputRow(label, input, index == m.setup.Field)
	}
	return setupPanel("L A S T . F M  A C C O U N T", rows)
}

func renderSetupStorage(m model) string {
	choices := credentialChoices()
	rows := make([]string, len(choices))
	for index, choice := range choices {
		marker := theme.SecondaryTextStyle.Render("○")
		label := theme.RowLabelStyle.Render(choice.label)
		if choice.value == m.setup.CredentialSource {
			marker = theme.AccentTextStyle.Render("●")
		}
		if index == m.setup.Field {
			label = theme.FocusedRowLabelStyle.Render(choice.label)
		}
		rows[index] = fitSetupValue(marker+" "+label+"  "+theme.MutedStyle.Render(choice.note), setupPanelWidth-4)
	}
	return setupPanel("C R E D E N T I A L  S T O R A G E", rows)
}

func renderSetupScrobbling(m model) string {
	rows := []string{
		setupRow("USE RECOMMENDED", setupOnOff(m.setup.Recommended), m.setup.Field == 0),
		setupRow("LOOP", fmt.Sprintf("%d", m.setup.Loop), m.setup.Field == 1),
		setupRow("INTERVAL", m.setup.Interval.String(), m.setup.Field == 2),
	}
	return setupPanel("S C R O B B L I N G", rows)
}

func renderSetupInterface(m model) string {
	rows := []string{
		setupRow("MOUSE SUPPORT", setupOnOff(m.setup.MouseEnabled), m.setup.Field == 0),
		setupRow("COMPACT HEADER", setupOnOff(m.setup.CompactHeader), m.setup.Field == 1),
		setupRow("NOTIFICATIONS", setupOnOff(m.setup.Notify), m.setup.Field == 2),
	}
	return setupPanel("I N T E R F A C E", rows)
}

func renderSetupReview(m model) string {
	font := m.setup.SelectedFont().Name
	if font == "" {
		font = "current font"
	}
	rows := []string{
		setupRow("ACCOUNT", m.setup.Account.Username, false),
		setupRow("CREDENTIALS", m.setup.CredentialSource, false),
		setupRow("LOOP", fmt.Sprintf("%d", m.setup.ApplyConfig().DefaultLoop), false),
		setupRow("INTERVAL", m.setup.ApplyConfig().DefaultInterval.String(), false),
		setupRow("MOUSE", setupOnOff(m.setup.MouseEnabled), false),
		setupRow("NOTIFICATIONS", setupOnOff(m.setup.Notify), false),
		setupRow("COMPACT", setupOnOff(m.setup.CompactHeader), false),
		setupRow("NERD FONT", font, false),
		setupRow("TERMINAL DEFAULT", setupTerminalChoice(m), false),
	}
	return setupPanel("R E V I E W", rows)
}

func renderSetupApply(m model) string {
	result := m.setup.ApplyResult
	rows := []string{
		setupStatusRow("Nerd Font", result.FontStatus, m.setup.Applying),
		setupStatusRow("terminal", result.TerminalStatus, m.setup.Applying),
		setupStatusRow("configuration", result.Configuration, m.setup.Applying),
		setupStatusRow("credentials", result.Credentials, m.setup.Applying),
		setupStatusRow("Last.fm API", result.Connection, m.setup.Applying),
		setupStatusRow("authentication", result.Authentication, m.setup.Applying),
	}
	if m.setup.Applying {
		rows[0] = m.spinner.View() + "  " + theme.RowLabelStyle.Render("Nerd Font") + "  " + theme.RowValueStyle.Render("working")
	}
	if result.Error != nil {
		rows = append(rows, theme.ErrorStyle.Render(truncateToWidth("setup failed: "+result.Error.Error(), setupPanelWidth-4)))
	}
	return setupPanel("C O N N E C T I O N  T E S T", rows)
}

func renderSetupComplete() string {
	return setupPanel("S E T U P  C O M P L E T E  "+theme.IconSuccess, []string{
		"Last.fm Scrobbler is ready to use.",
		"",
		"You can change these options later in Settings.",
	})
}
