//go:build windows

package platform

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type windowsPicker struct{}

func newPicker() Picker { return windowsPicker{} }

func (windowsPicker) Available() bool {
	_, err := exec.LookPath("powershell.exe")
	return err == nil
}

func (p windowsPicker) Description() string { return PickerDescriptionFor("windows", p.Available()) }

func (windowsPicker) Pick(ctx context.Context, kind PickerKind, prompt string) (string, error) {
	if !(windowsPicker{}).Available() {
		return "", fmt.Errorf("%s", PickerDescriptionFor("windows", false))
	}
	quotedPrompt := strings.ReplaceAll(prompt, "'", "''")
	script := "$ErrorActionPreference='Stop'; Add-Type -AssemblyName System.Windows.Forms; "
	if kind == PickerFolder {
		script += "$dialog=New-Object System.Windows.Forms.FolderBrowserDialog; $dialog.Description='" + quotedPrompt + "'; "
	} else {
		script += "$dialog=New-Object System.Windows.Forms.OpenFileDialog; $dialog.Title='" + quotedPrompt + "'; $dialog.Filter='All files (*.*)|*.*'; "
	}
	script += "if($dialog.ShowDialog() -eq [System.Windows.Forms.DialogResult]::OK){Write-Output $dialog."
	if kind == PickerFolder {
		script += "SelectedPath"
	} else {
		script += "FileName"
	}
	script += "}"
	out, err := exec.CommandContext(ctx, "powershell.exe", "-NoProfile", "-STA", "-NonInteractive", "-Command", script).Output()
	if err != nil {
		return "", fmt.Errorf("open Windows picker: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
