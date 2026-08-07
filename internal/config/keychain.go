package config

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

const keychainService = "lastfm-scrobbler"

type KeychainSecrets struct {
	APISecret  string
	Password   string
	SessionKey string
}

func keychainAccount(profile, key string) string {
	profile = sanitizeProfileName(profile)
	if profile == "" {
		profile = "default"
	}
	return profile + ":" + key
}

func readKeychainValue(profile, key string) (string, error) {
	if runtime.GOOS != "darwin" {
		return "", errors.New("macOS Keychain is only available on macOS")
	}
	cmd := exec.Command("security", "find-generic-password", "-s", keychainService, "-a", keychainAccount(profile, key), "-w")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func writeKeychainValue(profile, key, value string) error {
	if runtime.GOOS != "darwin" {
		return errors.New("macOS Keychain is only available on macOS")
	}
	account := keychainAccount(profile, key)
	if strings.TrimSpace(value) == "" {
		_ = exec.Command("security", "delete-generic-password", "-s", keychainService, "-a", account).Run()
		return nil
	}
	cmd := exec.Command("security", "add-generic-password", "-U", "-s", keychainService, "-a", account, "-w", value)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("save %s to Keychain: %v: %s", key, err, strings.TrimSpace(string(out)))
	}
	return nil
}

// LoadKeychainSecrets reads optional credentials from macOS Keychain.
func LoadKeychainSecrets(profile string) (KeychainSecrets, error) {
	if runtime.GOOS != "darwin" {
		return KeychainSecrets{}, errors.New("macOS Keychain is unavailable")
	}
	var secrets KeychainSecrets
	secrets.APISecret, _ = readKeychainValue(profile, "api-secret")
	secrets.Password, _ = readKeychainValue(profile, "password")
	secrets.SessionKey, _ = readKeychainValue(profile, "session-key")
	if secrets.APISecret == "" && secrets.Password == "" && secrets.SessionKey == "" {
		return secrets, errors.New("no Keychain credentials found")
	}
	return secrets, nil
}

// SaveKeychainSecrets writes sensitive credentials to macOS Keychain.
func SaveKeychainSecrets(profile string, secrets KeychainSecrets) error {
	if err := writeKeychainValue(profile, "api-secret", secrets.APISecret); err != nil {
		return err
	}
	if err := writeKeychainValue(profile, "password", secrets.Password); err != nil {
		return err
	}
	return writeKeychainValue(profile, "session-key", secrets.SessionKey)
}

func SaveSessionKey(profile, sessionKey string) error {
	return writeKeychainValue(profile, "session-key", strings.TrimSpace(sessionKey))
}

func PersistSessionKey(cfg Config, sessionKey string) error {
	if strings.TrimSpace(cfg.SessionKey) != "" || strings.TrimSpace(sessionKey) == "" {
		return nil
	}
	return SaveSessionKey(cfg.Profile, sessionKey)
}
