

![header](assets/header.png)

<h1 align="center">𝗟𝗔𝗦𝗧.𝗙𝗠 𝗦𝗖𝗥𝗢𝗕𝗕𝗟𝗘𝗥</h1>

<p align="center">
  <img src="https://img.shields.io/badge/Last.fm-Torch%20Red-f8211c?style=for-the-badge&logo=last.fm&logoColor=white" alt="Last.fm">
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

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Features

| Area | What it provides |
| --- | --- |
| **Manual scrobbling** | Resolve an `Artist - Album`, choose tracks, preview the queue, and scrobble with configurable loops and intervals. |
| **Discography curation** | Filter, sort, clean, select, and queue multiple albums from an artist discography. |
| **Import workflows** | Load TXT, CSV, TSV, JSON, M3U/M3U8 playlists, album folders, or artist folders. |
| **Recovery** | Keep history, resume interrupted queues, edit saved sessions, or perform exact re-runs. |
| **Automation** | Use stable JSON-capable CLI commands from Keyboard Maestro, shell scripts, and launch agents. |
| **Terminal UI** | A centered 67-column Bubble Tea interface with Torch Red active controls and mouse support. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Installation & Build

Requirements:

- Go 1.24.2 or newer
- Internet access during the first build so Go can download Bubble Tea modules
- A Nerd Font for the intended icons

If the repository is published at `github.com/deathrashed/lastfm-scrobbler`,
install the latest tagged command directly with Go:

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
  -ldflags "-X github.com/deathrashed/lastfm-scrobbler/internal/version.Version=v10.0.0 \
            -X github.com/deathrashed/lastfm-scrobbler/internal/version.Commit=$(git rev-parse --short HEAD) \
            -X github.com/deathrashed/lastfm-scrobbler/internal/version.Repository=OWNER/REPOSITORY" \
  -o bin/scrobbler ./cmd/scrobbler
