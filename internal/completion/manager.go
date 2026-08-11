package completion

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Status string

const (
	StatusInstalled        Status = "installed"
	StatusAlreadyInstalled Status = "already installed"
	StatusUpdated          Status = "updated"
	StatusUpdateAvailable  Status = "update available"
	StatusNotInstalled     Status = "not installed"
	StatusManual           Status = "manual configuration required"
)

type InstallResult struct {
	Shell  Shell
	Status Status
	Path   string
}

type Manager struct {
	Home string
	GOOS string
}

func NewManager(home, goos string) Manager {
	if strings.TrimSpace(home) == "" {
		home, _ = os.UserHomeDir()
	}
	if strings.TrimSpace(goos) == "" {
		goos = runtime.GOOS
	}
	return Manager{Home: home, GOOS: goos}
}

func DefaultManager() Manager {
	home, _ := os.UserHomeDir()
	return NewManager(home, runtime.GOOS)
}

func DetectShell() Shell {
	for _, key := range []string{"SHELL", "ComSpec", "PSShell"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			if shell, err := parseShellExecutable(value); err == nil {
				return shell
			}
		}
	}
	if runtime.GOOS == "windows" {
		return ShellPowerShell
	}
	return ShellBash
}

func parseShellExecutable(value string) (Shell, error) {
	value = strings.TrimSpace(value)
	if separator := strings.LastIndexAny(value, `/\`); separator >= 0 {
		value = value[separator+1:]
	}
	value = strings.TrimSuffix(strings.ToLower(value), ".exe")
	return ParseShell(value)
}

func (m Manager) Path(shell Shell) string {
	shell, _ = ParseShell(shell.String())
	switch shell {
	case ShellZsh:
		return filepath.Join(m.Home, ".zfunc", "_scrobbler")
	case ShellBash:
		return filepath.Join(m.Home, ".local", "share", "bash-completion", "completions", "scrobbler")
	case ShellFish:
		return filepath.Join(m.Home, ".config", "fish", "completions", "scrobbler.fish")
	case ShellPowerShell:
		return filepath.Join(m.Home, ".config", "powershell", "completions", "scrobbler.ps1")
	default:
		return ""
	}
}

func (m Manager) ProfilePath(shell Shell) string {
	shell, _ = ParseShell(shell.String())
	switch shell {
	case ShellZsh:
		return filepath.Join(m.Home, ".zshrc")
	case ShellPowerShell:
		if m.GOOS == "windows" {
			return filepath.Join(m.Home, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		}
		return filepath.Join(m.Home, ".config", "powershell", "Microsoft.PowerShell_profile.ps1")
	default:
		return ""
	}
}

func (m Manager) Status(shell Shell) Status {
	parsed, err := ParseShell(shell.String())
	if err != nil {
		return StatusManual
	}
	path := m.Path(parsed)
	data, err := os.ReadFile(path)
	if err != nil {
		return StatusNotInstalled
	}
	generated, err := Generate(parsed.String())
	if err != nil || string(data) != generated {
		return StatusUpdateAvailable
	}
	profile := m.ProfilePath(parsed)
	if profile != "" {
		profileData, readErr := os.ReadFile(profile)
		if readErr != nil || !bytesContainMarker(profileData) {
			return StatusManual
		}
	}
	return StatusInstalled
}

func (m Manager) InstallName(name string) (InstallResult, error) {
	shell := DetectShell()
	if strings.TrimSpace(name) != "" {
		parsed, err := ParseShell(name)
		if err != nil {
			return InstallResult{}, err
		}
		shell = parsed
	}
	return m.Install(shell)
}

func (m Manager) Install(shell Shell) (InstallResult, error) {
	parsed, err := ParseShell(shell.String())
	if err != nil {
		return InstallResult{}, err
	}
	script, err := Generate(parsed.String())
	if err != nil {
		return InstallResult{}, err
	}
	path := m.Path(parsed)
	if path == "" {
		return InstallResult{}, errors.New("completion path is unavailable")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return InstallResult{}, fmt.Errorf("create completion directory: %w", err)
	}
	previous, readErr := os.ReadFile(path)
	if readErr != nil && !os.IsNotExist(readErr) {
		return InstallResult{}, fmt.Errorf("read completion file: %w", readErr)
	}
	if err := os.WriteFile(path, []byte(script), 0600); err != nil {
		return InstallResult{}, fmt.Errorf("write completion file: %w", err)
	}
	if profile := m.ProfilePath(parsed); profile != "" {
		if err := ensureProfile(profile, parsed, path); err != nil {
			return InstallResult{}, err
		}
	}
	result := StatusInstalled
	switch {
	case m.Status(parsed) == StatusManual:
		result = StatusManual
	case len(previous) == 0:
		result = StatusInstalled
	case string(previous) == script:
		result = StatusAlreadyInstalled
	default:
		result = StatusUpdated
	}
	return InstallResult{Shell: parsed, Status: result, Path: path}, nil
}

const profileMarker = "# lastfm-scrobbler completion"

func ensureProfile(path string, shell Shell, completionPath string) error {
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read shell profile: %w", err)
	}
	if bytesContainMarker(data) {
		return nil
	}
	if len(data) > 0 {
		backup := path + ".scrobbler-backup"
		if _, statErr := os.Stat(backup); os.IsNotExist(statErr) {
			if err := os.WriteFile(backup, data, 0600); err != nil {
				return fmt.Errorf("backup shell profile: %w", err)
			}
		}
	}
	line := profileLine(shell, completionPath)
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		data = append(data, '\n')
	}
	data = append(data, []byte(line+"\n")...)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create shell profile directory: %w", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("write shell profile: %w", err)
	}
	return nil
}

func profileLine(shell Shell, path string) string {
	switch shell {
	case ShellZsh:
		return fmt.Sprintf("fpath=(%s $fpath); autoload -Uz compinit && compinit %s", shellQuote(path), profileMarker)
	case ShellPowerShell:
		return fmt.Sprintf(". %s %s", shellQuote(path), profileMarker)
	default:
		return ""
	}
}

func shellQuote(path string) string {
	return "'" + strings.ReplaceAll(path, "'", "'\\''") + "'"
}

func bytesContainMarker(data []byte) bool {
	return strings.Contains(string(data), profileMarker)
}
