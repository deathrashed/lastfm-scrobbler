package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/sessionstore"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

func renderResultsView(m model) string {
	selected := 0
	if len(m.results) > 0 {
		selected = m.resultsCursor + 1
	}
	counter := renderSelectedBadge(selected, len(m.results))
	list := renderSimpleAlbumList(m.results, m.resultsCursor, 13)
	var detail string
	if len(m.results) > 0 {
		album := m.results[m.resultsCursor]
		detail = renderInfoBox("MATCH", album.Artist+" — "+album.Title, "", 65, false)
	} else {
		detail = renderInfoBox("MATCH", "No album matches", "", 65, false)
	}
	return lipgloss.JoinVertical(lipgloss.Left,
		centerToHeader(detail),
		centerToHeader(counter),
		centerToHeader(list),
		"",
	)
}

func renderSimpleAlbumList(albums []lastfm.Album, cursor, maxRows int) string {
	const totalWidth = 65
	contentWidth := totalWidth - 4
	if len(albums) == 0 {
		return renderPanelBox([]string{theme.MutedStyle.Render("No albums found.")}, totalWidth, theme.BorderStyle)
	}
	cursor = minInt(maxInt(cursor, 0), len(albums)-1)
	start := visibleStart(cursor, len(albums), maxRows)
	end := minInt(len(albums), start+maxRows)
	visible := end - start
	thumb := scrollbarThumb(start, len(albums), visible)
	rows := make([]string, 0, visible)
	for row, index := 0, start; index < end; index, row = index+1, row+1 {
		mark := "  "
		if index == cursor {
			mark = theme.PromptStyle.Render("▸ ")
		}
		album := albums[index]
		text := album.Artist + " — " + album.Title
		text = truncateToWidth(text, contentWidth-5)
		scroll := " "
		if len(albums) > visible {
			scroll = theme.MutedStyle.Render("│")
			if row == thumb {
				scroll = theme.KeyStyle.Render("█")
			}
		}
		left := mark + theme.AlbumStyle.Render(text)
		gap := contentWidth - lipgloss.Width(left) - 1
		if gap < 1 {
			gap = 1
		}
		rows = append(rows, left+strings.Repeat(" ", gap)+scroll)
	}
	return renderPanelBox(rows, totalWidth, theme.BorderStyle)
}

