# <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="24" height="24" alt="Last.fm icon"> TUI controls and workflows

The Bubble Tea interface uses a fixed-width layout with Torch Red active
controls, centered panels, footer hints, and mouse support. Run `scrobbler`
without a subcommand to open it.

The UI requires a terminal at least 67 columns wide. Compact Header is only
enabled by the `SCROBBLE_COMPACT_HEADER` setting; it does not switch on
automatically for narrow terminals. Compact mode uses a fixed four-line header,
while the full header retains the profile URL and artwork.

<p align="center">
  <a href="../README.md">Overview</a> •
  <a href="CLI.md">CLI</a> •
  <a href="CONFIGURATION.md">Configuration</a> •
  <a href="AUTOMATION.md">Automation</a>
</p>

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Global controls

| Key | Action |
| --- | --- |
| `↑ ↓ ← →` / `J K` | Navigate lists and tabs. |
| `Enter` | Confirm, save, open, or continue. |
| `Esc` | Go back or cancel. |
| `Q` / `Ctrl+C` | Quit. Active sessions cancel promptly and retain recovery state. |
| `?` | Open the contextual help overlay. |
| Mouse wheel | Navigate long lists when mouse support is enabled. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Dashboard

| Key | Destination |
| --- | --- |
| `M` | Manual album entry. |
| `D` | Artist discography. |
| `F` | File, playlist, or folder import. |
| `H` | Session history and recovery. |
| `P` | Saved profiles. |
| `C` | Configuration. |
| `I` | In-app Info and documentation. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Info

The Info page is a compact reference for the main workflows and local data
used by the application. Move between sections with `←` and `→`:

| Section | Covers |
| --- | --- |
| **Modes** | Manual entry, discography curation, file imports, and similar albums. |
| **Automation** | Headless CLI commands, connection tests, completions, mouse support, and update checks. |
| **Data** | History, recovery, profiles, diagnostics, and local storage. |
| **Curation** | Track selection, loop controls, saved-queue editing, and exports. |
| **Imports** | TXT, CSV, TSV, JSON, M3U/M3U8, album folders, artist folders, and the native picker. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Manual

Enter `Artist - Album`, choose a matching Last.fm result when necessary, then
select tracks and preview the queue before starting the scrobble.

| Key | Action |
| --- | --- |
| `Enter` | Search or continue. |
| `Space` | Toggle a track. |
| `A` | Select all or select none. |
| `-` / `+` | Change the global album loop. |
| `[` / `]` | Change the current-album loop. |
| `E` | Export the queue. |
| `S` | Find similar albums. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Discography

Search an artist, filter or sort the returned albums, clean obvious duplicate
editions, select albums, and load their tracks into one queue.

| Key | Action |
| --- | --- |
| `Space` | Toggle an album. |
| `A` | Select all or select none. |
| `/` or `F` | Filter album titles. |
| `C` | Toggle duplicate/reissue cleanup. |
| `S` | Change sorting mode. |
| `-` / `+` | Change the global loop. |
| `Enter` | Load selected albums. |

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> File import

The File page supports TXT, CSV, TSV, JSON, M3U/M3U8, one album folder, or an
artist folder containing album folders. Press `O` to open the native picker,
then `Enter` to import the selected path.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Completion and history

After a queue finishes or is cancelled:

| Key | Action |
| --- | --- |
| `R` | Edit a saved queue before re-running it. |
| `Shift+R` | Perform an exact re-run without editing. |
| `E` | Export a queue or history entry. |
| `S` | Find similar albums. |
| `H` | Return to History. |
| `Esc` | Return to the Dashboard. |

Interrupted active queues are offered for resume at the next launch.

With the full header, the Last.fm profile URL is both an OSC 8 hyperlink and a
mouse target. Moving over it highlights the URL and clicking opens it. Compact
Header intentionally has no URL or URL hit area.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Config

The first row contains direct values. The second row opens larger tools.

| Key | Action |
| --- | --- |
| `←` / `→` | Move within the current row. |
| `↑` / `↓` | Change rows. |
| `Tab` / `Shift+Tab` | Move between username/password or key/secret fields. |
| `Enter` | Save an editable field or open a utility. |
| `Ctrl+P` | Open the credentials path editor. |
| `Ctrl+G` | Open Advanced. |
| `Ctrl+O` | Open Info. |

When a text field is focused, printable characters are sent to the field. This
means usernames, passwords, API keys, paths, and filters can contain letters
such as `a`, `h`, `p`, and `i` without activating navigation.

Advanced includes connection testing, diagnostics, update checking, retry
settings, duplicate protection, notifications, exports, and mouse support.
