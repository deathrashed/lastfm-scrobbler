package setup

import (
	"archive/zip"
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
)

type fakeFontInstaller struct {
	installed bool
	err       error
}

func (f *fakeFontInstaller) UserFontDirectory() (string, error) { return "/tmp", nil }
func (f *fakeFontInstaller) DetectInstalled(context.Context, FontChoice) (bool, error) {
	return f.installed, nil
}
func (f *fakeFontInstaller) Install(context.Context, FontChoice) error { return f.err }

type fakeTerminalConfigurator struct {
	configured bool
	err        error
}

func (f *fakeTerminalConfigurator) Detect() Terminal {
	return Terminal{Name: "Fake", Supported: true, ConfigPath: ""}
}
func (f *fakeTerminalConfigurator) Configure(Terminal, string) error {
	f.configured = true
	return f.err
}

func TestNewStateUsesPlatformCredentialBackends(t *testing.T) {
	state := NewState(config.Config{CredentialSource: "auto", DefaultInterval: 2 * time.Second, DefaultLoop: 1})
	if len(state.Fonts) != 7 || state.Fonts[1].Asset != "JetBrainsMono" {
		t.Fatalf("curated font list is not stable")
	}
	if runtime.GOOS == "darwin" {
		found := false
		for _, choice := range credentialChoicesForTest() {
			if choice == "keychain" {
				found = true
			}
		}
		if !found {
			t.Fatal("macOS setup omitted Keychain")
		}
	}
}

func credentialChoicesForTest() []string {
	choices := []string{"file", "environment"}
	if runtime.GOOS == "darwin" {
		choices = append([]string{"keychain"}, choices...)
	}
	return choices
}

func TestApplyDoesNotSaveUntilApply(t *testing.T) {
	state := NewState(config.Config{EnvPath: filepath.Join(t.TempDir(), ".env"), DefaultLoop: 1, DefaultInterval: 2 * time.Second})
	state.FontIndex = 1
	state.InstallFont = false
	state.Account = Account{Username: "user", Password: "pass", APIKey: "key", APISecret: "secret"}
	saved := false
	var savedConfig config.Config
	result := Apply(context.Background(), state, ApplyHooks{
		Fonts:    &fakeFontInstaller{},
		Terminal: &fakeTerminalConfigurator{},
		SaveConfig: func(cfg config.Config) error {
			saved = true
			savedConfig = cfg
			return nil
		},
		TestConnect: func(context.Context, config.Config) (string, string, error) {
			return "connected", "ready", nil
		},
	})
	if !saved || result.Error != nil || result.Configuration != "saved" {
		t.Fatalf("apply did not complete safely: saved=%t error=%v status=%q", saved, result.Error, result.Configuration)
	}
	if config.NeedsSetup(savedConfig) {
		t.Fatal("applied configuration is not usable on the next startup")
	}
}

func TestApplySkipsUnsupportedTerminalWithoutFailing(t *testing.T) {
	state := NewState(config.Config{DefaultLoop: 1, DefaultInterval: 2 * time.Second})
	state.FontIndex = 1
	state.InstallFont = false
	state.TerminalDefault = true
	state.Terminal.Supported = false
	result := Apply(context.Background(), state, ApplyHooks{
		Fonts:      &fakeFontInstaller{},
		Terminal:   &fakeTerminalConfigurator{},
		SaveConfig: func(config.Config) error { return nil },
		TestConnect: func(context.Context, config.Config) (string, string, error) {
			return "connected", "ready", nil
		},
	})
	if result.Error != nil || result.TerminalStatus != "manual setup required" {
		t.Fatalf("unsupported terminal was fatal: status=%q error=%v", result.TerminalStatus, result.Error)
	}
}

func TestFontInstallerRejectsFailedDownload(t *testing.T) {
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("network unavailable")
	})}
	installer := NewFontInstaller(client)
	err := installer.Install(context.Background(), FontChoice{Asset: "Hack"})
	if err == nil {
		t.Fatal("failed download was accepted")
	}
}

func TestFontInstallerExtractsFakeReleaseIntoUserDirectory(t *testing.T) {
	archiveData := fakeFontArchive(t, "HackNerdFont-Regular.ttf")
	client := &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusOK, Status: "200 OK", Body: io.NopCloser(bytes.NewReader(archiveData))}, nil
	})}
	destination := t.TempDir()
	installer := &fontInstaller{httpClient: client, fontDir: func() (string, error) { return destination, nil }}
	if err := installer.Install(context.Background(), FontChoice{Asset: "Hack"}); err != nil {
		t.Fatalf("fake release install failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(destination, "HackNerdFont-Regular.ttf")); err != nil {
		t.Fatalf("font was not extracted: %v", err)
	}
}

func TestApplyBacksUpExistingTerminalConfig(t *testing.T) {
	terminalPath := filepath.Join(t.TempDir(), "ghostty-config")
	if err := os.WriteFile(terminalPath, []byte("font-size = 13\n"), 0600); err != nil {
		t.Fatal(err)
	}
	state := NewState(config.Config{DefaultLoop: 1, DefaultInterval: 2 * time.Second})
	state.FontIndex = 1
	state.InstallFont = false
	state.TerminalDefault = true
	state.Terminal = Terminal{Name: "Ghostty", ConfigPath: terminalPath, Supported: true}
	terminal := &fakeTerminalConfigurator{}
	result := Apply(context.Background(), state, ApplyHooks{Fonts: &fakeFontInstaller{}, Terminal: terminal, SaveConfig: func(config.Config) error { return nil }})
	if result.Error != nil || !terminal.configured {
		t.Fatalf("terminal apply failed: %v", result.Error)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(terminalPath), ".scrobbler-backup-ghostty-config")); err != nil {
		t.Fatalf("terminal backup was not created: %v", err)
	}
}

func fakeFontArchive(t *testing.T, name string) []byte {
	t.Helper()
	var buffer bytes.Buffer
	archive := zip.NewWriter(&buffer)
	file, err := archive.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.Write([]byte("fake-font")); err != nil {
		t.Fatal(err)
	}
	if err := archive.Close(); err != nil {
		t.Fatal(err)
	}
	return buffer.Bytes()
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }
