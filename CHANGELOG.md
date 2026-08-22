# Changelog

## v1.2.0 — 2026-08-22

### Added

- In-app Last.fm reauthentication. When a signed request fails with
  `Invalid session key` (Last.fm API error 9), the scrobble run pauses instead
  of retrying with the dead key: completed tracks are never resubmitted, the
  remaining queue is preserved, and the scrobble screen offers
  `a re-authenticate`. **Settings → Account** gains an auth status row plus a
  `RE-AUTHENTICATE` / `GRANT LAST.FM PERMISSION` action.
- Dedicated auth screen following the app's attached-capsule grammar:
  `STATUS ❯ WAITING / FETCHING / AUTHENTICATED / FAILED / EXPIRED` and
  `ACCOUNT ❯ name` capsules, a state-driven CURRENT ACTION section, per-step
  state glyphs (`✓ ● ○ ✗`), state-specific actions
  (`OPEN LAST.FM ↗`, `GET SESSION KEY`, `START AGAIN ↻`, context-aware
  `RETURN TO …`), and an optional RESULT capsule that summarizes API errors as
  `Error N · reason` without exposing tokens or secrets.
- Desktop-style browser authorization: a fresh `auth.getToken` request opens
  `https://www.last.fm/api/auth/` with the configured API key and pending
  token; after granting, `enter` exchanges the same token through
  `auth.getSession`.
- Exchanged sessions persist immediately through the existing credential-source
  rules (macOS Keychain for `auto`/`keychain`, credentials file for `file`),
  the live client is rebuilt in place, and no restart is required. Returning
  from auth restores the interrupted workflow — including paused scrobble runs,
  which show preserved progress such as `RETURN ❯ SCROBBLING • 7 / 12
  preserved`.

### Changed

- Adaptive header system inspired by compact height-aware layouts: tall
  terminals use the framed Last.fm hero wordmark (red full-block glyphs with
  dim box-drawing shadow), medium terminals keep the classic header, and short
  terminals fall back to the compact header automatically. Enabling Compact
  Header still forces compact mode at every size.
- Hero hierarchy now centers on live Last.fm state: `• SCROBBLER •` sits above
  the wordmark, the configured `last.fm/user/...` URL is embedded in the top
  frame and remains clickable, and current/recent `Artist - Track` activity
  becomes the lower frame caption.
- Increased muted-text contrast while preserving the red/white identity; footer
  hover help collapses from two rows to one or zero on short terminals so the
  active workflow keeps more vertical space.
- Auth screen mouse regions match visible buttons exactly (`open`,
  `get-session`, `retry`, dynamic `return`); keyboard and mouse invoke the same
  underlying actions.

### Fixed

- `auth.getToken` and `auth.getSession` requests are now signed with the API
  secret, as Last.fm requires.

## v1.1.0 — 2026-08-12

### Added

- Cross-platform first-run setup wizard with staged Review / Apply flow,
  platform-aware credential choices, user-scope Nerd Font installation, and
  Ghostty configuration where supported.
- Optional full-header Last.fm activity from `user.getRecentTracks`, with
  animated current-track status and recent-track fallback.
- Unified File source and PATH workflow for list files, playlists, album
  folders, and artist folders, with native or detected platform pickers.
- Discography curation, Last Session reruns, six Settings sections, and
  per-user zsh, bash, fish, and PowerShell completion installation.

### Changed

- Responsive TUI now adapts from 67 to 127 application columns, centers on
  wider terminals, and keeps dense working panels bounded and aligned.
- Release packaging now supports macOS Apple Silicon and Intel, Linux x86_64
  and ARM64, and Windows x64 archives with checksums and completion files.
- GitHub Releases is the default update source for official builds.

### Fixed

- Corrected responsive card proportions, File source and attached-badge
  geometry, Discography filter connectors, results heights, mouse regions,
  shell-completion layout, and full-header activity alignment.
- Removed the Dashboard `p profiles` shortcut; Profiles remain under
  **Settings → Profiles**.

## v1.0.1 — 2026-08-07

- Corrected release metadata and checksum paths for the published macOS
  binaries.

## v1.0.0 — 2026-08-07

- Initial tagged release of Last.fm Scrobbler.
