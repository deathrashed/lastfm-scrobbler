package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

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
		options = append(options, tea.WithMouseCellMotion())
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
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		loaded, err := client.GetAlbumTracks(ctx, artist, album)
		cancel()
		if err != nil {
			fmt.Println("Error:", err)
			continue
		}
		ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
		err = client.Authenticate(ctx)
		cancel()
		if err != nil {
			fmt.Println("Authentication:", err)
			continue
		}
		limit := cfg.DefaultLimit
		if limit <= 0 || limit > len(loaded.Tracks) {
			limit = len(loaded.Tracks)
		}
		for loop := 0; loop < max(1, cfg.DefaultLoop); loop++ {
			for index, track := range loaded.Tracks[:limit] {
				ctx, cancel = context.WithTimeout(context.Background(), 45*time.Second)
				err = client.Scrobble(ctx, loaded.Artist, track.Title, loaded.Title, time.Now().Unix())
				cancel()
				if err != nil {
					fmt.Printf("Failed: %s — %v\n", track.Title, err)
				} else {
					fmt.Printf("[%d/%d] %s\n", index+1, limit, track.Title)
				}
				if index < limit-1 && cfg.DefaultInterval > 0 {
					time.Sleep(cfg.DefaultInterval)
				}
			}
		}
		fmt.Println("Done.")
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

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
