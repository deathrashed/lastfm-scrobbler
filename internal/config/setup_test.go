package config

import "testing"

func TestNeedsSetupUsesCompleteCredentialSet(t *testing.T) {
	base := Config{APIKey: "key", APISecret: "secret"}
	if !NeedsSetup(base) {
		t.Fatal("missing authentication credentials should require setup")
	}
	base.Username = "user"
	base.Password = "pass"
	if NeedsSetup(base) {
		t.Fatal("username/password credentials should be usable")
	}
	base.Password = ""
	base.SessionKey = "session"
	if NeedsSetup(base) {
		t.Fatal("session-key credentials should be usable")
	}
}