```

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Quick start

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

The project-local `.env` is the normal save and load location. Values missing
from it may be read from `~/.env`. Set
`LASTFM_ENV_FILE=/absolute/path/to/file.env` when an explicit credentials file
is required. `.env` is ignored by Git; `.env.example` is the safe template to
commit.

> [!TIP]
> Use `./bin/scrobbler test --json` after setup. It checks Last.fm access and
> authentication readiness without submitting a scrobble.

> [!WARNING]
> Keep `.env`, `~/.env`, and exported credentials files private. Use the
> redacted diagnostics bundle when sharing troubleshooting information.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Screenshots

The interface keeps the same Torch Red, fixed-width visual language across
search, selection, configuration, recovery, and scrobbling workflows. The
gallery follows the recommended walkthrough order.

<table>
  <tr>
    <td width="33%" valign="top">
      <a href="assets/screenshots/1-dashboard-menu.png"><img src="assets/screenshots/1-dashboard-menu.png" alt="Last.fm Scrobbler dashboard menu"></a>
      <p align="center"><strong>1. Dashboard</strong><br><sub>Search • Select • Scrobble</sub></p>
    </td>
    <td width="33%" valign="top">
      <a href="assets/screenshots/2-info-menu.png"><img src="assets/screenshots/2-info-menu.png" alt="Last.fm Scrobbler Info menu"></a>
      <p align="center"><strong>2. Info</strong><br><sub>Modes, automation, data, curation, and imports</sub></p>
    </td>
    <td width="33%" valign="top">
      <a href="assets/screenshots/3-help-menu.png"><img src="assets/screenshots/3-help-menu.png" alt="Last.fm Scrobbler help menu"></a>
      <p align="center"><strong>3. Help</strong><br><sub>Contextual controls at a glance</sub></p>
    </td>
  </tr>
  <tr>
    <td valign="top">
      <a href="assets/screenshots/4-config-menu.png"><img src="assets/screenshots/4-config-menu.png" alt="Last.fm Scrobbler configuration menu"></a>
      <p align="center"><strong>4. Config</strong><br><sub>Settings and credentials</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/5-file-menu.png"><img src="assets/screenshots/5-file-menu.png" alt="Last.fm Scrobbler file import menu"></a>
      <p align="center"><strong>5. File</strong><br><sub>Lists, playlists, and folders</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/6-history.png"><img src="assets/screenshots/6-history.png" alt="Last.fm Scrobbler session history"></a>
      <p align="center"><strong>6. History</strong><br><sub>Recovery, exports, and re-runs</sub></p>
    </td>
  </tr>
  <tr>
    <td valign="top">
      <a href="assets/screenshots/7-discography-search.png"><img src="assets/screenshots/7-discography-search.png" alt="Last.fm Scrobbler discography search"></a>
      <p align="center"><strong>7. Discography search</strong><br><sub>Find an artist's albums</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/8-discography-filter.png"><img src="assets/screenshots/8-discography-filter.png" alt="Last.fm Scrobbler discography filter"></a>
      <p align="center"><strong>8. Discography filter</strong><br><sub>Filter, clean, sort, and select</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/9-discography-progress.png"><img src="assets/screenshots/9-discography-progress.png" alt="Last.fm Scrobbler discography progress"></a>
      <p align="center"><strong>9. Discography progress</strong><br><sub>Build and scrobble the queue</sub></p>
    </td>
  </tr>
  <tr>
    <td valign="top">
      <a href="assets/screenshots/10-manual-album-selection.png"><img src="assets/screenshots/10-manual-album-selection.png" alt="Last.fm Scrobbler manual album selection"></a>
      <p align="center"><strong>10. Manual album selection</strong><br><sub>Choose tracks from an album</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/11-manual-album-preferences.png"><img src="assets/screenshots/11-manual-album-preferences.png" alt="Last.fm Scrobbler manual album preferences"></a>
      <p align="center"><strong>11. Manual preferences</strong><br><sub>Adjust loops and queue options</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/12-manual-progress.png"><img src="assets/screenshots/12-manual-progress.png" alt="Last.fm Scrobbler manual scrobble progress"></a>
      <p align="center"><strong>12. Manual progress</strong><br><sub>Track status and completion</sub></p>
    </td>
  </tr>
  <tr>
    <td valign="top">
      <a href="assets/screenshots/13-manual-complete.png"><img src="assets/screenshots/13-manual-complete.png" alt="Last.fm Scrobbler completed manual scrobble"></a>
      <p align="center"><strong>13. Manual complete</strong><br><sub>Review the finished session</sub></p>
    </td>
    <td valign="top">
      <a href="assets/screenshots/14-config-advanced.png"><img src="assets/screenshots/14-config-advanced.png" alt="Last.fm Scrobbler advanced configuration"></a>
      <p align="center"><strong>14. Advanced config</strong><br><sub>Reliability, diagnostics, and updates</sub></p>
    </td>
    <td valign="top"></td>
  </tr>
</table>

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Main TUI workflows

### Manual

Enter `Artist - Album`, select the correct search result when necessary, choose
individual tracks, adjust global or per-album loops, inspect the dry-run
preview, and start scrobbling.

### Discography

Search an artist, filter and sort the returned albums, hide obvious reissues or
duplicates, select any combination, load their tracks, and build one queue.
Long discographies and track lists include scroll indicators.

### File and folder import

The File page makes every import source visible:

- TXT, CSV, TSV, or JSON album lists
- M3U and M3U8 playlists
- one `Artist/Album` folder
- an artist folder containing album subfolders

The native macOS picker is available with `O`.

### Similar albums

Press `S` from an album result, queue preview, or completed session to retrieve
album suggestions based on similar artists.

### History, recovery, and re-running

Completed and cancelled sessions are recorded locally. History can re-open a
saved queue for editing before it runs again. `Shift+R` performs an exact
re-run without editing. An interrupted active queue is offered for resume at
the next launch.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> TUI controls

| Context | Controls |
| --- | --- |
| **Global** | Arrow keys or `J/K` navigate, `Enter` confirms, `Esc` goes back, `Q` or `Ctrl+C` quits, and `?` opens help. |
| **Dashboard** | `M` Manual, `D` Discography, `F` File, `H` History, `P` Profiles, `C` Config, `I` Info. |
| **Track selection** | `Space` toggles a track, `A` selects all, `-/+` changes loops, `Enter` previews, `S` finds similar albums. |
| **Config** | Arrows navigate, `Tab` changes fields, `Enter` saves or opens, `Ctrl+P` edits the credentials path, `Ctrl+G` opens Advanced, and `Ctrl+O` opens Info. |

When a text field is focused, printable characters are passed to the field so
usernames, passwords, API keys, paths, and filters can contain normal letters
without triggering navigation shortcuts.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="22" height="22" alt="Last.fm icon"> Configuration interface

Config uses two centered rows:

```text
L O O P • I N T E R V A L • U S E R N A M E • A P I
A D V A N C E D • H I S T O R Y • P R O F I L E S
```

Advanced contains:

- retry count and delay
- duplicate-scrobble guard
- completion notifications
- compact header
- discography cleanup
- export directory
- credential source
- mouse support
- update URL
- connection test
- diagnostics bundle
- update checker

The Connection Test validates the public API and checks authentication
readiness without sending a scrobble. Username/password mode can obtain a
mobile session; an existing session key is reported as configured and is
validated by the first signed write. Diagnostics creates a ZIP with runtime details,
redacted configuration, a history summary, and the tail of the application
log. Passwords, API secrets, session keys, and complete API keys are excluded.

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

### Environment file precedence

For normal `auto` mode, values are resolved in this order:

1. Real process environment variables.
2. The project-local `.env` next to the executable.
3. Missing values from `~/.env`.
4. A selected or remembered credentials file.
5. macOS Keychain for missing secret values.
6. Built-in defaults for non-secret settings.

`LASTFM_ENV_FILE=/absolute/path/to/file.env` forces a specific file. The Config
screen can also change and remember the credentials path. `API_SECRET`,
`LASTFM_API_SECRET`, and `LASTFM_SHARED_SECRET` are accepted as aliases.

The project `.env` is written with owner-only permissions and is ignored by
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

Generate completion from the binary so it always matches the installed
commands.

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

## Mouse support

Mouse support is enabled by default and can be disabled in Advanced or with
`SCROBBLE_MOUSE=false`.

- click Dashboard modes, File import sources, and Config tabs
- use the mouse wheel to navigate results, discographies, tracks, History,
  Profiles, and Advanced
- keyboard controls remain available everywhere

Bubble Tea mouse capture can interfere with normal terminal text selection in
some terminals. Disable the setting when native selection is more important.

## Update checking

The checker does not assume a repository. Configure either:

```env
SCROBBLER_UPDATE_URL=https://api.github.com/repos/OWNER/REPOSITORY/releases/latest
```

or inject `Repository=OWNER/REPOSITORY` at build time. A custom endpoint may
return GitHub-style JSON (`tag_name`, `html_url`, `body`) or simple JSON
(`version`, `url`, `notes`). The checker only reports availability; it does not
replace the running binary.

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
Last.fm client ── search / album tracks / discographies / similar albums
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
current value in Config → Credentials Path, or force a path explicitly:

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
- [TUI controls and workflows](docs/TUI.md)
- [Release checklist](docs/RELEASING.md)
- [Last.fm colour reference](docs/Last.fm-colors.html)

## License

This project uses the **Do What The Fuck You Want To Public License, Version
2**, supplied in [LICENSE](LICENSE).
