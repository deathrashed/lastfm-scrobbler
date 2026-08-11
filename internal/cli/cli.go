package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/completion"
	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/connection"
	"github.com/deathrashed/lastfm-scrobbler/internal/diagnostics"
	"github.com/deathrashed/lastfm-scrobbler/internal/importer"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/logging"
	"github.com/deathrashed/lastfm-scrobbler/internal/platform"
	"github.com/deathrashed/lastfm-scrobbler/internal/scrobbler"
	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
	"github.com/deathrashed/lastfm-scrobbler/internal/updater"
	"github.com/deathrashed/lastfm-scrobbler/internal/version"
)

type Result struct {
	Command   string `json:"command"`
	Albums    int    `json:"albums"`
	Tracks    int    `json:"tracks"`
	Scrobbled int    `json:"scrobbled"`
	Failed    int    `json:"failed"`
	Skipped   int    `json:"skipped_duplicates,omitempty"`
	DryRun    bool   `json:"dry_run"`
	Message   string `json:"message,omitempty"`
}

type commonOptions struct {
	loop     int
	limit    int
	interval time.Duration
	dryRun   bool
	json     bool
}

func IsCommand(args []string) bool {
	if len(args) == 0 {
		return false
	}
	switch args[0] {
	case "tui", "setup", "manual", "file", "discography", "similar", "test", "diagnostics", "check-update", "completion", "version", "help", "--help", "-h", "--version":
		return true
	default:
		return strings.HasPrefix(args[0], "-")
	}
}

func Run(args []string, cfg config.Config, client lastfm.Client, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "tui" {
		return -1
	}
	command := args[0]
	arguments := args[1:]
	switch command {
	case "setup":
		return -1
	case "help", "--help", "-h":
		PrintHelp(stdout)
		return 0
	case "version", "--version":
		fmt.Fprintf(stdout, "%s (%s)\n", version.Version, version.Commit)
		return 0
	case "completion":
		if len(arguments) > 0 && arguments[0] == "install" {
			if len(arguments) > 2 {
				fmt.Fprintln(stderr, "usage: scrobbler completion install [zsh|bash|fish|powershell]")
				return 2
			}
			name := ""
			if len(arguments) == 2 {
				name = arguments[1]
			}
			result, err := completion.DefaultManager().InstallName(name)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 2
			}
			fmt.Fprintf(stdout, "%s %s completion at %s\n", result.Status, result.Shell, result.Path)
			return 0
		}
		if len(arguments) != 1 {
			fmt.Fprintln(stderr, "usage: scrobbler completion <zsh|bash|fish|powershell>")
			return 2
		}
		script, err := Completion(arguments[0])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return 2
		}
		fmt.Fprint(stdout, script)
		return 0
	case "test":
		return runConnectionTest(cfg, client, stdout, stderr, arguments)
	case "diagnostics":
		return runDiagnostics(cfg, stdout, stderr, arguments)
	case "check-update":
		return runUpdateCheck(cfg, stdout, stderr, arguments)
	case "manual":
		return runManual(cfg, client, stdout, stderr, arguments)
	case "file":
		return runFile(cfg, client, stdout, stderr, arguments)
	case "discography":
		return runDiscography(cfg, client, stdout, stderr, arguments)
	case "similar":
		return runSimilar(client, stdout, stderr, arguments)
	default:
		fmt.Fprintf(stderr, "unknown command %q\n\n", command)
		PrintHelp(stderr)
		return 2
	}
}

func PrintHelp(w io.Writer) {
	fmt.Fprint(w, `Last.fm Scrobbler

Usage:
  scrobbler                         Launch the styled TUI
  scrobbler setup                   Run the first-time setup wizard
  scrobbler manual [options]        Scrobble one Artist - Album
  scrobbler file [options] PATH     Load TXT/CSV/TSV/JSON/M3U/M3U8/folders
  scrobbler discography [options] ARTIST
  scrobbler similar ARTIST          List similar album suggestions
  scrobbler test [--json]           Test API lookup and authentication
  scrobbler diagnostics             Export a redacted diagnostics ZIP
  scrobbler check-update            Check the configured release endpoint
  scrobbler completion SHELL        Print zsh, bash, fish, or PowerShell completion
  scrobbler completion install      Install completion for the detected shell
  scrobbler version                 Print the application version

Common automation options:
  --loop N                          Album loop count
  --limit N                         Tracks per album; 0 means all
  --interval DURATION               Delay such as 2s or 500ms
  --dry-run                         Resolve and print without scrobbling
  --json                            Machine-readable output

Manual examples:
  scrobbler manual "Slayer - Hell Awaits"
  scrobbler manual --artist Slayer --album "Hell Awaits" --loop 2

Last.fm top-album examples:
  scrobbler discography "Demolition Hammer"            # list results
  scrobbler discography "Demolition Hammer" --all
  scrobbler discography "Demolition Hammer" --albums "Epidemic of Violence,Tortured Existence"
  scrobbler discography "Demolition Hammer" --first 3 --dry-run

Environment:
  API_KEY / LASTFM_API_KEY
  API_SECRET / LASTFM_API_SECRET
  LASTFM_USERNAME, LASTFM_PASSWORD, LASTFM_SESSION_KEY
  LASTFM_ENV_FILE, LASTFM_PROFILE, LASTFM_CREDENTIAL_SOURCE
  SCROBBLER_UPDATE_URL
`)
}

