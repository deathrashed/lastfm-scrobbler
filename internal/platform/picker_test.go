package platform

import (
	"strings"
	"testing"
)

func TestPickerDescriptionsReflectPlatformAndAvailability(t *testing.T) {
	tests := []struct {
		name      string
		goos      string
		available bool
		want      string
	}{
		{name: "macOS", goos: "darwin", available: true, want: "native macOS"},
		{name: "Windows", goos: "windows", available: true, want: "native Windows"},
		{name: "Linux", goos: "linux", available: true, want: "desktop"},
		{name: "unavailable", goos: "linux", want: "picker unavailable"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := PickerDescriptionFor(test.goos, test.available)
			if !strings.Contains(got, test.want) {
				t.Fatalf("PickerDescriptionFor(%q, %t) = %q, want %q", test.goos, test.available, got, test.want)
			}
		})
	}
}
