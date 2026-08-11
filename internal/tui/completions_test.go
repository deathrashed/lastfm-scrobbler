package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/completion"
	"github.com/deathrashed/lastfm-scrobbler/internal/config"
)

func TestCompletionScreenPreservesPanelWidth(t *testing.T) {
	for _, shell := range completion.Shells {
		m := model{stage: stageCompletions, completionShell: shell, completionStatus: completion.StatusNotInstalled}
		for lineNumber, line := range strings.Split(renderCompletionsView(m), "\n") {
			trimmed := strings.TrimSpace(stripANSI(line))
			if got := lipgloss.Width(trimmed); got != m.panelWidth() {
				t.Fatalf("shell %q line %d panel width = %d, want %d: %q", shell, lineNumber+1, got, m.panelWidth(), trimmed)
			}
		}
	}
}

func TestCompletionScreenMouseSelectionDoesNotMoveOnHover(t *testing.T) {
	m := model{stage: stageCompletions, completionShell: completion.ShellZsh, completionStatus: completion.StatusNotInstalled, cfg: config.Config{MouseEnabled: true}}
	var target mouseRegion
	for _, region := range m.screenRegions() {
		if region.id == "completion:shell:fish" {
			target = region
			break
		}
	}
	if target.height == 0 {
		t.Fatal("fish completion region missing")
	}
	updated, _ := m.updateMouse(tea.MouseMsg{X: target.x, Y: target.y, Action: tea.MouseActionMotion})
	got := updated.(model)
	if got.completionShell != completion.ShellZsh || got.hoverRegion != target.id {
		t.Fatalf("hover changed completion selection: shell=%q hover=%q", got.completionShell, got.hoverRegion)
	}
}

func TestCompletionScreenUsesPlainResponsiveBodyPanel(t *testing.T) {
	for _, width := range []int{67, 127} {
		m := model{width: width, stage: stageCompletions, completionShell: completion.ShellZsh, completionStatus: completion.StatusNotInstalled}
		plain := stripANSI(renderCompletionsView(m))
		if strings.Contains(plain, "S H E L L  C O M P L E T I O N S") {
			t.Fatalf("width %d still renders obsolete nested shell-completions title:\n%s", width, plain)
		}
		lines := strings.Split(plain, "\n")
		want := m.panelWidth()
		for lineNumber, line := range lines {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if got := lipgloss.Width(trimmed); got > want {
				t.Fatalf("width %d line %d body width = %d, exceeds %d: %q", width, lineNumber+1, got, want, trimmed)
			}
		}
	}
}

func TestCompletionShellRowsPreserveFullStatuses(t *testing.T) {
	for _, status := range []completion.Status{
		completion.StatusInstalled,
		completion.StatusAlreadyInstalled,
		completion.StatusUpdated,
		completion.StatusUpdateAvailable,
		completion.StatusNotInstalled,
		completion.StatusManual,
	} {
		row := stripANSI(completionShellRow(completion.ShellPowerShell, status, true, 61))
		if !strings.Contains(row, string(status)) {
			t.Fatalf("status %q truncated from completion row %q", status, row)
		}
	}
}
