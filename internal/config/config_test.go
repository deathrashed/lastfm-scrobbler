package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadReadsDotEnvAndEnvironmentOverrides(t *testing.T) {
	dir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	envPath := filepath.Join(dir, ".env")
	content := "API_KEY=file-key\nAPI_SECRET=file-secret\nLASTFM_USERNAME=file-user\nLASTFM_PASSWORD=file-pass\nSCROBBLE_INTERVAL=3\nSCROBBLE_LOOP=2\n"
	if err := os.WriteFile(envPath, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	t.Setenv("LASTFM_API_KEY", "environment-key")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "environment-key" || cfg.APISecret != "file-secret" || cfg.Username != "file-user" || cfg.Password != "file-pass" {
		t.Fatalf("cfg = %#v", cfg)
	}
	if cfg.DefaultInterval != 3*time.Second || cfg.DefaultLoop != 2 {
		t.Fatalf("defaults = %s, %d", cfg.DefaultInterval, cfg.DefaultLoop)
	}
	resolvedEnvPath, err := filepath.EvalSymlinks(cfg.EnvPath)
	if err != nil {
		t.Fatal(err)
	}
	resolvedExpectedPath, err := filepath.EvalSymlinks(envPath)
	if err != nil {
		t.Fatal(err)
	}
	if resolvedEnvPath != resolvedExpectedPath {
		t.Fatalf("EnvPath = %q", cfg.EnvPath)
	}
}

func TestSaveUsesLoadedEnvPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	cfg := Config{
		APIKey:          "key",
		APISecret:       "secret",
		Username:        "deathrashed",
		Password:        "password",
		DefaultInterval: 2 * time.Second,
		DefaultLoop:     3,
		EnvPath:         path,
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	loaded, err := readEnvFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if loaded["LASTFM_USERNAME"] != "deathrashed" || loaded["SCROBBLE_LOOP"] != "3" {
		t.Fatalf("loaded = %#v", loaded)
	}
}

func TestRememberEnvPathAndLoadFromPath(t *testing.T) {
	configHome := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", configHome)
	t.Setenv("LASTFM_ENV_FILE", "")

	path := filepath.Join(t.TempDir(), "credentials.env")
	content := "API_KEY=key\nAPI_SECRET=secret\nLASTFM_USERNAME=user\nLASTFM_SESSION_KEY=session\n"
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	if err := RememberEnvPath(path); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.EnvPath != path || cfg.SessionKey != "session" {
		t.Fatalf("cfg = %#v", cfg)
	}
}

func TestLoadUsesProjectEnvAndHomeEnvForMissingValues(t *testing.T) {
	projectDir := t.TempDir()
	homeDir := t.TempDir()
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWD) }()
	if err := os.Chdir(projectDir); err != nil {
		t.Fatal(err)
	}
	expectedProjectPath, err := filepath.Abs(".env")
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HOME", homeDir)
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("LASTFM_ENV_FILE", "")
	t.Setenv("LASTFM_PROFILE", "")
	for _, key := range []string{"API_KEY", "API_SECRET", "LASTFM_API_KEY", "LASTFM_API_SECRET", "LASTFM_SHARED_SECRET", "LASTFM_USERNAME", "LASTFM_PASSWORD", "LASTFM_SESSION_KEY"} {
		t.Setenv(key, "")
	}

	projectPath := filepath.Join(projectDir, ".env")
	if err := os.WriteFile(projectPath, []byte("API_KEY=project-key\nLASTFM_USERNAME=project-user\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(homeDir, ".env"), []byte("LASTFM_SHARED_SECRET=home-secret\nLASTFM_PASSWORD=home-pass\n"), 0600); err != nil {
		t.Fatal(err)
	}

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	resolvedEnvPath, err := filepath.EvalSymlinks(cfg.EnvPath)
	if err != nil {
		t.Fatal(err)
	}
	resolvedExpectedPath, err := filepath.EvalSymlinks(expectedProjectPath)
	if err != nil {
		t.Fatal(err)
	}
	if resolvedEnvPath != resolvedExpectedPath || cfg.APIKey != "project-key" || cfg.Username != "project-user" || cfg.APISecret != "home-secret" || cfg.Password != "home-pass" {
		t.Fatalf("cfg = %#v", cfg)
	}
}

func TestLoadFromPathSupportsLastFMSharedSecret(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("LASTFM_SHARED_SECRET=shared-secret\n"), 0600); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"API_SECRET", "LASTFM_API_SECRET", "LASTFM_SHARED_SECRET"} {
		t.Setenv(key, "")
	}

	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APISecret != "shared-secret" {
		t.Fatalf("APISecret = %q", cfg.APISecret)
	}
}

func TestCredentialSourceEnvironmentDoesNotFallBackToFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, ".env")
	if err := os.WriteFile(path, []byte("LASTFM_CREDENTIAL_SOURCE=environment\nAPI_KEY=file-key\nAPI_SECRET=file-secret\nLASTFM_USERNAME=file-user\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LASTFM_API_KEY", "environment-key")
	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "environment-key" {
		t.Fatalf("APIKey = %q", cfg.APIKey)
	}
	if cfg.APISecret != "" || cfg.Username != "" {
		t.Fatalf("environment source fell back to file: %#v", cfg)
	}
}

func TestCredentialSourceFileIgnoresEnvironmentCredentialOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("LASTFM_CREDENTIAL_SOURCE=file\nAPI_KEY=file-key\nLASTFM_USERNAME=file-user\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LASTFM_API_KEY", "environment-key")
	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.APIKey != "file-key" || cfg.Username != "file-user" {
		t.Fatalf("cfg = %#v", cfg)
	}
}

func TestSaveDoesNotPersistEnvironmentCredentials(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	cfg := Config{APIKey: "key", APISecret: "secret", Username: "user", Password: "pass", SessionKey: "session", CredentialSource: "environment", EnvPath: path}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	values, err := readEnvFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"API_KEY", "API_SECRET", "LASTFM_USERNAME", "LASTFM_PASSWORD", "LASTFM_SESSION_KEY"} {
		if values[key] != "" {
			t.Fatalf("%s was persisted: %q", key, values[key])
		}
	}
}

func TestSaveDoesNotPersistAutoEnvironmentOverrides(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".env")
	if err := os.WriteFile(path, []byte("LASTFM_CREDENTIAL_SOURCE=auto\nAPI_KEY=file-key\nLASTFM_USERNAME=file-user\n"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("LASTFM_API_KEY", "environment-key")
	t.Setenv("LASTFM_USERNAME", "environment-user")
	cfg, err := LoadFromPath(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := Save(cfg); err != nil {
		t.Fatal(err)
	}
	values, err := readEnvFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if values["API_KEY"] != "" || values["LASTFM_USERNAME"] != "" {
		t.Fatalf("environment overrides were persisted: %#v", values)
	}
}
