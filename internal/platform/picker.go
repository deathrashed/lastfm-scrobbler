package platform

import (
	"context"
	"errors"
	"strings"
)

type PickerKind string

const (
	PickerFile   PickerKind = "file"
	PickerFolder PickerKind = "folder"
)

type Picker interface {
	Available() bool
	Description() string
	Pick(context.Context, PickerKind, string) (string, error)
}

func DefaultPicker() Picker { return newPicker() }

func Pick(kind PickerKind, prompt string) (string, error) {
	return DefaultPicker().Pick(context.Background(), kind, prompt)
}

func PickerAvailable() bool { return DefaultPicker().Available() }

func PickerDescription() string { return DefaultPicker().Description() }

func PickerDescriptionFor(goos string, available bool) string {
	if !available {
		return "picker unavailable; enter a path manually"
	}
	switch strings.ToLower(strings.TrimSpace(goos)) {
	case "darwin":
		return "open the native macOS file/folder picker"
	case "windows":
		return "open the native Windows file/folder picker"
	case "linux":
		return "open the desktop file/folder picker"
	default:
		return "open the platform file/folder picker"
	}
}

func unsupportedPicker(goos string) Picker {
	return unavailablePicker{description: PickerDescriptionFor(goos, false)}
}

type unavailablePicker struct{ description string }

func (p unavailablePicker) Available() bool     { return false }
func (p unavailablePicker) Description() string { return p.description }
func (p unavailablePicker) Pick(context.Context, PickerKind, string) (string, error) {
	return "", errors.New(p.description)
}
