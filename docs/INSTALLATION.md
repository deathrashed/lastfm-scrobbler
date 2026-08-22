# Installation

Last.fm Scrobbler is distributed as self-contained release binaries and a
tested Homebrew formula. Go is not required for those paths; Go 1.24.2 or
newer is required only for `go install` and source builds.

## Homebrew on macOS

```bash
brew tap deathrashed/tap
brew install deathrashed/tap/scrobbler
scrobbler --version
```

The formula supports Apple Silicon and Intel Macs and installs the Zsh, Bash,
and Fish completions alongside the binary. PowerShell completion remains
available from the binary generator or TUI installer.

## GitHub release binaries

Download [v1.2.0](https://github.com/deathrashed/lastfm-scrobbler/releases/tag/v1.2.0)
and verify the matching entry in `checksums.txt`:

| Platform | Archive |
| --- | --- |
| macOS Apple Silicon | `scrobbler-v1.2.0-darwin-arm64.tar.gz` |
| macOS Intel | `scrobbler-v1.2.0-darwin-amd64.tar.gz` |
| Linux x86_64 | `scrobbler-v1.2.0-linux-amd64.tar.gz` |
| Linux ARM64 | `scrobbler-v1.2.0-linux-arm64.tar.gz` |
| Windows x64 | `scrobbler-v1.2.0-windows-amd64.zip` |

Unix archives contain `scrobbler`, `LICENSE`, and all four completion files.
The Windows archive contains `scrobbler.exe`, `LICENSE`, and the same
completion files. Place the executable on your user PATH.

## Go install and source builds

Go 1.24.2 or newer is required for these paths:

```bash
go install github.com/deathrashed/lastfm-scrobbler/cmd/scrobbler@latest
```

Or build a checkout:

```bash
git clone https://github.com/deathrashed/lastfm-scrobbler.git
cd lastfm-scrobbler
mkdir -p bin
go build -buildvcs=false -o bin/scrobbler ./cmd/scrobbler
```

On Windows, use the equivalent `.exe` output path and Go's normal Windows
environment/path conventions.

## First launch

Run the installed command:

```bash
scrobbler
```

When no usable configuration exists, the first-run wizard opens before the
Dashboard. It detects macOS, Linux, or Windows, stages credentials, defaults,
font installation, and supported terminal configuration until **Review →
Apply**, and can be canceled safely before Apply. Run `scrobbler setup` later
to open it explicitly.

After setup:

```bash
scrobbler --version
scrobbler --help
scrobbler test --json
```

The connection test checks API and authentication readiness without submitting
a scrobble. See [Configuration](CONFIGURATION.md) for credentials paths and
source precedence.
