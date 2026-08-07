package config

import "testing"

func TestSessionKeyDestinationByCredentialSource(t *testing.T) {
	for source, want := range map[string]string{
		"auto":        "keychain",
		"keychain":    "keychain",
		"file":        "file",
		"environment": "none",
	} {
		if got := sessionKeyDestination(source); got != want {
			t.Fatalf("sessionKeyDestination(%q) = %q, want %q", source, got, want)
		}
	}
}
