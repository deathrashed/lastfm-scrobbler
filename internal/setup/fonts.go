package setup

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const nerdFontsDownloadBase = "https://github.com/ryanoasis/nerd-fonts/releases/latest/download/"

type FontChoice struct {
	Name   string
	Asset  string
	Family string
}

var curatedFonts = []FontChoice{
	{Name: "Keep my current font"},
	{Name: "JetBrainsMono Nerd Font Mono", Asset: "JetBrainsMono", Family: "JetBrainsMono Nerd Font Mono"},
	{Name: "Hack Nerd Font Mono", Asset: "Hack", Family: "Hack Nerd Font Mono"},
	{Name: "FiraCode Nerd Font Mono", Asset: "FiraCode", Family: "FiraCode Nerd Font Mono"},
	{Name: "Meslo Nerd Font Mono", Asset: "Meslo", Family: "Meslo Nerd Font Mono"},
	{Name: "CaskaydiaMono Nerd Font", Asset: "CaskaydiaMono", Family: "CaskaydiaMono Nerd Font"},
	{Name: "Browse all Nerd Fonts..."},
}

type FontInstaller interface {
	UserFontDirectory() (string, error)
	DetectInstalled(context.Context, FontChoice) (bool, error)
	Install(context.Context, FontChoice) error
}

type fontInstaller struct {
	httpClient *http.Client
	fontDir    func() (string, error)
}

func NewFontInstaller(client *http.Client) FontInstaller {
	if client == nil {
		client = &http.Client{}
	}
	return &fontInstaller{httpClient: client, fontDir: userFontDirectory}
}

func (f *fontInstaller) UserFontDirectory() (string, error) { return f.fontDir() }

func (f *fontInstaller) DetectInstalled(ctx context.Context, choice FontChoice) (bool, error) {
	if choice.Asset == "" {
		return false, nil
	}
	dir, err := f.UserFontDirectory()
	if err != nil {
		return false, err
	}
	entries, err := os.ReadDir(dir)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	want := strings.ToLower(choice.Asset)
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if strings.Contains(strings.ToLower(entry.Name()), want) {
			select {
			case <-ctx.Done():
				return false, ctx.Err()
			default:
			}
			return true, nil
		}
	}
	return false, nil
}

func (f *fontInstaller) Install(ctx context.Context, choice FontChoice) error {
	if choice.Asset == "" {
		return nil
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, nerdFontsDownloadBase+choice.Asset+".zip", nil)
	if err != nil {
		return fmt.Errorf("create Nerd Font request: %w", err)
	}
	response, err := f.httpClient.Do(request)
	if err != nil {
		return fmt.Errorf("download Nerd Font: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download Nerd Font: HTTP %s", response.Status)
	}
	archivePath, err := os.CreateTemp("", "scrobbler-font-*.zip")
	if err != nil {
		return fmt.Errorf("create font archive: %w", err)
	}
	archiveName := archivePath.Name()
	defer os.Remove(archiveName)
	if _, err := io.Copy(archivePath, response.Body); err != nil {
		archivePath.Close()
		return fmt.Errorf("download Nerd Font archive: %w", err)
	}
	if err := archivePath.Close(); err != nil {
		return fmt.Errorf("close Nerd Font archive: %w", err)
	}
	dir, err := f.UserFontDirectory()
	if err != nil {
		return err
	}
	return extractFontArchive(ctx, archiveName, choice, dir)
}

func extractFontArchive(ctx context.Context, archiveName string, choice FontChoice, dir string) error {
	archive, err := zip.OpenReader(archiveName)
	if err != nil {
		return fmt.Errorf("open Nerd Font archive: %w", err)
	}
	defer archive.Close()
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create user font directory: %w", err)
	}
	installed := false
	for _, file := range archive.File {
		if err := contextErr(ctx); err != nil {
			return err
		}
		if file.FileInfo().IsDir() || !isFontFile(file.Name) || !strings.Contains(strings.ToLower(file.Name), strings.ToLower(choice.Asset)) {
			continue
		}
		if err := extractZipFont(file, dir); err != nil {
			return err
		}
		installed = true
	}
	if !installed {
		return fmt.Errorf("Nerd Font archive did not contain %s font files", choice.Asset)
	}
	return refreshFontCache(dir)
}

func extractZipFont(file *zip.File, dir string) error {
	reader, err := file.Open()
	if err != nil {
		return err
	}
	defer reader.Close()
	name := filepath.Base(file.Name)
	destination := filepath.Join(dir, name)
	output, err := os.OpenFile(destination, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("create font file: %w", err)
	}
	if _, err := io.Copy(output, reader); err != nil {
		output.Close()
		return fmt.Errorf("install font file: %w", err)
	}
	if err := output.Close(); err != nil {
		return err
	}
	return registerInstalledFont(destination)
}

func refreshFontCache(dir string) error {
	if runtime.GOOS == "linux" {
		if fcCache, err := exec.LookPath("fc-cache"); err == nil {
			if err := exec.Command(fcCache, "-f", dir).Run(); err != nil {
				return fmt.Errorf("refresh font cache: %w", err)
			}
		}
	}
	return nil
}

func userFontDirectory() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Fonts"), nil
	case "windows":
		if local := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); local != "" {
			return filepath.Join(local, "Microsoft", "Windows", "Fonts"), nil
		}
		return filepath.Join(home, "AppData", "Local", "Microsoft", "Windows", "Fonts"), nil
	default:
		return filepath.Join(home, ".local", "share", "fonts"), nil
	}
}

func isFontFile(name string) bool {
	ext := strings.ToLower(filepath.Ext(name))
	return ext == ".ttf" || ext == ".otf"
}

func contextErr(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		return nil
	}
}
