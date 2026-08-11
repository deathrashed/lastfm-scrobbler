# <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="24" height="24" alt="Last.fm icon"> Automation and Keyboard Maestro

The headless CLI is preferred over driving the TUI with simulated keystrokes.
It has stable exit codes, optional JSON output, dry-run support, and the same
configuration as the interactive app.

The legacy `NO_TUI=1` or `--no-tui` input loop routes each entered album through
the same headless command/session runner. It therefore uses the same
authentication, retry, duplicate, history, recovery, notification, and
cancellation behavior.

<p align="center">
  <a href="../README.md">Overview</a> •
  <a href="TUI.md">TUI controls</a> •
  <a href="CONFIGURATION.md">Configuration</a> •
  <a href="CLI.md">CLI reference</a>
</p>

> [!TIP]
> GUI automation does not always inherit an interactive shell's `PATH` or
> environment. Use an absolute binary path and an explicit `LASTFM_ENV_FILE`.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Automation principles

| Principle | Recommendation |
| --- | --- |
| **Validate first** | Run `--dry-run --json` before a real scrobble. |
| **Use absolute paths** | Point wrappers at the checkout-local binary, input file, and env file. |
| **Quote all values** | Artist, album, path, and profile names can contain spaces or shell metacharacters. |
| **Keep secrets out of arguments** | Store credentials in the ignored `.env`, `~/.env`, Keychain, or an explicit credentials file. |
| **Check exit codes** | `0` succeeds, `1` reports an operational failure, and `2` reports invalid usage. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Keyboard Maestro: manual album

Use an **Execute Shell Script** action:

```bash
#!/bin/zsh
set -euo pipefail

SCROBBLER="/path/to/lastfm-scrobbler/bin/scrobbler"
ENV_FILE="/path/to/lastfm-scrobbler/.env"
ARTIST="$KMVAR_Artist"
ALBUM="$KMVAR_Album"

LASTFM_ENV_FILE="$ENV_FILE" "$SCROBBLER" manual \
  --artist "$ARTIST" \
  --album "$ALBUM" \
  --loop "${KMVAR_Loop:-1}" \
  --limit "${KMVAR_Limit:-0}" \
  --interval "${KMVAR_Interval:-2s}" \
  --json
```

The JSON response can be stored in a Keyboard Maestro variable.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Validate before running

```bash
scrobbler manual "Slayer - Hell Awaits" --dry-run --json
```

A macro can parse the output, show a confirmation dialog, then run the command
again without `--dry-run`.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> File or folder workflow

```bash
scrobbler file "$KMVAR_Path" --loop 1 --interval 2s --json
```

This works with a dragged list file, M3U playlist, album folder, or artist
folder.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Scheduled connection check

```bash
scrobbler test --json > "$HOME/Library/Logs/scrobbler-connection.json"
```

A non-zero exit code means the API or authentication test failed.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Diagnostics macro

```bash
bundle=$(scrobbler diagnostics --json)
printf '%s\n' "$bundle"
```

The bundle contains redacted data, but should still be reviewed before sharing.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Environment isolation

Set a credentials file for one macro without changing the saved default:

```bash
LASTFM_ENV_FILE="$HOME/.config/lastfm-scrobbler/archive.env" \
  scrobbler manual "Artist - Album"
```

Or select a named profile:

```bash
LASTFM_PROFILE=archive scrobbler file albums.txt
```

## Shell completions and updates

Generate or install per-user completion for zsh, bash, fish, or PowerShell:

```text
scrobbler completion powershell
scrobbler completion install
scrobbler completion install fish
```

The installer is idempotent and may require a shell/profile reload. The normal
update source is the project's GitHub Releases API, so automation can call
`scrobbler check-update` without setting an update URL.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> LaunchAgent or cron

Use absolute paths for the binary, credentials file, and imports because
scheduled environments have a minimal `PATH`.

```bash
/usr/local/bin/scrobbler file /Users/name/Lists/albums.txt \
  --interval 2s --json
```
