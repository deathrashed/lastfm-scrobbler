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
			if got := lipgloss.Width(stripANSI(line)); got != 65 {
				t.Fatalf("shell %q line %d width = %d, want 65: %q", shell, lineNumber+1, got, stripANSI(line))
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
