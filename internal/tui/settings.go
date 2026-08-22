package tui

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

type settingsSection int

const (
	settingsAccount settingsSection = iota
	settingsScrobbling
	settingsHistory
	settingsTools
	settingsInterface
	settingsProfiles
)

const (
	settingsSideWidth   = 19
	settingsCenterWidth = 25
	settingsGridWidth   = 65
)

type settingsFocus int

const (
	settingsFocusContent settingsFocus = iota
	settingsFocusSections
)

type settingsRowKind int

const (
	settingsText settingsRowKind = iota
	settingsSecret
	settingsToggle
	settingsChoice
	settingsPathAction
	settingsAction
	settingsReadOnly
)

type settingsSectionSpec struct {
	Mode     string
	Label    string
	Subtitle string
	Icon     string
}

var settingsSectionSpecs = []settingsSectionSpec{
	{Mode: "account", Label: "A C C O U N T", Subtitle: "manage Last.fm credentials and account access", Icon: theme.IconProfile},
	{Mode: "scrobbling", Label: "S C R O B B L I N G", Subtitle: "configure queue and scrobble behaviour", Icon: theme.IconSettings},
	{Mode: "history", Label: "H I S T O R Y", Subtitle: "review, export, or re-run sessions", Icon: theme.IconHistory},
	{Mode: "tools", Label: "T O O L S", Subtitle: "maintenance, exports, connection and updates", Icon: theme.IconSettings},
	{Mode: "interface", Label: "I N T E R F A C E", Subtitle: "customize display and interaction behaviour", Icon: theme.IconSettings},
	{Mode: "profiles", Label: "P R O F I L E S", Subtitle: "switch and manage saved profiles", Icon: theme.IconProfile},
}

func init() {
	for mode, spec := range settingsHeaderSpecs() {
		compactHeaderSpecs[mode] = spec
	}
}

type settingsRowSpec struct {
	ID           string
	Label        string
	Description  string
	Placeholder  string
	Kind         settingsRowKind
	MaskOverview bool
}

var settingsRowsBySection = map[settingsSection][]settingsRowSpec{
	settingsAccount: {
		{ID: "username", Label: "LAST.FM USERNAME", Description: "Last.fm account used for scrobbles", Kind: settingsText},
		{ID: "password", Label: "LAST.FM PASSWORD", Description: "password used to obtain an authenticated session", Kind: settingsSecret, MaskOverview: true},
		{ID: "api-key", Label: "API KEY", Description: "Last.fm application API key", Kind: settingsText, MaskOverview: true},
		{ID: "api-secret", Label: "API SECRET", Description: "Last.fm application shared secret", Kind: settingsSecret, MaskOverview: true},
		{ID: "credential-source", Label: "CREDENTIAL SOURCE", Description: "choose auto, file, environment, or keychain", Kind: settingsChoice},
		{ID: "credential-path", Label: "CREDENTIAL PATH", Description: "credentials file used for file-backed settings", Placeholder: "~/.config/lastfm-scrobbler/.env", Kind: settingsPathAction},
		{ID: "auth-status", Label: "AUTH STATUS", Description: "current Last.fm authentication state", Kind: settingsReadOnly},
		{ID: "reauthenticate", Label: "RE-AUTHENTICATE", Description: "start a fresh Last.fm authorization flow", Kind: settingsAction},
	},
	settingsScrobbling: {
		{ID: "loop", Label: "LOOP", Description: "how many times to scrobble each selected album", Kind: settingsText},
		{ID: "interval", Label: "INTERVAL", Description: "delay between scrobbles in the active queue", Placeholder: "2s", Kind: settingsText},
		{ID: "retry-count", Label: "RETRY COUNT", Description: "automatic retries after a failed Last.fm request", Kind: settingsText},
		{ID: "retry-delay", Label: "RETRY DELAY", Description: "pause before each retry", Placeholder: "2s", Kind: settingsText},
		{ID: "duplicate-guard", Label: "DUPLICATE GUARD", Description: "skip matching recent scrobbles; 0 disables it", Placeholder: "0s", Kind: settingsText},
		{ID: "clean-top-albums", Label: "CLEAN DISCOGRAPHY", Description: "hide obvious reissues, demos and duplicate editions", Kind: settingsToggle},
	},
	settingsTools: {
		{ID: "export-dir", Label: "EXPORT DIR", Description: "folder used for JSON, CSV, TXT, M3U8, and diagnostics", Kind: settingsText},
		{ID: "update-url", Label: "UPDATE SOURCE", Description: "GitHub Releases by default; custom JSON endpoint when needed", Placeholder: "GitHub Releases (default)", Kind: settingsText},
		{ID: "connection-test", Label: "CONNECTION TEST", Description: "verify API lookup and authentication readiness", Kind: settingsAction},
		{ID: "diagnostics", Label: "DIAGNOSTICS BUNDLE", Description: "export logs and redacted configuration for troubleshooting", Kind: settingsAction},
		{ID: "completions", Label: "INSTALL COMPLETIONS", Description: "install zsh, bash, fish, or PowerShell completion", Kind: settingsAction},
		{ID: "check-updates", Label: "CHECK FOR UPDATES", Description: "compare this build with the configured release endpoint", Kind: settingsAction},
	},
	settingsInterface: {
		{ID: "notifications", Label: "NOTIFICATIONS", Description: "send a completion notification when supported", Kind: settingsToggle},
		{ID: "now-playing", Label: "NOW PLAYING", Description: "show current or recent Last.fm activity in the full header", Kind: settingsToggle},
		{ID: "compact-header", Label: "COMPACT HEADER", Description: "use the four-line compact header", Kind: settingsToggle},
		{ID: "mouse-support", Label: "MOUSE SUPPORT", Description: "enable clickable controls and mouse-wheel navigation", Kind: settingsToggle},
	},
}

