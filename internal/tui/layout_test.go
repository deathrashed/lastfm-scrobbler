package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func assertBlockWidth(t *testing.T, block string, want int) {
	t.Helper()
	for lineNumber, line := range strings.Split(block, "\n") {
		if got := lipgloss.Width(stripANSI(line)); got != want {
			t.Fatalf("line %d width = %d, want %d\n%q", lineNumber+1, got, want, line)
		}
	}
}

func TestMultiFieldConfigBoxNeverBreaksBorder(t *testing.T) {
	for _, fields := range [][]configRenderField{
		{
			{Label: "LASTFM USERNAME", Value: strings.Repeat("u", 120), Active: true},
			{Label: "LASTFM PASSWORD", Value: strings.Repeat("•", 120)},
		},
		{
			{Label: "API KEY", Value: strings.Repeat("a", 120)},
			{Label: "API SECRET", Value: strings.Repeat("•", 120), Active: true},
		},
	} {
		assertBlockWidth(t, renderMultiFieldBox(fields, 65), 65)
	}
}

func TestHeaderAlwaysFitsExactMockupWidth(t *testing.T) {
	for _, mode := range []string{"", "manual", "discography", "file", "config", "advanced", "history", "profiles"} {
		header := RenderHeader(140, stageInput, mode, "deathrashed", "", false)
		lines := strings.Split(header, "\n")
		for lineNumber, line := range lines {
			got := lipgloss.Width(stripANSI(line))
			if got > 67 {
				t.Fatalf("mode %q line %d width = %d, exceeds 67\n%q", mode, lineNumber+1, got, line)
			}
		}
		for _, lineNumber := range []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9} {
			if got := lipgloss.Width(stripANSI(lines[lineNumber])); got != 67 {
				t.Fatalf("mode %q structural line %d width = %d, want 67", mode, lineNumber+1, got)
			}
		}
	}
}

func TestDashboardAndConfigRowsFitHeader(t *testing.T) {
	dashboard := joinThreeLineBoxes([]string{
		renderExactBox("M A N U A L", 19, true),
		renderExactBox("D I S C O G R A P H Y", 25, false),
		renderExactBox("F I L E", 18, false),
	}, "•")
	for _, line := range strings.Split(dashboard, "\n") {
		if got := lipgloss.Width(line); got > headerContentWidth {
			t.Fatalf("dashboard width = %d, exceeds %d: %q", got, headerContentWidth, line)
		}
	}

	config := joinThreeLineBoxes([]string{
		renderExactBox("L O O P", 11, false),
		renderExactBox("I N T E R V A L", 19, false),
		renderExactBox("U S E R N A M E", 19, true),
		renderExactBox("A P I", 9, false),
	}, "•")
	for _, line := range strings.Split(config, "\n") {
		if got := lipgloss.Width(line); got > headerContentWidth {
			t.Fatalf("config width = %d, exceeds %d: %q", got, headerContentWidth, line)
		}
	}
}

func TestExpandedConfigAndFootersFitHeader(t *testing.T) {
	m := model{}
	m.cfg.Username = strings.Repeat("username", 20)
	m.cfg.Password = strings.Repeat("password", 20)
	m.cfg.APIKey = strings.Repeat("a", 120)
	m.cfg.APISecret = strings.Repeat("s", 120)
	m.cfg.Profile = "default"
	m.profiles = []string{"default", "archive"}
	m.configInput = newTextInput(1024, 44)
	for index := 0; index < 7; index++ {
		m.configIndex = index
		m.configFieldIndex = 0
		m.loadConfigField()
		view := renderConfigView(m)
		for lineNumber, line := range strings.Split(view, "\n") {
			if got := lipgloss.Width(stripANSI(line)); got > headerContentWidth {
				t.Fatalf("config index %d line %d width = %d, exceeds %d\n%q", index, lineNumber+1, got, headerContentWidth, line)
			}
		}
	}

	for _, stage := range []stage{stageInput, stageImportSource, stageConfig, stageInfo} {
		m.stage = stage
		for _, line := range strings.Split(renderFooter(m), "\n") {
			if got := lipgloss.Width(stripANSI(line)); got > headerContentWidth {
				t.Fatalf("footer stage %d width = %d, exceeds %d: %q", stage, got, headerContentWidth, line)
			}
		}
	}
}
