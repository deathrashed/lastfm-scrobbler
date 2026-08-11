//go:build linux

package platform

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type linuxPicker struct{ tool string }

func newPicker() Picker {
	for _, tool := range []string{"zenity", "kdialog", "yad"} {
		if _, err := exec.LookPath(tool); err == nil {
			return linuxPicker{tool: tool}
		}
	}
	return linuxPicker{}
}

func (p linuxPicker) Available() bool { return p.tool != "" }

func (p linuxPicker) Description() string { return PickerDescriptionFor("linux", p.Available()) }

func (p linuxPicker) Pick(ctx context.Context, kind PickerKind, prompt string) (string, error) {
	if !p.Available() {
		return "", fmt.Errorf("%s", p.Description())
	}
	args := []string{}
	switch p.tool {
	case "zenity":
		args = []string{"--file-selection", "--title", prompt}
		if kind == PickerFolder {
			args = append(args, "--directory")
		}
	case "kdialog":
		if kind == PickerFolder {
			args = []string{"--getexistingdirectory", "", "--title", prompt}
		} else {
			args = []string{"--getopenfilename", "", "--title", prompt}
		}
	case "yad":
		args = []string{"--file-selection", "--title", prompt}
		if kind == PickerFolder {
			args = append(args, "--directory")
		}
	}
	out, err := exec.CommandContext(ctx, p.tool, args...).Output()
	if err != nil {
		return "", fmt.Errorf("open Linux picker: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
