

![header](assets/header.png)

<h1 align="center">𝗟𝗔𝗦𝗧.𝗙𝗠 𝗦𝗖𝗥𝗢𝗕𝗕𝗟𝗘𝗥</h1>

<p align="center">
  <img src="https://img.shields.io/badge/Last.fm-Scrobbler-f8211c?style=for-the-badge&logo=last.fm&logoColor=white" alt="Last.fm">
  <img src="https://img.shields.io/badge/Go-1.24.2%2B-f8211c?style=for-the-badge&logo=go&logoColor=white" alt="Go 1.24.2 or newer">
  <img src="https://img.shields.io/badge/License-WTFPL%202-f8211c?style=for-the-badge&logo=open-source-initiative&logoColor=white" alt="WTFPL 2 License">
</p>


<p align="center">
  A fixed-width terminal application for searching, curating, previewing, and
  scrobbling Last.fm album queues.
</p>

<p align="center">
  <a href="#features">Features</a> •
  <a href="#installation--build">Installation</a> •
  <a href="#quick-start">Quick start</a> •
  <a href="#screenshots">Screenshots</a> •
  <a href="#headless-cli">CLI</a> •
  <a href="#credentials">Configuration</a> •
  <a href="#troubleshooting">Troubleshooting</a>
</p>

The visual system is deliberately consistent across every screen:

- 67-column header and content area
- white structural borders
- Last.fm Torch Red (`#f8211c`) active controls
- centered panels and footer hints
- wrapped or clipped long text that cannot break a border
- Nerd Font icons with plain-text fallbacks where practical

The TUI requires at least 67 terminal columns. Compact Header is a user
setting, not an automatic narrow-terminal fallback. It normally uses a compact
four-line header; Manual and Discography add one centered `ARTIST ❯` metadata
row after an artist has been resolved. Full-header profile URLs highlight on
hover and open on click; compact mode has no profile URL.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Features

| Area | What it provides |
| --- | --- |
| **Manual scrobbling** | Search by artist, album, or both, choose tracks, preview the queue, and scrobble with configurable loops and intervals. |
| **Discography curation** | Filter, sort, clean, select, and queue albums returned by Last.fm's top-albums endpoint. |
| **Import workflows** | Load TXT, CSV, TSV, JSON, M3U/M3U8 playlists, album folders, or artist folders. |
| **Recovery** | Keep history, resume interrupted queues, edit saved sessions, or perform exact re-runs. |
| **Automation** | Use stable JSON-capable CLI commands from Keyboard Maestro, shell scripts, and launch agents. |
| **Terminal UI** | A centered 67-column Bubble Tea interface with Torch Red active controls and mouse support. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Installation & Build

Requirements:

- Go 1.24.2 or newer
- Internet access during the first build so Go can download Bubble Tea modules
- A Nerd Font is recommended for the intended icons (any compatible Nerd Font works)

### First-run setup

On a new installation, `scrobbler` opens a cross-platform setup wizard before
the dashboard. It collects Last.fm credentials, stores them in the selected
credential source, chooses sensible scrobbling and interface defaults, tests
the connection, and can optionally download an official Nerd Font release and
install it at user scope. Run `scrobbler setup` later to review the wizard
again; leaving it before **Apply** does not write credentials, font files, or
terminal configuration.

The wizard works on macOS, Linux, and Windows without requiring Homebrew,
Chocolatey, Scoop, apt, dnf, or another package manager. It supports macOS
Keychain where available and otherwise offers the existing credentials-file or
environment-variable backends. Automatic terminal font configuration is
limited to detected Ghostty configurations; other terminals receive manual
instructions while font installation can still complete.

Install the latest tagged command directly with Go:

```bash
go install github.com/deathrashed/lastfm-scrobbler/cmd/scrobbler@latest
```

```bash
cd /path/to/lastfm-scrobbler
mkdir -p bin
go build -buildvcs=false -o bin/scrobbler ./cmd/scrobbler
./bin/scrobbler
```

A versioned release build can inject source information for update checking:

```bash
go build \
  -ldflags "-X github.com/deathrashed/lastfm-scrobbler/internal/version.Version=v1.0.0 \
            -X github.com/deathrashed/lastfm-scrobbler/internal/version.Commit=$(git rev-parse --short HEAD) \
            -X github.com/deathrashed/lastfm-scrobbler/internal/version.Repository=deathrashed/lastfm-scrobbler" \
  -o bin/scrobbler ./cmd/scrobbler
```

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Quick Start