func PrintSetupHelp(w io.Writer) {
	fmt.Fprintln(w, "Usage: scrobbler setup")
	fmt.Fprintln(w, "Run the interactive first-run setup wizard.")
}

func addCommon(fs *flag.FlagSet, cfg config.Config, opts *commonOptions) {
	fs.IntVar(&opts.loop, "loop", max(1, cfg.DefaultLoop), "album loop count")
	fs.IntVar(&opts.limit, "limit", max(0, cfg.DefaultLimit), "tracks per album; 0 means all")
	fs.DurationVar(&opts.interval, "interval", cfg.DefaultInterval, "delay between scrobbles")
	fs.BoolVar(&opts.dryRun, "dry-run", false, "resolve queue without scrobbling")
	fs.BoolVar(&opts.json, "json", false, "machine-readable JSON output")
}

func runManual(cfg config.Config, client lastfm.Client, stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("manual", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var opts commonOptions
	var artist, album string
	addCommon(fs, cfg, &opts)
	fs.StringVar(&artist, "artist", "", "artist name")
	fs.StringVar(&album, "album", "", "album title")
	if err := fs.Parse(normalizeArgs(args, map[string]bool{"artist": true, "album": true, "loop": true, "limit": true, "interval": true})); err != nil {
		return 2
	}
	if (artist == "" || album == "") && fs.NArg() > 0 {
		joined := strings.Join(fs.Args(), " ")
		first, second, ok := splitArtistAlbum(joined)
		if ok {
			artist, album = first, second
		}
	}
	if strings.TrimSpace(artist) == "" || strings.TrimSpace(album) == "" {
		fmt.Fprintln(stderr, "manual requires --artist and --album, or one \"Artist - Album\" argument")
		return 2
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	loaded, err := client.GetAlbumTracks(ctx, artist, album)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return executeAlbums("manual", cfg, client, []lastfm.Album{loaded}, opts, stdout, stderr)
}

func runFile(cfg config.Config, client lastfm.Client, stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("file", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var opts commonOptions
	addCommon(fs, cfg, &opts)
	if err := fs.Parse(normalizeArgs(args, map[string]bool{"loop": true, "limit": true, "interval": true})); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		fmt.Fprintln(stderr, "usage: scrobbler file [options] PATH")
		return 2
	}
	targets, err := importer.Load(config.ExpandPath(fs.Arg(0)))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	albums, err := loadTargets(client, targets)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	return executeAlbums("file", cfg, client, albums, opts, stdout, stderr)
}

func runDiscography(cfg config.Config, client lastfm.Client, stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("discography", flag.ContinueOnError)
	fs.SetOutput(stderr)
	var opts commonOptions
	var all bool
	var names string
	var first int
	var clean bool
	addCommon(fs, cfg, &opts)
	fs.BoolVar(&all, "all", false, "select every returned album")
	fs.StringVar(&names, "albums", "", "comma-separated album titles")
	fs.IntVar(&first, "first", 0, "select the first N albums")
	fs.BoolVar(&clean, "clean", cfg.CleanDiscography, "remove obvious reissues and duplicates")
	if err := fs.Parse(normalizeArgs(args, map[string]bool{"loop": true, "limit": true, "interval": true, "albums": true, "first": true})); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(stderr, "usage: scrobbler discography [options] ARTIST")
		return 2
	}
	artist := strings.Join(fs.Args(), " ")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	found, err := client.GetDiscography(ctx, artist)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if clean {
		found = cleanAlbums(found)
	}
	if !all && strings.TrimSpace(names) == "" && first <= 0 {
		for i, album := range found {
			fmt.Fprintf(stdout, "%3d  %s\n", i+1, album.Title)
		}
		fmt.Fprintln(stdout, "\nSelect with --all, --first N, or --albums \"Title One,Title Two\".")
		return 0
	}
	selected := selectDiscography(found, all, names, first)
	if len(selected) == 0 {
		fmt.Fprintln(stderr, "no albums matched the requested selection")
		return 1
	}
	loaded := make([]lastfm.Album, 0, len(selected))
	for _, album := range selected {
		full, fetchErr := client.GetAlbumTracks(ctx, album.Artist, album.Title)
		if fetchErr != nil {
			fmt.Fprintf(stderr, "%s: %v\n", album.Title, fetchErr)
			return 1
		}
		loaded = append(loaded, full)
	}
	return executeAlbums("discography", cfg, client, loaded, opts, stdout, stderr)
}

func runSimilar(client lastfm.Client, stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("similar", flag.ContinueOnError)
	fs.SetOutput(stderr)
	limit := fs.Int("limit", 20, "maximum suggestions")
	jsonOut := fs.Bool("json", false, "machine-readable JSON")
	if err := fs.Parse(normalizeArgs(args, map[string]bool{"limit": true})); err != nil {
		return 2
	}
	if fs.NArg() < 1 {
		fmt.Fprintln(stderr, "usage: scrobbler similar [--limit N] ARTIST")
		return 2
	}
	artist := strings.Join(fs.Args(), " ")
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	albums, err := client.GetSimilarAlbums(ctx, artist, *limit)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *jsonOut {
		_ = json.NewEncoder(stdout).Encode(albums)
		return 0
	}
	for i, album := range albums {
		fmt.Fprintf(stdout, "%3d  %s — %s\n", i+1, album.Artist, album.Title)
	}
	return 0
}

func loadTargets(client lastfm.Client, targets []importer.Target) ([]lastfm.Album, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()
	out := make([]lastfm.Album, 0, len(targets))
	for i, target := range targets {
		album, err := client.GetAlbumTracks(ctx, target.Artist, target.Album)
		if err != nil {
			return nil, fmt.Errorf("entry %d (%s — %s): %w", i+1, target.Artist, target.Album, err)
		}
		out = append(out, album)
	}
	return out, nil
}

func executeAlbums(command string, cfg config.Config, client lastfm.Client, albums []lastfm.Album, opts commonOptions, stdout, stderr io.Writer) int {
	queue := buildQueue(albums, opts.loop, opts.limit)
	result := Result{Command: command, Albums: len(albums), Tracks: len(queue), DryRun: opts.dryRun}
	if len(queue) == 0 {
		fmt.Fprintln(stderr, "queue is empty")
		return 1
	}
	if opts.dryRun {
		result.Message = "queue resolved; no scrobbles sent"
		return emitResult(result, opts.json, stdout)
	}
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	authCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	if err := client.Authenticate(authCtx); err != nil {
		cancel()
		fmt.Fprintln(stderr, err)
		return 1
	}
	cancel()
	if sessionClient, ok := client.(interface{ SessionKey() string }); ok {
		_ = config.PersistSessionKey(cfg, sessionClient.SessionKey())
	}
	if cfg.DuplicateGuard > 0 && strings.TrimSpace(cfg.Username) != "" {
		recent, err := client.GetRecentTracks(ctx, cfg.Username, time.Now().Add(-cfg.DuplicateGuard))
		if err == nil {
			seen := map[string]bool{}
			for _, track := range recent {
				seen[recentKey(track.Artist, track.Title, track.Album)] = true
			}
			filtered := queue[:0]
			for _, item := range queue {
				if seen[recentKey(item.Artist, item.Title, item.Album)] || seen[recentKey(item.Artist, item.Title, "")] {
					result.Skipped++
					continue
				}
				filtered = append(filtered, item)
			}
			queue = filtered
			result.Tracks = len(queue)
		}
	}
	if len(queue) == 0 {
		fmt.Fprintln(stderr, "queue is empty after duplicate protection")
		return 1
	}
	store := sessionstore.New(config.DataDir())
	record := recordFromQueue(command, cfg, queue, opts)
	record.SkippedDuplicates = result.Skipped
	_ = store.SavePending(record)
	for index, item := range queue {
		if err := ctx.Err(); err != nil {
			record.Status = "cancelled"
			_ = store.SavePending(record)
			_ = store.Append(record)
			return 1
		}
		_, err := scrobbler.RunOne(ctx, client, scrobbler.Track{Artist: item.Artist, Title: item.Title, Album: item.Album}, scrobbler.Options{Retries: cfg.RetryCount, RetryDelay: cfg.RetryDelay})
		if errors.Is(err, context.Canceled) {
			record.Status = "cancelled"
			_ = store.SavePending(record)
			_ = store.Append(record)
			return 1
		}
		if err != nil {
			result.Failed++
			fmt.Fprintf(stderr, "failed: %s — %v\n", item.Title, err)
			record.Failures++
		} else {
			result.Scrobbled++
			if !opts.json {
				fmt.Fprintf(stdout, "[%d/%d] %s — %s\n", index+1, len(queue), item.Artist, item.Title)
			}
		}
		record.Completed = index + 1
		_ = store.SavePending(record)
		if index < len(queue)-1 && opts.interval > 0 {
			if err := scrobbler.Wait(ctx, opts.interval); err != nil {
				record.Status = "cancelled"
				_ = store.SavePending(record)
				_ = store.Append(record)
				return 1
			}
		}
	}
	if result.Failed > 0 {
		result.Message = "session completed with failures"
	} else {
		result.Message = "session complete"
	}
	record.Status = "complete"
	record.CompletedAt = time.Now()
	_ = store.Append(record)
	_ = store.ClearPending()
	emitResult(result, opts.json, stdout)
	if cfg.Notify {
		message := fmt.Sprintf("Scrobbled %d track(s)", result.Scrobbled)
		if result.Failed > 0 {
			message += fmt.Sprintf("; %d failed", result.Failed)
		}
		_ = platform.Notify("Last.fm Scrobbler", message)
	}
	if result.Failed > 0 {
		return 1
	}
	return 0
}

func recentKey(artist, title, album string) string {
	return strings.ToLower(strings.TrimSpace(artist) + "\x00" + strings.TrimSpace(title) + "\x00" + strings.TrimSpace(album))
}

func recordFromQueue(command string, cfg config.Config, queue []queueTrack, opts commonOptions) sessionstore.Record {
	started := time.Now()
	record := sessionstore.Record{ID: sessionstore.NewID(started), Mode: command, Profile: cfg.Profile, StartedAt: started, Status: "pending", Loop: opts.loop, Interval: opts.interval, Limit: strconv.Itoa(opts.limit)}
	record.Queue = make([]sessionstore.Track, 0, len(queue))
	for _, item := range queue {
		record.Queue = append(record.Queue, sessionstore.Track{
			Artist: item.Artist, Title: item.Title, Album: item.Album,
			AlbumIndex: item.AlbumIndex, AlbumTotal: item.AlbumTotal,
			TrackIndex: item.TrackIndex, TrackTotal: item.TrackTotal,
			LoopIndex: item.LoopIndex, LoopTotal: item.LoopTotal,
		})
	}
	return record
}

type queueTrack struct {
	Artist, Album, Title string
	AlbumIndex           int
	AlbumTotal           int
	TrackIndex           int
	TrackTotal           int
	LoopIndex            int
	LoopTotal            int
}

func buildQueue(albums []lastfm.Album, loops, limit int) []queueTrack {
	loops = max(1, loops)
	var queue []queueTrack
	for albumIndex, album := range albums {
		tracks := album.Tracks
		if limit > 0 && limit < len(tracks) {
			tracks = tracks[:limit]
		}
		for loop := 0; loop < loops; loop++ {
			for trackIndex, track := range tracks {
				queue = append(queue, queueTrack{
					Artist: album.Artist, Album: album.Title, Title: track.Title,
					AlbumIndex: albumIndex + 1, AlbumTotal: len(albums),
					TrackIndex: trackIndex + 1, TrackTotal: len(tracks),
					LoopIndex: loop + 1, LoopTotal: loops,
				})
			}
		}
	}
	return queue
}

func emitResult(result Result, jsonOut bool, stdout io.Writer) int {
	if jsonOut {
		_ = json.NewEncoder(stdout).Encode(result)
	} else {
		fmt.Fprintf(stdout, "%s: %d album(s), %d queue item(s), %d scrobbled, %d failed\n", result.Message, result.Albums, result.Tracks, result.Scrobbled, result.Failed)
	}
	return 0
}

func runConnectionTest(cfg config.Config, client lastfm.Client, stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "machine-readable JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	report := connection.Test(context.Background(), cfg, client)
	if *jsonOut {
		_ = json.NewEncoder(stdout).Encode(report)
	} else {
		for _, item := range report.Items {
			status := "OK"
			if item.Skipped {
				status = "SKIP"
			} else if !item.OK {
				status = "FAIL"
			}
			fmt.Fprintf(stdout, "%-12s %-5s %s\n", item.Label, status, item.Detail)
		}
	}
	if !report.OK() {
		return 1
	}
	return 0
}

func runDiagnostics(cfg config.Config, stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("diagnostics", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "machine-readable JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	history, _ := sessionstore.New(config.DataDir()).LoadHistory()
	path, err := diagnostics.Create(cfg, history, logging.Path(), version.Version, version.Commit)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *jsonOut {
		_ = json.NewEncoder(stdout).Encode(map[string]string{"path": path})
	} else {
		fmt.Fprintln(stdout, path)
	}
	return 0
}

func runUpdateCheck(cfg config.Config, stdout, stderr io.Writer, args []string) int {
	fs := flag.NewFlagSet("check-update", flag.ContinueOnError)
	fs.SetOutput(stderr)
	jsonOut := fs.Bool("json", false, "machine-readable JSON")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	result, err := (updater.Checker{}).Check(context.Background(), version.Version, cfg.UpdateURL, version.Repository)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 1
	}
	if *jsonOut {
		_ = json.NewEncoder(stdout).Encode(result)
		return 0
	}
	if result.Available {
		fmt.Fprintf(stdout, "Update available: %s → %s\n", result.Current, result.Latest)
		if result.URL != "" {
			fmt.Fprintln(stdout, result.URL)
		}
	} else {
		fmt.Fprintf(stdout, "Up to date: %s\n", result.Current)
	}
	return 0
}

func splitArtistAlbum(value string) (string, string, bool) {
	for _, separator := range []string{" — ", " - ", "\t", "|"} {
		parts := strings.SplitN(value, separator, 2)
		if len(parts) == 2 {
			left, right := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			if left != "" && right != "" {
				return left, right, true
			}
		}
	}
	return "", "", false
}

func cleanAlbums(albums []lastfm.Album) []lastfm.Album {
	seen := map[string]bool{}
	out := make([]lastfm.Album, 0, len(albums))
	for _, album := range albums {
		lower := strings.ToLower(album.Title)
		noisy := false
		for _, marker := range []string{"reissue", "remaster", "bonus", "deluxe", "anniversary", "demo", "live", "anthology", "compilation", "disc 1", "disc 2"} {
			if strings.Contains(lower, marker) {
				noisy = true
				break
			}
		}
		if noisy {
			continue
		}
		key := strings.Join(strings.Fields(strings.NewReplacer("'", "", "-", " ", "—", " ", ":", " ", "(", " ", ")", " ").Replace(lower)), " ")
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, album)
	}
	return out
}

