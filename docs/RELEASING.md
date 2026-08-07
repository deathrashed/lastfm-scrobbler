# <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="24" height="24" alt="Last.fm icon"> Releasing

Releases are built by GitHub Actions from version tags. Local credentials and
build outputs stay outside the tracked source tree.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Before tagging

```bash
test -z "$(gofmt -l .)"
go test ./...
go vet ./...
go build -buildvcs=false ./cmd/scrobbler
if git ls-files | grep -Eq '(^|/)(\.env|bin/|dist/)'; then
  echo "tracked credentials or build output found" >&2
  exit 1
fi
```

Confirm that `.env`, `bin/`, `dist/`, logs, ZIP files, and coverage output are
ignored and absent from the tracked file list. Never create a public source ZIP
by archiving the working directory directly; use GitHub's tracked source archive
or `git archive`.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Create a release

```bash
git tag -a v1.0.0 -m "Release v1.0.0"
git push origin v1.0.0
```

The release workflow builds:

```text
scrobbler-darwin-arm64
scrobbler-darwin-amd64
scrobbler-v<version>-darwin-arm64.tar.gz
scrobbler-v<version>-darwin-amd64.tar.gz
checksums.txt
```

Each archive contains the binary, `LICENSE`, and Zsh, Bash, and Fish
completion files. Checksums use release-asset basenames so `shasum -a 256 -c
checksums.txt` works directly from a downloaded release directory.

It injects the tag, commit SHA, and repository into the binary and publishes
the files to the GitHub Release. The update checker can then use the repository
metadata or a configured `SCROBBLER_UPDATE_URL`.

## <img src="https://api.iconify.design/selfhst:last-fm.svg?color=f8211c" width="20" height="20" alt="Last.fm icon"> Installation paths

The public module path supports:

```bash
go install github.com/deathrashed/lastfm-scrobbler/cmd/scrobbler@latest
```

A Homebrew formula belongs in a separate `deathrashed/homebrew-tap` repository
and should be added after the first tagged release has stable asset URLs and
checksums.
