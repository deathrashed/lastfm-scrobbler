package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/completion"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

const completionRowsStartOffset = 3

func renderCompletionsView(m model) string {
	manager := completion.DefaultManager()
	width := m.panelWidth()
	contentWidth := maxInt(1, width-4)
	rows := []string{
		setupRow("DETECTED SHELL", completion.DetectShell().String(), false, width),
		setupRow("STATUS", string(m.completionStatusFor(manager, m.completionShell)), false, width),
	}
	for index, shell := range completion.Shells {
		selected := shell == m.completionShell
		status := manager.Status(shell)
		if index == completionShellIndex(m.completionShell) {
			status = m.completionStatusFor(manager, shell)
		}
		rows = append(rows, completionShellRow(shell, status, selected, contentWidth))
	}
	if m.completionMessage != "" {
		rows = append(rows, theme.SuccessStyle.Render(truncateToWidth(m.completionMessage, contentWidth)))
	}
	// The page header already carries the attached C O M P L E T I O N S title.
	// The body is therefore a normal working panel, not a second nested setup
	// title. This keeps the screen consistent with the approved reference view.
	return m.centerToApp(renderPanelBox(rows, width, theme.BorderStyle))
}

func completionShellRow(shell completion.Shell, status completion.Status, selected bool, width int) string {
	marker := theme.SecondaryTextStyle.Render("○")
	labelStyle := theme.RowLabelStyle
	if selected {
		marker = theme.AccentTextStyle.Render("●")
		labelStyle = theme.FocusedRowLabelStyle
	}
	left := marker + " " + labelStyle.Render(shell.String())
	rightText := string(status)
	availableRight := maxInt(1, width-lipgloss.Width(left)-1)
	rightText = truncateToWidth(rightText, availableRight)
	right := theme.MutedStyle.Render(rightText)
	gap := maxInt(1, width-lipgloss.Width(left)-lipgloss.Width(right))
	return fitStyled(left+strings.Repeat(" ", gap)+right, width)
}

func (m model) completionStatusText() string {
	return string(m.completionStatusFor(completion.DefaultManager(), m.completionShell))
}

func (m model) completionStatusFor(manager completion.Manager, shell completion.Shell) completion.Status {
	if shell == m.completionShell && m.completionStatus != "" {
		return m.completionStatus
	}
	return manager.Status(shell)
}

func completionScreenRegions(m model, bodyY int) []mouseRegion {
	regions := make([]mouseRegion, 0, len(completion.Shells))
	for index, shell := range completion.Shells {
		regions = append(regions, mouseRegion{
			id:     "completion:shell:" + shell.String(),
			x:      m.workX(),
			y:      bodyY + completionRowsStartOffset + index,
			width:  m.panelWidth(),
			height: 1,
		})
	}
	return regions
}
