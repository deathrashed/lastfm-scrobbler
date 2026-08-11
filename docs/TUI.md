# <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="24" height="24" alt="Last.fm icon"> TUI controls and workflows

The Bubble Tea interface uses a responsive 67–127-cell application with Torch
Red active controls, centered panels, contextual footer hints, and optional
mouse support. At 67 columns it retains the approved compact design. As the
terminal grows, the outer header can expand to 127 cells while dense working
surfaces grow only as far as they remain visually cohesive (currently 103
cells) and stay centered inside the application. Natural-width navigation
cards use bounded spacing rather than stretching to the edges. Above 127 cells
the application remains 127 cells wide and centers itself. Terminal resizing
updates rendered geometry and mouse regions live. Run `scrobbler` without a
subcommand to open it.

The UI requires a terminal at least 67 columns wide. Compact Header is enabled
only by `SCROBBLE_COMPACT_HEADER` or **Settings → Interface → Compact Header**;
it does not switch on automatically for narrow terminals. Compact mode uses a
four-line header until Manual or Discography has resolved an artist; then it
adds one centered `ARTIST ❯ NAME` metadata row inside the responsive working
surface. The full header retains the profile URL, artwork, and its existing
attached artist badge.

## First-run setup wizard

An unconfigured interactive launch opens the setup wizard. Run `scrobbler
setup` to start it explicitly later. The flow covers system detection, a
curated optional Nerd Font selection, Last.fm account credentials, supported
credential storage, recommended scrobbling defaults, interface preferences,
review, and Apply/Connection Test. The review page is always before the first
persistent write. Escaping or quitting earlier leaves the real configuration,
font directory, and terminal configuration untouched.

The wizard is cross-platform on macOS, Linux, and Windows. It uses the
existing macOS Keychain backend only where available, and otherwise offers
credentials-file or environment-variable storage. Nerd Fonts are downloaded
from official latest-release family archives and installed at user scope.
Ghostty is the only terminal with an automatic font configuration adapter at
present; unsupported terminals are reported as manual setup rather than being
edited heuristically.

## Nerd Font icons

A Nerd Font is recommended for the intended TUI appearance because the screens
use Nerd Font glyphs for icons. Any compatible Nerd Font is acceptable. On
macOS with Homebrew, one optional choice is:

```bash
brew install --cask font-jetbrains-mono-nerd-font
```

Then select `JetBrainsMono Nerd Font` (or another installed Nerd Font) in your
terminal settings. Core functionality remains available if an icon renders
incorrectly. The setup wizard can download an official release archive without
requiring Homebrew or another package manager, and a Nerd Font remains
optional.

<p align="center">
  <a href="../README.md">Overview</a> •
  <a href="CLI.md">CLI</a> •
  <a href="CONFIGURATION.md">Configuration</a> •
  <a href="AUTOMATION.md">Automation</a>
</p>

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Global controls

| Key | Action |
| --- | --- |
| `↑ ↓ ← →` / `J K` | Navigate the active list, grid, or tabs. |
| `Enter` | Confirm, save, open, or continue. |
| `Esc` | Go back or cancel. |
| `Q` / `Ctrl+C` | Quit. Active sessions cancel promptly and retain recovery state. |
| `?` | Open the contextual Help overlay. |
| Mouse wheel | Navigate the active list/section when Mouse Support is enabled. |

When Mouse Support is enabled, visible section boxes, tabs, cards, list rows,
editable areas, action panels, footer actions, and the Help close hint accept
mouse input. Keyboard controls remain available everywhere. The full-header
Last.fm URL is Torch Red at rest, turns white on hover, and opens when clicked;
Compact Header intentionally has no URL hit area.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Dashboard

| Key | Destination |
| --- | --- |
| `M` | Manual album entry. |
| `D` | Discography results for an artist, sourced from Last.fm top albums. |
| `F` | File, playlist, or folder import. |
| `H` | History inside the Settings shell. |
| `S` | Settings, opening on Scrobbling. |
| `I` | In-app Info and documentation. |
| `R` | Last Session, for an exact rerun or edit-first rerun. |

The Dashboard footer currently reads:

```text
enter select • → ↑ navigate ↓ ← • s settings
i info • h history • m d quick f q • r rerun • ? help
```

Profiles are opened through **Settings → Profiles**, not a Dashboard shortcut.

### Full-header activity