func settingsSpec(section settingsSection) settingsSectionSpec {
	if int(section) < 0 || int(section) >= len(settingsSectionSpecs) {
		return settingsSectionSpecs[settingsScrobbling]
	}
	return settingsSectionSpecs[section]
}

func settingsRows(section settingsSection) []settingsRowSpec {
	return settingsRowsBySection[section]
}

func settingsSectionFromMode(mode string) (settingsSection, bool) {
	for index, spec := range settingsSectionSpecs {
		if spec.Mode == mode {
			return settingsSection(index), true
		}
	}
	return settingsScrobbling, false
}

func (m model) inSettingsArea() bool {
	return m.stage == stageConfig || m.stage == stageHistory || m.stage == stageProfiles
}

func (m model) currentSettingsSection() settingsSection {
	if m.stage == stageHistory {
		return settingsHistory
	}
	if m.stage == stageProfiles {
		return settingsProfiles
	}
	if section, ok := settingsSectionFromMode(m.modeChoice); ok {
		return section
	}
	if m.settingsSection >= settingsAccount && m.settingsSection <= settingsProfiles {
		return m.settingsSection
	}
	return settingsScrobbling
}

func (m model) openSettings() (tea.Model, tea.Cmd) {
	return m.openSettingsSection(settingsScrobbling, settingsFocusContent)
}

func (m model) openSettingsSection(section settingsSection, focus settingsFocus) (tea.Model, tea.Cmd) {
	previousSection := m.settingsSection
	previousRow := m.settingsRow
	if m.stage == stageConfig {
		m.commitSettingsField()
	}
	m.settingsSection = section
	m.settingsFocus = focus
	m.configStatus = ""
	m.err = nil
	m.returnStage = stageInput
	spec := settingsSpec(section)
	m.modeChoice = spec.Mode

	switch section {
	case settingsHistory:
		m.stage = stageHistory
		m.historyCursor = minInt(m.historyCursor, maxInt(0, len(m.history)-1))
		m.historyStatus = ""
		m.configInput.Blur()
		return m, nil
	case settingsProfiles:
		m.stage = stageProfiles
		m.profileCursor = indexOf(m.profiles, m.cfg.Profile)
		if m.profileCursor < 0 {
			m.profileCursor = 0
		}
		m.configInput.Blur()
		return m, nil
	default:
		m.stage = stageConfig
		rows := settingsRows(section)
		if previousSection == section && previousRow >= 0 && previousRow < len(rows) {
			m.settingsRow = previousRow
		} else {
			m.settingsRow = 0
		}
		m.loadSettingsField()
		if focus == settingsFocusContent && m.settingsRowEditable() {
			return m, m.configInput.Focus()
		}
		m.configInput.Blur()
		return m, nil
	}
}

