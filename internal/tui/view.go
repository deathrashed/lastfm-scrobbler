package tui

import (
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

const headerContentWidth = 67

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)

func (m model) View() string {
	if m.width > 0 && m.width < 40 {
		return theme.ErrorStyle.Render("Terminal too narrow (need ≥40 cols)")
	}

	var body string
	if m.helpVisible {
		body = renderHelpView(m)
	} else {
		switch m.stage {
		case stageInput:
			body = renderInputView(m)
		case stageImportSource:
			body = renderImportSourceView(m)
		case stageSearch:
			body = renderSearchView(m)
		case stageResults:
			body = renderResultsView(m)
		case stageDiscographySelect:
			body = renderDiscographySelectView(m)
		case stageTrackSelect:
			body = renderTrackSelectView(m)
		case stagePreview:
			body = renderPreviewView(m)
		case stageConfig:
			body = renderConfigView(m)
		case stageAdvancedConfig:
			body = renderAdvancedConfigView(m)
		case stageEnvPath:
			body = renderEnvPathView(m)
		case stageScrobbling:
			body = renderScrobblingView(m)
		case stageDone:
			body = renderDoneView(m)
		case stageHistory:
			body = renderHistoryView(m)
		case stageRecovery:
			body = renderRecoveryView(m)
		case stageSimilarSelect:
			body = renderSimilarSelectView(m)
		case stageProfiles:
			body = renderProfilesView(m)
		case stageProfileName:
			body = renderProfileNameView(m)
		case stageInfo:
			body = renderInfoView(m)
		case stageConnectionTest:
			body = renderConnectionTestView(m)
		case stageDiagnostics:
			body = renderDiagnosticsView(m)
		case stageUpdateCheck:
			body = renderUpdateCheckView(m)
		}
	}

	parts := []string{
		RenderHeaderWithHover(m.width, m.stage, m.modeChoice, m.cfg.Username, m.headerSettingsLine(), m.cfg.CompactHeader, m.headerURLHover),
		body,
	}
	if !m.helpVisible {
		parts = append(parts, centerToHeader(renderFooter(m)))
	}
	if m.err != nil {
		parts = append(parts, centerToHeader(theme.ErrorStyle.Render(fmt.Sprintf("%s %s", theme.IconError, m.err.Error()))))
	}
	return strings.Join(parts, "\n")
}

func hintKey(value string) string  { return theme.KeyStyle.Render(value) }
func hintSep() string              { return theme.SepStyle.Render(" • ") }
func hintText(value string) string { return theme.MutedStyle.Render(value) }
func hint(parts ...string) string  { return strings.Join(parts, "") }

