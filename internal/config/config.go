package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Config holds the scrobbler configuration. Real process environment
// variables take precedence over values loaded from a credentials file.
type Config struct {
	APIKey           string
	APISecret        string
	Username         string
	Password         string
	SessionKey       string
	DefaultInterval  time.Duration
	DefaultLimit     int
	DefaultLoop      int
	RetryCount       int
	RetryDelay       time.Duration
	DuplicateGuard   time.Duration
	Notify           bool
	CompactHeader    bool
	CleanDiscography bool
	ExportDir        string
	EnvPath          string
	Profile          string
	CredentialSource string // auto, file, environment, keychain
	UseKeychain      bool
	MouseEnabled     bool
	UpdateURL        string

	apiKeyFromEnvironment    bool
	apiSecretFromEnvironment bool
	usernameFromEnvironment  bool
	passwordFromEnvironment  bool
	sessionFromEnvironment   bool
	apiSecretFromKeychain    bool
	passwordFromKeychain     bool
	sessionFromKeychain      bool
}

// Load discovers a credentials file and then overlays real environment
// variables. LASTFM_ENV_FILE is the strongest file-location override.
func Load() (Config, error) { return loadFromDiscoveredPath() }

// LoadFromPath loads one explicit credentials file while still allowing real
// environment variables to override its values.
func LoadFromPath(path string) (Config, error) {
	path = ExpandPath(path)
	if strings.TrimSpace(path) == "" {
		return Config{}, fmt.Errorf("credentials path is empty")
	}
	return loadConfig(path)
}

// LoadProfile loads a named profile from the app config directory.
func LoadProfile(name string) (Config, error) {
	name = sanitizeProfileName(name)
	if name == "" {
		return Config{}, fmt.Errorf("profile name is empty")
	}
	path := ProfilePath(name)
	if _, err := os.Stat(path); err != nil {
		return Config{}, err
	}
	cfg, err := loadConfig(path)
	if err != nil {
		return Config{}, err
	}
	cfg.Profile = name
	return cfg, nil
}

func loadFromDiscoveredPath() (Config, error) {
	return loadConfigWithFallback(findEnvFile(), homeEnvFile())
}

func defaultConfig() Config {
	exportDir := ""
	if home, err := os.UserHomeDir(); err == nil {
		exportDir = filepath.Join(home, "Downloads")
	}
	return Config{
		DefaultInterval:  2 * time.Second,
		DefaultLimit:     0,
		DefaultLoop:      1,
		RetryCount:       2,
		RetryDelay:       2 * time.Second,
		DuplicateGuard:   0,
		Notify:           true,
		CompactHeader:    false,
		CleanDiscography: false,
		ExportDir:        exportDir,
		Profile:          "default",
		CredentialSource: "auto",
		MouseEnabled:     true,
	}
}

func loadConfig(envPath string) (Config, error) {
	return loadConfigWithFallback(envPath, "")
}