func (m model) switchSettingsSection(section settingsSection) (tea.Model, tea.Cmd) {
	return m.openSettingsSection(section, settingsFocusSections)
}

func settingsGridMove(section settingsSection, key string) settingsSection {
	index := int(section)
	row, col := index/3, index%3
	switch key {
	case "left":
		col = (col + 2) % 3
	case "right":
		col = (col + 1) % 3
	case "up", "down":
		row = 1 - row
	}
	return settingsSection(row*3 + col)
}

func (m model) updateSettingsGrid(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "left", "right", "up", "down", "h", "j", "k", "l":
		key := msg.String()
		switch key {
		case "h":
			key = "left"
		case "l":
			key = "right"
		case "k":
			key = "up"
		case "j":
			key = "down"
		}
		return m.switchSettingsSection(settingsGridMove(m.currentSettingsSection(), key))
	case "enter", "tab", "shift+tab":
		m.settingsFocus = settingsFocusContent
		if m.stage == stageConfig {
			m.loadSettingsField()
			if m.settingsRowEditable() {
				return m, m.configInput.Focus()
			}
		}
		return m, nil
	case "esc":
		return m.leaveSettings()
	case "q", "ctrl+c":
		return m, tea.Quit
	}
	return m, nil
}

func (m model) leaveSettings() (tea.Model, tea.Cmd) {
	if m.stage == stageConfig {
		m.commitSettingsField()
	}
	m.stage = stageInput
	m.modeChoice = ""
	m.returnStage = stageInput
	m.settingsFocus = settingsFocusContent
	m.configInput.Blur()
	return m, nil
}

func (m model) settingsCurrentRow() (settingsRowSpec, bool) {
	rows := settingsRows(m.currentSettingsSection())
	if len(rows) == 0 || m.settingsRow < 0 || m.settingsRow >= len(rows) {
		return settingsRowSpec{}, false
	}
	return rows[m.settingsRow], true
}

func (m model) settingsRowEditable() bool {
	row, ok := m.settingsCurrentRow()
	if !ok {
		return false
	}
	return row.Kind == settingsText || row.Kind == settingsSecret || row.Kind == settingsToggle || row.Kind == settingsChoice
}

func (m *model) loadSettingsField() {
	row, ok := m.settingsCurrentRow()
	if !ok || !m.settingsRowEditable() {
		m.configInput.SetValue("")
		m.configInput.Blur()
		return
	}
	m.configInput.EchoMode = textinput.EchoNormal
	m.configInput.EchoCharacter = '•'
	m.configInput.Width = maxInt(10, 55-len([]rune(row.Label)))
	m.configInput.Placeholder = row.Placeholder
	m.configInput.SetValue(m.settingsRawValue(row))
	if row.Kind == settingsSecret {
		m.configInput.EchoMode = textinput.EchoPassword
	}
	m.configInput.CursorEnd()
}

func (m model) settingsRawValue(row settingsRowSpec) string {
	switch row.ID {
	case "username":
		return m.cfg.Username
	case "password":
		return m.cfg.Password
	case "api-key":
		return m.cfg.APIKey
	case "api-secret":
		return m.cfg.APISecret
	case "credential-source":
		return m.cfg.CredentialSource
	case "loop":
		return strconv.Itoa(m.cfg.DefaultLoop)
	case "interval":
		return m.cfg.DefaultInterval.String()
	case "retry-count":
		return strconv.Itoa(m.cfg.RetryCount)
	case "retry-delay":
		return m.cfg.RetryDelay.String()
	case "duplicate-guard":
		return m.cfg.DuplicateGuard.String()
	case "clean-top-albums":
		return boolWord(m.cfg.CleanDiscography)
	case "export-dir":
		return m.cfg.ExportDir
	case "update-url":
		return m.cfg.UpdateURL
	case "notifications":
		return boolWord(m.cfg.Notify)
	case "now-playing":
		return boolWord(m.cfg.NowPlaying)
	case "compact-header":
		return boolWord(m.cfg.CompactHeader)
	case "mouse-support":
		return boolWord(m.cfg.MouseEnabled)
	}
	return ""
}

