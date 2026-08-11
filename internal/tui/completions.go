package tui

import (
	tea "github.com/charmbracelet/bubbletea"

	"github.com/deathrashed/lastfm-scrobbler/internal/completion"
)

type completionInstallMsg struct {
	result completion.InstallResult
	err    error
}

func (m model) openCompletions() (tea.Model, tea.Cmd) {
	m.stage = stageCompletions
	m.modeChoice = "completions"
	m.returnStage = stageConfig
	m.settingsSection = settingsTools
	m.completionShell = completion.DetectShell()
	m.completionStatus = completion.DefaultManager().Status(m.completionShell)
	m.completionMessage = ""
	m.err = nil
	return m, nil
}

func (m model) updateCompletions(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	shells := completion.Shells
	shellIndex := completionShellIndex(m.completionShell)
	switch msg.String() {
	case "up", "k":
		shellIndex = (shellIndex + len(shells) - 1) % len(shells)
		m.completionShell = shells[shellIndex]
		m.refreshCompletionStatus()
	case "down", "j":
		shellIndex = (shellIndex + 1) % len(shells)
		m.completionShell = shells[shellIndex]
		m.refreshCompletionStatus()
	case "enter":
		return m, m.installCompletionCmd()
	case "r":
		m.refreshCompletionStatus()
	case "esc":
		return m.openSettingsSection(settingsTools, settingsFocusContent)
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m *model) refreshCompletionStatus() {
	m.completionStatus = completion.DefaultManager().Status(m.completionShell)
}

func completionShellIndex(shell completion.Shell) int {
	for index, candidate := range completion.Shells {
		if candidate == shell {
			return index
		}
	}
	return 0
}

func (m model) installCompletionCmd() tea.Cmd {
	shell := m.completionShell
	return func() tea.Msg {
		result, err := completion.DefaultManager().Install(shell)
		return completionInstallMsg{result: result, err: err}
	}
}
