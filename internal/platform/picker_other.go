//go:build !darwin && !linux && !windows

package platform

import "runtime"

func newPicker() Picker { return unsupportedPicker(runtime.GOOS) }