func (m model) settingsOverviewValue(row settingsRowSpec) string {
	switch row.ID {
	case "auth-status":
		return m.authStatusOverview()
	case "credential-path":
		if strings.TrimSpace(m.cfg.EnvPath) == "" {
			return "not configured"
		}
		return truncateToWidth(m.cfg.EnvPath, 30)
	case "connection-test", "diagnostics", "completions", "check-updates":
		return "ENTER"
	}
	value := m.settingsRawValue(row)
	if row.MaskOverview {
		return maskValue(value)
	}
	if row.ID == "update-url" && strings.TrimSpace(value) == "" {
		return "GitHub Releases"
	}
	if row.ID == "export-dir" || row.ID == "update-url" {
		return truncateToWidth(value, 28)
	}
	return value
}

func (m *model) commitSettingsField() {
	row, ok := m.settingsCurrentRow()
	if !ok || !m.settingsRowEditable() {
		return
	}
	value := strings.TrimSpace(m.configInput.Value())
	switch row.ID {
	case "username":
		if value != m.cfg.Username {
			m.cfg.MarkCredentialEdited("username")
		}
		m.cfg.Username = value
	case "password":
		if value != m.cfg.Password {
			m.cfg.MarkCredentialEdited("password")
		}
		m.cfg.Password = value
	case "api-key":
		if value != m.cfg.APIKey {
			m.cfg.MarkCredentialEdited("api_key")
		}
		m.cfg.APIKey = value
	case "api-secret":
		if value != m.cfg.APISecret {
			m.cfg.MarkCredentialEdited("api_secret")
		}
		m.cfg.APISecret = value
	case "credential-source":
		source := strings.ToLower(value)
		if indexOf([]string{"auto", "file", "environment", "keychain"}, source) >= 0 {
			m.cfg.CredentialSource = source
		}
	case "loop":
		if number, err := strconv.Atoi(value); err == nil && number > 0 {
			m.cfg.DefaultLoop = number
			m.loopCount = number
		}
	case "interval":
		if duration, ok := parseSettingsDuration(value); ok {
			m.cfg.DefaultInterval = duration
			m.interval = duration
		}
	case "retry-count":
		if number, err := strconv.Atoi(value); err == nil && number >= 0 {
			m.cfg.RetryCount = number
		}
	case "retry-delay":
		if duration, err := time.ParseDuration(value); err == nil && duration >= 0 {
			m.cfg.RetryDelay = duration
		}
	case "duplicate-guard":
		if value == "0" || strings.EqualFold(value, "off") {
			m.cfg.DuplicateGuard = 0
		} else if duration, err := time.ParseDuration(value); err == nil && duration >= 0 {
			m.cfg.DuplicateGuard = duration
		}
	case "clean-top-albums":
		m.cfg.CleanDiscography = parseToggle(value, m.cfg.CleanDiscography)
		m.discographyClean = m.cfg.CleanDiscography
	case "export-dir":
		if value != "" {
			m.cfg.ExportDir = config.ExpandPath(value)
		}
	case "update-url":
		m.cfg.UpdateURL = value
	case "notifications":
		m.cfg.Notify = parseToggle(value, m.cfg.Notify)
	case "now-playing":
		m.cfg.NowPlaying = parseToggle(value, m.cfg.NowPlaying)
	case "compact-header":
		m.cfg.CompactHeader = parseToggle(value, m.cfg.CompactHeader)
	case "mouse-support":
		m.cfg.MouseEnabled = parseToggle(value, m.cfg.MouseEnabled)
	}
}

func parseSettingsDuration(value string) (time.Duration, bool) {
	if duration, err := time.ParseDuration(value); err == nil {
		return duration, true
	}
	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		return time.Duration(seconds * float64(time.Second)), true
	}
	return 0, false
}

func (m *model) selectSettingsRow(index int) tea.Cmd {
	rows := settingsRows(m.currentSettingsSection())
	if len(rows) == 0 {
		return nil
	}
	index = maxInt(0, minInt(index, len(rows)-1))
	m.commitSettingsField()
	m.settingsRow = index
	m.settingsFocus = settingsFocusContent
	m.configStatus = ""
	m.loadSettingsField()
	if m.settingsRowEditable() {
		return m.configInput.Focus()
	}
	m.configInput.Blur()
	return nil
}

