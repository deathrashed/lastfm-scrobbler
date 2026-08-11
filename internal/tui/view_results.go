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
	detail := renderManualResultDetail(m)
	list := renderSimpleAlbumListWithResults(m.results, m.resultsCursor, manualResultsMaxRows(m), m.hoverRegion, "results", m.panelWidth())
	return lipgloss.JoinVertical(lipgloss.Left,
		m.centerToApp(detail),
		m.centerToApp(list),
		"",
	)
}

func renderManualResultDetail(m model) string {
	if len(m.results) > 0 {
		album := m.results[minInt(maxInt(m.resultsCursor, 0), len(m.results)-1)]
		return renderInfoBox("MATCH", album.Artist+" — "+album.Title, "", m.panelWidth(), false)
	}
	return renderInfoBox("MATCH", "No album matches", "", m.panelWidth(), false)
}

func manualResultsMaxRows(m model) int {
	if m.height <= 0 {
		return 32
	}

	// Do not estimate this screen with a magic "roughly N rows" budget. The
	// full Manual header can grow when Now Playing and the resolved artist badge
	// are present, while MATCH can itself wrap for an unusually long result.
	// Count the actual fixed-height pieces and give only the remaining terminal
	// rows to the result list. This prevents a tall result list from scrolling
	// the top of the header out of Ghostty.
	detailRows := len(strings.Split(renderManualResultDetail(m), "\n"))
	footerRows := len(strings.Split(renderFooter(m), "\n"))
	const (
		resultsAttachmentChrome = 4 // list top + three attached RESULTS rows
		bodyTrailingBlank       = 1
	)
	reserved := detailRows + resultsAttachmentChrome + bodyTrailingBlank + footerRows
	if m.err != nil {
		reserved++
	}
	available := m.height - m.headerHeight() - reserved
	return maxInt(1, minInt(32, available))
}

func renderSimpleAlbumList(albums []lastfm.Album, cursor, maxRows int, hoverRegion, hoverPrefix string, totalWidth int) string {
	rows := simpleAlbumListRows(albums, cursor, maxRows, hoverRegion, hoverPrefix, totalWidth)
	if len(rows) == 0 {
		return renderPanelBox([]string{theme.MutedStyle.Render("No albums found.")}, totalWidth, theme.BorderStyle)
	}
	return renderPanelBox(rows, totalWidth, theme.BorderStyle)
}

func renderSimpleAlbumListWithResults(albums []lastfm.Album, cursor, maxRows int, hoverRegion, hoverPrefix string, totalWidth int) string {
	rows := simpleAlbumListRows(albums, cursor, maxRows, hoverRegion, hoverPrefix, totalWidth)
	if len(rows) == 0 {
		rows = []string{theme.MutedStyle.Render("No albums found.")}
	}
	return renderPanelBoxWithBadgeAttachment(rows, totalWidth, renderCountContent("RESULTS", len(albums)), theme.BorderStyle)
}

