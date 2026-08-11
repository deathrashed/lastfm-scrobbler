package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/deathrashed/lastfm-scrobbler/internal/config"
)

func TestUnifiedFileViewPreservesFixedWidth(t *testing.T) {
	for index := range fileSourceSpecs {
		m := model{
			stage:             stageImportSource,
			modeChoice:        "file",
			importSourceIndex: index,
			searchInput:       newTextInput(512, 48),
			cfg:               config.Config{MouseEnabled: true},
		}
		m.searchInput.SetValue(strings.Repeat("C:\\Music\\Artist\\", 8))
		for lineNumber, line := range strings.Split(renderImportSourceView(m), "\n") {
			if got := lipgloss.Width(stripANSI(line)); got > headerContentWidth {
				t.Fatalf("source %d line %d width = %d, exceeds %d: %q", index, lineNumber+1, got, headerContentWidth, stripANSI(line))
			}
		}
		if got := lipgloss.Width(stripANSI(renderFilePathView(m))); got != 65 {
			t.Fatalf("source %d path panel width = %d, want 65", index, got)
		}
	}
}

func TestUnifiedFileViewUsesDynamicTypesBadge(t *testing.T) {
	want := []string{"TYPES ❯ TXT • CSV • TSV • JSON", "TYPES ❯ M3U • M3U8", "TYPE ❯ FOLDER", "TYPE ❯ FOLDER"}
	for index, value := range want {
		m := model{importSourceIndex: index, searchInput: newTextInput(128, 48)}
		plain := stripANSI(renderFilePathView(m))
		if !strings.Contains(plain, value) {
			t.Fatalf("source %d badge = %q, want %q", index, plain, value)
		}
	}
}

func TestFilePathMouseRegionUsesUnifiedPanelGeometry(t *testing.T) {
	for _, compact := range []bool{false, true} {
		m := model{stage: stageImportSource, modeChoice: "file", cfg: config.Config{CompactHeader: compact, MouseEnabled: true}, searchInput: newTextInput(128, 48)}
		var pathRegion mouseRegion
		for _, region := range m.screenRegions() {
			if region.id == "file:path" {
				pathRegion = region
			}
		}
		if pathRegion.height == 0 || pathRegion.y != m.headerHeight()+filePathPanelOffset {
			t.Fatalf("compact=%t path region = %+v, want y=%d and nonzero height", compact, pathRegion, m.headerHeight()+filePathPanelOffset)
		}
		updated, _ := m.updateMouse(tea.MouseMsg{X: pathRegion.x + 2, Y: pathRegion.y + 1, Action: tea.MouseActionPress, Button: tea.MouseButtonLeft})
		if !updated.(model).filePathFocused {
			t.Fatalf("compact=%t clicking unified PATH panel did not focus the path input", compact)
		}
	}
}