func (m *model) adjustSettingsRow(delta int) bool {
	row, ok := m.settingsCurrentRow()
	if !ok {
		return false
	}
	switch row.ID {
	case "clean-top-albums":
		m.cfg.CleanDiscography = !m.cfg.CleanDiscography
		m.discographyClean = m.cfg.CleanDiscography
	case "notifications":
		m.cfg.Notify = !m.cfg.Notify
	case "now-playing":
		m.cfg.NowPlaying = !m.cfg.NowPlaying
	case "compact-header":
		m.cfg.CompactHeader = !m.cfg.CompactHeader
	case "mouse-support":
		m.cfg.MouseEnabled = !m.cfg.MouseEnabled
	case "credential-source":
		sources := []string{"auto", "file", "environment", "keychain"}
		index := indexOf(sources, m.cfg.CredentialSource)
		if index < 0 {
			index = 0
		}
		index = (index + delta + len(sources)) % len(sources)
		m.cfg.CredentialSource = sources[index]
	default:
		return false
	}
	m.loadSettingsField()
	return true
}

func (m *model) activitySettingChanged(rowID string) tea.Cmd {
	if rowID == "now-playing" || rowID == "compact-header" {
		return m.restartActivity()
	}
	return nil
}

func (m model) openSettingsRowAction() (tea.Model, tea.Cmd) {
	row, ok := m.settingsCurrentRow()
	if !ok {
		return m, nil
	}
	switch row.ID {
	case "credential-path":
		return m.openEnvPath()
	case "connection-test":
		return m.openConnectionTest()
	case "diagnostics":
		return m.openDiagnostics()
	case "completions":
		return m.openCompletions()
	case "check-updates":
		return m.openUpdateCheck()
	case "reauthenticate":
		return m.reauthenticate()
	}
	return m, nil
}

func (m model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.settingsFocus == settingsFocusSections {
		return m.updateSettingsGrid(msg)
	}

	rows := settingsRows(m.currentSettingsSection())
	if len(rows) == 0 {
		return m, nil
	}
	row, _ := m.settingsCurrentRow()

	switch msg.String() {
	case "tab", "shift+tab":
		m.commitSettingsField()
		m.settingsFocus = settingsFocusSections
		m.configInput.Blur()
		return m, nil
	case "up", "k":
		if m.settingsRow == 0 {
			m.commitSettingsField()
			m.settingsFocus = settingsFocusSections
			m.configInput.Blur()
			return m, nil
		}
		cmd := m.selectSettingsRow(m.settingsRow - 1)
		return m, cmd
	case "down", "j":
		if m.settingsRow < len(rows)-1 {
			cmd := m.selectSettingsRow(m.settingsRow + 1)
			return m, cmd
		}
		return m, nil
	case "left":
		if m.adjustSettingsRow(-1) {
			return m, tea.Batch(m.configInput.Focus(), m.activitySettingChanged(row.ID))
		}
	case "right", " ":
		if m.adjustSettingsRow(1) {
			return m, tea.Batch(m.configInput.Focus(), m.activitySettingChanged(row.ID))
		}
	case "o":
		if row.ID == "export-dir" {
			return m, pickFolderCmd("Choose an export folder", "export")
		}
	case "enter":
		if row.Kind == settingsAction || row.Kind == settingsPathAction {
			return m.openSettingsRowAction()
		}
		m.commitSettingsField()
		m.loadSettingsField()
		updated, cmd := m.saveConfig()
		return updated, tea.Batch(cmd, m.activitySettingChanged(row.ID))
	case "esc":
		return m.leaveSettings()
	case "q", "ctrl+c":
		return m, tea.Quit
	}

	if !m.settingsRowEditable() {
		return m, nil
	}
	var cmd tea.Cmd
	m.configInput, cmd = m.configInput.Update(msg)
	return m, cmd
}

func (m model) settingsFooterAction() string {
	row, ok := m.settingsCurrentRow()
	if !ok {
		return "open"
	}
	if row.Kind == settingsAction || row.Kind == settingsPathAction {
		return "open"
	}
	return "save"
}

