# Changelog

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
