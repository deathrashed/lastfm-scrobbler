package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/platform"
)

func TestInfoUsesCurrentTerminologyAndPickerCapability(t *testing.T) {
	for index := 0; index < 5; index++ {
		view := stripANSI(renderInfoView(model{infoIndex: index}))
		for lineNumber, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(line); got > headerContentWidth {
				t.Fatalf("Info section %d line %d width = %d: %q", index, lineNumber+1, got, line)
			}
		}
	}
	modes := stripANSI(renderInfoView(model{infoIndex: 0}))
	if !strings.Contains(modes, "DISCOGRAPHY") || strings.Contains(modes, "TOP ALBUMS") {
		t.Fatalf("Modes terminology is stale: %q", modes)
	}
	automation := stripANSI(renderInfoView(model{infoIndex: 1}))
	if !strings.Contains(automation, "COMPLETIONS") || !strings.Contains(automation, "PowerShell") {
		t.Fatalf("Automation completion wording is stale: %q", automation)
	}
	imports := stripANSI(renderInfoView(model{infoIndex: 4}))
	if !strings.Contains(imports, platform.PickerDescription()) {
		t.Fatalf("Imports picker wording = %q, want %q", imports, platform.PickerDescription())
	}
}