func (m model) settingsFooterSpec() [][]footerItem {
	a := footerActionFor
	s := footerStatic
	enter := func(label, description string) footerItem { return a("footer:enter", "enter", label, description) }
	esc := func(label, description string) footerItem { return a("footer:esc", "esc", label, description) }
	if m.settingsFocus == settingsFocusSections {
		return [][]footerItem{{
			s("↑ ↓ ← →", " section"),
			enter(" content", "move focus into the selected Settings section"),
			a("footer:tab", "tab", " content", "move focus into the selected Settings section"),
			esc(" back", "return to the main menu"),
		}}
	}
	row, _ := m.settingsCurrentRow()
	line := []footerItem{s("↑ ↓", " navigate")}
	if row.Kind == settingsToggle || row.Kind == settingsChoice {
		line = append(line, s("← →", " adjust"))
	}
	actionDescription := "save the highlighted setting"
	if m.settingsFooterAction() == "open" {
		actionDescription = "open or run the highlighted setting"
	}
	line = append(line,
		enter(" "+m.settingsFooterAction(), actionDescription),
		a("footer:tab", "tab", " sections", "move focus to the Settings sections"),
		esc(" back", "return to the main menu"),
	)
	if row.ID == "export-dir" {
		return [][]footerItem{line, {a("footer:o", "o", " folder", "choose an export folder")}}
	}
	return [][]footerItem{line}
}

func renderSettingsView(m model) string {
	rows := settingsRows(m.currentSettingsSection())
	if len(rows) == 0 {
		return renderSettingsShell(m, "")
	}
	if m.settingsRow < 0 || m.settingsRow >= len(rows) {
		m.settingsRow = 0
	}
	listRows := make([]string, 0, len(rows))
	for index, row := range rows {
		listRows = append(listRows, renderSettingsOverviewRow(m, row, index))
	}
	list := renderPanelBox(listRows, m.panelWidth(), theme.BorderStyle)
	row := rows[m.settingsRow]
	detail := renderSettingsDetail(m, row)
	lines := []string{
		m.centerToApp(list),
		m.centerToApp(detail),
		m.centerToApp(theme.MutedStyle.Render(row.Description)),
		"",
	}
	if m.configStatus != "" {
		lines = append(lines, m.centerToApp(theme.SuccessStyle.Render(m.configStatus)), "")
	}
	content := lipgloss.JoinVertical(lipgloss.Left, lines...)
	return renderSettingsShell(m, content)
}

func renderSettingsShell(m model, content string) string {
	grid := m.centerToApp(renderSettingsGrid(m))
	if content == "" {
		return grid
	}
	return lipgloss.JoinVertical(lipgloss.Left, grid, content)
}

func renderSettingsGrid(m model) string {
	boxes := make([]string, len(settingsSectionSpecs))
	widths := []int{settingsSideWidth, settingsCenterWidth, settingsSideWidth, settingsSideWidth, settingsCenterWidth, settingsSideWidth}
	current := m.currentSettingsSection()
	for index, spec := range settingsSectionSpecs {
		boxes[index] = renderSettingsSectionBox(spec.Label, widths[index], settingsSection(index) == current, m.hoverRegion == "settings:section:"+strconv.Itoa(index), m.settingsFocus == settingsFocusSections)
	}
	first := joinResponsiveBoxes(boxes[:3], widths[:3], m.panelWidth(), theme.SepStyle.Render("•"))
	second := joinResponsiveBoxes(boxes[3:], widths[3:], m.panelWidth(), theme.SepStyle.Render("•"))
	return first + "\n" + second
}

func renderSettingsSectionBox(label string, totalWidth int, selected, hovered, gridFocused bool) string {
	_ = gridFocused // selected section remains red regardless of which focus zone is active.
	return renderChoiceBox(label, totalWidth, selected, hovered && !selected)
}

