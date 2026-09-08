package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/setup"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

const setupPanelWidth = 65

func renderSetupView(m model) string {
	content := renderSetupBody(m)
	return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(content), m.centerToApp(setupProgress(m.setup.Page)))
}

func renderSetupBody(m model) string {
	var content string
	switch m.setup.Page {
	case setup.PageWelcome:
		content = renderSetupWelcome(m)
	case setup.PageSystem:
		content = renderSetupSystem(m)
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
	case setup.PageComplete:
		content = renderSetupComplete(m)
	}
	return content
}

func renderSetupWelcome(m model) string {
	lines := []string{
		centerText(theme.MutedStyle.Render("first-time configuration wizard"), m.panelWidth()-4),
		centerToWidth(renderExactBox("GET STARTED  ❯", 20, true), m.panelWidth()-4),
	}
	return setupPanelForModel(m, "W E L C O M E", lines)
}

func renderSetupSystem(m model) string {
	env := m.setup.Environment
	rows := []string{
		setupRow("OS", env.OperatingSystem, false, m.panelWidth()),
		setupRow("ARCH", env.Architecture, false, m.panelWidth()),
		setupRow("TERM", env.Terminal, false, m.panelWidth()),
		setupRow("PKG", env.PackageManager, false, m.panelWidth()),
	}
	return setupPanelForModel(m, "S Y S T E M", rows)
}

func renderSetupAccount(m model) string {
	labels := []string{"USERNAME", "PASSWORD", "API KEY", "SECRET"}
	rows := make([]string, len(labels))
	for index, label := range labels {
		input := m.setupInputs[index].View()
		rows[index] = setupInputRow(label, input, index == m.setup.Field, m.panelWidth())
	}
	return setupPanelForModel(m, "L A S T . F M  A C C O U N T", rows)
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
		rows[index] = fitSetupValue(marker+" "+label+"  "+theme.MutedStyle.Render(choice.note), m.panelWidth()-4)
	}
	return setupPanelForModel(m, "C R E D E N T I A L  S T O R A G E", rows)
}

func renderSetupScrobbling(m model) string {
	width := setupNarrowPanelWidth(m, 39)
	rows := []string{
		setupRow("USE RECOMMENDED", setupOnOff(m.setup.Recommended), m.setup.Field == 0, width),
		setupRow("LOOP", fmt.Sprintf("%d", m.setup.Loop), m.setup.Field == 1, width),
		setupRow("INTERVAL", m.setup.Interval.String(), m.setup.Field == 2, width),
	}
	return setupPanel("S C R O B B L I N G", rows, width)
}

func renderSetupInterface(m model) string {
	width := setupNarrowPanelWidth(m, 37)
	rows := []string{
		setupRow("MOUSE SUPPORT", setupOnOff(m.setup.MouseEnabled), m.setup.Field == 0, width),
		setupRow("COMPACT HEADER", setupOnOff(m.setup.CompactHeader), m.setup.Field == 1, width),
		setupRow("NOTIFICATIONS", setupOnOff(m.setup.Notify), m.setup.Field == 2, width),
	}
	return setupPanel("I N T E R F A C E", rows, width)
}

func renderSetupReview(m model) string {
	rows := []string{
		setupRow("ACCOUNT", m.setup.Account.Username, false, m.panelWidth()),
		setupRow("CREDENTIALS", m.setup.CredentialSource, false, m.panelWidth()),
		setupRow("LOOP", fmt.Sprintf("%d", m.setup.ApplyConfig().DefaultLoop), false, m.panelWidth()),
		setupRow("INTERVAL", m.setup.ApplyConfig().DefaultInterval.String(), false, m.panelWidth()),
		setupRow("MOUSE", setupOnOff(m.setup.MouseEnabled), false, m.panelWidth()),
		setupRow("NOTIFICATIONS", setupOnOff(m.setup.Notify), false, m.panelWidth()),
		setupRow("COMPACT", setupOnOff(m.setup.CompactHeader), false, m.panelWidth()),
	}
	return setupPanelForModel(m, "R E V I E W", rows)
}

func renderSetupComplete(m model) string {
	width := setupNarrowPanelWidth(m, 37)
	button := centerToWidth(renderExactBox("E N J O Y  ❯", 18, true), width-4)
	return setupPanel("C O M P L E T E D", []string{button}, width)
}