func loadConfigWithFallback(envPath, fallbackPath string) (Config, error) {
	cfg := defaultConfig()

	fileValues := map[string]string{}
	if envPath != "" {
		values, err := readEnvFile(envPath)
		if err != nil {
			return cfg, err
		}
		fileValues = values
		cfg.EnvPath = envPath
	} else {
		cwd, _ := os.Getwd()
		cfg.EnvPath = filepath.Join(cwd, ".env")
	}
	if fallbackPath != "" && filepath.Clean(fallbackPath) != filepath.Clean(envPath) {
		if fallbackValues, err := readEnvFile(fallbackPath); err == nil {
			for key, value := range fallbackValues {
				if strings.TrimSpace(fileValues[key]) == "" {
					fileValues[key] = value
				}
			}
		}
	}

	getEnv := func(keys ...string) string {
		for _, key := range keys {
			if value, ok := os.LookupEnv(key); ok && strings.TrimSpace(value) != "" {
				return strings.TrimSpace(value)
			}
		}
		return ""
	}
	getFile := func(keys ...string) string {
		for _, key := range keys {
			if value := strings.TrimSpace(fileValues[key]); value != "" {
				return value
			}
		}
		return ""
	}
	get := func(keys ...string) string {
		if value := getEnv(keys...); value != "" {
			return value
		}
		return getFile(keys...)
	}

	cfg.Profile = firstNonEmpty(get("LASTFM_PROFILE", "SCROBBLE_PROFILE"), rememberedProfile(), "default")
	cfg.CredentialSource = strings.ToLower(firstNonEmpty(get("LASTFM_CREDENTIAL_SOURCE", "SCROBBLE_CREDENTIAL_SOURCE"), "auto"))
	if cfg.CredentialSource != "auto" && cfg.CredentialSource != "file" && cfg.CredentialSource != "environment" && cfg.CredentialSource != "keychain" {
		cfg.CredentialSource = "auto"
	}

	credentialValue := func(keys ...string) string {
		switch cfg.CredentialSource {
		case "environment":
			return getEnv(keys...)
		case "file":
			return getFile(keys...)
		case "keychain":
			// API keys and usernames are not Keychain-only values. Let a
			// process override the credentials file while keeping secrets in
			// Keychain.
			return firstNonEmpty(getEnv(keys...), getFile(keys...))
		default:
			return get(keys...)
		}
	}

	cfg.APIKey = credentialValue("API_KEY", "LASTFM_API_KEY")
	cfg.APISecret = credentialValue("API_SECRET", "LASTFM_API_SECRET", "LASTFM_SHARED_SECRET")
	cfg.Username = credentialValue("LASTFM_USERNAME", "LASTFM_USER")
	cfg.Password = credentialValue("LASTFM_PASSWORD")
	cfg.SessionKey = credentialValue("LASTFM_SESSION_KEY", "SESSION_KEY")
	if cfg.CredentialSource == "environment" || cfg.CredentialSource == "auto" {
		cfg.apiKeyFromEnvironment = strings.TrimSpace(getEnv("API_KEY", "LASTFM_API_KEY")) != ""
		cfg.apiSecretFromEnvironment = strings.TrimSpace(getEnv("API_SECRET", "LASTFM_API_SECRET", "LASTFM_SHARED_SECRET")) != ""
		cfg.usernameFromEnvironment = strings.TrimSpace(getEnv("LASTFM_USERNAME", "LASTFM_USER")) != ""
		cfg.passwordFromEnvironment = strings.TrimSpace(getEnv("LASTFM_PASSWORD")) != ""
		cfg.sessionFromEnvironment = strings.TrimSpace(getEnv("LASTFM_SESSION_KEY", "SESSION_KEY")) != ""
	}

	// Auto fills only missing secrets from Keychain. Keychain mode makes the
	// Keychain authoritative for secret values, while environment/file still
	// supply the public API key and username.
	if cfg.CredentialSource == "auto" || cfg.CredentialSource == "keychain" {
		secrets, err := LoadKeychainSecrets(cfg.Profile)
		if err == nil {
			if cfg.CredentialSource == "keychain" || cfg.APISecret == "" {
				cfg.APISecret = secrets.APISecret
				cfg.apiSecretFromKeychain = secrets.APISecret != ""
			}
			if cfg.CredentialSource == "keychain" || cfg.Password == "" {
				cfg.Password = secrets.Password
				cfg.passwordFromKeychain = secrets.Password != ""
			}
			if cfg.CredentialSource == "keychain" || cfg.SessionKey == "" {
				cfg.SessionKey = secrets.SessionKey
				cfg.sessionFromKeychain = secrets.SessionKey != ""
			}
			cfg.UseKeychain = secrets.APISecret != "" || secrets.Password != "" || secrets.SessionKey != ""
		}
	}

	if value := get("SCROBBLE_INTERVAL", "INTERVAL"); value != "" {
		cfg.DefaultInterval = parseDuration(value, cfg.DefaultInterval)
	}
	if value := get("SCROBBLE_LIMIT", "LIMIT"); value != "" {
		cfg.DefaultLimit = parseInt(value, cfg.DefaultLimit)
	}
	if value := get("SCROBBLE_LOOP", "LOOP_COUNT", "LOOP"); value != "" {
		cfg.DefaultLoop = parseInt(value, cfg.DefaultLoop)
	}
	if value := get("SCROBBLE_RETRIES", "RETRY_COUNT"); value != "" {
		cfg.RetryCount = parseInt(value, cfg.RetryCount)
	}
	if value := get("SCROBBLE_RETRY_DELAY", "RETRY_DELAY"); value != "" {
		cfg.RetryDelay = parseDuration(value, cfg.RetryDelay)
	}
	if value := get("SCROBBLE_DUPLICATE_GUARD", "DUPLICATE_GUARD"); value != "" {
		cfg.DuplicateGuard = parseDuration(value, cfg.DuplicateGuard)
	}
	if value := get("SCROBBLE_NOTIFY", "NOTIFY"); value != "" {
		cfg.Notify = parseBool(value, cfg.Notify)
	}
	if value := get("SCROBBLE_COMPACT_HEADER", "COMPACT_HEADER"); value != "" {
		cfg.CompactHeader = parseBool(value, cfg.CompactHeader)
	}
	if value := get("SCROBBLE_CLEAN_DISCOGRAPHY", "CLEAN_DISCOGRAPHY"); value != "" {
		cfg.CleanDiscography = parseBool(value, cfg.CleanDiscography)
	}
	if value := get("SCROBBLE_EXPORT_DIR", "EXPORT_DIR"); value != "" {
		cfg.ExportDir = ExpandPath(value)
	}
	if value := get("SCROBBLE_MOUSE", "MOUSE_ENABLED"); value != "" {
		cfg.MouseEnabled = parseBool(value, cfg.MouseEnabled)
	}
	cfg.UpdateURL = get("SCROBBLER_UPDATE_URL", "SCROBBLE_UPDATE_URL")

	if cfg.DefaultLoop < 1 {
		cfg.DefaultLoop = 1
	}
	if cfg.DefaultLimit < 0 {
		cfg.DefaultLimit = 0
	}
	if cfg.RetryCount < 0 {
		cfg.RetryCount = 0
	}
	if cfg.RetryDelay < 0 {
		cfg.RetryDelay = 0
	}
	if cfg.DuplicateGuard < 0 {
		cfg.DuplicateGuard = 0
	}
	if cfg.ExportDir == "" {
		cfg.ExportDir = DataDir()
	}

	return cfg, nil
}