func renderSettingsOverviewRow(m model, row settingsRowSpec, index int) string {
	focused := m.settingsFocus == settingsFocusContent && m.settingsRow == index
	hovered := m.hoverRegion == "settings:row:"+strconv.Itoa(index)
	marker := "  "
	if focused {
		marker = theme.PromptStyle.Render("❯ ")
	}
	labelStyle := theme.RowLabelStyle
	arrowStyle := theme.RowArrowStyle
	valueStyle := theme.RowValueStyle
	if focused {
		labelStyle = theme.FocusedRowLabelStyle
		arrowStyle = theme.FocusedRowArrowStyle
		valueStyle = theme.FocusedRowValueStyle
	} else if hovered {
		labelStyle = theme.HoverRowLabelStyle
		arrowStyle = theme.HoverRowArrowStyle
		valueStyle = theme.HoverRowValueStyle
	}
	left := marker + labelStyle.Render(row.Label+" ") + arrowStyle.Render("❯")
	right := valueStyle.Render(m.settingsOverviewValue(row))
	contentWidth := maxInt(1, m.panelWidth()-4)
	gap := contentWidth - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		gap = 1
	}
	return fitStyled(left+strings.Repeat(" ", gap)+right, contentWidth)
}

func renderSettingsEditValue(m model, row settingsRowSpec) string {
	if row.Kind == settingsText || row.Kind == settingsSecret {
		// textinput.View() supplies the real blinking cursor. The input's
		// TextStyle/Cursor style is configured centrally in newTextInput().
		return m.configInput.View()
	}
	// Toggles and choices are adjusted with arrows, so they should not pretend
	// to be text editors with a caret.
	return m.configInput.Value()
}

func renderSettingsDetail(m model, row settingsRowSpec) string {
	active := m.settingsFocus == settingsFocusContent
	hovered := m.hoverRegion == "settings:detail"
	if row.Kind == settingsReadOnly {
		return renderSettingsDetailLine(row.Label, m.settingsOverviewValue(row), "", active, hovered, false, m.panelWidth())
	}
	if row.Kind == settingsAction {
		return renderSettingsDetailLine(row.Label, "", "ENTER", active, hovered, true, m.panelWidth())
	}
	if row.Kind == settingsPathAction {
		value := m.cfg.EnvPath
		if strings.TrimSpace(value) == "" {
			value = row.Placeholder
		}
		return renderSettingsDetailLine(row.Label, truncateToWidth(value, m.panelWidth()-27), "ENTER", active, hovered, true, m.panelWidth())
	}
	value := renderSettingsEditValue(m, row)
	if strings.TrimSpace(stripANSI(value)) == "" {
		value = theme.MutedStyle.Render(row.Placeholder)
	}
	accentBorder := row.Kind == settingsToggle || row.Kind == settingsChoice
	return renderSettingsDetailLine(row.Label, value, "", active, hovered, accentBorder, m.panelWidth())
}

func renderSettingsDetailLine(label, value, right string, active, hovered, accentBorder bool, totalWidth int) string {
	border := theme.BorderStyle
	if accentBorder && (active || hovered) {
		border = theme.InnerBorderStyle
	}
	labelStyle := theme.RowLabelStyle
	arrowStyle := theme.RowArrowStyle
	valueStyle := theme.RowValueStyle
	if active {
		labelStyle = theme.FocusedRowLabelStyle
		arrowStyle = theme.FocusedRowArrowStyle
		valueStyle = theme.FocusedRowValueStyle
	} else if hovered {
		labelStyle = theme.HoverRowLabelStyle
		arrowStyle = theme.HoverRowArrowStyle
		valueStyle = theme.HoverRowValueStyle
	}
	left := labelStyle.Render(label+" ") + arrowStyle.Render("❯")
	middle := valueStyle.Render(value)
	rightText := ""
	if right != "" {
		rightStyle := theme.MutedStyle
		if active || hovered {
			rightStyle = theme.SuccessStyle
		}
		rightText = rightStyle.Render(right)
	}
	contentWidth := maxInt(1, totalWidth-4)
	leftAndValue := left
	if strings.TrimSpace(stripANSI(value)) != "" {
		leftAndValue += " " + middle
	}
	if rightText != "" {
		gap := contentWidth - lipgloss.Width(leftAndValue) - lipgloss.Width(rightText)
		if gap < 1 {
			gap = 1
		}
		leftAndValue += strings.Repeat(" ", gap) + rightText
	}
	return renderPanelBox([]string{fitStyled(leftAndValue, contentWidth)}, totalWidth, border)
}

func settingsSectionContentStartY() int { return 6 }

func settingsRowPanelHeight(m model) int {
	return len(settingsRows(m.currentSettingsSection())) + 2
}

func (m model) settingsDetailY() int {
	return m.headerHeight() + settingsSectionContentStartY() + settingsRowPanelHeight(m)
}

