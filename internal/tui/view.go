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

const headerContentWidth = minAppWidth

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;?]*[ -/]*[@-~]`)
var oscPattern = regexp.MustCompile(`\x1b\][^\x1b]*(\x07|\x1b\\)`)

func (m model) View() string {
	if m.width > 0 && m.width < minAppWidth {
		return theme.ErrorStyle.Render("Terminal too narrow\nLast.fm Scrobbler requires at least 67 columns.")
	}

	body := m.renderBody()

	parts := []string{
		m.renderHeader(),
		body,
	}
	if !m.helpVisible {
		parts = append(parts, m.centerToApp(renderFooter(m)))
	}
	if m.err != nil {
		parts = append(parts, m.centerToApp(theme.ErrorStyle.Render(fmt.Sprintf("%s %s", theme.IconError, m.err.Error()))))
	}
	view := strings.Join(parts, "\n")
	if offset := m.appX(); offset > 0 {
		prefix := strings.Repeat(" ", offset)
		lines := strings.Split(view, "\n")
		for index, line := range lines {
			lines[index] = prefix + line
		}
		view = strings.Join(lines, "\n")
	}
	return view
}

func (m model) renderBody() string {
	if m.helpVisible {
		return renderHelpView(m)
	}
	var body string
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
		body = renderSettingsView(m)
	case stageEnvPath:
		body = renderEnvPathView(m)
	case stageScrobbling:
		body = renderScrobblingView(m)
	case stageDone:
		body = renderDoneView(m)
	case stageHistory:
		body = renderSettingsShell(m, renderHistoryView(m))
	case stageLastSession:
		body = renderLastSessionView(m)
	case stageRecovery:
		body = renderRecoveryView(m)
	case stageSimilarSelect:
		body = renderSimilarSelectView(m)
	case stageProfiles:
		body = renderSettingsShell(m, renderProfilesView(m))
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
	case stageCompletions:
		body = renderCompletionsView(m)
	case stageAuth:
		body = m.renderAuthView()
	case stageSetup:
		body = renderSetupView(m)
	}
	return body
}

func hintKey(value string) string  { return theme.KeyStyle.Render(value) }
func hintSep() string              { return theme.SepStyle.Render(" • ") }
func hintText(value string) string { return theme.MutedStyle.Render(value) }
func hint(parts ...string) string  { return strings.Join(parts, "") }

func renderFooter(m model) string {
	spec := footerSpec(m)
	// Footer geometry is deliberately stable: action rows never change height
	// when the mouse enters or leaves an item. Two reserved help rows sit below
	// the controls on every screen, so hover descriptions cannot displace or
	// hide a second action row.
	lines := make([]string, 0, len(spec)+2)
	for _, items := range spec {
		hoverGroup := ""
		for _, item := range items {
			if item.interactive && item.id == m.hoverRegion {
				hoverGroup = item.group
				break
			}
		}
		parts := make([]string, 0, len(items)*2-1)
		for index, item := range items {
			if index > 0 && !item.tight {
				parts = append(parts, hintSep())
			}

			key := hintKey(item.key)
			label := hintText(item.label)
			if item.group != "" && hoverGroup == item.group {
				if item.interactive && item.id == m.hoverRegion {
					key = theme.SelectedModeStyle.Render(item.key)
				} else if !item.interactive {
					label = theme.AccentTextStyle.Render(item.label)
				}
			} else if item.interactive && item.id == m.hoverRegion {
				label = theme.PrimaryTextStyle.Render(item.label)
			}
			parts = append(parts, hint(key, label))
		}
		lines = append(lines, strings.Join(parts, ""))
	}

	description, accent := footerHoverDescription(m, spec)
	switch footerDetailRows(m) {
	case 0:
		// Short terminals keep only actionable controls. Hover details remain
		// available again as soon as there is enough vertical room.
	case 1:
		parts := make([]string, 0, 2)
		if description != "" {
			parts = append(parts, theme.PrimaryTextStyle.Render(description))
		}
		if accent != "" {
			parts = append(parts, theme.AccentTextStyle.Render(accent))
		}
		line := strings.Join(parts, theme.MutedStyle.Render("  •  "))
		lines = append(lines, fitStyled(line, m.contentWidth()))
	default:
		descriptionLine := ""
		accentLine := ""
		if description != "" {
			descriptionLine = theme.PrimaryTextStyle.Render(truncateToWidth(description, m.contentWidth()))
		}
		if accent != "" {
			accentLine = theme.AccentTextStyle.Render(truncateToWidth(accent, m.contentWidth()))
		}
		lines = append(lines, descriptionLine, accentLine)
	}
	return strings.Join(lines, "\n")
}

func footerDetailRows(m model) int {
	if m.height > 0 {
		switch {
		case m.height <= 24:
			return 0
		case m.height <= 32:
			return 1
		}
	}
	return 2
}

func footerHoverDescription(m model, spec [][]footerItem) (string, string) {
	if !strings.HasPrefix(m.hoverRegion, "footer:") {
		return "", ""
	}
	for _, items := range spec {
		for _, item := range items {
			if item.id == m.hoverRegion {
				return item.description, item.descriptionAccent
			}
		}
	}
	return "", ""
}

func centerToHeader(value string) string {
	return centerToWidth(value, minAppWidth)
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
	return renderExactBoxWithMnemonic(label, totalWidth, selected, false)
}

func renderDashboardBox(label string, totalWidth int, selected bool) string {
	return renderExactBoxWithMnemonicHover(label, totalWidth, selected, true, false)
}

func renderExactBoxWithMnemonic(label string, totalWidth int, selected, mnemonic bool) string {
	return renderExactBoxWithMnemonicHover(label, totalWidth, selected, mnemonic, false)
}

func renderExactBoxWithMnemonicHover(label string, totalWidth int, selected, mnemonic, hovered bool) string {
	border := theme.BorderStyle
	if selected || hovered {
		border = theme.InnerBorderStyle
	}
	innerWidth := maxInt(1, totalWidth-2)
	labelStyle := theme.ModeStyle
	mnemonicStyle := labelStyle
	if selected || hovered {
		labelStyle = theme.SelectedModeStyle
		mnemonicStyle = labelStyle
	}
	labelText := label
	if mnemonic && len(label) > 0 {
		if selected {
			mnemonicStyle = theme.SelectedMnemonicStyle
		} else {
			mnemonicStyle = theme.MnemonicStyle
		}
		labelText = mnemonicStyle.Render(label[:1]) + labelStyle.Render(label[1:])
	} else {
		labelText = labelStyle.Render(label)
	}
	return strings.Join([]string{
		border.Render("╭" + strings.Repeat("─", innerWidth) + "╮"),
		border.Render("│") + centerText(labelText, innerWidth) + border.Render("│"),
		border.Render("╰" + strings.Repeat("─", innerWidth) + "╯"),
	}, "\n")
}

// renderChoiceBox is the shared card language for section/source choosers:
// muted when idle, Torch Red text on hover, and a Torch Red border with bold
// white text when selected. Hover never changes keyboard selection.
func renderChoiceBox(label string, totalWidth int, selected, hovered bool) string {
	border := theme.BorderStyle
	labelStyle := theme.SecondaryTextStyle
	switch {
	case selected:
		border = theme.InnerBorderStyle
		labelStyle = theme.SelectedModeStyle
	case hovered:
		labelStyle = theme.AccentTextStyle
	}
	innerWidth := maxInt(1, totalWidth-2)
	return strings.Join([]string{
		border.Render("╭" + strings.Repeat("─", innerWidth) + "╮"),
		border.Render("│") + centerText(labelStyle.Render(label), innerWidth) + border.Render("│"),
		border.Render("╰" + strings.Repeat("─", innerWidth) + "╯"),
	}, "\n")
}

func renderTextBox(label, value, placeholder string, totalWidth int, active bool) string {
	// A red border is reserved for selected navigation/actions and validation
	// errors. Text editing keeps the structural border white and communicates
	// focus through the label/arrow plus the textinput's blinking cursor.
	border := theme.BorderStyle
	innerWidth := maxInt(4, totalWidth-2)
	contentWidth := maxInt(1, innerWidth-2)
	shown := value
	if strings.TrimSpace(stripANSI(value)) == "" {
		shown = theme.MutedStyle.Render(placeholder)
	}
	labelStyle := theme.RowLabelStyle
	arrowStyle := theme.RowArrowStyle
	if active {
		labelStyle = theme.FocusedRowLabelStyle
		arrowStyle = theme.FocusedRowArrowStyle
	}
	prefix := labelStyle.Render(label+" ") + arrowStyle.Render("❯ ")
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
			prefix := theme.SummaryLabelStyle.Render(label+" ") + theme.SummaryArrowStyle.Render("❯ ")
			content = prefix + theme.PrimaryTextStyle.Render(line)
		} else {
			content = strings.Repeat(" ", prefixWidth) + theme.PrimaryTextStyle.Render(line)
		}
		if index == len(wrapped)-1 && right != "" {
			remaining := contentWidth - lipgloss.Width(content) - lipgloss.Width(right)
			if remaining < 1 {
				remaining = 1
			}
			content += strings.Repeat(" ", remaining) + theme.SummaryMetaStyle.Render(right)
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

// renderAttachedTitlePanel keeps the established detached/attached title
// language while giving the body panel its own semantic identity. The title
// is centered mathematically and the panel remains an exact totalWidth cells.
func renderAttachedTitlePanel(title string, lines []string, totalWidth int) string {
	border := theme.BorderStyle
	innerWidth := maxInt(4, totalWidth-2)
	contentWidth := maxInt(1, innerWidth-2)
	titleContent := theme.HeaderTextStyle.Render(title)
	titleInnerWidth := lipgloss.Width(" " + title + " ")
	titleTotalWidth := titleInnerWidth + 2
	if titleTotalWidth > totalWidth-4 {
		available := maxInt(1, totalWidth-8)
		title = truncateToWidth(title, available)
		titleContent = theme.HeaderTextStyle.Render(title)
		titleInnerWidth = lipgloss.Width(" " + title + " ")
		titleTotalWidth = titleInnerWidth + 2
	}
	start := maxInt(1, (totalWidth-titleTotalWidth)/2)
	end := maxInt(1, totalWidth-titleTotalWidth-start)

	topTitle := border.Render("╭" + strings.Repeat("─", titleInnerWidth) + "╮")
	line0 := strings.Repeat(" ", start) + topTitle + strings.Repeat(" ", end)
	line1 := border.Render("╭"+strings.Repeat("─", maxInt(0, start-1))+"┤") +
		" " + titleContent + " " +
		border.Render("├"+strings.Repeat("─", maxInt(0, end-1))+"╮")
	closure := border.Render("╰" + strings.Repeat("─", titleInnerWidth) + "╯")
	line2 := border.Render("│") + strings.Repeat(" ", maxInt(0, start-1)) + closure + strings.Repeat(" ", maxInt(0, end-1)) + border.Render("│")

	out := []string{fitStyled(line0, totalWidth), fitStyled(line1, totalWidth), fitStyled(line2, totalWidth)}
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
	content := renderSelectedContent(selected, total)
	width := maxInt(26, lipgloss.Width(content)+6)
	return renderPanelBox([]string{centerText(content, width-4)}, width, theme.BorderStyle)
}

func renderSelectedContent(selected, total int) string {
	return theme.SummaryLabelStyle.Render("SELECTED ") +
		theme.SummaryArrowStyle.Render("❯") + "   " +
		theme.AccentTextStyle.Render(fmt.Sprintf("%d", selected)) + " " +
		theme.MutedStyle.Render("/") + " " +
		theme.AccentTextStyle.Render(fmt.Sprintf("%d", total))
}

func renderCountContent(label string, value int) string {
	return theme.SummaryLabelStyle.Render(label+" ") +
		theme.SummaryArrowStyle.Render("❯") + "  " +
		theme.AccentTextStyle.Render(fmt.Sprintf("%d", value))
}

func renderPanelBoxWithBadgeAttachment(lines []string, totalWidth int, content string, border lipgloss.Style) string {
	innerWidth := maxInt(4, totalWidth-2)
	contentWidth := maxInt(1, innerWidth-2)
	out := []string{border.Render("╭" + strings.Repeat("─", innerWidth) + "╮")}
	for _, line := range lines {
		line = fitStyled(line, contentWidth)
		out = append(out, border.Render("│")+" "+padRight(line, contentWidth)+" "+border.Render("│"))
	}

	badgeWidth := maxInt(18, lipgloss.Width(content)+4)
	badgeWidth = minInt(badgeWidth, totalWidth-10)
	badgeInner := badgeWidth - 2
	rightTail := 4
	leftDash := totalWidth - 4 - badgeInner - rightTail
	if leftDash < 4 {
		leftDash = 4
		rightTail = maxInt(1, totalWidth-4-badgeInner-leftDash)
	}

	badgeTop := border.Render("╭" + strings.Repeat("─", badgeInner) + "╮")
	topRightSpaces := maxInt(0, innerWidth-leftDash-badgeWidth)
	out = append(out,
		border.Render("│")+strings.Repeat(" ", leftDash)+badgeTop+strings.Repeat(" ", topRightSpaces)+border.Render("│"),
	)

	badgeMiddle := centerText(content, badgeInner)
	out = append(out,
		border.Render("╰"+strings.Repeat("─", leftDash)+"┤")+
			badgeMiddle+
			border.Render("├"+strings.Repeat("─", rightTail)+"╯"),
	)

	badgeBottom := border.Render("╰" + strings.Repeat("─", badgeInner) + "╯")
	bottom := strings.Repeat(" ", leftDash+1) + badgeBottom
	bottom += strings.Repeat(" ", maxInt(0, totalWidth-lipgloss.Width(bottom)))
	out = append(out, bottom)
	return strings.Join(out, "\n")
}

func renderPanelBoxWithTwoBadgeAttachments(lines []string, totalWidth int, leftContent, rightContent string, border lipgloss.Style) string {
	innerWidth := maxInt(4, totalWidth-2)
	contentWidth := maxInt(1, innerWidth-2)
	out := []string{border.Render("╭" + strings.Repeat("─", innerWidth) + "╮")}
	for _, line := range lines {
		line = fitStyled(line, contentWidth)
		out = append(out, border.Render("│")+" "+padRight(line, contentWidth)+" "+border.Render("│"))
	}

	leftWidth := maxInt(18, lipgloss.Width(leftContent)+4)
	rightWidth := maxInt(18, lipgloss.Width(rightContent)+4)
	gap := 1
	groupWidth := leftWidth + gap + rightWidth
	if groupWidth > innerWidth-8 {
		over := groupWidth - (innerWidth - 8)
		trimLeft := over / 2
		trimRight := over - trimLeft
		leftWidth = maxInt(14, leftWidth-trimLeft)
		rightWidth = maxInt(14, rightWidth-trimRight)
		groupWidth = leftWidth + gap + rightWidth
	}
	leftDash := maxInt(4, (totalWidth-2-groupWidth)/2)
	rightDash := maxInt(4, totalWidth-2-groupWidth-leftDash)
	if 1+leftDash+groupWidth+rightDash+1 != totalWidth {
		rightDash = maxInt(1, totalWidth-2-groupWidth-leftDash)
	}

	leftTop := border.Render("╭" + strings.Repeat("─", leftWidth-2) + "╮")
	rightTop := border.Render("╭" + strings.Repeat("─", rightWidth-2) + "╮")
	topGroup := leftTop + " " + rightTop
	insideRight := maxInt(0, innerWidth-leftDash-lipgloss.Width(topGroup))
	out = append(out, border.Render("│")+strings.Repeat(" ", leftDash)+topGroup+strings.Repeat(" ", insideRight)+border.Render("│"))

	leftMiddle := border.Render("┤") + centerText(leftContent, leftWidth-2) + border.Render("├")
	rightMiddle := border.Render("┤") + centerText(rightContent, rightWidth-2) + border.Render("├")
	out = append(out,
		border.Render("╰"+strings.Repeat("─", leftDash))+
			leftMiddle+border.Render("─")+rightMiddle+
			border.Render(strings.Repeat("─", rightDash)+"╯"),
	)

	leftBottom := border.Render("╰" + strings.Repeat("─", leftWidth-2) + "╯")
	rightBottom := border.Render("╰" + strings.Repeat("─", rightWidth-2) + "╯")
	bottom := strings.Repeat(" ", leftDash+1) + leftBottom + " " + rightBottom
	bottom += strings.Repeat(" ", maxInt(0, totalWidth-lipgloss.Width(bottom)))
	out = append(out, bottom)
	return strings.Join(out, "\n")
}

func renderPanelBoxWithSelectedAttachment(lines []string, totalWidth, selected, total int, border lipgloss.Style) string {
	innerWidth := maxInt(4, totalWidth-2)
	contentWidth := maxInt(1, innerWidth-2)
	out := []string{border.Render("╭" + strings.Repeat("─", innerWidth) + "╮")}
	for _, line := range lines {
		line = fitStyled(line, contentWidth)
		out = append(out, border.Render("│")+" "+padRight(line, contentWidth)+" "+border.Render("│"))
	}

	content := renderSelectedContent(selected, total)
	badgeWidth := maxInt(24, lipgloss.Width(content)+4)
	badgeWidth = minInt(badgeWidth, totalWidth-10)
	badgeInner := badgeWidth - 2
	rightTail := 4
	leftDash := totalWidth - 4 - badgeInner - rightTail
	if leftDash < 4 {
		leftDash = 4
		rightTail = maxInt(1, totalWidth-4-badgeInner-leftDash)
	}

	badgeTop := border.Render("╭" + strings.Repeat("─", badgeInner) + "╮")
	topRightSpaces := maxInt(0, innerWidth-leftDash-badgeWidth)
	out = append(out,
		border.Render("│")+strings.Repeat(" ", leftDash)+badgeTop+strings.Repeat(" ", topRightSpaces)+border.Render("│"),
	)

	badgeMiddle := centerText(content, badgeInner)
	out = append(out,
		border.Render("╰"+strings.Repeat("─", leftDash)+"┤")+
			badgeMiddle+
			border.Render("├"+strings.Repeat("─", rightTail)+"╯"),
	)

	badgeBottom := border.Render("╰" + strings.Repeat("─", badgeInner) + "╯")
	bottom := strings.Repeat(" ", leftDash+1) + badgeBottom
	bottom += strings.Repeat(" ", maxInt(0, totalWidth-lipgloss.Width(bottom)))
	out = append(out, bottom)
	return strings.Join(out, "\n")
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

func stripANSI(value string) string {
	return oscPattern.ReplaceAllString(ansiPattern.ReplaceAllString(value, ""), "")
}

func albumInfoLine(album lastfm.Album) string {
	return album.Artist + " — " + album.Title + fmt.Sprintf(" (%d tracks)", len(album.Tracks))
}
