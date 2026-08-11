package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	minAppWidth = 67
	maxAppWidth = 127
)

type layoutDensity int

const (
	densityCompact layoutDensity = iota
	densityComfortable
	densityWide
)

func appWidth(terminalWidth int) int {
	if terminalWidth < minAppWidth {
		return minAppWidth
	}
	if terminalWidth > maxAppWidth {
		return maxAppWidth
	}
	return terminalWidth
}

func appOffset(terminalWidth int) int {
	width := appWidth(terminalWidth)
	if terminalWidth <= width {
		return 0
	}
	return (terminalWidth - width) / 2
}

func densityFor(width int) layoutDensity {
	switch appWidth(width) {
	case 104, 105, 106, 107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124, 125, 126, 127:
		return densityWide
	case 80, 81, 82, 83, 84, 85, 86, 87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103:
		return densityComfortable
	default:
		return densityCompact
	}
}

func (m model) appWidth() int { return appWidth(m.width) }

func (m model) appX() int { return appOffset(m.width) }

func (m model) contentWidth() int { return maxInt(1, m.appWidth()-2) }

func (m model) panelWidth() int { return m.contentWidth() }

func (m model) density() layoutDensity { return densityFor(m.width) }

func (m model) nowPlayingEnabled() bool { return !m.cfg.CompactHeader && m.cfg.NowPlaying }

func (m *model) resizeInputs() {
	width := maxInt(1, m.panelWidth()-4)
	m.searchInput.Width = width
	m.filterInput.Width = maxInt(1, width-4)
	m.profileInput.Width = maxInt(1, width-4)
	m.configInput.Width = maxInt(1, width-4)
	m.envInput.Width = maxInt(1, width-4)
	for index := range m.setupInputs {
		m.setupInputs[index].Width = maxInt(1, width-4)
	}
}

func (m model) centerToApp(value string) string {
	if value == "" {
		return ""
	}
	lines := splitLines(value)
	for index, line := range lines {
		lines[index] = centerText(line, m.appWidth())
	}
	return joinLines(lines)
}

func centerToWidth(value string, width int) string {
	if value == "" {
		return ""
	}
	lines := splitLines(value)
	for index, line := range lines {
		lines[index] = centerText(line, width)
	}
	return joinLines(lines)
}

func splitLines(value string) []string {
	lines := make([]string, 0, 1)
	start := 0
	for index, r := range value {
		if r == '\n' {
			lines = append(lines, value[start:index])
			start = index + 1
		}
	}
	return append(lines, value[start:])
}

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	result := lines[0]
	for _, line := range lines[1:] {
		result += "\n" + line
	}
	return result
}

func responsiveValueWidth(totalWidth, leftWidth, rightWidth int) int {
	return maxInt(1, totalWidth-leftWidth-rightWidth-2)
}

func responsiveText(value string, width int) string {
	return truncateToWidth(value, maxInt(1, width))
}

func (m model) visibleRows(reserved, minimum, maximum int) int {
	if m.height <= 0 {
		return maximum
	}
	available := m.height - m.headerHeight() - reserved
	return maxInt(minimum, minInt(maximum, available))
}

func dashboardCardPositions(panelWidth int) []int {
	positions, _ := responsiveCardPositions([]int{19, 25, 18}, panelWidth)
	for index := range positions {
		positions[index]++
	}
	return positions
}

func responsiveCardPositions(widths []int, panelWidth int) ([]int, int) {
	if len(widths) == 0 {
		return nil, 0
	}
	minimum := 1
	for _, width := range widths {
		minimum += width
	}
	minimum += len(widths) - 1
	extra := maxInt(0, panelWidth-minimum)
	gap := 1 + extra/maxInt(1, len(widths)-1)
	groupWidth := 0
	for _, width := range widths {
		groupWidth += width
	}
	groupWidth += gap * (len(widths) - 1)
	start := maxInt(0, (panelWidth-groupWidth)/2)
	positions := make([]int, len(widths))
	for index := range widths {
		if index > 0 {
			positions[index] = positions[index-1] + widths[index-1] + gap
		}
		if index == 0 {
			positions[index] = start
		}
	}
	return positions, gap
}

func joinResponsiveBoxes(boxes []string, widths []int, panelWidth int, separator string) string {
	_, gap := responsiveCardPositions(widths, panelWidth)
	rows := make([][]string, len(boxes))
	maxRows := 0
	for index, box := range boxes {
		rows[index] = strings.Split(box, "\n")
		maxRows = maxInt(maxRows, len(rows[index]))
	}
	output := make([]string, 0, maxRows)
	for row := 0; row < maxRows; row++ {
		line := ""
		for index, width := range widths {
			if index > 0 {
				separatorText := strings.Repeat(" ", gap)
				if lipgloss.Width(separator) <= gap {
					left := (gap - lipgloss.Width(separator)) / 2
					separatorText = strings.Repeat(" ", left) + separator + strings.Repeat(" ", gap-lipgloss.Width(separator)-left)
				}
				line += separatorText
			}
			if row < len(rows[index]) {
				line += fitStyled(rows[index][row], width)
			} else {
				line += strings.Repeat(" ", width)
			}
		}
		if lipgloss.Width(line) < panelWidth {
			line += strings.Repeat(" ", panelWidth-lipgloss.Width(line))
		}
		output = append(output, line)
	}
	return strings.Join(output, "\n")
}

func centeredGroupStart(containerWidth, groupWidth int) int {
	return maxInt(0, (containerWidth-groupWidth)/2)
}

func displayWidth(value string) int { return lipgloss.Width(value) }
