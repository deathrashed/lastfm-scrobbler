package setup

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type Terminal struct {
	Name        string
	ConfigPath  string
	Supported   bool
	CurrentFont string
}

type TerminalConfigurator interface {
	Detect() Terminal
	Configure(Terminal, string) error
}

type terminalConfigurator struct{}

func NewTerminalConfigurator() TerminalConfigurator { return terminalConfigurator{} }

func (terminalConfigurator) Detect() Terminal {
	name := detectTerminalName()
	terminal := Terminal{Name: name}
	if strings.EqualFold(name, "Ghostty") {
		terminal.ConfigPath = ghosttyConfigPath()
		terminal.Supported = true
	}
	return terminal
}

func (terminalConfigurator) Configure(terminal Terminal, family string) error {
	if !terminal.Supported {
		return fmt.Errorf("automatic font configuration is not supported for %s", terminal.Name)
	}
	if strings.TrimSpace(family) == "" {
		return fmt.Errorf("terminal font family is empty")
	}
	if err := os.MkdirAll(filepath.Dir(terminal.ConfigPath), 0700); err != nil {
		return err
	}
	data, err := os.ReadFile(terminal.ConfigPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(data)
	lines := strings.Split(strings.ReplaceAll(content, "\r\n", "\n"), "\n")
	replaced := false
	for index, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "font-family") {
			lines[index] = "font-family = " + family
			replaced = true
		}
	}
	if !replaced {
		if content != "" && !strings.HasSuffix(content, "\n") {
			content += "\n"
		}
		content += "font-family = " + family + "\n"
	} else {
		content = strings.Join(lines, "\n")
	}
	return os.WriteFile(terminal.ConfigPath, []byte(content), 0600)
}

func detectTerminalName() string {
	if strings.TrimSpace(os.Getenv("WT_SESSION")) != "" {
		return "Windows Terminal"
	}
	if strings.TrimSpace(os.Getenv("GHOSTTY_RESOURCES_DIR")) != "" || os.Getenv("TERM_PROGRAM") == "ghostty" {
		return "Ghostty"
	}
	if strings.TrimSpace(os.Getenv("KITTY_WINDOW_ID")) != "" {
		return "kitty"
	}
	if strings.TrimSpace(os.Getenv("WEZTERM_EXECUTABLE")) != "" {
		return "WezTerm"
	}
	if value := strings.TrimSpace(os.Getenv("TERM_PROGRAM")); value != "" {
		return value
	}
	return "unknown"
}

func ghosttyConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "config"
	}
	if runtime.GOOS == "darwin" {
		return filepath.Join(home, "Library", "Application Support", "com.mitchellh.ghostty", "config")
	}
	if runtime.GOOS == "windows" {
		base := os.Getenv("APPDATA")
		if base == "" {
			base = filepath.Join(home, "AppData", "Roaming")
		}
		return filepath.Join(base, "ghostty", "config")
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "ghostty", "config")
}
