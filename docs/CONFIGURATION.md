# <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="24" height="24" alt="Last.fm icon"> Configuration and credentials

Configuration is shared by the TUI, headless CLI, Keyboard Maestro, and shell
automation. Source-tree builds may use a project `.env`; installed binaries
use `~/.config/lastfm-scrobbler/.env` by default. Credential files are
owner-only and ignored by Git.

<p align="center">
  <a href="../README.md">Overview</a> •
  <a href="TUI.md">TUI controls</a> •
  <a href="CLI.md">CLI reference</a> •
  <a href="AUTOMATION.md">Automation</a>
</p>

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Quick setup

Launch an unconfigured installation to open the first-run wizard. The same
wizard can be started explicitly with `scrobbler setup`. It keeps credentials,
preferences, font installation, and terminal changes staged until **Apply**;
leaving the wizard earlier does not write a partial configuration.

For a source-tree build, a project `.env` remains available as an advanced
manual alternative:

```bash
cp .env.example .env
chmod 600 .env
```

For a normal authenticated session:

```env
API_KEY=your-lastfm-api-key
API_SECRET=your-lastfm-api-secret
LASTFM_USERNAME=your-username
LASTFM_SESSION_KEY=your-session-key
```

Alternatively, omit `LASTFM_SESSION_KEY` and provide `LASTFM_PASSWORD`; the
connection test can obtain a mobile session when the account permits it. A
newly acquired session key is persisted according to the selected credential
source, as described below.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> File precedence

For normal `auto` mode, `LASTFM_ENV_FILE` is authoritative even when its file
does not exist yet. An existing credentials path selected in **Settings →
Account** is used before automatic discovery. Otherwise the resolver checks the
checkout-root `.env` for the conventional `<checkout>/bin/scrobbler` source
binary, then current-working-directory compatibility locations, then
`~/.config/lastfm-scrobbler/.env` for installed or release binaries. During
automatic discovery, missing values may come from `~/.env`; real process
environment variables overlay file values, and missing secret values may then
come from macOS Keychain. `LASTFM_PROFILE` selects a named profile.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> TUI Settings sections

The TUI exposes configuration through **Settings** (`S` from the Dashboard):

- **Account** — username/password, API key/secret, credential source/path.
- **Scrobbling** — loop, interval, retry behavior, duplicate guard, Discography cleanup.
- **Tools** — export directory, update source, connection test, diagnostics, completions, updates.
- **Interface** — notifications, Now Playing, Compact Header, Mouse Support.
- **History** and **Profiles** share the same Settings shell but retain their existing data/actions.

Moving these values between TUI sections does not change their environment
variable names, config-file format, or precedence rules.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Last.fm variables

| Variable | Purpose |
| --- | --- |
| `API_KEY`, `LASTFM_API_KEY` | Last.fm application key. |
| `API_SECRET`, `LASTFM_API_SECRET`, `LASTFM_SHARED_SECRET` | Last.fm application/shared secret. |
| `LASTFM_USERNAME`, `LASTFM_USER` | Last.fm account name. |
| `LASTFM_PASSWORD` | Password used to obtain a mobile session. |
| `LASTFM_SESSION_KEY`, `SESSION_KEY` | Existing authenticated session. |
| `LASTFM_PROFILE` | Named account profile; defaults to `default`. |
| `LASTFM_CREDENTIAL_SOURCE` | `auto`, `environment`, `file`, or `keychain`. |
| `LASTFM_ENV_FILE` | Explicit credentials/settings file. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Scrobble defaults

```env
SCROBBLE_INTERVAL=2s
SCROBBLE_LIMIT=0
SCROBBLE_LOOP=1
```

`SCROBBLE_LIMIT=0` means all tracks. The interval accepts Go durations such as
`500ms`, `2s`, or `1m`.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Reliability and safety

```env
SCROBBLE_RETRIES=2
SCROBBLE_RETRY_DELAY=2s
SCROBBLE_DUPLICATE_GUARD=0
```

The duplicate guard accepts durations such as `5m` or `1h`; zero disables it.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Interface and output

```env
SCROBBLE_NOTIFY=true
SCROBBLE_NOW_PLAYING=true
SCROBBLE_COMPACT_HEADER=false
SCROBBLE_CLEAN_DISCOGRAPHY=false
SCROBBLE_EXPORT_DIR=~/Downloads
SCROBBLE_MOUSE=true
SCROBBLER_UPDATE_URL=
```

`SCROBBLE_CLEAN_DISCOGRAPHY` is retained as the compatibility/environment
variable name; the TUI labels this setting **Clean Discography**.

`SCROBBLE_NOW_PLAYING` defaults to true. When enabled, the full header reads
one item from Last.fm `user.getRecentTracks` and displays the current track or
the most recent scrobble. Compact Header never displays activity. The request
is supplemental, refreshes about every 30 seconds, and never submits playback
state.

`SCROBBLER_UPDATE_URL` is optional. An empty value uses the official GitHub
Releases source for `deathrashed/lastfm-scrobbler`; set it only for a custom
JSON-compatible endpoint.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Profiles and Keychain

Profiles are stored below `~/.config/lastfm-scrobbler/profiles/` and can be
selected in the TUI or with `LASTFM_PROFILE`. They are useful when switching
between Last.fm accounts or automation contexts.

Credential sources:

| Source | Behavior |
| --- | --- |
| `auto` | Process environment overrides files; missing secrets may come from Keychain. |
| `environment` | Read credentials only from real process environment variables. |
| `file` | Read credentials only from the selected environment file. |
| `keychain` | Public values come from environment/file; secret values come from macOS Keychain. |

The wizard only offers macOS Keychain on macOS, where the existing backend is
implemented. Linux and Windows use the existing owner-only credentials file or
environment-variable backends; no unimplemented Secret Service or Windows
Credential Manager adapter is claimed.

Environment and Keychain overrides are runtime values, not replacements for
the values already stored in the credentials file. When the TUI saves an
unrelated setting, the original file fallback is preserved. Explicitly editing
a credential field marks the new value as intentional.

Session-key persistence follows the source:

| Source | Newly acquired session key |
| --- | --- |
| `auto` | Save to macOS Keychain. |
| `keychain` | Save to macOS Keychain. |
| `file` | Save to the selected credentials file. |
| `environment` | Do not persist automatically. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Security notes

- Credential files are written with owner-only permissions.
- API secret, password, and session key may be stored in macOS Keychain.
- Diagnostics bundles redact passwords, API secrets, session keys, and complete
  API keys.
- Process environment variables can be exposed to child processes and system
  inspection tools; use Keychain or an owner-only file for long-lived secrets.

## First-run font and terminal setup

The wizard detects the operating-system family, architecture, terminal, and
available package manager for context. Package managers are never required.
The curated Nerd Font choices download official release-family archives from
the Nerd Fonts latest release and install selected Mono font files at user
scope. Linux refreshes the font cache when `fc-cache` is available; Windows
registers installed files in the current-user font registry. No root or
administrator privileges are required by the default paths.

Automatic terminal configuration is currently implemented for detected
Ghostty configuration files only, with a backup made before editing. Other
terminals remain usable but are reported as manual setup; selecting a font
does not silently edit arbitrary terminal configuration files.
