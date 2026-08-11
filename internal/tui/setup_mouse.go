package tui

import (
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) updateSetupMouse(region mouseRegion) (tea.Model, tea.Cmd) {
	parts := strings.Split(region.id, ":")
	if len(parts) < 2 {
		return m, nil
	}
	switch parts[1] {
	case "continue":
		return m.setupContinue()
	case "font":
		if len(parts) == 3 {
			index, err := strconv.Atoi(parts[2])
			if err == nil && index >= 0 && index < len(m.setup.Fonts) {
				m.setup.Field = index
				m.setup.FontIndex = index
			}
		}
	case "storage":
		if len(parts) == 3 {
			index, err := strconv.Atoi(parts[2])
			choices := credentialChoices()
			if err == nil && index >= 0 && index < len(choices) {
				m.setup.Field = index
				m.setup.CredentialSource = choices[index].value
			}
		}
	case "account":
		if len(parts) == 3 {
			index, err := strconv.Atoi(parts[2])
			if err == nil && index >= 0 && index < len(m.setupInputs) {
				m.setup.Field = index
				return m.focusSetupInput()
			}
		}
	case "toggle":
		if len(parts) == 3 && parts[2] == "install" {
			m.setup.Field = len(m.setup.Fonts)
			m.toggleSetupValue()
		} else if len(parts) == 3 && parts[2] == "terminal" {
			m.setup.Field = len(m.setup.Fonts) + 1
			m.toggleSetupValue()
		} else if len(parts) == 3 {
			index, err := strconv.Atoi(parts[2])
			if err == nil {
				m.setup.Field = index
				m.toggleSetupValue()
			}
		}
	}
	return m, nil
}

func (m model) setupMouseMove(delta int) (tea.Model, tea.Cmd) {
	m.moveSetupCursor(delta)
	return m, nil
}