func renderFooter(m model) string {
	switch m.stage {
	case stageInput:
		lineOne := hint(hintKey("→/↑/↓/←"), hintText(" navigate"), hintSep(), hintKey("enter"), hintText(" select"), hintSep(), hintKey("M-D-F"), hintText(" quick"))
		lineTwo := hint(hintKey("h"), hintText(" history"), hintSep(), hintKey("p"), hintText(" profiles"), hintSep(), hintKey("c"), hintText(" config"), hintSep(), hintKey("i"), hintText(" info"), hintSep(), hintKey("?"), hintText(" help"))
		return lineOne + "\n" + lineTwo
	case stageImportSource:
		return hint(hintKey("→/↑/↓/←"), hintText(" navigate"), hintSep(), hintKey("enter"), hintText(" choose"), hintSep(), hintKey("o"), hintText(" picker"), hintSep(), hintKey("esc"), hintText(" menu"), hintSep(), hintKey("?"), hintText(" help"))
	case stageSearch:
		extra := ""
		if m.modeChoice == "file" {
			extra = hint(hintSep(), hintKey("o"), hintText(" picker"))
		}
		return hint(hintKey("enter"), hintText(" continue"), extra, hintSep(), hintKey("esc"), hintText(" back"), hintSep(), hintKey("ctrl+c"), hintText(" quit"))
	case stageResults:
		return hint(hintKey("↑/↓"), hintText(" navigate"), hintSep(), hintKey("enter"), hintText(" select"), hintSep(), hintKey("s"), hintText(" similar"), hintSep(), hintKey("esc"), hintText(" back"))
	case stageDiscographySelect:
		if m.discographyFiltering {
			return hint(hintKey("type"), hintText(" filter"), hintSep(), hintKey("enter"), hintText(" apply"), hintSep(), hintKey("esc"), hintText(" cancel"))
		}
		return hint(hintKey("↑/↓"), hintText(" navigate"), hintSep(), hintKey("space"), hintText(" check"), hintSep(), hintKey("a"), hintText(" all"), hintSep(), hintKey("c"), hintText(" clean"), hintSep(), hintKey("/"), hintText(" filter"), hintSep(), hintKey("s"), hintText(" sort"), hintSep(), hintKey("enter"), hintText(" continue"))
	case stageTrackSelect:
		return hint(hintKey("↑/↓"), hintText(" navigate"), hintSep(), hintKey("space"), hintText(" check"), hintSep(), hintKey("a"), hintText(" all"), hintSep(), hintKey("-/+"), hintText(" loop"), hintSep(), hintKey("[/]"), hintText(" album loop"), hintSep(), hintKey("enter"), hintText(" preview"), hintSep(), hintKey("s"), hintText(" similar"))
	case stagePreview:
		return hint(hintKey("enter"), hintText(" start"), hintSep(), hintKey("e"), hintText(" export"), hintSep(), hintKey("s"), hintText(" similar"), hintSep(), hintKey("esc"), hintText(" edit"), hintSep(), hintKey("?"), hintText(" help"))
	case stageConfig:
		action := "save"
		if m.configIndex >= 4 {
			action = "open"
		}
		lineOne := hint(hintKey("→/↑/↓/←"), hintText(" navigate"), hintSep(), hintKey("enter"), hintText(" "+action), hintSep(), hintKey("tab"), hintText(" field"), hintSep(), hintKey("ctrl+p"), hintText(" credentials path"))
		lineTwo := hint(hintKey("ctrl+g"), hintText(" advanced"), hintSep(), hintKey("ctrl+o"), hintText(" info"), hintSep(), hintKey("esc"), hintText(" back"))
		return lineOne + "\n" + lineTwo
	case stageAdvancedConfig:
		action := "save"
		if advancedAction(m.advancedIndex) {
			action = "open"
		}
		return hint(hintKey("↑/↓"), hintText(" navigate"), hintSep(), hintKey("←/→"), hintText(" toggle"), hintSep(), hintKey("enter"), hintText(" "+action), hintSep(), hintKey("o"), hintText(" folder"), hintSep(), hintKey("esc"), hintText(" back"))
	case stageEnvPath:
		return hint(hintKey("enter"), hintText(" save"), hintSep(), hintKey("o"), hintText(" picker"), hintSep(), hintKey("esc"), hintText(" back"))
	case stageScrobbling:
		return hint(hintKey("esc"), hintText(" cancel"), hintSep(), hintKey("q"), hintText(" quit + resume later"))
	case stageDone:
		lineOne := hint(hintKey("enter"), hintText(" another"), hintSep(), hintKey("r"), hintText(" edit + re-run"), hintSep(), hintKey("R"), hintText(" exact re-run"), hintSep(), hintKey("e"), hintText(" export"))
		lineTwo := hint(hintKey("s"), hintText(" similar"), hintSep(), hintKey("h"), hintText(" history"), hintSep(), hintKey("esc"), hintText(" menu"), hintSep(), hintKey("q"), hintText(" quit"))
		return lineOne + "\n" + lineTwo
	case stageHistory:
		return hint(hintKey("↑/↓"), hintText(" navigate"), hintSep(), hintKey("enter/r"), hintText(" edit + re-run"), hintSep(), hintKey("R"), hintText(" exact re-run"), hintSep(), hintKey("e"), hintText(" export"), hintSep(), hintKey("d"), hintText(" delete"), hintSep(), hintKey("esc"), hintText(" menu"))
	case stageRecovery:
		return hint(hintKey("enter"), hintText(" resume"), hintSep(), hintKey("r"), hintText(" restart"), hintSep(), hintKey("d"), hintText(" discard"), hintSep(), hintKey("q"), hintText(" quit"))
	case stageSimilarSelect:
		return hint(hintKey("↑/↓"), hintText(" navigate"), hintSep(), hintKey("enter"), hintText(" load"), hintSep(), hintKey("esc"), hintText(" back"))
	case stageProfiles:
		return hint(hintKey("↑/↓"), hintText(" navigate"), hintSep(), hintKey("enter"), hintText(" load"), hintSep(), hintKey("n"), hintText(" new"), hintSep(), hintKey("s"), hintText(" save"), hintSep(), hintKey("d"), hintText(" delete"), hintSep(), hintKey("esc"), hintText(" menu"))
	case stageProfileName:
		return hint(hintKey("enter"), hintText(" create"), hintSep(), hintKey("esc"), hintText(" back"))
	case stageInfo:
		return hint(hintKey("←/→"), hintText(" section"), hintSep(), hintKey("esc"), hintText(" back"), hintSep(), hintKey("?"), hintText(" quick help"), hintSep(), hintKey("q"), hintText(" quit"))
	case stageConnectionTest:
		return hint(hintKey("r"), hintText(" re-test"), hintSep(), hintKey("esc"), hintText(" advanced"), hintSep(), hintKey("q"), hintText(" quit"))
	case stageDiagnostics:
		return hint(hintKey("enter"), hintText(" export"), hintSep(), hintKey("o"), hintText(" open folder"), hintSep(), hintKey("esc"), hintText(" advanced"), hintSep(), hintKey("q"), hintText(" quit"))
	case stageUpdateCheck:
		return hint(hintKey("r"), hintText(" check again"), hintSep(), hintKey("o"), hintText(" open release"), hintSep(), hintKey("esc"), hintText(" advanced"), hintSep(), hintKey("q"), hintText(" quit"))
	}
	return ""
}

