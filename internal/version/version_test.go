package version

import "testing"

func TestResolveVersionPrefersExplicitRelease(t *testing.T) {
	if got := ResolveVersion("v1.0.0", "v1.1.0"); got != "v1.0.0" {
		t.Fatalf("ResolveVersion = %q, want v1.0.0", got)
	}
}

func TestResolveVersionUsesBuildInfo(t *testing.T) {
	if got := ResolveVersion("development", "v1.0.0"); got != "v1.0.0" {
		t.Fatalf("ResolveVersion = %q, want v1.0.0", got)
	}
}

func TestResolveVersionUsesDevelopmentFallback(t *testing.T) {
	for _, buildInfo := range []string{"", "(devel)"} {
		if got := ResolveVersion("development", buildInfo); got != "development" {
			t.Fatalf("ResolveVersion(%q) = %q, want development", buildInfo, got)
		}
	}
}