func (m model) settingsRowY(index int) int {
	return m.headerHeight() + settingsSectionContentStartY() + 1 + index
}

func (m model) settingsGridRegion(section settingsSection) mouseRegion {
	widths := []int{settingsSideWidth, settingsCenterWidth, settingsSideWidth}
	row, col := int(section)/3, int(section)%3
	positions, _ := responsiveCardPositions(widths, m.panelWidth())
	x := m.workX() + positions[col]
	return mouseRegion{
		id:     "settings:section:" + strconv.Itoa(int(section)),
		x:      x,
		y:      m.headerHeight() + row*3,
		width:  widths[col],
		height: 3,
	}
}

func (m model) settingsSectionRegions() []mouseRegion {
	regions := make([]mouseRegion, 0, len(settingsSectionSpecs))
	for section := settingsAccount; section <= settingsProfiles; section++ {
		regions = append(regions, m.settingsGridRegion(section))
	}
	return regions
}

func (m model) settingsRowRegions() []mouseRegion {
	rows := settingsRows(m.currentSettingsSection())
	regions := make([]mouseRegion, 0, len(rows)+1)
	for index := range rows {
		regions = append(regions, mouseRegion{
			id:     "settings:row:" + strconv.Itoa(index),
			x:      m.workX(),
			y:      m.settingsRowY(index),
			width:  m.panelWidth(),
			height: 1,
		})
	}
	if len(rows) > 0 {
		regions = append(regions, mouseRegion{id: "settings:detail", x: m.workX(), y: m.settingsDetailY(), width: m.panelWidth(), height: 3, message: keyMessage("enter")})
	}
	return regions
}

func (m model) updateSettingsMouseRegion(region mouseRegion) (tea.Model, tea.Cmd) {
	switch {
	case strings.HasPrefix(region.id, "settings:section:"):
		index, _ := strconv.Atoi(strings.TrimPrefix(region.id, "settings:section:"))
		return m.openSettingsSection(settingsSection(index), settingsFocusSections)
	case strings.HasPrefix(region.id, "settings:row:"):
		index, _ := strconv.Atoi(strings.TrimPrefix(region.id, "settings:row:"))
		cmd := m.selectSettingsRow(index)
		return m, cmd
	case region.id == "settings:detail":
		m.settingsFocus = settingsFocusContent
		row, ok := m.settingsCurrentRow()
		if !ok {
			return m, nil
		}
		if row.Kind == settingsAction || row.Kind == settingsPathAction {
			return m.openSettingsRowAction()
		}
		if row.Kind == settingsToggle || row.Kind == settingsChoice {
			m.adjustSettingsRow(1)
			return m, tea.Batch(m.configInput.Focus(), m.activitySettingChanged(row.ID))
		}
		return m, m.configInput.Focus()
	}
	return m, nil
}

func (m model) settingsMouseMove(delta int) (tea.Model, tea.Cmd) {
	if m.settingsFocus == settingsFocusSections {
		index := (int(m.currentSettingsSection()) + delta + len(settingsSectionSpecs)) % len(settingsSectionSpecs)
		return m.openSettingsSection(settingsSection(index), settingsFocusSections)
	}
	rows := settingsRows(m.currentSettingsSection())
	if len(rows) == 0 {
		return m, nil
	}
	index := (m.settingsRow + delta + len(rows)) % len(rows)
	cmd := m.selectSettingsRow(index)
	return m, cmd
}

func settingsHeaderSpecs() map[string]compactHeaderSpec {
	result := make(map[string]compactHeaderSpec, len(settingsSectionSpecs))
	for _, spec := range settingsSectionSpecs {
		result[spec.Mode] = compactHeaderSpec{Title: spec.Label, Subtitle: spec.Subtitle, Icon: spec.Icon}
	}
	return result
}

func settingsSectionNames() []string {
	result := make([]string, 0, len(settingsSectionSpecs))
	for _, spec := range settingsSectionSpecs {
		result = append(result, strings.ReplaceAll(spec.Label, " ", ""))
	}
	return result
}

func (m model) settingsDebugString() string {
	return fmt.Sprintf("section=%s focus=%d row=%d", settingsSpec(m.currentSettingsSection()).Mode, m.settingsFocus, m.settingsRow)
}