func centerToHeader(value string) string {
	if value == "" {
		return ""
	}
	lines := strings.Split(value, "\n")
	for index, line := range lines {
		lines[index] = centerText(line, headerContentWidth)
	}
	return strings.Join(lines, "\n")
}

func (m model) headerSettingsLine() string {
	switch m.stage {
	case stageTrackSelect, stagePreview, stageScrobbling, stageDone:
		loop := fmt.Sprintf("loop %d", m.loopCount)
		if m.mixedLoops() {
			loop = "loop mixed"
		}
		return fmt.Sprintf("%s|interval %s|limit %s", loop, m.interval.String(), m.limitLabel())
	default:
		return ""
	}
}

func renderExactBox(label string, totalWidth int, selected bool) string {
	border := theme.BorderStyle
	if selected {
		border = theme.InnerBorderStyle
	}
	innerWidth := maxInt(1, totalWidth-2)
	labelStyle := theme.ModeStyle
	mnemonicStyle := theme.MnemonicStyle
	if selected {
		labelStyle = theme.SelectedModeStyle
		mnemonicStyle = theme.SelectedMnemonicStyle
	}
	labelText := label
	if len(label) > 0 {
		labelText = mnemonicStyle.Render(label[:1]) + labelStyle.Render(label[1:])
	}
	return strings.Join([]string{
		border.Render("╭" + strings.Repeat("─", innerWidth) + "╮"),
		border.Render("│") + centerText(labelText, innerWidth) + border.Render("│"),
		border.Render("╰" + strings.Repeat("─", innerWidth) + "╯"),
	}, "\n")
}

func renderTextBox(label, value, placeholder string, totalWidth int, active bool) string {
	border := theme.BorderStyle
	if active {
		border = theme.InnerBorderStyle
	}
	innerWidth := maxInt(4, totalWidth-2)
	contentWidth := maxInt(1, innerWidth-2)
	shown := value
	if strings.TrimSpace(stripANSI(value)) == "" {
		shown = theme.MutedStyle.Render(placeholder)
	}
	prefix := theme.KeyStyle.Render(label + " ❯ ")
	available := maxInt(1, contentWidth-lipgloss.Width(prefix))
	shown = fitStyled(shown, available)
	return renderPanelBox([]string{prefix + shown}, totalWidth, border)
}

func renderInfoBox(label, value, right string, totalWidth int, active bool) string {
	border := theme.BorderStyle
	if active {
		border = theme.InnerBorderStyle
	}
	innerWidth := maxInt(4, totalWidth-2)
	contentWidth := maxInt(1, innerWidth-2)
	prefixPlain := label + " ❯ "
	prefixWidth := lipgloss.Width(prefixPlain)
	valueWidth := maxInt(1, contentWidth-prefixWidth)
	reserve := 0
	if right != "" {
		reserve = lipgloss.Width(right) + 1
	}
	wrapped := wrapWithLastReserve(strings.TrimSpace(value), valueWidth, reserve)
	if len(wrapped) == 0 {
		wrapped = []string{""}
	}
	lines := make([]string, 0, len(wrapped))
	for index, line := range wrapped {
		var content string
		if index == 0 {
			content = theme.KeyStyle.Render(prefixPlain) + theme.AlbumStyle.Render(line)
		} else {
			content = strings.Repeat(" ", prefixWidth) + theme.AlbumStyle.Render(line)
		}
		if index == len(wrapped)-1 && right != "" {
			remaining := contentWidth - lipgloss.Width(content) - lipgloss.Width(right)
			if remaining < 1 {
				remaining = 1
			}
			content += strings.Repeat(" ", remaining) + theme.MutedStyle.Render(right)
		}
		lines = append(lines, fitStyled(content, contentWidth))
	}
	return renderPanelBox(lines, totalWidth, border)
}