1. Create the project-local environment file:

   ```bash
   cp .env.example .env
   chmod 600 .env
   ```

2. Add an API key and secret, then choose either a Last.fm password or an
   existing session key:

   ```env
   API_KEY=your-lastfm-api-key
   API_SECRET=your-lastfm-api-secret
   LASTFM_USERNAME=your-username
   LASTFM_SESSION_KEY=your-session-key
   ```

3. Build and launch:

   ```bash
   go build -buildvcs=false -o bin/scrobbler ./cmd/scrobbler
   ./bin/scrobbler
   ```

4. Verify the connection without submitting a scrobble:

   ```bash
   ./bin/scrobbler test
   ```

For a new installation, replace the manual file setup with:

```bash
./bin/scrobbler setup
```

Source-tree builds can use the project-local `.env`. Installed binaries use
`~/.config/lastfm-scrobbler/.env` by default. Values missing from an
automatically discovered file may be read from `~/.env`. Set
`LASTFM_ENV_FILE=/absolute/path/to/file.env` to make a specific credentials
file authoritative, including before it exists. `.env` is ignored by Git;
`.env.example` is the safe template to commit.

> [!TIP]
> Use `./bin/scrobbler test --json` after setup. It checks Last.fm access and
> authentication readiness without submitting a scrobble.

> [!WARNING]
> Keep `.env`, `~/.env`, and exported credentials files private. Use the
> redacted diagnostics bundle when sharing troubleshooting information.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Screenshots

The interface keeps the same Torch Red, fixed-width visual language across
search, selection, settings, recovery, and scrobbling workflows. The gallery
uses current captures and follows the recommended walkthrough order.

