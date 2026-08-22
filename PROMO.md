# lastfm-scrobbler — Promotion Post Draft

> Draft for sharing on Reddit / HN / r/selfhosted / r/commandline / r/golang.
> Repo: https://github.com/deathrashed/lastfm-scrobbler
> Screenshots: https://github.com/deathrashed/lastfm-scrobbler#screenshots

---

## Main post (Reddit — r/commandline / r/selfhosted friendly)

**Title:** I built a cross-platform Last.fm scrobbler with a TUI, CLI, and automation support — `scrobbler`

**Body:**

I got tired of Last.fm scrobbling being a "background daemon you set up once and
forget" — so I built a scrobbler you actually *use*. It's a responsive terminal
app (Bubble Tea) that lets you search, curate, preview, and scrobble album
queues interactively, plus a headless CLI for automation.

**What it does:**

- **Manual scrobbling** — search by artist, album, or both, pick tracks, preview
  the queue, scrobble with configurable loops and intervals
- **Discography curation** — pull an artist's top albums, filter/sort/clean,
  select any combination, and build one queue
- **Import workflows** — TXT, CSV, TSV, JSON, M3U/M3U8 playlists, album folders,
  or artist folders
- **Recovery** — history, resume interrupted queues, edit saved sessions, exact
  re-runs
- **Automation** — stable JSON-capable CLI commands for Keyboard Maestro, shell
  scripts, and launch agents
- **A real TUI** — live-resizes from 67–127 columns, centers on wider terminals,
  mouse support, Nerd Font icons, Last.fm Torch Red theme

**Install:**

```bash
# macOS
brew install deathrashed/tap/scrobbler

# or anywhere
go install github.com/deathrashed/lastfm-scrobbler/cmd/scrobbler@latest
```

Prebuilt binaries for macOS (Apple Silicon + Intel), Linux (x86_64 + ARM64), and
Windows x64 — no Go or package manager required. First-run wizard handles
credentials (macOS Keychain supported), and `scrobbler test --json` verifies
your connection without submitting a scrobble.

Headless example:

```bash
scrobbler manual "Slayer - Hell Awaits" --loop 2
scrobbler discography "Demolition Hammer" --first 3 --dry-run
scrobbler file ~/Downloads/albums.txt --dry-run
```

WTFPL-2 licensed. Feedback, issues, and feature requests welcome!

---

## Posting log & status

- **2026-08-19 — r/commandline post attempt.** Posted the main version. Reddit
  flagged the account for "suspicious activity" (automated security check —
  password reset required, done). The `github-guard` bot scored the repo 0/1:
  **"New Repository (under 30 days old)"** — first commit is 2026-08-07, so the
  repo is 11 days old. r/commandline rule #5: *"No new projects newer than 30
  days."* The 30-day mark is **2026-09-06** — repost to r/commandline after
  that date. Other subreddits (r/golang, r/selfhosted megathread, r/lastfm)
  have no age rule and can be posted now.

---

## Variant — r/golang (technical angle)

**Title:** [Showoff] Last.fm scrobbler in Go — Bubble Tea TUI + headless CLI sharing one core

**Body:**

Built a cross-platform Last.fm scrobbler in Go. The interesting bit: the TUI
(Bubble Tea) and the headless CLI share the same configuration, Last.fm client,
queue-building, history, export, and diagnostics packages — so interactive and
automation workflows stay behaviorally aligned.

- Responsive TUI that live-resizes 67–127 columns and centers on wider terminals
- Stable JSON output for automation (`--dry-run`, `--json`, exit codes 0/1/2)
- macOS Keychain credential backend, cross-platform first-run wizard
- Shell completions for Zsh, Bash, Fish, PowerShell
- Redacted diagnostics bundles for safe troubleshooting
- `go test ./...` + `go vet ./...` clean; tests cover parsing, imports, exports,
  recovery, queue construction, layout, and diagnostics redaction

