package setup

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
)

type ApplyHooks struct {
	Fonts       FontInstaller
	Terminal    TerminalConfigurator
	SaveConfig  func(config.Config) error
	TestConnect func(context.Context, config.Config) (string, string, error)
}

func Apply(ctx context.Context, state State, hooks ApplyHooks) ApplyResult {
	result := ApplyResult{}
	cfg := state.ApplyConfig()
	if hooks.SaveConfig == nil {
		hooks.SaveConfig = config.Save
	}
	if hooks.Fonts == nil {
		hooks.Fonts = NewFontInstaller(nil)
	}
	if hooks.Terminal == nil {
		hooks.Terminal = NewTerminalConfigurator()
	}

	choice := state.SelectedFont()
	if !state.HasFont() || !state.InstallFont {
		result.FontStatus = "skipped"
	} else if err := hooks.Fonts.Install(ctx, choice); err != nil {
		result.FontStatus = "failed"
		result.Error = err
		return result
	} else {
		result.FontStatus = "installed"
	}

	terminalBackup := ""
	if state.TerminalDefault && state.HasFont() {
		if !state.Terminal.Supported {
			result.TerminalStatus = "manual setup required"
		} else if backup, err := backupTerminalConfig(state.Terminal.ConfigPath); err != nil {
			result.TerminalStatus = "failed"
			result.Error = err
			return result
		} else {
			terminalBackup = backup
			if err := hooks.Terminal.Configure(state.Terminal, choice.Family); err != nil {
				result.TerminalStatus = "failed"
				result.Error = err
				return result
			}
			result.TerminalStatus = "configured"
		}
	} else {
		result.TerminalStatus = "skipped"
	}

	if err := hooks.SaveConfig(cfg); err != nil {
		result.Configuration = "failed"
		result.Error = err
		if terminalBackup != "" {
			_ = restoreTerminalConfig(state.Terminal.ConfigPath, terminalBackup)
		}
		return result
	}
	result.Configuration = "saved"
	if strings.EqualFold(cfg.CredentialSource, "environment") {
		result.Credentials = "environment variables"
	} else {
		result.Credentials = "stored securely"
	}

	if hooks.TestConnect == nil {
		result.Connection = "skipped"
		result.Authentication = "skipped"
		return result
	}
	connectionStatus, authenticationStatus, err := hooks.TestConnect(ctx, cfg)
	result.Connection = connectionStatus
	result.Authentication = authenticationStatus
	if err != nil {
		result.Error = err
	}
	return result
}

func backupTerminalConfig(path string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	backup := filepath.Join(filepath.Dir(path), ".scrobbler-backup-"+filepath.Base(path))
	if _, err := os.Stat(backup); err == nil {
		return backup, nil
	}
	if err := os.WriteFile(backup, data, 0600); err != nil {
		return "", fmt.Errorf("backup terminal configuration: %w", err)
	}
	return backup, nil
}

func restoreTerminalConfig(path, backup string) error {
	data, err := os.ReadFile(backup)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}
