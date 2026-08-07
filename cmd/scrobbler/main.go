package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/deathrashed/lastfm-scrobbler/internal/cli"
	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/logging"
	"github.com/deathrashed/lastfm-scrobbler/internal/tui"
	"github.com/deathrashed/lastfm-scrobbler/internal/version"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}
	_ = logging.Init(config.DataDir())
	defer logging.Close()
	logging.Event("startup", map[string]any{"version": version.Version, "profile": cfg.Profile})

	client := lastfm.New(cfg.APIKey, cfg.APISecret, cfg.Username, cfg.Password, cfg.SessionKey)
	args := os.Args[1:]
	if len(args) > 0 && args[0] == "--no-tui" {
		runPlainMode(cfg, client)
		return
	}
	if cli.IsCommand(args) {
		code := cli.Run(args, cfg, client, os.Stdout, os.Stderr)
		if code >= 0 {
			os.Exit(code)
		}
	}
	if os.Getenv("NO_TUI") == "1" || os.Getenv("TERM") == "dumb" {
		runPlainMode(cfg, client)
		return
	}

	m := tui.New(cfg, client)
	options := []tea.ProgramOption{tea.WithAltScreen()}
	if cfg.MouseEnabled {
		options = append(options, tea.WithMouseAllMotion())
	}
	p := tea.NewProgram(m, options...)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		select {
		case <-sigCh:
			logging.Printf("received termination signal")
			cancel()
			p.Send(tea.Quit())
		case <-ctx.Done():
		}
	}()

	if _, err := p.Run(); err != nil {
		logging.Printf("TUI failure: %v", err)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// runPlainMode remains as a small interactive fallback for dumb terminals.
// Automation should use the documented manual/file/discography subcommands.
func runPlainMode(cfg config.Config, client lastfm.Client) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Scrobbler (plain mode)")
	fmt.Println("Enter Artist - Album, or q to quit.")
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			return
		}
		line = strings.TrimSpace(line)
		if line == "q" || line == "quit" {
			return
		}
		artist, album, ok := splitArtistAlbum(line)
		if !ok {
			fmt.Println("Use: Artist - Album")
			continue
		}
		code := cli.Run([]string{"manual", "--artist", artist, "--album", album}, cfg, client, os.Stdout, os.Stderr)
		if code != 0 {
			fmt.Printf("Command exited with status %d\n", code)
		}
	}
}

func splitArtistAlbum(value string) (string, string, bool) {
	for _, separator := range []string{" — ", " - ", "\t", "|"} {
		parts := strings.SplitN(value, separator, 2)
		if len(parts) == 2 {
			artist, album := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if artist != "" && album != "" {
				return artist, album, true
			}
		}
	}
	return "", "", false
}
