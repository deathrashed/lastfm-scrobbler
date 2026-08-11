## Highlights

Last.fm Scrobbler v1.1.0 expands the TUI and release packages across macOS,
Linux, and Windows while keeping the existing Manual scrobble workflow intact.

## Cross-platform setup

The first-run Setup Wizard detects macOS, Linux, or Windows, stages changes
until Review / Apply, installs a supported Nerd Font at user scope, and
configures Ghostty where supported. Credential storage choices follow the
platform capabilities. Canceling before Apply is safe and leaves the existing
configuration untouched.

## Responsive TUI

The application now resizes live from 67 to 127 columns and centers itself on
wider terminals. Working panels use bounded responsive widths, with improved
height usage, mouse geometry, card alignment, attached badges, and narrow-mode
behavior. Profiles remain available under **Settings → Profiles**.

## Last.fm activity

Optional **Settings → Interface → Now Playing** reads `user.getRecentTracks`
for the full header, shows an animated current-track icon or a distinct recent
scrobble fallback, and refreshes about every 30 seconds. It only reads Last.fm
activity; it does not submit `track.updateNowPlaying` or any other playback
state.

## File and Discography workflows

File now combines source selection and PATH entry for TXT, CSV, TSV, JSON,
M3U/M3U8, album folders, and artist folders, with native or detected
cross-platform pickers. Discography adds user-facing filter, sort, clean,
multiselect, connected filter layout, and long-title handling. It is sourced
from Last.fm's `artist.getTopAlbums` API and is not a canonical complete
discography.

## Automation and completions

Last Session provides an exact rerun and an edit-first rerun without an
accidental immediate scrobble. Completion generation and per-user installation
support zsh, bash, fish, and PowerShell. Settings includes diagnostics,
connection testing, GitHub Releases update checks, and six organized sections.

## Fixes

The release includes responsive layout and alignment fixes across Dashboard,
Manual, Discography, File, Settings, Info, Help, Shell Completions, History,
Last Session, and progress/completed states.

## Platforms

Supported release assets are:

- macOS Apple Silicon (`darwin-arm64`)
- macOS Intel (`darwin-amd64`)
- Linux x86_64 (`linux-amd64`)
- Linux ARM64 (`linux-arm64`)
- Windows x64 (`windows-amd64`)

## Installation / upgrade

Download the matching archive, verify it with `checksums.txt`, and place the
binary on your PATH. The archives include `LICENSE` and all four completion
files. Alternatively:

```bash
go install github.com/deathrashed/lastfm-scrobbler/cmd/scrobbler@v1.1.0
```
