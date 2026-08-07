package version

import (
	"runtime/debug"
	"strings"
)

const DefaultRepository = "deathrashed/lastfm-scrobbler"

var (
	Version    = "development"
	Commit     = "development"
	Repository = DefaultRepository
	UpdateURL  = ""
)

func init() {
	var buildVersion string
	if info, ok := debug.ReadBuildInfo(); ok {
		buildVersion = info.Main.Version
	}
	Version = ResolveVersion(Version, buildVersion)
	if strings.TrimSpace(Repository) == "" {
		Repository = DefaultRepository
	}
}

func ResolveVersion(explicit, buildInfo string) string {
	if value := strings.TrimSpace(explicit); value != "" && value != "development" {
		return value
	}
	if value := strings.TrimSpace(buildInfo); value != "" && value != "(devel)" {
		return value
	}
	return "development"
}
