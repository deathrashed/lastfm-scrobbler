# Installation

Last.fm Scrobbler is distributed as a self-contained release binary or can be
built from source. Release binaries do not require Go. Source builds require Go
1.24.2 or newer.

## Release binaries

Download the asset for macOS, Linux, or Windows from the project's GitHub
Releases page and place `scrobbler` (or `scrobbler.exe`) somewhere on your
user PATH. The first interactive launch opens the setup wizard when no usable
configuration exists.

## Build from source

```bash
go install github.com/deathrashed/lastfm-scrobbler/cmd/scrobbler@latest
```

Or build a checkout:

```bash
go build -buildvcs=false -o bin/scrobbler ./cmd/scrobbler
```

On Windows, use the equivalent `.exe` output path and Go's normal Windows
environment/path conventions. No package manager is required on any platform.

There is no published Homebrew, Chocolatey, Scoop, apt, dnf, or pacman package
promised by this repository. Use a release asset, `go install`, or a source
build until a package is published and documented here.

## Verify

```text
scrobbler --version
scrobbler --help
scrobbler setup
```

Use `scrobbler test --json` after setup to check API and authentication
readiness without submitting a scrobble.
