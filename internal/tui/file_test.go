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

func TestFileSourceRowsKeepCompactOneCellSeparatorAtWideWidth(t *testing.T) {
	m := model{
		width:             127,
		stage:             stageImportSource,
		modeChoice:        "file",
		importSourceIndex: 0,
		searchInput:       newTextInput(512, 48),
	}
	plain := stripANSI(renderImportSourceView(m))
	if got := strings.Count(plain, "│•│"); got != 2 {
		t.Fatalf("File source row compact separator count = %d, want 2:\n%s", got, plain)
	}
	if strings.Contains(plain, "│ • │") {
		t.Fatalf("File source rows grew spaces around the separator:\n%s", plain)
	}
	if !strings.Contains(plain, "L I S T  F I L E") {
		t.Fatalf("File list label spacing regressed:\n%s", plain)
	}

	positions := fileSourceCardPositions(m.panelWidth())
	divider := fileSourceDividerX(m.panelWidth())
	for _, pair := range [][2]int{{0, 1}, {2, 3}} {
		left, right := pair[0], pair[1]
		if got := positions[left][0] + fileSourceSpecs[left].width; got != divider {
			t.Fatalf("left File card %d ends at %d, want divider %d", left, got, divider)
		}
		if got := positions[right][0]; got != divider+1 {
			t.Fatalf("right File card %d starts at %d, want %d", right, got, divider+1)
		}
	}
}

func TestFileSourceRowsShareOneDividerColumn(t *testing.T) {
	m := model{width: 67, stage: stageImportSource, modeChoice: "file", importSourceIndex: 0, searchInput: newTextInput(128, 48)}
	plain := stripANSI(renderImportSourceView(m))
	lines := strings.Split(plain, "\n")
	var dividerColumns []int
	for _, line := range lines {
		if at := strings.Index(line, "•"); at >= 0 {
			dividerColumns = append(dividerColumns, lipgloss.Width(line[:at]))
		}
	}
	if len(dividerColumns) < 2 {
		t.Fatalf("file source divider bullets=%v, want at least two\n%s", dividerColumns, plain)
	}
	if dividerColumns[0] != dividerColumns[1] {
		t.Fatalf("file source dividers do not align: %v\n%s", dividerColumns[:2], plain)
	}
}