func selectDiscography(found []lastfm.Album, all bool, names string, first int) []lastfm.Album {
	if all {
		return found
	}
	if first > 0 {
		return found[:min(first, len(found))]
	}
	wanted := map[string]bool{}
	for _, name := range strings.Split(names, ",") {
		name = strings.ToLower(strings.TrimSpace(name))
		if name != "" {
			wanted[name] = true
		}
	}
	var out []lastfm.Album
	for _, album := range found {
		lower := strings.ToLower(strings.TrimSpace(album.Title))
		if wanted[lower] {
			out = append(out, album)
		}
	}
	return out
}

func Completion(shell string) (string, error) {
	return completion.Generate(shell)
}

// normalizeArgs keeps standard flag.FlagSet parsing while allowing options to
// appear before or after positional arguments. valueFlags contains flag names
// that consume the following argument when they are not written as --name=value.
func normalizeArgs(args []string, valueFlags map[string]bool) []string {
	flags := make([]string, 0, len(args))
	positionals := make([]string, 0, len(args))
	for index := 0; index < len(args); index++ {
		argument := args[index]
		if argument == "--" {
			positionals = append(positionals, args[index+1:]...)
			break
		}
		if !strings.HasPrefix(argument, "-") || argument == "-" {
			positionals = append(positionals, argument)
			continue
		}
		flags = append(flags, argument)
		name := strings.TrimLeft(argument, "-")
		if before, _, found := strings.Cut(name, "="); found {
			name = before
			continue
		}
		if valueFlags[name] && index+1 < len(args) {
			index++
			flags = append(flags, args[index])
		}
	}
	return append(flags, positionals...)
}

func ParseInt(value string, fallback int) int {
	number, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return number
}

func SortedKeys(values map[string]bool) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func IsUsageError(err error) bool {
	return errors.Is(err, flag.ErrHelp)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