<table>
  <tr>
    <td width="33%" valign="top">
      <a href="assets/screenshots/1-dashboard.png"><img src="assets/screenshots/1-dashboard.png" alt="Last.fm Scrobbler dashboard"></a>
      <p align="center"><strong>Dashboard</strong><br><sub>Choose Manual, Discography, or File</sub></p>
    </td>
    <td width="33%" valign="top">
      <a href="assets/screenshots/2-manual-search.png"><img src="assets/screenshots/2-manual-search.png" alt="Last.fm Scrobbler Manual search"></a>
      <p align="center"><strong>Manual search</strong><br><sub>Search by artist, album, or both</sub></p>
    </td>
    <td width="33%" valign="top">
      <a href="assets/screenshots/3-manual-results.png"><img src="assets/screenshots/3-manual-results.png" alt="Last.fm Scrobbler Manual search results"></a>
      <p align="center"><strong>Manual results</strong><br><sub>Resolve the matching artist and album</sub></p>
    </td>
  </tr>
  <tr>
    <td valign="top">
      <a href="assets/screenshots/4-manual-select.png"><img src="assets/screenshots/4-manual-select.png" alt="Last.fm Scrobbler Manual track selection"></a>
      <p align="center"><strong>Manual track selection</strong><br><sub>Select tracks with album-specific controls</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/5-manual-queue.png"><img src="assets/screenshots/5-manual-queue.png" alt="Last.fm Scrobbler Manual queue preview"></a>
      <p align="center"><strong>Manual queue</strong><br><sub>Review tracks, interval, loops, and ETA</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/6-manual-run.png"><img src="assets/screenshots/6-manual-run.png" alt="Last.fm Scrobbler Manual scrobble progress"></a>
      <p align="center"><strong>Manual run</strong><br><sub>Watch track progress and ETA</sub></p>
    </td>
  </tr>
  <tr>
    <td valign="top">
      <a href="assets/screenshots/7-manual-done.png"><img src="assets/screenshots/7-manual-done.png" alt="Last.fm Scrobbler completed Manual scrobble"></a>
      <p align="center"><strong>Manual complete</strong><br><sub>Confirm completion and rerun or export</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/8-discography-search.png"><img src="assets/screenshots/8-discography-search.png" alt="Last.fm Scrobbler Discography search"></a>
      <p align="center"><strong>Discography search</strong><br><sub>Resolve an artist before loading results</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/9-discography-filter.png"><img src="assets/screenshots/9-discography-filter.png" alt="Last.fm Scrobbler Discography filter"></a>
      <p align="center"><strong>Discography filter</strong><br><sub>Filter Discography results while browsing</sub></p>
    </td>
  </tr>
  <tr>
    <td valign="top">
      <a href="assets/screenshots/10-discography-select.png"><img src="assets/screenshots/10-discography-select.png" alt="Last.fm Scrobbler Discography track selection"></a>
      <p align="center"><strong>Discography track selection</strong><br><sub>Select tracks across multiple albums</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/11-discography-queue.png"><img src="assets/screenshots/11-discography-queue.png" alt="Last.fm Scrobbler Discography queue preview"></a>
      <p align="center"><strong>Discography queue</strong><br><sub>Review selected album names and queue totals</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/12-discography-run.png"><img src="assets/screenshots/12-discography-run.png" alt="Last.fm Scrobbler Discography scrobble progress"></a>
      <p align="center"><strong>Discography run</strong><br><sub>Track multi-album scrobble progress</sub></p>
    </td>
  </tr>
  <tr>
    <td valign="top">
      <a href="assets/screenshots/13-discography-similar-albums.png"><img src="assets/screenshots/13-discography-similar-albums.png" alt="Last.fm Scrobbler similar albums from Discography"></a>
      <p align="center"><strong>Discography similar albums</strong><br><sub>Discover related albums based on the resolved artist</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/14-file-select.png"><img src="assets/screenshots/14-file-select.png" alt="Last.fm Scrobbler File source selection"></a>
      <p align="center"><strong>File source selection</strong><br><sub>Choose a list, playlist, or folder source</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/15-file-path.png"><img src="assets/screenshots/15-file-path.png" alt="Last.fm Scrobbler File path input"></a>
      <p align="center"><strong>File path</strong><br><sub>Enter or select an import path</sub></p>
    </td>
  </tr>
  <tr>
    <td valign="top">
      <a href="assets/screenshots/16-compact-header.png"><img src="assets/screenshots/16-compact-header.png" alt="Last.fm Scrobbler compact header"></a>
      <p align="center"><strong>Compact header</strong><br><sub>Four-line header before artist resolution</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/17-compact-header-artist.png"><img src="assets/screenshots/17-compact-header-artist.png" alt="Last.fm Scrobbler compact header with artist context"></a>
      <p align="center"><strong>Compact header with artist</strong><br><sub>Resolved artist metadata inside Manual</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/18-history.png"><img src="assets/screenshots/18-history.png" alt="Last.fm Scrobbler session history"></a>
      <p align="center"><strong>History</strong><br><sub>Review, export, delete, or rerun sessions</sub></p>
    </td>
  </tr>
  <tr>
    <td valign="top">
      <a href="assets/screenshots/19-settings.png"><img src="assets/screenshots/19-settings.png" alt="Last.fm Scrobbler unified Settings"></a>
      <p align="center"><strong>Settings</strong><br><sub>Six-section configuration and utilities shell</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/20-info.png"><img src="assets/screenshots/20-info.png" alt="Last.fm Scrobbler Info reference"></a>
      <p align="center"><strong>Info</strong><br><sub>Workflows, controls, data, and imports</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/21-help.png"><img src="assets/screenshots/21-help.png" alt="Last.fm Scrobbler Help screen"></a>
      <p align="center"><strong>Help</strong><br><sub>Keyboard and mouse controls with clickable close</sub></p>
    </td>
  </tr>
</table>

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Main TUI Workflows

### Manual

Search by artist, album, or both, select the correct result when necessary, choose
individual tracks, adjust loops and interval controls, inspect the dry-run
preview, and start scrobbling. Manual result lists show a compact attached
`RESULTS` count rather than a multiselect counter.

### Discography

Search an artist, filter and sort the returned albums, hide obvious reissues or
duplicates, select any combination, load their tracks, and build one queue. The
Discography chooser integrates `SORT`, `FILTER`, and `CLEAN` into the top border with
`RESULTS` and `SELECTED` attached below; long filters expand into a connected
wide input. The source is Last.fm's `artist.getTopAlbums` result, not a
canonical complete discography; long results and track lists include scroll
indicators.

### File and folder import

The File page keeps source selection and PATH entry on one screen. It supports:

- TXT, CSV, TSV, or JSON album lists
- M3U and M3U8 playlists
- one `Artist/Album` folder
- an artist folder containing album subfolders

