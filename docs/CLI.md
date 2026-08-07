# <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="24" height="24" alt="Last.fm icon"> CLI reference

The headless CLI shares configuration and queue behavior with the TUI. It is
the preferred surface for scripts and automation because it provides stable
exit codes, optional JSON output, duplicate protection, recovery records, and
completion notifications.

<p align="center">
  <a href="../README.md">Overview</a> •
  <a href="TUI.md">TUI controls</a> •
  <a href="CONFIGURATION.md">Configuration</a> •
  <a href="AUTOMATION.md">Automation</a>
</p>

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Command overview

| Command | Purpose |
| --- | --- |
| `scrobbler` / `scrobbler tui` | Open the Bubble Tea TUI. |
| `scrobbler manual` | Load and scrobble one album. |
| `scrobbler file` | Import albums from a list, playlist, or folder. |
| `scrobbler discography` | List or scrobble an artist discography. |
| `scrobbler similar` | Print similar album suggestions. |
| `scrobbler test` | Test Last.fm API access and authentication readiness. |
| `scrobbler diagnostics` | Export a redacted diagnostics ZIP. |
| `scrobbler check-update` | Query the configured release endpoint. |
| `scrobbler completion` | Print zsh, bash, or fish completion. |
| `scrobbler version` | Print version and build commit. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Common options

These options apply to `manual`, `file`, and `discography`:

| Option | Meaning |
| --- | --- |
| `--loop N` | Album loop count. |
| `--limit N` | Tracks per album; `0` means all tracks. |
| `--interval DURATION` | Delay between scrobbles, such as `2s`. |
| `--dry-run` | Resolve and print the queue without scrobbling. |
| `--json` | Emit machine-readable output where supported. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Manual

```bash
scrobbler manual "Artist - Album"
scrobbler manual --artist "Artist" --album "Album"
```

Options:

```text
--artist NAME
--album TITLE
--loop N
--limit N
--interval DURATION
--dry-run
--json
```

A dry run resolves the album and exact queue but does not authenticate or send
anything.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> File

```bash
scrobbler file [options] PATH
```

`PATH` can be TXT, CSV, TSV, JSON, M3U/M3U8, an album folder, or an artist
folder. Import errors identify the specific entry that failed.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Discography

Listing results without scrobbling:

```bash
scrobbler discography "Artist"
```

Selection options:

```text
--all                    use all returned albums
--first N                use the first N results
--albums "A,B,C"         exact comma-separated album titles
--clean                  remove obvious duplicate editions and reissues
```

Common loop, limit, interval, dry-run, and JSON options also apply.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Similar

```bash
scrobbler similar "Artist" --limit 20
scrobbler similar "Artist" --json
```

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Connection test

```bash
scrobbler test
scrobbler test --json
```

The test checks:

1. API key presence
2. a real read-only Last.fm request
3. API secret presence
4. authentication readiness: username/password mode obtains a mobile session;
   an existing session key is reported as configured and is validated by the
   first signed write

It does not submit a scrobble.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Diagnostics

```bash
scrobbler diagnostics
scrobbler diagnostics --json
```

The JSON form returns:

```json
{"path":"/Users/name/Downloads/lastfm-scrobbler-diagnostics-20260807-120000.zip"}
```

Scrobbling commands also write pending and completed session records to the
normal data directory, so interrupted CLI sessions can be inspected alongside
TUI sessions.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Update checker

```bash
scrobbler check-update
scrobbler check-update --json
```

Configure `SCROBBLER_UPDATE_URL` or inject a repository with build-time
`-ldflags`.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Exit codes

```text
0  success
1  API, authentication, filesystem, or scrobble failure
2  command syntax or option error
```