```bash
go install github.com/deathrashed/lastfm-scrobbler/cmd/scrobbler@latest
```

https://github.com/deathrashed/lastfm-scrobbler

---

## Variant — r/selfhosted (New Project Megathread format)

> Post this as a top-level comment in the weekly New Project Megathread.
> Follows the subreddit's recommended template exactly.

**Project Name:** scrobbler

**Repo/Website Link:** [https://github.com/deathrashed/lastfm-scrobbler](https://github.com/deathrashed/lastfm-scrobbler)

**Description:**

A cross-platform Last.fm scrobbler you actually *use* instead of a background
daemon you set up once and forget. It's a responsive terminal app (Go + Bubble
Tea) that lets you search, curate, preview, and scrobble album queues
interactively, plus a headless CLI for automation.

Features:

- **Manual scrobbling** — search by artist, album, or both, pick tracks, preview
  the queue, scrobble with configurable loops and intervals
- **Discography curation** — pull an artist's top albums, filter/sort/clean,
  select any combination, and build one queue
- **Import workflows** — TXT, CSV, TSV, JSON, M3U/M3U8 playlists, album folders,
  or artist folders
- **Recovery** — session history, resume interrupted queues, edit saved
  sessions, exact re-runs
- **Automation** — stable JSON-capable CLI commands for shell scripts, launch
  agents, and Keyboard Maestro
- **A real TUI** — live-resizes from 67–127 columns, centers on wider terminals,
  mouse support, Nerd Font icons, Last.fm Torch Red theme
- **Privacy** — no cloud, no account beyond your Last.fm credentials (stored in
  macOS Keychain or an owner-only file)

**Deployment:**

This is a client-side tool, not a server — there's no Docker image because
there's nothing to host. It runs on macOS, Linux, and Windows, and is released
with prebuilt binaries for each platform (Apple Silicon + Intel, x86_64 +
ARM64, Windows x64) — no Go or package manager required.

```bash
# macOS
brew install deathrashed/tap/scrobbler

# or anywhere
go install github.com/deathrashed/lastfm-scrobbler/cmd/scrobbler@latest
```

First-run wizard handles credentials, and `scrobbler test --json` verifies your
connection without submitting a scrobble. Installation and usage are documented
in the README.

**AI Involvement:**

AI-assisted development. The core logic — Last.fm API client, queue building,
import/export parsing, recovery, and the CLI — is hand-written. AI was used for
scaffolding, boilerplate, and debugging support during development.

WTFPL-2 licensed. Feedback, issues, and feature requests welcome!

---

## Variant — r/lastfm (community angle)

**Title:** I made a TUI scrobbler for people who actually curate their scrobbles

**Body:**

Most scrobblers are passive — they just record what you play. I wanted one that
lets you *choose* what gets scrobbled: search an artist, preview the queue,
scrobble with loops, import playlists or album folders, and even re-run or
resume past sessions. It's a terminal app (works on macOS, Linux, Windows).

```bash
brew install deathrashed/tap/scrobbler   # macOS
```

Screenshots in the README. Happy to take feedback!

https://github.com/deathrashed/lastfm-scrobbler

---

## Posting tips

- **r/commandline** — post the main version; mention the responsive TUI + mouse
  support; screenshots help a lot
- **r/golang** — use the technical variant; be ready for "why not just use
  lastfm-cli?" — answer: interactive curation + recovery + automation
- **r/selfhosted** — emphasize no cloud dependency, local credentials, WTFPL
- **r/lastfm** — community angle; screenshots of the Torch Red UI will land well
- **Hacker News** — use the main post with a "Show HN:" prefix; keep the body
  tight, lead with the screenshots link
- Post screenshots as an image/gallery where the platform allows (Reddit image
  posts outperform link posts)
- Don't post the same text verbatim everywhere — subreddits penalize
  cross-posted duplicates; use the variants above