Press `O` for the platform picker when available, or enter the path manually.
The screen shows the accepted formats in a dynamic `TYPE`/`TYPES` attachment.

### Similar albums

Press `S` from an album result, queue preview, or completed session to retrieve
album suggestions based on similar artists.

### History, recovery, and re-running

Completed and cancelled sessions are recorded locally. History can re-open a
saved queue for editing before it runs again. `Shift+R` performs an exact
re-run without editing. An interrupted active queue is offered for resume at
the next launch.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> TUI Controls

| Context | Controls |
| --- | --- |
| **Global** | Arrow keys or `J/K` navigate, `Enter` confirms, `Esc` goes back, `Q` or `Ctrl+C` quits, and `?` opens help. |
| **Dashboard** | `M` Manual, `D` Discography, `F` File, `H` History, `P` Profiles, `S` Settings, `I` Info. |
| **Track selection** | `Space` toggles a track, `A` selects all, `-/+` changes the current album loop, `Enter` continues, `S` finds similar albums; mouse footer controls also adjust interval and navigation directly. |
| **Settings** | `Tab` moves between section navigation and section content; arrows navigate the active zone; `Enter` saves, opens, or runs the selected item. |

When a text field is focused, printable characters are passed to the field so
usernames, passwords, API keys, paths, and filters can contain normal letters
without triggering navigation shortcuts.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Settings Interface

Press `S` from the Dashboard to open Settings. The six sections share one fixed
65-cell navigation grid:

```text
╭─────────────────╮ ╭───────────────────────╮ ╭─────────────────╮
│  A C C O U N T  │•│  S C R O B B L I N G  │•│  H I S T O R Y  │
╰─────────────────╯ ╰───────────────────────╯ ╰─────────────────╯
╭─────────────────╮ ╭───────────────────────╮ ╭─────────────────╮
│    T O O L S    │•│   I N T E R F A C E   │•│ P R O F I L E S │
╰─────────────────╯ ╰───────────────────────╯ ╰─────────────────╯
```

The sections are grouped by purpose:

- **Account** — Last.fm username/password, API key/secret, credential source,
  and credential path.
- **Scrobbling** — loop, interval, retries, duplicate guard, and Discography
  cleanup.
- **History** — saved sessions, edit/exact re-runs, export, and delete.
- **Tools** — export directory, GitHub Releases/custom update source, connection
  test, diagnostics, shell completions, and update checking.
- **Interface** — notifications, Compact Header, and Mouse Support.
- **Profiles** — load, create, save, and delete named profiles.

Settings has two keyboard focus zones. `Tab` or `Shift+Tab` moves from section
content to the six-section grid; arrow keys then choose a section, and `Enter`
or `Tab` returns to its content. Pressing `Up` from the first content row also
returns to the section grid. Mouse users can click a section or setting row
directly.

Settings rows deliberately use a quieter color hierarchy: idle labels are
white, the `❯` accent is Torch Red, and values are muted. The focused row uses
a leading `❯`, turns its label Torch Red, and turns the row arrow/value white;
mouse hover uses the same red-label emphasis without moving keyboard focus.

Navigation cards use a separate state language: idle labels are muted, hover
turns the label Torch Red, and the selected card uses a Torch Red border with
bold white text. Text inputs keep a white structural border and show focus with
a red label, white arrow, and blinking red cursor rather than a red input box.

The Connection Test validates the public API and checks authentication
readiness without sending a scrobble. Diagnostics creates a ZIP with runtime
details, redacted configuration, a history summary, and the tail of the
application log. Passwords, API secrets, session keys, and complete API keys
are excluded.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Credentials

Read-only Last.fm operations need an API key. Scrobbling needs an API key, API
secret, and authenticated session.

Either configure an existing session:

```env
API_KEY=...
API_SECRET=...
LASTFM_USERNAME=...
LASTFM_SESSION_KEY=...
```

or let the app obtain a mobile session:

```env
API_KEY=...
API_SECRET=...
LASTFM_USERNAME=...
LASTFM_PASSWORD=...
```

When `LASTFM_SESSION_KEY` is present, the password is not required for normal
scrobbling.

Credential sources:

- `auto`: process environment overrides file values; missing secrets may come
  from macOS Keychain
