package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

func renderTrackSelectView(m model) string {
	refs := m.flattenedTracks()
	total := len(refs)
	selected := m.selectedTrackCount()
	albumText := ""
	albumCount := fmt.Sprintf("(%d tracks)", total)
	if len(m.selectedAlbums) > 1 {
		albumText = fmt.Sprintf("%s — %d selected albums", m.selectedAlbums[0].Artist, len(m.selectedAlbums))
	} else if len(m.selectedAlbums) == 1 {
		albumText = m.selectedAlbums[0].Artist + " — " + m.selectedAlbums[0].Title
	} else {
		albumText = m.selectedAlbum.Artist + " — " + m.selectedAlbum.Title
	}
	albumBox := renderInfoBox("ALBUM", albumText, albumCount, 65, false)
	selectedBox := renderSelectedBadge(selected, total)
	trackBox := renderTrackList(m, refs)
	var loopInfo string
	if len(refs) > 0 {
		ref := refs[minInt(m.trackCursor, len(refs)-1)]
		if len(m.selectedAlbums) > 1 {
			loopInfo = renderInfoBox("ALBUM LOOP", ref.Album.Title, fmt.Sprintf("%d", m.loopForAlbum(ref.AlbumIndex)), 65, false)
		}
	}
	parts := []string{centerToHeader(albumBox), centerToHeader(selectedBox)}
	if loopInfo != "" {
		parts = append(parts, centerToHeader(loopInfo))
	}
	parts = append(parts, centerToHeader(trackBox), "")
	if m.exportStatus != "" {
		parts = append(parts, centerToHeader(theme.SuccessStyle.Render(m.exportStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func renderTrackList(m model, refs []trackRef) string {
	const totalWidth = 65
	maxRows := 13
	if m.height > 0 {
		maxRows = maxInt(6, minInt(16, m.height-21))
	}
	if len(refs) == 0 {
		return renderPanelBox([]string{theme.MutedStyle.Render("No tracks were returned for this album.")}, totalWidth, theme.BorderStyle)
	}
	contentWidth := totalWidth - 4
	start := visibleStart(m.trackCursor, len(refs), maxRows)
	end := minInt(len(refs), start+maxRows)
	visible := end - start
	thumb := scrollbarThumb(start, len(refs), visible)
	rows := make([]string, 0, visible)
	for row, index := 0, start; index < end; index, row = index+1, row+1 {
		ref := refs[index]
		cursor := "  "
		if index == m.trackCursor {
			cursor = theme.PromptStyle.Render("▸ ")
		}
		check := theme.BorderStyle.Render("☐")
		if m.trackSelected[ref.GlobalIndex] {
			check = theme.KeyStyle.Render("☑")
		}
		number := fmt.Sprintf("%2d", ref.TrackIndex+1)
		title := ref.Track.Title
		if len(m.selectedAlbums) > 1 {
			title += " — " + ref.Album.Title
		}
		title = truncateToWidth(title, contentWidth-11)
		scroll := " "
		if len(refs) > visible {
			scroll = theme.MutedStyle.Render("│")
			if row == thumb {
				scroll = theme.KeyStyle.Render("█")
			}
		}
		left := cursor + check + "  " + theme.MutedStyle.Render(number) + "  " + theme.AlbumStyle.Render(title)
		gap := contentWidth - lipgloss.Width(left) - 1
		if gap < 1 {
			gap = 1
		}
		rows = append(rows, left+strings.Repeat(" ", gap)+scroll)
	}
	return renderPanelBox(rows, totalWidth, theme.BorderStyle)
}

func renderScrobblingView(m model) string { return renderScrobbleStatus(m, false) }
func renderDoneView(m model) string       { return renderScrobbleStatus(m, true) }

func renderScrobbleStatus(m model, complete bool) string {
	total := len(m.scrobbleQueue)
	completed := minInt(m.scrobbleIdx, total)
	item, hasItem := m.displayQueueItem(complete)
	var sections []string
	if hasItem && (m.modeChoice == "discography" || item.AlbumTotal > 1) {
		sections = append(sections, centerToHeader(renderInfoBox("ARTIST", item.Artist, "", 65, false)))
		sections = append(sections, centerToHeader(renderInfoBox("ALBUM", item.Album, fmt.Sprintf("%d / %d", item.AlbumIndex, item.AlbumTotal), 65, false)))
		trackTitle := item.Title
		trackCount := fmt.Sprintf("%d / %d", item.TrackIndex, item.TrackTotal)
		if complete {
			trackTitle = "Complete"
			trackCount = fmt.Sprintf("%d / %d", item.TrackTotal, item.TrackTotal)
		}
		sections = append(sections, centerToHeader(renderInfoBox("TRACK", trackTitle, trackCount, 65, false)))
	} else if hasItem {
		sections = append(sections, centerToHeader(renderInfoBox("ALBUM", item.Artist+" — "+item.Album, "", 65, false)))
		status := item.Title
		if complete {
			status = "Complete"
		}
		sections = append(sections, centerToHeader(renderInfoBox("SCROBBLING", status, fmt.Sprintf("%d / %d", completed, total), 65, false)))
	} else {
		sections = append(sections, centerToHeader(renderInfoBox("SCROBBLING", "No tracks queued", "0 / 0", 65, false)))
	}

	percent := 0.0
	if total > 0 {
		percent = float64(completed) / float64(total)
	}
	if complete && total > 0 {
		percent = 1
		completed = total
	}
	sections = append(sections, centerToHeader(renderProgressBox(m, percent)))
	eta := "DONE"
	if !complete {
		eta = formatDuration(time.Duration(maxInt(0, total-completed)) * m.interval)
	}
	stats := joinThreeLineBoxes([]string{renderStatBox("ETA", eta, 19), renderStatBox("TOTAL", fmt.Sprintf("%d / %d", completed, total), 19)}, theme.SepStyle.Render("•"))
	sections = append(sections, centerToHeader(stats), "")
	if m.skippedDuplicates > 0 {
		sections = append(sections, centerToHeader(theme.WarnStyle.Render(fmt.Sprintf("%d recent duplicate(s) skipped", m.skippedDuplicates))), "")
	}
	if len(m.failures) > 0 {
		sections = append(sections, centerToHeader(theme.WarnStyle.Render(fmt.Sprintf("%s %d scrobble(s) failed after retries", theme.IconError, len(m.failures)))), "")
	}
	if complete && m.exportStatus != "" {
		sections = append(sections, centerToHeader(theme.SuccessStyle.Render(m.exportStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m model) displayQueueItem(complete bool) (queuedTrack, bool) {
	if len(m.scrobbleQueue) == 0 {
		return queuedTrack{}, false
	}
	if complete {
		return m.scrobbleQueue[len(m.scrobbleQueue)-1], true
	}
	index := m.scrobbleIdx
	if index >= len(m.scrobbleQueue) {
		index = len(m.scrobbleQueue) - 1
	}
	return m.scrobbleQueue[index], true
}

func renderProgressBox(m model, percent float64) string {
	const totalWidth = 65
	contentWidth := totalWidth - 4
	percent = maxFloat(0, minFloat(1, percent))
	percentText := fmt.Sprintf("%3.0f%%", percent*100)
	prefix := " " + theme.KeyStyle.Render(theme.IconTimer) + "  "
	barWidth := contentWidth - lipgloss.Width(prefix) - lipgloss.Width(percentText) - 2
	if barWidth < 10 {
		barWidth = 10
	}
	filled := int(percent * float64(barWidth))
	if percent >= 1 {
		filled = barWidth
	}
	markers := progressMarkers(m.scrobbleQueue, barWidth)
	var bar strings.Builder
	for index := 0; index < barWidth; index++ {
		if markers[index] {
			bar.WriteString(theme.SepStyle.Render("|"))
			continue
		}
		if index < filled {
			bar.WriteString(theme.KeyStyle.Render("█"))
		} else {
			bar.WriteString(theme.MutedStyle.Render("░"))
		}
	}
	line := prefix + bar.String() + "  " + theme.AlbumStyle.Render(percentText)
	return renderPanelBox([]string{line}, totalWidth, theme.BorderStyle)
}

func progressMarkers(queue []queuedTrack, barWidth int) map[int]bool {
	markers := map[int]bool{}
	if len(queue) < 2 || barWidth < 3 {
		return markers
	}
	var boundaries []int
	for index := 1; index < len(queue); index++ {
		if queue[index-1].AlbumIndex != queue[index].AlbumIndex {
			boundaries = append(boundaries, index)
		}
	}
	sort.Ints(boundaries)
	for _, boundary := range boundaries {
		cell := int(float64(boundary) / float64(len(queue)) * float64(barWidth))
		if cell > 0 && cell < barWidth-1 {
			markers[cell] = true
		}
	}
	return markers
}

func renderStatBox(label, value string, totalWidth int) string {
	innerWidth := totalWidth - 2
	text := theme.KeyStyle.Render(label+" ❯ ") + theme.AlbumStyle.Render(value)
	return strings.Join([]string{
		theme.BorderStyle.Render("╭" + strings.Repeat("─", innerWidth) + "╮"),
		theme.BorderStyle.Render("│") + centerText(text, innerWidth) + theme.BorderStyle.Render("│"),
		theme.BorderStyle.Render("╰" + strings.Repeat("─", innerWidth) + "╯"),
	}, "\n")
}

func formatDuration(duration time.Duration) string {
	duration = duration.Round(time.Second)
	hours := duration / time.Hour
	duration -= hours * time.Hour
	minutes := duration / time.Minute
	duration -= minutes * time.Minute
	seconds := duration / time.Second
	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}
	if minutes > 0 {
		return fmt.Sprintf("%dm%ds", minutes, seconds)
	}
	return fmt.Sprintf("%ds", seconds)
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
