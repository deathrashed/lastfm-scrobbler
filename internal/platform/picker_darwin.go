//go:build darwin

package platform

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type darwinPicker struct{}

func newPicker() Picker { return darwinPicker{} }

func (darwinPicker) Available() bool {
	_, err := exec.LookPath("osascript")
	return err == nil
}

func (p darwinPicker) Description() string { return PickerDescriptionFor("darwin", p.Available()) }

func (darwinPicker) Pick(ctx context.Context, kind PickerKind, prompt string) (string, error) {
	command := "choose file"
	if kind == PickerFolder {
		command = "choose folder"
	}
	script := fmt.Sprintf(`set chosenItem to %s with prompt %q
POSIX path of chosenItem`, command, prompt)
	out, err := exec.CommandContext(ctx, "osascript", "-e", script).Output()
	if err != nil {
		return "", fmt.Errorf("open macOS picker: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
