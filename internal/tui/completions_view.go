package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/completion"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

const completionRowsStartOffset = 4

func renderCompletionsView(m model) string {
	manager := completion.DefaultManager()
	rows := []string{
		setupRow("DETECTED SHELL", completion.DetectShell().String(), false),
		setupRow("STATUS", m.completionStatusText(), false),
	}
	for index, shell := range completion.Shells {
		marker := theme.SecondaryTextStyle.Render("○")
		label := theme.RowLabelStyle.Render(shell.String())
		if shell == m.completionShell {
			marker = theme.AccentTextStyle.Render("●")
			label = theme.FocusedRowLabelStyle.Render(shell.String())
		}
		status := manager.Status(shell)
		if index == completionShellIndex(m.completionShell) {
			status = m.completionStatus
		}
		rows = append(rows, fitStyled(marker+" "+label+strings.Repeat(" ", maxInt(1, 52-lipgloss.Width(marker+" "+label)))+theme.MutedStyle.Render(string(status)), 61))
	}
	if m.completionMessage != "" {
		rows = append(rows, theme.SuccessStyle.Render(truncateToWidth(m.completionMessage, 61)))
	}
	return setupPanel("S H E L L  C O M P L E T I O N S", rows)
}

func (m model) completionStatusText() string {
	return string(m.completionStatus)
}

func completionScreenRegions(m model, bodyY int) []mouseRegion {
	regions := make([]mouseRegion, 0, len(completion.Shells))
	for index, shell := range completion.Shells {
		regions = append(regions, mouseRegion{
			id:     "completion:shell:" + shell.String(),
			x:      1,
			y:      bodyY + completionRowsStartOffset + index,
			width:  65,
			height: 1,
		})
	}
	return regions
}
