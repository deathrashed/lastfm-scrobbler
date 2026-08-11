package tui

import (
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

const (
	fileSourceRowHeight = 3
	filePathPanelOffset = fileSourceRowHeight*2 + 2
)

type fileSourceSpec struct {
	label       string
	width       int
	description string
	placeholder string
	types       string
}

var fileSourceSpecs = []fileSourceSpec{
	{label: "L I S T  F I L E", width: 21, description: "load albums from a TXT, CSV, TSV, or JSON list", placeholder: "/path/to/TXT, CSV, TSV, or JSON list", types: "TYPES ❯ TXT • CSV • TSV • JSON"},
	{label: "P L A Y L I S T", width: 19, description: "load albums from an M3U or M3U8 playlist", placeholder: "/path/to/M3U or M3U8 playlist", types: "TYPES ❯ M3U • M3U8"},
	{label: "A L B U M   F O L D E R", width: 27, description: "scan one album folder", placeholder: "/path/to/Artist/Album folder", types: "TYPE ❯ FOLDER"},
	{label: "A R T I S T   F O L D E R", width: 29, description: "scan album folders inside one artist folder", placeholder: "/path/to/Artist folder", types: "TYPE ❯ FOLDER"},
}

func fileSourceSpecFor(index int) fileSourceSpec {
	if index < 0 || index >= len(fileSourceSpecs) {
		return fileSourceSpecs[0]
	}
	return fileSourceSpecs[index]
}

func fileSourceRegions(bodyY, panelWidth int) []mouseRegion {
	regions := make([]mouseRegion, 0, len(fileSourceSpecs)+1)
	positions := fileSourceCardPositions(panelWidth)
	for index, spec := range fileSourceSpecs {
		regions = append(regions, mouseRegion{
			id:     "import:" + strconv.Itoa(index),
			x:      positions[index][0],
			y:      bodyY + positions[index][1],
			width:  spec.width,
			height: fileSourceRowHeight,
		})
	}
	regions = append(regions, mouseRegion{id: "file:path", x: 0, y: bodyY + filePathPanelOffset, width: panelWidth, height: 5})
	return regions
}

func fileSourceDividerX(panelWidth int) int {
	// File has intentionally asymmetric card widths. The visual anchor is the
	// one-cell separator between the two columns, not the bounding box of each
	// row. Keeping this divider at one fixed x-coordinate makes LIST/PLAYLIST
	// and ALBUM/ARTIST line up exactly like the original design.
	return maxInt(0, (panelWidth+1)/2)
}

func fileSourceCardPositions(panelWidth int) [][2]int {
	positions := make([][2]int, 0, len(fileSourceSpecs))
	dividerX := fileSourceDividerX(panelWidth)
	for row, indexes := range [][2]int{{0, 1}, {2, 3}} {
		leftWidth := fileSourceSpecs[indexes[0]].width
		rightWidth := fileSourceSpecs[indexes[1]].width
		leftX := maxInt(0, dividerX-leftWidth)
		rightX := minInt(maxInt(0, panelWidth-rightWidth), dividerX+1)
		positions = append(positions,
			[2]int{leftX, row * fileSourceRowHeight},
			[2]int{rightX, row * fileSourceRowHeight},
		)
	}
	return positions
}

func renderFileTypesContent(value string) string {
	parts := strings.SplitN(value, " ❯ ", 2)
	if len(parts) != 2 {
		return theme.SummaryLabelStyle.Render(value)
	}
	return theme.SummaryLabelStyle.Render(parts[0]+" ") +
		theme.SummaryArrowStyle.Render("❯") + " " +
		theme.SummaryValueStyle.Render(parts[1])
}

func renderFilePathView(m model) string {
	spec := fileSourceSpecFor(m.importSourceIndex)
	value := m.searchInput.View()
	if strings.TrimSpace(stripANSI(value)) == "" {
		value = theme.MutedStyle.Render(spec.placeholder)
	}
	prefix := theme.RowLabelStyle.Render("PATH ") + theme.RowArrowStyle.Render("❯ ")
	if m.filePathFocused {
		prefix = theme.FocusedRowLabelStyle.Render("PATH ") + theme.FocusedRowArrowStyle.Render("❯ ")
	}
	value = fitStyled(value, maxInt(1, m.panelWidth()-4-lipgloss.Width(prefix)))
	return renderPanelBoxWithBadgeAttachment(
		[]string{prefix + value},
		m.panelWidth(),
		renderFileTypesContent(spec.types),
		theme.BorderStyle,
	)
}
