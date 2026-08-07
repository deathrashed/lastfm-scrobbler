package version

// These values may be replaced at build time with -ldflags, for example:
//
//	-X github.com/deathrashed/lastfm-scrobbler/internal/version.Version=v10.1.0
//	-X github.com/deathrashed/lastfm-scrobbler/internal/version.Commit=$(git rev-parse --short HEAD)
//	-X github.com/deathrashed/lastfm-scrobbler/internal/version.Repository=owner/repository
var (
	Version    = "v10.0.0"
	Commit     = "development"
	Repository = ""
	UpdateURL  = ""
)
