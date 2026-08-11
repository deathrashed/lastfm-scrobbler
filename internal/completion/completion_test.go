package completion

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateSupportsEveryShell(t *testing.T) {
	for _, shell := range Shells {
		script, err := Generate(shell.String())
		if err != nil {
			t.Fatalf("Generate(%q): %v", shell, err)
		}
		if strings.TrimSpace(script) == "" {
			t.Fatalf("Generate(%q) returned an empty script", shell)
		}
	}
}

func TestParseShellAcceptsPowerShellAlias(t *testing.T) {
	got, err := ParseShell("pwsh")
	if err != nil || got != ShellPowerShell {
		t.Fatalf("ParseShell(pwsh) = %q, %v", got, err)
	}
}

func TestManagerInstallIsIdempotentAndBacksUpProfiles(t *testing.T) {
	home := t.TempDir()
	profile := filepath.Join(home, ".zshrc")
	if err := os.WriteFile(profile, []byte("export PATH=\"$PATH\"\n"), 0600); err != nil {
		t.Fatal(err)
	}
	manager := NewManager(home, "darwin")

	first, err := manager.Install(ShellZsh)
	if err != nil || first.Status != StatusInstalled {
		t.Fatalf("first install = %+v, %v", first, err)
	}
	second, err := manager.Install(ShellZsh)
	if err != nil || second.Status != StatusAlreadyInstalled {
		t.Fatalf("second install = %+v, %v", second, err)
	}
	profileData, err := os.ReadFile(profile)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Count(string(profileData), profileMarker) != 1 {
		t.Fatalf("profile marker count = %d, want 1", strings.Count(string(profileData), profileMarker))
	}
	if _, err := os.Stat(profile + ".scrobbler-backup"); err != nil {
		t.Fatalf("profile backup missing: %v", err)
	}
}

func TestManagerReportsManualProfileConfiguration(t *testing.T) {
	home := t.TempDir()
	manager := NewManager(home, "darwin")
	path := manager.Path(ShellZsh)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		t.Fatal(err)
	}
	script, err := Generate("zsh")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(script), 0600); err != nil {
		t.Fatal(err)
	}
	if got := manager.Status(ShellZsh); got != StatusManual {
		t.Fatalf("Status(zsh) = %q, want manual configuration required", got)
	}
}

func TestDetectShellParsesExecutablePath(t *testing.T) {
	for value, want := range map[string]Shell{
		"/usr/local/bin/zsh":                      ShellZsh,
		`C:\\Program Files\\PowerShell\\pwsh.exe`: ShellPowerShell,
	} {
		t.Setenv("SHELL", value)
		if got := DetectShell(); got != want {
			t.Fatalf("DetectShell(%q) = %q, want %q", value, got, want)
		}
	}
}
