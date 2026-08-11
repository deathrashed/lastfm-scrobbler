package setup

import (
	"bufio"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type Environment struct {
	OperatingSystem string
	Distribution    string
	Architecture    string
	Terminal        string
	PackageManager  string
}

func DetectEnvironment() Environment {
	env := Environment{
		OperatingSystem: operatingSystem(runtime.GOOS),
		Architecture:    runtime.GOARCH,
		Terminal:        detectTerminal(),
		PackageManager:  detectPackageManager(),
	}
	if runtime.GOOS == "linux" {
		env.Distribution = detectDistribution()
	}
	return env
}

func operatingSystem(goos string) string {
	switch goos {
	case "darwin":
		return "macOS"
	case "windows":
		return "Windows"
	case "linux":
		return "Linux"
	default:
		return "unknown"
	}
}

func detectTerminal() string {
	if strings.TrimSpace(os.Getenv("WT_SESSION")) != "" {
		return "Windows Terminal"
	}
	if strings.TrimSpace(os.Getenv("GHOSTTY_RESOURCES_DIR")) != "" || os.Getenv("TERM_PROGRAM") == "ghostty" {
		return "Ghostty"
	}
	if strings.TrimSpace(os.Getenv("KITTY_WINDOW_ID")) != "" || strings.HasPrefix(os.Getenv("TERM"), "xterm-kitty") {
		return "kitty"
	}
	if strings.TrimSpace(os.Getenv("WEZTERM_EXECUTABLE")) != "" || os.Getenv("TERM_PROGRAM") == "WezTerm" {
		return "WezTerm"
	}
	if strings.TrimSpace(os.Getenv("ALACRITTY_SOCKET")) != "" || os.Getenv("TERM_PROGRAM") == "Alacritty" {
		return "Alacritty"
	}
	if value := strings.TrimSpace(os.Getenv("TERM_PROGRAM")); value != "" {
		return value
	}
	if value := strings.TrimSpace(os.Getenv("TERM")); value != "" {
		return value
	}
	return "unknown"
}

func detectPackageManager() string {
	candidates := []struct {
		name string
		bin  string
	}{
		{"Homebrew", "brew"},
		{"winget", "winget"},
		{"Chocolatey", "choco"},
		{"Scoop", "scoop"},
		{"dnf", "dnf"},
		{"apt", "apt"},
		{"pacman", "pacman"},
	}
	for _, candidate := range candidates {
		if _, err := exec.LookPath(candidate.bin); err == nil {
			return candidate.name
		}
	}
	return "unknown"
}

func detectDistribution() string {
	file, err := os.Open("/etc/os-release")
	if err != nil {
		return "unknown"
	}
	defer file.Close()
	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}
		values[key] = strings.Trim(strings.TrimSpace(value), "\"")
	}
	if value := strings.TrimSpace(values["PRETTY_NAME"]); value != "" {
		return value
	}
	if value := strings.TrimSpace(values["NAME"]); value != "" {
		return value
	}
	return "unknown"
}
