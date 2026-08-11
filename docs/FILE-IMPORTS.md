# File imports

The TUI File screen combines source selection and PATH entry. Choose one of
four sources, enter a path, or press `O` to use the platform picker when one is
available.

| Source | Accepted input |
| --- | --- |
| LIST FILE | TXT, CSV, TSV, or JSON album list |
| PLAYLIST | M3U or M3U8 playlist |
| ALBUM FOLDER | One album folder containing audio files |
| ARTIST FOLDER | An artist folder containing album subfolders |

## List formats

Plain text entries use the parser's `Artist - Album`, `Artist — Album`, tab, or
`Artist | Album` forms, one per line. CSV and TSV files may use `artist` (or
`artist_name`) and `album` (or `album_name`/`title`) columns. JSON accepts an
array of target objects, an array of `Artist - Album` strings, or an object
with an `albums` array.

M3U/M3U8 entries may contain the same artist/album text or audio paths. Folder
imports infer artist and album names from the directory structure and audio
files. Parsing semantics are shared by the TUI and `scrobbler file PATH`.

## Picker behavior

macOS uses the native file or folder picker. Windows uses a per-user
PowerShell/Windows Forms dialog. Linux uses an available `zenity`, `kdialog`,
or `yad` dialog. If none is available, the screen says so and PATH entry
remains fully usable. File and playlist sources request a file picker; album
and artist folders request a folder picker.