- `environment`: read credentials only from the real process environment
- `file`: read credentials from the selected environment file
- `keychain`: public values come from environment/file and secrets come from
  macOS Keychain

Saving an unrelated setting preserves file fallback values when an environment
or Keychain override is active. Newly acquired session keys go to Keychain for
`auto` and `keychain`, to the credentials file for `file`, and nowhere
automatically for `environment`.

### Environment file precedence

For normal `auto` mode, values are resolved in this order:

1. Real process environment variables.
2. The project-local `.env` next to the executable.
3. Missing values from `~/.env`.
4. A selected or remembered credentials file.
5. macOS Keychain for missing secret values.
6. Built-in defaults for non-secret settings.

`LASTFM_ENV_FILE=/absolute/path/to/file.env` forces a specific file. **Settings →
Account → Credential Path** can also change and remember the credentials path. `API_SECRET`,
`LASTFM_API_SECRET`, and `LASTFM_SHARED_SECRET` are accepted as aliases.

Credential files are written with owner-only permissions and are ignored by
Git. Never replace `.env.example` with real credentials or commit a credentials
file.

See [docs/CONFIGURATION.md](docs/CONFIGURATION.md) for every variable and its
precedence.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Headless CLI

Launching `scrobbler` without a command opens the TUI. Subcommands are stable,
non-interactive entry points suitable for Keyboard Maestro and shell scripts.

```bash
scrobbler manual "Slayer - Hell Awaits"
scrobbler manual --artist Slayer --album "Hell Awaits" --loop 2
scrobbler file ~/Downloads/albums.txt --dry-run
scrobbler discography "Demolition Hammer" --first 3 --dry-run
scrobbler discography "Demolition Hammer" --albums "Epidemic of Violence,Tortured Existence"
scrobbler similar "Demolition Hammer" --limit 20
scrobbler test --json
scrobbler diagnostics --json
scrobbler check-update
```

Common options for `manual`, `file`, and `discography`:

```text
--loop N
--limit N
--interval 2s
--dry-run
--json
```

The CLI returns exit code `0` for success, `1` for an operational failure, and
`2` for invalid command usage. JSON output is intended for automation systems
that need to inspect totals or exported paths.

Detailed command documentation and Keyboard Maestro examples are in
[docs/CLI.md](docs/CLI.md) and [docs/AUTOMATION.md](docs/AUTOMATION.md).

## Shell completion

Generate or install completion from the binary so it always matches the
installed commands. Installation is per-user and idempotent; it never requires
root or administrator access.

```bash
scrobbler completion install
scrobbler completion install zsh
scrobbler completion install bash
scrobbler completion install fish
scrobbler completion install powershell
```

### Zsh

```bash
mkdir -p ~/.zfunc
scrobbler completion zsh > ~/.zfunc/_scrobbler
fpath=(~/.zfunc $fpath)
autoload -Uz compinit && compinit
```

### Bash

```bash
scrobbler completion bash > ~/.scrobbler-completion.bash
source ~/.scrobbler-completion.bash
```

### Fish

```bash
mkdir -p ~/.config/fish/completions
scrobbler completion fish > ~/.config/fish/completions/scrobbler.fish
```

### PowerShell

```powershell
scrobbler completion powershell > $HOME/.config/powershell/completions/scrobbler.ps1
```

The TUI equivalent is **Settings → Tools → Install Completions**. A profile
reload may be needed after installation.

## Mouse support

Mouse support is enabled by default and can be disabled in **Settings → Interface** or with
`SCROBBLE_MOUSE=false`.

- click Dashboard modes, File import sources, Settings sections and rows
- use the mouse wheel to navigate results, Discography lists, tracks, Settings,
  History, and Profiles
- click Info tabs, action cards, editable fields, the Help close hint, and contextual footer actions
- hover feedback turns interactive card/list text Torch Red while preserving
  keyboard selection; full-header profile URLs are red at rest and white while
  hovered
- keyboard controls remain available everywhere

With the full header, the Last.fm URL is Torch Red at rest, turns white on
hover, and opens when clicked. Compact Header does not expose a URL target.

Bubble Tea mouse capture can interfere with normal terminal text selection in
some terminals. Disable the setting when native selection is more important.

## Update checking

