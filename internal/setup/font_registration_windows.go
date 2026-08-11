//go:build windows

package setup

import (
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

func registerInstalledFont(path string) error {
	key, _, err := registry.CreateKey(registry.CURRENT_USER, `Software\Microsoft\Windows NT\CurrentVersion\Fonts`, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	name := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	return key.SetStringValue(name, path)
}
