package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

const (
	minAppWidth  = 67
	maxAppWidth  = 127
	maxWorkWidth = 103
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
	switch width = appWidth(width); {
	case width >= 104:
		return densityWide
	case width >= 80:
		return densityComfortable
	default:
		return densityCompact
	}
}

func (m model) appWidth() int { return appWidth(m.width) }

func (m model) appX() int { return appOffset(m.width) }

func (m model) contentWidth() int { return maxInt(1, m.appWidth()-2) }

func workWidth(appWidth int) int {
	return minInt(maxInt(1, appWidth-2), maxWorkWidth)
}

func (m model) panelWidth() int { return workWidth(m.appWidth()) }

// Documentation surfaces are deliberately allowed to use the whole responsive
// application content width. Their rows benefit from horizontal space, while
// their navigation/control elements still keep bounded natural-size gaps.
func (m model) documentationPanelWidth() int { return m.contentWidth() }
func (m model) infoPanelWidth() int          { return m.documentationPanelWidth() }
func (m model) helpPanelWidth() int          { return m.documentationPanelWidth() }

func (m model) infoX() int {
	return maxInt(0, (m.appWidth()-m.infoPanelWidth())/2)
}

// workX is the horizontal origin of the centered working-content area inside
// the responsive application. The outer header may grow to 127 cells, while
// dense working panels intentionally stop at maxWorkWidth so the wide layout
// keeps the same proportions as the original 67-cell design.
func (m model) workX() int {
	return maxInt(0, (m.appWidth()-m.panelWidth())/2)
}

func (m model) density() layoutDensity { return densityFor(m.width) }

func (m model) nowPlayingEnabled() bool { return !m.compactHeaderEnabled() && m.cfg.NowPlaying }

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
	return positions
}

func responsiveCardPositions(widths []int, panelWidth int) ([]int, int) {
	if len(widths) == 0 {
		return nil, 0
	}
	minimum := 0
	for _, width := range widths {
		minimum += width
	}
	gap := responsiveCardGap(panelWidth)
	if len(widths) > 1 {
		maxGap := maxInt(1, (panelWidth-minimum)/(len(widths)-1))
		gap = minInt(gap, maxGap)
	}
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

func responsiveCardGap(panelWidth int) int {
	switch {
	case panelWidth >= 102:
		return 5
	case panelWidth >= 78:
		return 3
	default:
		return 1
	}
}

func joinResponsiveBoxes(boxes []string, widths []int, panelWidth int, separator string) string {
	positions, gap := responsiveCardPositions(widths, panelWidth)
	rows := make([][]string, len(boxes))
	maxRows := 0
	for index, box := range boxes {
		rows[index] = strings.Split(box, "\n")
		maxRows = maxInt(maxRows, len(rows[index]))
	}
	output := make([]string, 0, maxRows)
	for row := 0; row < maxRows; row++ {
		line := strings.Repeat(" ", maxInt(0, positions[0]))
		for index, width := range widths {
			if index > 0 {
				separatorText := strings.Repeat(" ", gap)
				// Navigation separators belong on the content row only. Rendering
				// them on every line produces a vertical column of bullets between
				// otherwise horizontal three-line cards.
				if row == maxRows/2 && lipgloss.Width(separator) <= gap {
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
