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
	trackBox := renderTrackList(m, refs, selected, total)
	parts := make([]string, 0, 4)

	if len(m.selectedAlbums) > 1 {
		taskText := fmt.Sprintf("%d selected albums", len(m.selectedAlbums))
		if m.modeChoice == "discography" && strings.TrimSpace(m.currentArtist()) != "" {
			taskText = fmt.Sprintf("%s — %d selected albums", m.currentArtist(), len(m.selectedAlbums))
		}
		parts = append(parts, m.centerToApp(renderInfoBox("TASK", taskText, fmt.Sprintf("(%d tracks)", total), m.panelWidth(), false)))
		if len(refs) > 0 {
			ref := refs[minInt(m.trackCursor, len(refs)-1)]
			albumText := ref.Album.Title
			if m.modeChoice != "discography" {
				albumText = ref.Album.Artist + " — " + ref.Album.Title
			}
			parts = append(parts, m.centerToApp(renderInfoBox("ALBUM", albumText, fmt.Sprintf("%d", m.loopForAlbum(ref.AlbumIndex)), m.panelWidth(), false)))
		}
	} else {
		album := m.selectedAlbum
		if len(m.selectedAlbums) == 1 {
			album = m.selectedAlbums[0]
		}
		albumText := album.Title
		if m.modeChoice != "manual" && strings.TrimSpace(album.Artist) != "" {
			albumText = album.Artist + " — " + album.Title
		}
		parts = append(parts, m.centerToApp(renderInfoBox("ALBUM", albumText, fmt.Sprintf("(%d tracks)", total), m.panelWidth(), false)))
	}

	parts = append(parts, m.centerToApp(trackBox), "")
	if m.exportStatus != "" {
		parts = append(parts, m.centerToApp(theme.SuccessStyle.Render(m.exportStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func renderTrackList(m model, refs []trackRef, selected, total int) string {
	totalWidth := m.panelWidth()
	maxRows := trackListMaxRows(m)
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
		focused := index == m.trackCursor
		hovered := m.hoverRegion == "tracks:"+fmt.Sprintf("%d", index)
		cursor := "  "
		if focused {
			cursor = theme.PromptStyle.Render("❯ ")
		}
		check := theme.SecondaryTextStyle.Render("○")
		if m.trackSelected[ref.GlobalIndex] {
			check = theme.AccentTextStyle.Render("●")
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
		numberStyle := theme.SecondaryTextStyle
		titleStyle := theme.PrimaryTextStyle
		if focused {
			numberStyle = theme.FocusedRowValueStyle
			titleStyle = theme.FocusedRowLabelStyle
		} else if hovered {
			numberStyle = theme.HoverRowValueStyle
			titleStyle = theme.HoverRowLabelStyle
		}
		left := cursor + check + "  " + numberStyle.Render(number) + "  " + titleStyle.Render(title)
		gap := contentWidth - lipgloss.Width(left) - 1
		if gap < 1 {
			gap = 1
		}
		rows = append(rows, left+strings.Repeat(" ", gap)+scroll)
	}
	return renderPanelBoxWithSelectedAttachment(rows, totalWidth, selected, total, theme.BorderStyle)
}

func trackSelectListTopOffset(m model) int {
	refs := m.flattenedTracks()
	if len(m.selectedAlbums) > 1 {
		taskText := fmt.Sprintf("%d selected albums", len(m.selectedAlbums))
		if m.modeChoice == "discography" && strings.TrimSpace(m.currentArtist()) != "" {
			taskText = fmt.Sprintf("%s — %d selected albums", m.currentArtist(), len(m.selectedAlbums))
		}
		offset := renderedLineCount(renderInfoBox("TASK", taskText, fmt.Sprintf("(%d tracks)", len(refs)), m.panelWidth(), false))
		if len(refs) > 0 {
			ref := refs[minInt(m.trackCursor, len(refs)-1)]
			albumText := ref.Album.Title
			if m.modeChoice != "discography" {
				albumText = ref.Album.Artist + " — " + ref.Album.Title
			}
			offset += renderedLineCount(renderInfoBox("ALBUM", albumText, fmt.Sprintf("%d", m.loopForAlbum(ref.AlbumIndex)), m.panelWidth(), false))
		}
		return offset
	}
	album := m.selectedAlbum
	if len(m.selectedAlbums) == 1 {
		album = m.selectedAlbums[0]
	}
	albumText := album.Title
	if m.modeChoice != "manual" && strings.TrimSpace(album.Artist) != "" {
		albumText = album.Artist + " — " + album.Title
	}
	return renderedLineCount(renderInfoBox("ALBUM", albumText, fmt.Sprintf("(%d tracks)", len(refs)), m.panelWidth(), false))
}

func renderedLineCount(value string) int {
	if value == "" {
		return 0
	}
	return strings.Count(value, "\n") + 1
}

func trackListMaxRows(m model) int {
	available := m.height - m.headerHeight() - trackSelectListTopOffset(m) - 7
	if m.height <= 0 {
		return 32
	}
	return maxInt(6, minInt(32, available))
}

func renderScrobblingView(m model) string { return renderScrobbleStatus(m, false) }
func renderDoneView(m model) string       { return renderScrobbleStatus(m, true) }

func renderScrobbleStatus(m model, complete bool) string {
	total := len(m.scrobbleQueue)
	completed := minInt(m.scrobbleIdx, total)
	item, hasItem := m.displayQueueItem(complete)
	var sections []string
	artistInHeader := strings.TrimSpace(m.headerArtist()) != ""
	if hasItem && (m.modeChoice == "discography" || item.AlbumTotal > 1) {
		if !artistInHeader {
			sections = append(sections, m.centerToApp(renderInfoBox("ARTIST", item.Artist, "", m.panelWidth(), false)))
		}
		sections = append(sections, m.centerToApp(renderInfoBox("ALBUM", item.Album, fmt.Sprintf("%d / %d", item.AlbumIndex, item.AlbumTotal), m.panelWidth(), false)))
		trackTitle := item.Title
		trackCount := fmt.Sprintf("%d / %d", item.TrackIndex, item.TrackTotal)
		if complete {
			trackTitle = "Complete"
			trackCount = fmt.Sprintf("%d / %d", item.TrackTotal, item.TrackTotal)
		}
		sections = append(sections, m.centerToApp(renderInfoBox("TRACK", trackTitle, trackCount, m.panelWidth(), false)))
	} else if hasItem {
		albumText := item.Artist + " — " + item.Album
		if artistInHeader {
			albumText = item.Album
		}
		sections = append(sections, m.centerToApp(renderInfoBox("ALBUM", albumText, "", m.panelWidth(), false)))
		status := item.Title
		if complete {
			status = "Complete"
		}
		sections = append(sections, m.centerToApp(renderInfoBox("SCROBBLING", status, fmt.Sprintf("%d / %d", completed, total), m.panelWidth(), false)))
	} else {
		sections = append(sections, m.centerToApp(renderInfoBox("SCROBBLING", "No tracks queued", "0 / 0", m.panelWidth(), false)))
	}

	percent := 0.0
	if total > 0 {
		percent = float64(completed) / float64(total)
	}
	if complete && total > 0 {
		percent = 1
		completed = total
	}
	sections = append(sections, m.centerToApp(renderProgressBox(m, percent, complete)))
	eta := "DONE"
	if !complete {
		eta = formatDuration(time.Duration(maxInt(0, total-completed)) * m.interval)
	}
	stats := joinThreeLineBoxes([]string{renderStatBox("ETA", eta, 19), renderStatBox("TOTAL", fmt.Sprintf("%d / %d", completed, total), 19)}, theme.SepStyle.Render("•"))
	sections = append(sections, m.centerToApp(stats), "")
	if m.skippedDuplicates > 0 {
		sections = append(sections, m.centerToApp(theme.WarnStyle.Render(fmt.Sprintf("%d recent duplicate(s) skipped", m.skippedDuplicates))), "")
	}
	if len(m.failures) > 0 {
		sections = append(sections, m.centerToApp(theme.WarnStyle.Render(fmt.Sprintf("%s %d scrobble(s) failed after retries", theme.IconError, len(m.failures)))), "")
	}
	if complete && m.exportStatus != "" {
		sections = append(sections, m.centerToApp(theme.SuccessStyle.Render(m.exportStatus)), "")
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

func renderProgressBox(m model, percent float64, complete bool) string {
	totalWidth := m.panelWidth()
	contentWidth := totalWidth - 4
	percent = maxFloat(0, minFloat(1, percent))

	prefix := m.spinner.View() + " "
	suffix := theme.AlbumStyle.Render(fmt.Sprintf("%3.0f%%", percent*100))
	if complete {
		prefix = theme.CompleteStyle.Render(theme.IconSuccess) + "  " + theme.PrimaryTextStyle.Render(theme.IconDashboard) + "  "
		suffix = theme.CompleteStyle.Render("DONE")
	}

	barWidth := contentWidth - lipgloss.Width(prefix) - lipgloss.Width(suffix) - 2
	if barWidth < 10 {
		barWidth = 10
	}
	filled := int(percent * float64(barWidth))
	if percent >= 1 || complete {
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
	line := prefix + bar.String() + "  " + suffix
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
