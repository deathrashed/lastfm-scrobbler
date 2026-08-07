# <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="24" height="24" alt="Last.fm icon"> Configuration and credentials

Configuration is shared by the TUI, headless CLI, Keyboard Maestro, and shell
automation. Non-secret behavior lives alongside credentials in the project
`.env`; the file is owner-only and ignored by Git.

<p align="center">
  <a href="../README.md">Overview</a> •
  <a href="TUI.md">TUI controls</a> •
  <a href="CLI.md">CLI reference</a> •
  <a href="AUTOMATION.md">Automation</a>
</p>

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Quick setup

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
connection test can obtain a mobile session when the account permits it. On
macOS, a newly acquired session key is saved to Keychain so the password is no
longer required on later runs.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> File precedence

For normal `auto` mode, values are resolved in this order:

1. Real process environment variables.
2. The project-local `.env` next to the executable.
3. Missing values from `~/.env`.
4. A selected or remembered credentials file.
5. macOS Keychain for missing secret values.
6. Built-in defaults for non-secret settings.

`LASTFM_ENV_FILE` forces a particular file. `LASTFM_PROFILE` selects a named
profile. The Config screen can change and remember the credentials path.

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
SCROBBLE_COMPACT_HEADER=false
SCROBBLE_CLEAN_DISCOGRAPHY=false
SCROBBLE_EXPORT_DIR=~/Downloads
SCROBBLE_MOUSE=true
SCROBBLER_UPDATE_URL=
```

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

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Security notes

- Credential files are written with owner-only permissions.
- API secret, password, and session key may be stored in macOS Keychain.
- Diagnostics bundles redact passwords, API secrets, session keys, and complete
  API keys.
- Process environment variables can be exposed to child processes and system
  inspection tools; use Keychain or an owner-only file for long-lived secrets.
