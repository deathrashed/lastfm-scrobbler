package tui

import (
	"context"
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/connection"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/platform"
	"github.com/deathrashed/lastfm-scrobbler/internal/setup"
)

type setupApplyMsg struct {
	result setup.ApplyResult
	config config.Config
}

func (m model) updateSetup(msg tea.Msg) (tea.Model, tea.Cmd) {
	if apply, ok := msg.(setupApplyMsg); ok {
		m.setup.Applying = false
		m.setup.ApplyResult = apply.result
		if apply.result.Error == nil {
			m.cfg = apply.config
			m.client = lastfm.New(m.cfg.APIKey, m.cfg.APISecret, m.cfg.Username, m.cfg.Password, m.cfg.SessionKey)
			m.setup.Page = setup.PageComplete
		}
		return m, nil
	}
	key, ok := msg.(tea.KeyMsg)
	if !ok {
		return m, nil
	}
	if key.String() == "ctrl+c" {
		return m, tea.Quit
	}
	if m.setup.Page == setup.PageAccount {
		return m.updateSetupAccount(key)
	}
	switch key.String() {
	case "esc":
		return m.setupPrevious()
	case "up", "k":
		m.moveSetupCursor(-1)
	case "down", "j":
		m.moveSetupCursor(1)
	case "left":
		m.adjustSetupValue(-1)
	case "right":
		m.adjustSetupValue(1)
	case "space":
		m.toggleSetupValue()
	case "enter":
		return m.setupContinue()
	}
	return m, nil
}

func (m model) updateSetupAccount(key tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch key.String() {
	case "esc":
		return m.setupPrevious()
	case "enter":
		if m.setup.Field == len(m.setupInputs)-1 {
			return m.setupContinue()
		}
		m.setup.Field++
		return m.focusSetupInput()
	case "tab":
		m.setup.Field = (m.setup.Field + 1) % len(m.setupInputs)
		return m.focusSetupInput()
	case "shift+tab":
		m.setup.Field = (m.setup.Field + len(m.setupInputs) - 1) % len(m.setupInputs)
		return m.focusSetupInput()
	}
	input, command := m.setupInputs[m.setup.Field].Update(key)
	m.setupInputs[m.setup.Field] = input
	m.updateSetupAccountValue()
	return m, command
}

func (m model) focusSetupInput() (tea.Model, tea.Cmd) {
	commands := make([]tea.Cmd, 0, len(m.setupInputs))
	for index := range m.setupInputs {
		if index == m.setup.Field {
			commands = append(commands, m.setupInputs[index].Focus())
		} else {
			m.setupInputs[index].Blur()
		}
	}
	return m, tea.Batch(commands...)
}

func (m *model) updateSetupAccountValue() {
	m.setup.Account.Username = m.setupInputs[0].Value()
	m.setup.Account.Password = m.setupInputs[1].Value()
	m.setup.Account.APIKey = m.setupInputs[2].Value()
	m.setup.Account.APISecret = m.setupInputs[3].Value()
}

func (m *model) moveSetupCursor(delta int) {
	count := 1
	switch m.setup.Page {
	case setup.PageFont:
		count = len(m.setup.Fonts) + 2
	case setup.PageStorage:
		count = len(credentialChoices())
	case setup.PageScrobbling, setup.PageInterface:
		count = 3
	}
	m.setup.Field = (m.setup.Field + delta + count) % count
	if m.setup.Page == setup.PageFont && m.setup.Field < len(m.setup.Fonts) {
		m.setup.FontIndex = m.setup.Field
	}
	if m.setup.Page == setup.PageStorage {
		m.setup.CredentialSource = credentialChoices()[m.setup.Field].value
	}
}

func (m *model) toggleSetupValue() {
	m.adjustSetupValue(1)
}

