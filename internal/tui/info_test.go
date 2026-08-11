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

func TestInfoContentPanelActuallyGrowsWithResponsiveWorkWidth(t *testing.T) {
	panelBorderWidth := func(width int) int {
		plain := stripANSI(renderInfoView(model{width: width, infoIndex: 0}))
		largest := 0
		for _, line := range strings.Split(plain, "\n") {
			trimmed := strings.TrimSpace(line)
			if strings.HasPrefix(trimmed, "╭") && strings.HasSuffix(trimmed, "╮") {
				largest = maxInt(largest, lipgloss.Width(trimmed))
			}
		}
		return largest
	}
	if got := panelBorderWidth(67); got != 65 {
		t.Fatalf("Info 67-column panel width = %d, want 65", got)
	}
	if got := panelBorderWidth(127); got != 125 {
		t.Fatalf("Info 127-column panel width = %d, want 125", got)
	}
}
