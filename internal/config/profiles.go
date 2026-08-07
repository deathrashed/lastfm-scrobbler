package config

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

func sanitizeProfileName(name string) string {
	name = strings.TrimSpace(name)
	var out []rune
	for _, r := range name {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '-' || r == '_' {
			out = append(out, unicode.ToLower(r))
		}
	}
	return strings.Trim(string(out), "-_")
}

// ProfilePath returns the env-file location for a named profile.
func ProfilePath(name string) string {
	name = sanitizeProfileName(name)
	return filepath.Join(configDir(), "profiles", name+".env")
}

// ListProfiles returns all saved profile names, always including default.
func ListProfiles() ([]string, error) {
	dir := filepath.Join(configDir(), "profiles")
	entries, err := os.ReadDir(dir)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	seen := map[string]bool{"default": true}
	for _, entry := range entries {
		if entry.IsDir() || strings.ToLower(filepath.Ext(entry.Name())) != ".env" {
			continue
		}
		name := sanitizeProfileName(strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name())))
		if name != "" {
			seen[name] = true
		}
	}
	profiles := make([]string, 0, len(seen))
	for name := range seen {
		profiles = append(profiles, name)
	}
	sort.Strings(profiles)
	return profiles, nil
}

// SaveProfile saves cfg into a named profile env file.
func SaveProfile(name string, cfg Config) error {
	name = sanitizeProfileName(name)
	if name == "" {
		return fmt.Errorf("profile name is empty")
	}
	path := ProfilePath(name)
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	cfg.Profile = name
	cfg.EnvPath = path
	if err := Save(cfg); err != nil {
		return err
	}
	return RememberProfile(name)
}

// DeleteProfile removes a saved profile. The default profile cannot be deleted.
func DeleteProfile(name string) error {
	name = sanitizeProfileName(name)
	if name == "" || name == "default" {
		return fmt.Errorf("the default profile cannot be deleted")
	}
	return os.Remove(ProfilePath(name))
}