func findEnvFile() string {
	var candidates []string
	if explicit := strings.TrimSpace(os.Getenv("LASTFM_ENV_FILE")); explicit != "" {
		candidates = append(candidates, ExpandPath(explicit))
	}
	profile := firstNonEmpty(strings.TrimSpace(os.Getenv("LASTFM_PROFILE")), rememberedProfile())
	if profile != "" && profile != "default" {
		candidates = append(candidates, ProfilePath(profile))
	}

	if exe, err := os.Executable(); err == nil {
		binDir := filepath.Dir(exe)
		candidates = append(candidates,
			filepath.Join(binDir, ".env"),
			filepath.Join(binDir, "..", ".env"),
			filepath.Join(binDir, "..", "..", ".env"),
		)
	}

	cwd, _ := os.Getwd()
	candidates = append(candidates,
		filepath.Join(cwd, ".env"),
		filepath.Join(cwd, "go", ".env"),
		filepath.Join(cwd, "..", ".env"),
	)

	if remembered := rememberedEnvPath(); remembered != "" {
		candidates = append(candidates, remembered)
	}

	seen := map[string]bool{}
	for _, candidate := range candidates {
		candidate = ExpandPath(candidate)
		abs, err := filepath.Abs(candidate)
		if err != nil || seen[abs] {
			continue
		}
		seen[abs] = true
		if info, err := os.Stat(abs); err == nil && !info.IsDir() {
			return abs
		}
	}
	return ""
}

func homeEnvFile() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, ".env")
	if info, err := os.Stat(path); err == nil && !info.IsDir() {
		return path
	}
	return ""
}

// DataDir returns the app's persistent data directory.
func DataDir() string {
	if dir := strings.TrimSpace(os.Getenv("XDG_DATA_HOME")); dir != "" {
		return filepath.Join(ExpandPath(dir), "lastfm-scrobbler")
	}
	if dir := strings.TrimSpace(os.Getenv("XDG_CONFIG_HOME")); dir != "" {
		return filepath.Join(ExpandPath(dir), "lastfm-scrobbler")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "."
	}
	return filepath.Join(home, ".config", "lastfm-scrobbler")
}

func configDir() string { return DataDir() }

func rememberedEnvPath() string {
	data, err := os.ReadFile(filepath.Join(configDir(), "env-path"))
	if err != nil {
		return ""
	}
	return ExpandPath(strings.TrimSpace(string(data)))
}

// RememberEnvPath stores the user's chosen credentials-file location.
func RememberEnvPath(path string) error {
	path = ExpandPath(path)
	if path == "" {
		return fmt.Errorf("credentials path is empty")
	}
	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir(), "env-path"), []byte(path+"\n"), 0600)
}

func rememberedProfile() string {
	data, err := os.ReadFile(filepath.Join(configDir(), "profile"))
	if err != nil {
		return ""
	}
	return sanitizeProfileName(strings.TrimSpace(string(data)))
}

// RememberProfile stores the currently selected profile.
func RememberProfile(name string) error {
	name = sanitizeProfileName(name)
	if name == "" {
		return fmt.Errorf("profile name is empty")
	}
	if err := os.MkdirAll(configDir(), 0700); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(configDir(), "profile"), []byte(name+"\n"), 0600)
}

