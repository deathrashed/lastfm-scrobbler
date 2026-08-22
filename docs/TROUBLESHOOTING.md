# Troubleshooting

## Nerd Font icons look wrong

Install a compatible Nerd Font and select its Mono variant in the terminal
profile. The wizard can install a curated official font at user scope; package
managers are optional. If a terminal does not support automatic configuration,
follow its manual font instructions. The application remains usable with a
fallback font.

## Last.fm authentication fails

Run `scrobbler test --json`. Check the API key, API secret, username, and either
password or session key. Confirm that the selected credential source and
`LASTFM_ENV_FILE` point to the intended values. Never include secrets in a bug
report; share only a redacted diagnostics bundle.

## Scrobbling reports “Invalid session key” (error 9)

The stored session key expired or was revoked on Last.fm's side. Press `A` on
the paused scrobble screen or use **Settings → Account → RE-AUTHENTICATE**:
the app requests a fresh token, opens the Last.fm authorization page in your
browser, and exchanges the authorized token for a new session key after you
grant permission. The new key is saved through your normal credential source,
the running client updates immediately, and the interrupted queue keeps its
progress — already-scrobbled tracks are not repeated. If the browser cannot
open, copy the auth URL into a browser manually. Other Last.fm API errors are
not authentication failures and keep their existing behavior.

## Credentials appear missing

Review **Settings → Account → Credential Source** and **Credential Path**.
`auto` combines process environment, configured files, and macOS Keychain where
implemented. Linux and Windows use the file/environment backends. Credential
files should be owner-only and must not be committed.

## Completion does not load

Run `scrobbler completion install SHELL`, then reload the shell/profile. Check
the command's status message and ensure the generated file is in the documented
per-user location. Repeated installs are safe and do not duplicate marker
lines.

## The picker is unavailable

Enter PATH manually. On Linux, install or expose `zenity`, `kdialog`, or `yad`
if a desktop dialog is wanted. On Windows, ensure PowerShell and Windows Forms
are available. On macOS, confirm the process can launch `osascript`.

## Terminal font setup says manual

Font installation and terminal configuration are separate. Unsupported or
unknown terminal adapters intentionally do not edit arbitrary files. Configure
the terminal profile manually after the font is installed.

## Updates fail

The empty update source uses the official GitHub Releases API. Check network
access, proxy settings, and the current build's repository metadata. If using a
custom endpoint, verify its URL and response fields. No update URL is needed
for official builds.

## Now Playing is unavailable

Now Playing is a full-header-only convenience display. Check **Settings →
Interface → Now Playing**, confirm a Last.fm username and API key are
configured, and verify network access to Last.fm. A temporary
`user.getRecentTracks` failure is shown as unavailable without interrupting
Manual, Discography, File, or scrobbling.

## File imports fail

Confirm the path exists and matches the selected source type. Use the supported
TXT/CSV/TSV/JSON, M3U/M3U8, album-folder, or artist-folder forms documented in
[File imports](FILE-IMPORTS.md). Windows paths are accepted as Windows paths;
do not rewrite them with POSIX separators.