**Settings → Interface → Now Playing** is enabled by default and applies only
to the full header. The TUI reads one item through Last.fm's
[`user.getRecentTracks`](https://www.last.fm/api/show/user.getRecentTracks)
method. While this display is enabled, the profile URL is promoted into the
attached top badge above the activity row. A current track uses the animated
volume sequence, while the most recent non-playing scrobble uses a distinct
static history icon. Loading, no-history, and unavailable states keep the same
reserved activity row. The display refreshes about every 30 seconds and never
submits playback state.
Compact Header deliberately remains activity-free; it only adds its approved
Manual or Discography artist row after an artist is resolved.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Settings

Settings replaces the former Config/Advanced split with six first-class
sections. At the minimum width, the two rows use the familiar `19 / 25 / 19`
geometry shown below. On wider terminals the cards retain their natural widths
and the extra space becomes centered breathing room between them:

```text
╭─────────────────╮ ╭───────────────────────╮ ╭─────────────────╮
│  A C C O U N T  │•│  S C R O B B L I N G  │•│  H I S T O R Y  │
╰─────────────────╯ ╰───────────────────────╯ ╰─────────────────╯
╭─────────────────╮ ╭───────────────────────╮ ╭─────────────────╮
│    T O O L S    │•│   I N T E R F A C E   │•│ P R O F I L E S │
╰─────────────────╯ ╰───────────────────────╯ ╰─────────────────╯
```

Each section changes the main header title and subtitle. Account, Scrobbling,
Tools, and Interface show an overview list plus a detail/editor panel for the
selected row. History and Profiles preserve their existing list/action
workflows inside the same navigation shell.

| Section | Contents |
| --- | --- |
| **Account** | Last.fm username/password, API key/secret, credential source, credential path. |
| **Scrobbling** | Loop, interval, retry count/delay, duplicate guard, Clean Discography. |
| **History** | Saved sessions, edit/exact re-runs, export, delete. |
| **Tools** | Export directory, update source, connection test, diagnostics, completions, update check. |
| **Interface** | Notifications, Now Playing, Compact Header, Mouse Support. |
| **Profiles** | Load, create, save, and delete named profiles. |

### Keyboard focus

Settings has two explicit keyboard focus zones: the six-section grid and the
current section content.

- `Tab` or `Shift+Tab` from content focuses the section grid.
- `↑ ↓ ← →` in the grid changes section spatially.
- `Enter` or `Tab` on the grid returns focus to that section's content.
- `↑` from the first content row returns to the section grid.
- `↑` / `↓` inside content moves through rows.
- `←` / `→` adjusts toggles/choices; in text fields it retains normal cursor
  movement when no setting adjustment applies.
- `Enter` saves an editable value or opens/runs the selected action.
- `Esc` returns to the Dashboard.

Mouse users can click a section box or row directly. Clicking a toggle/choice
control changes it; clicking text content focuses the editor; clicking an
action panel invokes the same action as its keyboard control.

### Interaction colors

Rows use a deliberately restrained hierarchy:

- **Idle row:** title white, row `❯` Torch Red, value muted (`#736765`).
- **Keyboard-focused row:** leading `❯` and title Torch Red; row arrow and value
  white.
- **Mouse-hovered row:** title Torch Red; row arrow/value white without moving
  the keyboard cursor.

Navigation cards (Settings sections, File sources, and Info tabs) use a distinct
state language:

- **Idle card:** white border, muted label.
- **Hovered card:** white border, Torch Red label.
- **Selected card:** Torch Red border, bold white label.

Text-input borders stay white while editing. Focus is communicated by a Torch
Red label, white `❯`, and the text input's blinking Torch Red cursor. Torch Red
borders are therefore reserved for selected cards/actions and real attention
states rather than ordinary typing.

Multi-select lists use `○` for unselected and `●` for selected items, with a
leading `❯` for keyboard focus. Asynchronous Last.fm/diagnostic/update work uses
the custom Last.fm bounce spinner ` ∙ ∙ → ∙  ∙ → ∙ ∙  → ∙  ∙`.

Interactive footer actions expose contextual mouse help. Hover a footer action
to show a concise white description below the controls; when the explanation
names the current album, that dynamic album name is highlighted in Torch Red.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Info

The Info page is a compact reference for the main workflows and local data.
Move between sections with `←` and `→`, click a tab, or use the mouse wheel:

| Section | Covers |
| --- | --- |
| **Modes** | Manual entry, Discography curation, file imports, similar albums. |
| **Automation** | Headless CLI, Settings, connection tests, completions, mouse, updates. |
| **Data** | History, recovery, Profiles, Account credentials, local storage. |
| **Curation** | Track selection, loop controls, saved-queue editing, exports. |
| **Imports** | TXT, CSV, TSV, JSON, M3U/M3U8, album/artist folders, native picker. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Help

Press `?` where Help is available. Close it with `?`, `Esc`, or `Enter`.
With Mouse Support enabled, hover the final `close` hint to turn it white and
click it to return to the underlying screen.

Help advertises `S` for Settings on the Dashboard and no longer requires
separate Advanced or credentials-path shortcuts.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Manual

Search by artist, album, or both, choose a matching Last.fm result when necessary, then
select tracks and preview the queue before starting the scrobble.

Once an artist has been resolved, the full header grows a centered artist badge
beneath `M A N U A L`; its width follows the artist name. Compact mode adds the
centered `ARTIST ❯ NAME` row only for resolved Manual and Discography stages.
The search-result list attaches a `RESULTS` count to
its lower-right border instead of showing a misleading multiselect counter. The
track-selection list attaches its `SELECTED` counter the same way.

| Key | Action |
| --- | --- |
| `Enter` | Search or continue. |
| `Space` | Toggle a track. |
| `A` | Select all or select none. |
| `-` / `+` | Change the current album loop while selecting tracks. |
| Mouse footer `- interval +` | Decrease or increase the delay between track scrobbles. |
| Mouse footer `↑ navigate ↓` | Move directly through the track list. |
| `E` | Export the queue. |
| `S` | Find similar albums. |

The track-selection footer is also a mouse control surface:

```text
space check • s similar • a all • enter continue
    - interval + • ↑ navigate ↓ • - loop +
```

Click the individual `-`, `+`, `↑`, or `↓` symbols to adjust the value or move
the track cursor directly. In multi-album Discography workflows, `- loop +`
changes the loop count for the album containing the current track.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Discography

Search an artist, filter or sort the returned albums, clean obvious duplicate
editions, select albums, and load their tracks into one queue. The user-facing
feature is Discography; its source is Last.fm's `artist.getTopAlbums` result,
not a canonical complete discography.

After the artist is resolved, the full header carries the same dynamic artist
badge used by Manual. The album chooser is one integrated component: `SORT`,
`FILTER`, and `CLEAN` controls attach to its top border, while `RESULTS` and
`SELECTED` counts attach to the bottom. Short filter text stays in the compact
center control; editing or a filter too long to fit expands a connected wide
filter field beneath it. Multi-album track selection shows a `TASK` summary plus
the current `ALBUM` and its loop count. The preview queue lists the first five
queued album titles in order (with `… N more` when needed) before the centered
ALBUMS/TRACKS, INTERVAL/SCROBBLES, and ETA/LOOP summary cards.

| Key | Action |
| --- | --- |
| `Space` | Toggle an album. |
| `A` | Select all or select none. |
| `/` or `F` | Filter album titles. |
| `C` | Toggle duplicate/reissue cleanup. |
| `S` | Change sorting mode. |
| `-` / `+` | Change the current album loop. |
| `Enter` | Load selected albums. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> File import

The File page combines the four source cards and PATH editor. It supports TXT,
CSV, TSV, JSON, M3U/M3U8, one album folder, or an artist folder containing
album folders. Press `O` for the platform picker when available, or enter PATH
manually, then press `Enter` to import.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> History, recovery, and Profiles

`H` from the Dashboard opens History directly while retaining the Settings
section grid. Profiles remain available through **Settings → Profiles**.

`R` opens Last Session. `Enter` performs an exact rerun through the existing
queue path, `E` opens the existing edit-first track-selection path, and `Esc`
returns to the Dashboard. If no completed queue exists, the screen reports
that no previous session is available.

History controls:

| Key | Action |
| --- | --- |
| `Enter` / `R` | Edit a saved queue before re-running it. |
| `Shift+R` | Perform an exact re-run without editing. |
| `E` | Export the selected history entry. |
| `D` | Delete the selected history entry. |
| `Tab` | Focus the Settings section grid. |
| `Esc` | Return to the Dashboard. |

Profiles controls:

| Key | Action |
| --- | --- |
| `Enter` | Load the selected profile. |
| `N` | Create a profile. |
| `S` | Save the selected profile. |
| `D` | Delete the selected profile. |
| `Tab` | Focus the Settings section grid. |
| `Esc` | Return to the Dashboard. |

Interrupted active queues are offered for resume at the next launch.