// ExpandPath expands a leading ~/ and returns an absolute cleaned path when possible.
func ExpandPath(path string) string {
	path = strings.TrimSpace(strings.Trim(path, "\"'"))
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(home, strings.TrimPrefix(path, "~/"))
		}
	}
	if abs, err := filepath.Abs(path); err == nil {
		return filepath.Clean(abs)
	}
	return filepath.Clean(path)
}

func readEnvFile(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := map[string]string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		line = strings.TrimPrefix(line, "export ")
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}
		values[key] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func parseDuration(value string, fallback time.Duration) time.Duration {
	value = strings.TrimSpace(value)
	if value == "" || value == "0" {
		if value == "0" {
			return 0
		}
		return fallback
	}
	if duration, err := time.ParseDuration(value); err == nil {
		return duration
	}
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		return time.Duration(seconds * float64(time.Second))
	}
	return fallback
}

func parseInt(value string, fallback int) int {
	number, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil {
		return fallback
	}
	return number
}

func parseBool(value string, fallback bool) bool {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "1", "true", "yes", "y", "on", "enabled":
		return true
	case "0", "false", "no", "n", "off", "disabled":
		return false
	default:
		return fallback
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

// Save writes editable values back to cfg.EnvPath with owner-only permissions.
// When CredentialSource is keychain, secrets are stored in macOS Keychain and
// omitted from the credentials file.
func Save(cfg Config) error {
	path := ExpandPath(cfg.EnvPath)
	if strings.TrimSpace(path) == "" {
		cwd, _ := os.Getwd()
		path = filepath.Join(cwd, ".env")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}

	apiKey, apiSecret, username, password, sessionKey := cfg.APIKey, cfg.APISecret, cfg.Username, cfg.Password, cfg.SessionKey
	if cfg.CredentialSource == "keychain" {
		if err := SaveKeychainSecrets(cfg.Profile, KeychainSecrets{APISecret: apiSecret, Password: password, SessionKey: sessionKey}); err != nil {
			return err
		}
		apiSecret, password, sessionKey = "", "", ""
	}
	if cfg.CredentialSource == "environment" {
		apiKey, apiSecret, username, password, sessionKey = "", "", "", "", ""
	} else if cfg.CredentialSource == "auto" {
		if cfg.apiKeyFromEnvironment {
			apiKey = ""
		}
		if cfg.apiSecretFromEnvironment {
			apiSecret = ""
		}
		if cfg.usernameFromEnvironment {
			username = ""
		}
		if cfg.passwordFromEnvironment {
			password = ""
		}
		if cfg.sessionFromEnvironment {
			sessionKey = ""
		}
		if cfg.apiSecretFromKeychain {
			apiSecret = ""
		}
		if cfg.passwordFromKeychain {
			password = ""
		}
		if cfg.sessionFromKeychain {
			sessionKey = ""
		}
	}

	content := fmt.Sprintf(
		"API_KEY=%s\nAPI_SECRET=%s\nLASTFM_USERNAME=%s\nLASTFM_PASSWORD=%s\nLASTFM_SESSION_KEY=%s\nLASTFM_PROFILE=%s\nLASTFM_CREDENTIAL_SOURCE=%s\nSCROBBLE_INTERVAL=%s\nSCROBBLE_LIMIT=%d\nSCROBBLE_LOOP=%d\nSCROBBLE_RETRIES=%d\nSCROBBLE_RETRY_DELAY=%s\nSCROBBLE_DUPLICATE_GUARD=%s\nSCROBBLE_NOTIFY=%t\nSCROBBLE_COMPACT_HEADER=%t\nSCROBBLE_CLEAN_DISCOGRAPHY=%t\nSCROBBLE_EXPORT_DIR=%s\nSCROBBLE_MOUSE=%t\nSCROBBLER_UPDATE_URL=%s\n",
		apiKey,
		apiSecret,
		username,
		password,
		sessionKey,
		sanitizeProfileName(cfg.Profile),
		cfg.CredentialSource,
		cfg.DefaultInterval.String(),
		cfg.DefaultLimit,
		cfg.DefaultLoop,
		cfg.RetryCount,
		cfg.RetryDelay.String(),
		cfg.DuplicateGuard.String(),
		cfg.Notify,
		cfg.CompactHeader,
		cfg.CleanDiscography,
		cfg.ExportDir,
		cfg.MouseEnabled,
		cfg.UpdateURL,
	)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return err
	}
	_ = RememberProfile(cfg.Profile)
	return nil
}