func simpleAlbumListRows(albums []lastfm.Album, cursor, maxRows int, hoverRegion, hoverPrefix string, totalWidth int) []string {
	contentWidth := totalWidth - 4
	if len(albums) == 0 {
		return nil
	}
	cursor = minInt(maxInt(cursor, 0), len(albums)-1)
	start := visibleStart(cursor, len(albums), maxRows)
	end := minInt(len(albums), start+maxRows)
	visible := end - start
	thumb := scrollbarThumb(start, len(albums), visible)
	rows := make([]string, 0, visible)
	for row, index := 0, start; index < end; index, row = index+1, row+1 {
		focused := index == cursor
		hovered := hoverRegion == hoverPrefix+":"+fmt.Sprintf("%d", index)
		mark := "  "
		if focused {
			mark = theme.PromptStyle.Render("❯ ")
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
		textStyle := theme.PrimaryTextStyle
		if focused {
			textStyle = theme.FocusedRowLabelStyle
		} else if hovered {
			textStyle = theme.HoverRowLabelStyle
		}
		left := mark + textStyle.Render(text)
		gap := contentWidth - lipgloss.Width(left) - 1
		if gap < 1 {
			gap = 1
		}
		rows = append(rows, left+strings.Repeat(" ", gap)+scroll)
	}
	return rows
}

func renderPreviewView(m model) string {
	record := m.queueRecord("preview")
	albumCount := uniqueAlbumCount(record.Queue)
	trackCount := len(record.Queue)
	estimated := time.Duration(maxInt(0, trackCount-1)) * m.interval
	albumBox := renderPreviewQueueBox(m)
	preview := renderPreviewSummaryCards(
		fmt.Sprintf("%d", albumCount),
		fmt.Sprintf("%d", m.previewTrackCount()),
		m.interval.String(),
		fmt.Sprintf("%d", trackCount),
		formatDuration(estimated),
		previewLoopText(m),
	)
	lines := []string{m.centerToApp(albumBox), m.centerToApp(preview), ""}
	if m.searching {
		lines = append(lines, m.centerToApp(m.spinner.View()+" "+theme.MutedStyle.Render("Preparing authenticated session…")), "")
	}
	status := firstNonEmptyString(m.previewStatus, m.exportStatus)
	if status != "" {
		lines = append(lines, m.centerToApp(theme.SuccessStyle.Render(status)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

type previewAlbum struct {
	Artist string
	Title  string
}

func renderPreviewQueueBox(m model) string {
	const maxAlbums = 5
	totalWidth := m.panelWidth()
	contentWidth := totalWidth - 4
	prefixPlain := "QUEUE ❯ "
	prefixWidth := lipgloss.Width(prefixPlain)
	albums := previewAlbums(m)
	if len(albums) == 0 {
		return renderInfoBox("QUEUE", "Empty queue", "", totalWidth, false)
	}

	shown := minInt(len(albums), maxAlbums)
	rows := make([]string, 0, shown*2+1)
	for index := 0; index < shown; index++ {
		album := albums[index]
		value := album.Title
		if m.modeChoice != "manual" && m.modeChoice != "discography" {
			value = strings.TrimSpace(album.Artist + " — " + album.Title)
		}
		order := fmt.Sprintf("%d", index+1)
		available := maxInt(1, contentWidth-prefixWidth)
		wrapped := wrapWithLastReserve(value, available, lipgloss.Width(order)+1)
		if len(wrapped) == 0 {
			wrapped = []string{"(untitled album)"}
		}
		for lineIndex, line := range wrapped {
			prefix := strings.Repeat(" ", prefixWidth)
			if index == 0 && lineIndex == 0 {
				prefix = theme.SummaryLabelStyle.Render("QUEUE ") + theme.SummaryArrowStyle.Render("❯ ")
			}
			content := prefix + theme.PrimaryTextStyle.Render(line)
			if lineIndex == len(wrapped)-1 {
				gap := contentWidth - lipgloss.Width(content) - lipgloss.Width(order)
				if gap < 1 {
					gap = 1
				}
				content += strings.Repeat(" ", gap) + theme.SummaryValueStyle.Render(order)
			}
			rows = append(rows, fitStyled(content, contentWidth))
		}
	}
	if len(albums) > shown {
		more := fmt.Sprintf("… %d more", len(albums)-shown)
		rows = append(rows, strings.Repeat(" ", prefixWidth)+theme.SecondaryTextStyle.Render(more))
	}
	return renderPanelBox(rows, totalWidth, theme.BorderStyle)
}

func previewAlbums(m model) []previewAlbum {
	seen := map[string]bool{}
	albums := make([]previewAlbum, 0)
	for _, item := range m.scrobbleQueue {
		key := strings.ToLower(item.Artist + "\x00" + item.Album)
		if seen[key] {
			continue
		}
		seen[key] = true
		albums = append(albums, previewAlbum{Artist: item.Artist, Title: item.Album})
	}
	if len(albums) > 0 {
		return albums
	}
	for _, album := range m.selectedAlbums {
		key := strings.ToLower(album.Artist + "\x00" + album.Title)
		if seen[key] {
			continue
		}
		seen[key] = true
		albums = append(albums, previewAlbum{Artist: album.Artist, Title: album.Title})
	}
	return albums
}

func renderPreviewSummaryCards(albums, tracks, interval, scrobbles, eta, loop string) string {
	left := renderSummaryCard(17, [][2]string{{"ALBUMS", albums}, {"TRACKS", tracks}})
	middle := renderSummaryCard(21, [][2]string{{"INTERVAL", interval}, {"SCROBBLES", scrobbles}})
	right := renderSummaryCard(17, [][2]string{{"ETA", eta}, {"LOOP", loop}})
	return joinThreeLineBoxes([]string{left, middle, right}, " ")
}

func renderSummaryCard(totalWidth int, rows [][2]string) string {
	contentWidth := totalWidth - 4
	labelWidth := 0
	for _, row := range rows {
		labelWidth = maxInt(labelWidth, lipgloss.Width(row[0]))
	}
	lines := make([]string, 0, len(rows))
	for _, row := range rows {
		lines = append(lines, summaryLine(row[0], row[1], contentWidth, labelWidth))
	}
	return renderPanelBox(lines, totalWidth, theme.BorderStyle)
}

func summaryLine(label, value string, width, labelWidth int) string {
	left := theme.SecondaryTextStyle.Render(padRight(label, labelWidth)+" ") + theme.SummaryArrowStyle.Render("❯")
	gap := 2
	available := width - lipgloss.Width(left) - gap
	if available < lipgloss.Width(value) {
		gap = 1
		available = width - lipgloss.Width(left) - gap
	}
	if available < 1 {
		available = 1
	}
	right := theme.SummaryValueStyle.Render(truncateToWidth(value, available))
	return left + strings.Repeat(" ", gap) + right
}

func statLine(label, value string, width int) string {
	left := theme.SummaryLabelStyle.Render(label+" ") + theme.SummaryArrowStyle.Render("❯")
	right := theme.SummaryValueStyle.Render(value)
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

func uniqueAlbumCount(queue []sessionstore.Track) int {
	seen := map[string]bool{}
	for _, track := range queue {
		seen[strings.ToLower(track.Artist+"\x00"+track.Album)] = true
	}
	return len(seen)
}

func renderHistoryView(m model) string {
	totalWidth := m.panelWidth()
	contentWidth := totalWidth - 4
	if len(m.history) == 0 {
		return lipgloss.JoinVertical(lipgloss.Left,
			m.centerToApp(renderPanelBox([]string{theme.MutedStyle.Render("No completed sessions yet.")}, totalWidth, theme.BorderStyle)),
			"",
		)
	}
	maxRows := historyMaxRows(m)
	cursor := minInt(m.historyCursor, len(m.history)-1)
	start := visibleStart(cursor, len(m.history), maxRows)
	end := minInt(len(m.history), start+maxRows)
	visible := end - start
	thumb := scrollbarThumb(start, len(m.history), visible)
	rows := make([]string, 0, visible)
	for row, index := 0, start; index < end; index, row = index+1, row+1 {
		record := m.history[index]
		focused := m.settingsFocus == settingsFocusContent && index == cursor
		hovered := m.hoverRegion == "history:"+fmt.Sprintf("%d", index)
		mark := "  "
		if focused {
			mark = theme.PromptStyle.Render("❯ ")
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
		labelStyle := theme.PrimaryTextStyle
		statusStyle := theme.SecondaryTextStyle
		if focused {
			labelStyle = theme.FocusedRowLabelStyle
			statusStyle = theme.FocusedRowValueStyle
		} else if hovered {
			labelStyle = theme.HoverRowLabelStyle
			statusStyle = theme.HoverRowValueStyle
		}
		left := mark + labelStyle.Render(label)
		right := statusStyle.Render(status)
		gap := contentWidth - lipgloss.Width(left) - lipgloss.Width(right) - 1
		if gap < 1 {
			gap = 1
		}
		rows = append(rows, left+strings.Repeat(" ", gap)+right+scroll)
	}
	list := renderPanelBox(rows, totalWidth, theme.BorderStyle)
	record := m.history[cursor]
	detailRows := []string{
		statLine("DATE", record.StartedAt.Local().Format("2006-01-02 15:04"), contentWidth),
		statLine("MODE", record.Mode, contentWidth),
		statLine("TOTAL", fmt.Sprintf("%d", len(record.Queue)), contentWidth),
		statLine("FAILED", fmt.Sprintf("%d", record.Failures), contentWidth),
	}
	detail := renderPanelBox(detailRows, totalWidth, theme.BorderStyle)
	lines := []string{m.centerToApp(list), m.centerToApp(detail), ""}
	if m.historyStatus != "" {
		lines = append(lines, m.centerToApp(theme.SuccessStyle.Render(m.historyStatus)), "")
	}
	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}

func historyMaxRows(m model) int {
	// Settings grid (6), list/detail chrome, blank spacing and the two-line
	// History footer leave 21 non-list rows below the header.
	return m.visibleRows(21, 6, 32)
}

func lastSessionRecord(m model) (sessionstore.Record, bool) {
	if len(m.lastRecord.Queue) > 0 && m.lastRecord.Status == "complete" {
		return m.lastRecord, true
	}
	for _, record := range m.history {
		if record.Status == "complete" && len(record.Queue) > 0 {
			return record, true
		}
	}
	return sessionstore.Record{}, false
}

func renderLastSessionView(m model) string {
	record, ok := lastSessionRecord(m)
	if !ok {
		box := renderAttachedTitlePanel(
			"L A S T   S E S S I O N",
			[]string{"", centerText(theme.MutedStyle.Render("no previous session available"), m.panelWidth()-4), ""},
			m.panelWidth(),
		)
		return m.centerToApp(box)
	}
	contentWidth := m.panelWidth() - 4
	rows := make([]string, 0, 6)
	first := record.Queue[0]
	rows = append(rows, lastSessionLine("ARTIST", first.Artist, contentWidth))
	if uniqueAlbumCount(record.Queue) > 1 {
		rows = append(rows, lastSessionLine("ALBUMS", fmt.Sprintf("%d", uniqueAlbumCount(record.Queue)), contentWidth))
	} else {
		rows = append(rows, lastSessionLine("ALBUM", first.Album, contentWidth))
	}
	rows = append(rows,
		lastSessionLine("TRACKS", fmt.Sprintf("%d", len(record.Queue)), contentWidth),
		lastSessionLine("LOOP", fmt.Sprintf("%d", maxInt(1, record.Loop)), contentWidth),
		lastSessionLine("INTERVAL", record.Interval.String(), contentWidth),
	)
	return m.centerToApp(renderAttachedTitlePanel("L A S T   S E S S I O N", rows, m.panelWidth()))
}

func lastSessionLine(label, value string, width int) string {
	const labelWidth = 12
	left := theme.SummaryLabelStyle.Render(padRight(label, labelWidth)) + theme.SummaryArrowStyle.Render("❯ ")
	available := maxInt(1, width-lipgloss.Width(left))
	return left + theme.SummaryValueStyle.Render(truncateToWidth(value, available))
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
		return m.centerToApp(renderPanelBox([]string{theme.MutedStyle.Render("No unfinished session found.")}, m.panelWidth(), theme.BorderStyle))
	}
	record := *m.pending
	album := historyLabel(record)
	totalWidth := m.panelWidth()
	contentWidth := totalWidth - 4
	info := renderInfoBox("SESSION", album, fmt.Sprintf("%d / %d", record.Completed, len(record.Queue)), totalWidth, false)
	rows := []string{
		statLine("STATUS", "unfinished session found", contentWidth),
		statLine("STARTED", record.StartedAt.Local().Format("2006-01-02 15:04"), contentWidth),
		statLine("MODE", record.Mode, contentWidth),
		statLine("REMAINING", fmt.Sprintf("%d", maxInt(0, len(record.Queue)-record.Completed)), contentWidth),
	}
	box := renderPanelBox(rows, totalWidth, theme.BorderStyle)
	return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(info), m.centerToApp(box), "")
}

func renderHelpView(m model) string {
	width := m.helpPanelWidth()
	rowWidth := maxInt(1, width-4)
	rows := []string{
		helpRow("↑ / ↓ / ← / →", "navigate lists, tabs, and Settings sections", rowWidth),
		helpRow("enter", "confirm or continue", rowWidth),
		helpRow("space", "toggle a selected album or track", rowWidth),
		helpRow("a", "select all / select none", rowWidth),
		helpRow("- / +", "change the current album loop during track selection", rowWidth),
		helpRow("/", "filter long Discography lists", rowWidth),
		helpRow("c", "clean duplicate editions in Discography lists", rowWidth),
		helpRow("s", "Settings on Dashboard; similar/sort elsewhere", rowWidth),
		helpRow("e", "export the current queue or history entry", rowWidth),
		helpRow("h", "session history", rowWidth),
		helpRow("tab", "switch Settings section/content focus", rowWidth),
		helpRow("esc", "go back or cancel", rowWidth),
		helpRow("q", "quit; active sessions can resume later", rowWidth),
	}
	box := renderPanelBox(rows, width, theme.BorderStyle)
	closeStyle := theme.MutedStyle
	if m.hoverRegion == "help:close" {
		closeStyle = theme.SuccessStyle
	}
	closeLine := theme.MutedStyle.Render("press ? or esc to ") + closeStyle.Render("close")
	return lipgloss.JoinVertical(lipgloss.Left, m.centerToApp(box), m.centerToApp(closeLine), "")
}

func helpRow(key, description string, widths ...int) string {
	left := theme.KeyStyle.Render(key)
	right := theme.MutedStyle.Render(description)
	width := 59
	if len(widths) > 0 {
		width = maxInt(1, widths[0])
	}
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
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