func renderPreviewView(m model) string {
	record := m.queueRecord("preview")
	albumCount := uniqueAlbumCount(record.Queue)
	trackCount := len(record.Queue)
	estimated := time.Duration(maxInt(0, trackCount-1)) * m.interval
	rows := []string{
		statLine("ALBUMS", fmt.Sprintf("%d", albumCount), 59),
		statLine("TRACKS", fmt.Sprintf("%d", m.previewTrackCount()), 59),
		statLine("LOOPS", previewLoopText(m), 59),
		statLine("SCROBBLES", fmt.Sprintf("%d", trackCount), 59),
		statLine("INTERVAL", m.interval.String(), 59),
		statLine("ESTIMATED", formatDuration(estimated), 59),
	}
	preview := renderPanelBox(rows, 65, theme.BorderStyle)
	albumText := previewAlbumText(m)
	albumBox := renderInfoBox("QUEUE", albumText, fmt.Sprintf("%d album(s)", albumCount), 65, false)
	lines := []string{centerToHeader(albumBox), centerToHeader(preview), ""}
	if m.searching {
		lines = append(lines, centerToHeader(m.spinner.View()+" "+theme.MutedStyle.Render("Preparing authenticated session…")), "")
	}
	status := firstNonEmptyString(m.previewStatus, m.exportStatus)
	if status != "" {
		lines = append(lines, centerToHeader(theme.SuccessStyle.Render(status)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func statLine(label, value string, width int) string {
	left := theme.KeyStyle.Render(label + " ❯")
	right := theme.AlbumStyle.Render(value)
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func previewLoopText(m model) string {
	if m.mixedLoops() {
		return "per-album"
	}
	return fmt.Sprintf("%d", m.loopCount)
}

func previewAlbumText(m model) string {
	if len(m.selectedAlbums) == 1 {
		return m.selectedAlbums[0].Artist + " — " + m.selectedAlbums[0].Title
	}
	if len(m.selectedAlbums) > 1 {
		return fmt.Sprintf("%s — %d selected albums", m.selectedAlbums[0].Artist, len(m.selectedAlbums))
	}
	if len(m.scrobbleQueue) > 0 {
		return m.scrobbleQueue[0].Artist + " — " + m.scrobbleQueue[0].Album
	}
	return "Empty queue"
}

func uniqueAlbumCount(queue []sessionstore.Track) int {
	seen := map[string]bool{}
	for _, track := range queue {
		seen[strings.ToLower(track.Artist+"\x00"+track.Album)] = true
	}
	return len(seen)
}

func renderHistoryView(m model) string {
	const totalWidth = 65
	contentWidth := totalWidth - 4
	if len(m.history) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			centerToHeader(renderPanelBox([]string{theme.MutedStyle.Render("No completed sessions yet.")}, totalWidth, theme.BorderStyle)),
			"",
		)
	}
	maxRows := 13
	cursor := minInt(m.historyCursor, len(m.history)-1)
	start := visibleStart(cursor, len(m.history), maxRows)
	end := minInt(len(m.history), start+maxRows)
	visible := end - start
	thumb := scrollbarThumb(start, len(m.history), visible)
	rows := make([]string, 0, visible)
	for row, index := 0, start; index < end; index, row = index+1, row+1 {
		record := m.history[index]
		mark := "  "
		if index == cursor {
			mark = theme.PromptStyle.Render("▸ ")
		}
		label := historyLabel(record)
		label = truncateToWidth(label, contentWidth-8)
		status := strings.ToUpper(record.Status)
		scroll := " "
		if len(m.history) > visible {
			scroll = theme.MutedStyle.Render("│")
			if row == thumb {
				scroll = theme.KeyStyle.Render("█")
			}
		}
		left := mark + theme.AlbumStyle.Render(label)
		right := theme.MutedStyle.Render(status)
		gap := contentWidth - lipgloss.Width(left) - lipgloss.Width(right) - 1
		if gap < 1 {
			gap = 1
		}
		rows = append(rows, left+strings.Repeat(" ", gap)+right+scroll)
	}
	list := renderPanelBox(rows, totalWidth, theme.BorderStyle)
	record := m.history[cursor]
	detailRows := []string{
		statLine("DATE", record.StartedAt.Local().Format("2006-01-02 15:04"), 59),
		statLine("MODE", record.Mode, 59),
		statLine("TOTAL", fmt.Sprintf("%d", len(record.Queue)), 59),
		statLine("FAILED", fmt.Sprintf("%d", record.Failures), 59),
	}
	detail := renderPanelBox(detailRows, totalWidth, theme.BorderStyle)
	lines := []string{centerToHeader(list), centerToHeader(detail), ""}
	if m.historyStatus != "" {
		lines = append(lines, centerToHeader(theme.SuccessStyle.Render(m.historyStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func historyLabel(record sessionstore.Record) string {
	if len(record.Queue) == 0 {
		return record.StartedAt.Format("2006-01-02") + " — empty session"
	}
	first := record.Queue[0]
	albumCount := uniqueAlbumCount(record.Queue)
	if albumCount > 1 {
		return fmt.Sprintf("%s — %d albums", first.Artist, albumCount)
	}
	return first.Artist + " — " + first.Album
}

func renderRecoveryView(m model) string {
	if m.pending == nil {
		return centerToHeader(renderPanelBox([]string{theme.MutedStyle.Render("No unfinished session found.")}, 65, theme.BorderStyle))
	}
	record := *m.pending
	album := historyLabel(record)
	info := renderInfoBox("SESSION", album, fmt.Sprintf("%d / %d", record.Completed, len(record.Queue)), 65, false)
	rows := []string{
		statLine("STATUS", "unfinished session found", 59),
		statLine("STARTED", record.StartedAt.Local().Format("2006-01-02 15:04"), 59),
		statLine("MODE", record.Mode, 59),
		statLine("REMAINING", fmt.Sprintf("%d", maxInt(0, len(record.Queue)-record.Completed)), 59),
	}
	box := renderPanelBox(rows, 65, theme.BorderStyle)
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(info), centerToHeader(box), "")
}

func renderHelpView(m model) string {
	rows := []string{
		helpRow("↑ / ↓ / ← / →", "navigate lists and tabs"),
		helpRow("enter", "confirm or continue"),
		helpRow("space", "toggle a selected album or track"),
		helpRow("a", "select all / select none"),
		helpRow("/", "filter long discography lists"),
		helpRow("c", "clean duplicate editions in discography"),
		helpRow("s", "similar albums or sort, depending on screen"),
		helpRow("e", "export the current queue or history entry"),
		helpRow("h", "session history"),
		helpRow("p", "profiles"),
		helpRow("esc", "go back or cancel"),
		helpRow("q", "quit; active sessions can resume later"),
	}
	box := renderPanelBox(rows, 65, theme.BorderStyle)
	return lipgloss.JoinVertical(lipgloss.Left, centerToHeader(box), centerToHeader(theme.MutedStyle.Render("press ? or esc to close")), "")
}

func helpRow(key, description string) string {
	left := theme.KeyStyle.Render(key)
	right := theme.MutedStyle.Render(description)
	gap := 59 - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return left + strings.Repeat(" ", gap) + right
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