func (m *model) adjustSetupValue(direction int) {
	switch m.setup.Page {
	case setup.PageFont:
		switch m.setup.Field {
		case len(m.setup.Fonts):
			m.setup.InstallFont = !m.setup.InstallFont
		case len(m.setup.Fonts) + 1:
			m.setup.TerminalDefault = !m.setup.TerminalDefault
		}
	case setup.PageScrobbling:
		switch m.setup.Field {
		case 0:
			m.setup.Recommended = !m.setup.Recommended
		case 1:
			m.setup.Loop += direction
			if m.setup.Loop > 9 {
				m.setup.Loop = 1
			} else if m.setup.Loop < 1 {
				m.setup.Loop = 9
			}
		case 2:
			m.setup.Interval += time.Duration(direction) * 500 * time.Millisecond
			if m.setup.Interval > 10*time.Second {
				m.setup.Interval = 500 * time.Millisecond
			} else if m.setup.Interval < 500*time.Millisecond {
				m.setup.Interval = 10 * time.Second
			}
		}
	case setup.PageInterface:
		switch m.setup.Field {
		case 0:
			m.setup.MouseEnabled = !m.setup.MouseEnabled
		case 1:
			m.setup.CompactHeader = !m.setup.CompactHeader
		case 2:
			m.setup.Notify = !m.setup.Notify
		}
	}
}

func (m model) setupContinue() (tea.Model, tea.Cmd) {
	switch m.setup.Page {
	case setup.PageWelcome:
		m.setup.Page = setup.PageSystem
	case setup.PageSystem:
		m.setup.Page = setup.PageFont
	case setup.PageFont:
		if m.setup.SelectedFont().Name == "Browse all Nerd Fonts..." {
			return m, openNerdFontsURL()
		}
		m.setup.Page = setup.PageAccount
		m.setup.Field = 0
		return m.focusSetupInput()
	case setup.PageAccount:
		m.updateSetupAccountValue()
		m.setup.Page = setup.PageStorage
		m.setup.Field = credentialChoiceIndex(m.setup.CredentialSource)
	case setup.PageStorage:
		m.setup.Page = setup.PageScrobbling
		m.setup.Field = 0
	case setup.PageScrobbling:
		m.setup.Page = setup.PageInterface
		m.setup.Field = 0
	case setup.PageInterface:
		m.setup.Page = setup.PageReview
	case setup.PageReview:
		m.setup.Page = setup.PageApply
		m.setup.Applying = true
		return m, tea.Batch(m.setupApplyCommand(), m.spinner.Tick)
	case setup.PageApply:
		if m.setup.ApplyResult.Error != nil {
			m.setup.Applying = true
			return m, tea.Batch(m.setupApplyCommand(), m.spinner.Tick)
		}
	case setup.PageComplete:
		m.stage = stageInput
		m.modeChoice = ""
		m.setupOriginal = config.Config{}
	}
	return m, nil
}

func credentialChoiceIndex(value string) int {
	for index, choice := range credentialChoices() {
		if choice.value == value {
			return index
		}
	}
	return 0
}

func (m model) setupPrevious() (tea.Model, tea.Cmd) {
	if m.setup.Page == setup.PageWelcome {
		m.cfg = m.setupOriginal
		m.client = lastfm.New(m.cfg.APIKey, m.cfg.APISecret, m.cfg.Username, m.cfg.Password, m.cfg.SessionKey)
		m.stage = stageInput
		m.modeChoice = ""
		return m, nil
	}
	m.setup.Page--
	if m.setup.Page == setup.PageAccount {
		m.setup.Field = len(m.setupInputs) - 1
		return m.focusSetupInput()
	}
	m.setup.Field = 0
	return m, nil
}

func (m model) setupApplyCommand() tea.Cmd {
	state := m.setup
	return func() tea.Msg {
		result := setup.Apply(context.Background(), state, setup.ApplyHooks{
			SaveConfig: config.Save,
			TestConnect: func(ctx context.Context, cfg config.Config) (string, string, error) {
				client := lastfm.New(cfg.APIKey, cfg.APISecret, cfg.Username, cfg.Password, cfg.SessionKey)
				report := connection.TestWithoutPersistence(ctx, cfg, client)
				connectionStatus, authenticationStatus := "failed", "failed"
				for _, item := range report.Items {
					if item.Label == "READ API" && item.OK {
						connectionStatus = "connected"
					}
					if item.Label == "AUTH" && item.OK {
						authenticationStatus = "ready"
					}
				}
				if !report.OK() {
					return connectionStatus, authenticationStatus, fmt.Errorf("Last.fm connection test failed")
				}
				return connectionStatus, authenticationStatus, nil
			},
		})
		return setupApplyMsg{result: result, config: state.ApplyConfig()}
	}
}

func openNerdFontsURL() tea.Cmd {
	return func() tea.Msg {
		return headerURLMsg{err: platform.OpenURL("https://github.com/ryanoasis/nerd-fonts/releases/latest")}
	}
}