Official and normal development builds default to this project's public GitHub
Releases, so **Settings → Tools → Check for Updates** and `scrobbler
check-update` work without configuration. A custom endpoint may
return GitHub-style JSON (`tag_name`, `html_url`, `body`) or simple JSON
(`version`, `url`, `notes`). The checker only reports availability; it does not
replace the running binary.

Set `SCROBBLER_UPDATE_URL` only when an advanced custom endpoint is required;
the empty value means the official GitHub Releases source.

## Exports and diagnostics

Queue/session export formats:

- JSON
- CSV
- TXT
- M3U8

Diagnostics bundles are written to `SCROBBLE_EXPORT_DIR` and are safe to review
before sharing. Application logs rotate after approximately 2 MB.

## Persistent data

By default:

```text
~/.config/lastfm-scrobbler/
```

This contains History, pending-session recovery, profiles, remembered paths,
and `scrobbler.log`. `XDG_DATA_HOME` or `XDG_CONFIG_HOME` is respected.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Architecture

```text
Input
  ├─ Bubble Tea TUI
  ├─ Headless CLI / JSON output
  └─ Keyboard Maestro, shell, or launch-agent wrapper
          │
          ▼
Configuration ── project .env ── ~/.env fallback ── Keychain (optional)
          │
          ▼
Last.fm client ── search / album tracks / Discography / similar albums
          │
          ▼
Queue builder ── preview / dry-run / loop / limit / interval
          │
          ▼
Scrobbler ── history / recovery / exports / diagnostics
```

The TUI and CLI share the same configuration, Last.fm client, queue-building,
history, export, and diagnostics packages. This keeps interactive and
automation workflows behaviorally aligned.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Troubleshooting

<details>
  <summary><strong>Credentials are saved to the wrong file</strong></summary>

The default file is the project-local `.env` beside the executable. Check the
current value in **Settings → Account → Credential Path**, or force a path explicitly:

```bash
LASTFM_ENV_FILE=/absolute/path/to/file.env scrobbler test
```

Older remembered paths are used only after the project-local file has been
checked. `.env` is ignored by Git and should have owner-only permissions.

</details>

<details>
  <summary><strong>Connection test reports missing authentication</strong></summary>

Read-only lookups need `API_KEY`. Scrobbling also needs `API_SECRET` plus either
`LASTFM_SESSION_KEY` or `LASTFM_USERNAME` and `LASTFM_PASSWORD`. Run
`scrobbler test --json` to distinguish missing configuration from an API or
authentication failure.

</details>

<details>
  <summary><strong>Icons or mouse behavior look wrong</strong></summary>

Use a Nerd Font for the intended glyphs. Disable mouse capture with
`SCROBBLE_MOUSE=false` when terminal text selection is more important than
mouse navigation.

</details>

<details>
  <summary><strong>Automation receives human-readable output</strong></summary>

Use `--json` on supported commands and keep credentials outside command-line
arguments. The diagnostics command creates a redacted ZIP; inspect it before
sharing.

</details>

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Limitations

- The application requires network access to resolve Last.fm albums and submit
  scrobbles.
- It does not download music or repair Last.fm metadata; it works with album
  and track data returned by Last.fm.
- Dry runs resolve and display the queue but do not validate a successful
  scrobble.
- API limits, unavailable albums, renamed releases, and Last.fm authentication
  failures remain external service conditions.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Testing

```bash
go test ./...
go vet ./...
```

The project includes tests for Last.fm response parsing, configuration,
imports, exports, history/recovery, queue construction, fixed-width layouts,
connection reports, diagnostics redaction, update parsing, and editable
re-runs.

## Project documentation

- [CLI reference](docs/CLI.md)
- [Automation and Keyboard Maestro](docs/AUTOMATION.md)
- [Configuration and credentials](docs/CONFIGURATION.md)
- [Installation](docs/INSTALLATION.md)
- [Setup wizard](docs/SETUP.md)
- [Shell completions](docs/COMPLETIONS.md)
- [File imports](docs/FILE-IMPORTS.md)
- [Platform support](docs/PLATFORMS.md)
- [Updates](docs/UPDATES.md)
- [Troubleshooting](docs/TROUBLESHOOTING.md)
- [TUI controls and workflows](docs/TUI.md)
- [Release checklist](docs/RELEASING.md)
- [Last.fm colour reference](docs/Last.fm-colors.html)

## License

This project uses the **Do What The Fuck You Want To Public License, Version
2**, supplied in [LICENSE](LICENSE).
