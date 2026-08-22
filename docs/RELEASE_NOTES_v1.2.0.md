## Highlights

Last.fm Scrobbler v1.2.0 adds in-app Last.fm reauthentication and a refreshed,
height-aware header while keeping the existing scrobble workflows intact.

## Last.fm reauthentication

Sessions expire. Until now an expired session surfaced as
`last.fm API error 9: Invalid session key` on every scrobble with no way to
repair it inside the app.

When a signed request now fails with error 9, the run pauses instead of
retrying: already-scrobbled tracks are never resubmitted, the remaining queue
is preserved on disk, and the scrobble screen shows
`Last.fm authentication expired or was revoked` with a `a re-authenticate`
footer action.

Pressing it (or **Settings → Account → RE-AUTHENTICATE**) opens the new auth
screen:

1. A fresh token is requested via signed `auth.getToken`.
2. The browser opens at `https://www.last.fm/api/auth/` with the configured
   API key and pending token.
3. After granting permission, `enter` (`GET SESSION KEY`) exchanges the same
   token via signed `auth.getSession`.
4. The session key persists through the existing credential-source rules
   (Keychain for `auto`/`keychain`, credentials file for `file`), the live
   client is rebuilt immediately, and no restart is needed.
5. Auth returns to where you came from — Settings, the Dashboard, or the
   interrupted scrobble run, which shows preserved progress such as
   `RETURN ❯ SCROBBLING • 7 / 12 preserved`.

The auth screen uses the same attached-capsule grammar as the rest of the TUI:
a state-driven STATUS capsule (`WAITING`, `FETCHING`, `AUTHENTICATED`,
`FAILED`, `EXPIRED`), an ACCOUNT capsule, CURRENT ACTION instructions, step
state glyphs, and an optional RESULT capsule that summarizes API errors as
`Error N · reason`. Raw API messages and all credentials stay out of the main
UX; mouse regions match the visible buttons exactly.

## Adaptive header

Header density is chosen from terminal height: tall terminals get the framed
Last.fm hero wordmark — red full-block glyphs with a dim box-drawing shadow —
medium terminals keep the classic header, and short terminals fall back to the
compact header automatically. **Settings → Interface → Compact Header** still
forces compact mode at every size. In the hero layout the profile URL is
embedded in the top frame (still clickable) and current/recent activity from
the optional Now Playing setting becomes the lower frame caption. Muted-text
contrast is higher, and footer hover help collapses on short terminals so
workflows keep vertical space.

## Platforms

Supported release assets are unchanged:

- macOS Apple Silicon (`darwin-arm64`)
- macOS Intel (`darwin-amd64`)
- Linux x86_64 (`linux-amd64`)
- Linux ARM64 (`linux-arm64`)
- Windows x64 (`windows-amd64`)

Each archive contains the platform executable, `LICENSE`, and completion files
for zsh, bash, fish, and PowerShell; `checksums.txt` covers every asset.

## Installation / upgrade

Download the matching archive, verify it with `checksums.txt`, and place the
binary on your PATH. Alternatively:

```bash
go install github.com/deathrashed/lastfm-scrobbler/cmd/scrobbler@v1.2.0
```

Upgrading from v1.1.0 requires no configuration changes. If your stored session
key was revoked, start the app and follow the reauthenticate prompt.