func renderPanelBox(lines []string, totalWidth int, border lipgloss.Style) string {
	innerWidth := maxInt(4, totalWidth-2)
	contentWidth := maxInt(1, innerWidth-2)
	out := []string{border.Render("╭" + strings.Repeat("─", innerWidth) + "╮")}
	for _, line := range lines {
		line = fitStyled(line, contentWidth)
		out = append(out, border.Render("│")+" "+padRight(line, contentWidth)+" "+border.Render("│"))
	}
	out = append(out, border.Render("╰"+strings.Repeat("─", innerWidth)+"╯"))
	return strings.Join(out, "\n")
}

func fitStyled(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}
	return truncateStyled(value, width)
}

// truncateStyled clips a styled terminal string without counting ANSI control
// sequences as visible cells. This is deliberately used by every bordered
// component so a focused textinput can never push the right border outward.
func truncateStyled(value string, width int) string {
	if width <= 0 {
		return ""
	}
	var out strings.Builder
	visible := 0
	for index := 0; index < len(value); {
		if value[index] == '\x1b' {
			end := ansiSequenceEnd(value, index)
			out.WriteString(value[index:end])
			index = end
			continue
		}
		r, size := utf8.DecodeRuneInString(value[index:])
		if r == '\n' || r == '\r' {
			break
		}
		runeWidth := lipgloss.Width(string(r))
		if visible+runeWidth > width {
			break
		}
		out.WriteRune(r)
		visible += runeWidth
		index += size
	}
	// A reset prevents a clipped active style from leaking into padding or the
	// closing border. It has zero display width.
	out.WriteString("\x1b[0m")
	return out.String()
}

func ansiSequenceEnd(value string, start int) int {
	if start+1 >= len(value) {
		return len(value)
	}
	// CSI: ESC [ ... final-byte
	if value[start+1] == '[' {
		for index := start + 2; index < len(value); index++ {
			if value[index] >= 0x40 && value[index] <= 0x7e {
				return index + 1
			}
		}
		return len(value)
	}
	// OSC: ESC ] ... BEL or ST
	if value[start+1] == ']' {
		for index := start + 2; index < len(value); index++ {
			if value[index] == '\a' {
				return index + 1
			}
			if value[index] == '\x1b' && index+1 < len(value) && value[index+1] == '\\' {
				return index + 2
			}
		}
		return len(value)
	}
	return minInt(len(value), start+2)
}

func renderSelectedBadge(selected, total int) string {
	text := fmt.Sprintf("  SELECTED ❯   %d / %d  ", selected, total)
	width := maxInt(26, lipgloss.Width(text)+2)
	return renderExactBox(text, width, false)
}

func wrapWithLastReserve(text string, width, reserve int) []string {
	lines := wrapWords(text, width)
	if len(lines) == 0 || reserve <= 0 {
		return lines
	}
	lastWidth := maxInt(1, width-reserve)
	last := lines[len(lines)-1]
	if lipgloss.Width(last) <= lastWidth {
		return lines
	}
	rewrapped := wrapWords(last, lastWidth)
	return append(lines[:len(lines)-1], rewrapped...)
}

func wrapWords(text string, width int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if width < 1 {
		width = 1
	}
	words := strings.Fields(text)
	var lines []string
	current := ""
	for _, word := range words {
		if lipgloss.Width(word) > width {
			if current != "" {
				lines = append(lines, current)
				current = ""
			}
			parts := splitByWidth(word, width)
			if len(parts) > 1 {
				lines = append(lines, parts[:len(parts)-1]...)
			}
			current = parts[len(parts)-1]
			continue
		}
		candidate := word
		if current != "" {
			candidate = current + " " + word
		}
		if lipgloss.Width(candidate) <= width {
			current = candidate
		} else {
			lines = append(lines, current)
			current = word
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitByWidth(text string, width int) []string {
	var out []string
	current := ""
	for len(text) > 0 {
		r, size := utf8.DecodeRuneInString(text)
		text = text[size:]
		candidate := current + string(r)
		if current != "" && lipgloss.Width(candidate) > width {
			out = append(out, current)
			current = string(r)
		} else {
			current = candidate
		}
	}
	if current != "" {
		out = append(out, current)
	}
	return out
}

func truncateToWidth(text string, width int) string {
	if lipgloss.Width(text) <= width {
		return text
	}
	if width <= 1 {
		return "…"
	}
	parts := splitByWidth(text, width-1)
	if len(parts) == 0 {
		return "…"
	}
	return parts[0] + "…"
}

func stripANSI(value string) string { return ansiPattern.ReplaceAllString(value, "") }

func albumInfoLine(album lastfm.Album) string {
	return album.Artist + " — " + album.Title + fmt.Sprintf(" (%d tracks)", len(album.Tracks))
}
