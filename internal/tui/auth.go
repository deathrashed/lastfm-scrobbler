package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
	"github.com/deathrashed/lastfm-scrobbler/internal/platform"
	"github.com/deathrashed/lastfm-scrobbler/internal/theme"
)

// authState is the explicit authentication lifecycle state. Scattered bools
// made it impossible to tell "waiting for the browser" from "exchanging the
// token", so the workflow is modelled as a small state machine instead.
type authState int

const (
	// authUnknown is the idle state when no auth flow is in progress.
	authUnknown authState = iota
	// authInvalid means the persisted session key was rejected (eg. Last.fm
	// error 9), so the user must re-authenticate before scrobbling.
	authInvalid
	// authPending means a request token has been obtained and the browser is
	// awaiting the user's grant.
	authPending
	// authExchanging means GetSession is in flight for the pending token.
	authExchanging
	// authValid means the session was exchanged and persisted successfully.
	authValid
	// authFailed means the last exchange failed (bad/denied token, network).
	authFailed
)

const authModeChoice = "auth"

type authTokenMsg struct {
	token string
	err   error
}

type authSessionMsg struct {
	session lastfm.Session
	err     error
}

// authClient is the subset of lastfm.Client the auth workflow needs. It is an
// interface so tests can swap in a fake without constructing a full client.
type authClient interface {
	GetAuthToken(ctx context.Context) (string, error)
	GetSession(ctx context.Context, token string) (lastfm.Session, error)
	AuthURL(token string) string
}

func (m model) authClientFor() authClient {
	return m.client
}

// openBrowserURL is a seam so tests can capture the URL without launching a
// real browser. It defaults to the platform abstraction.
var openBrowserURL = platform.OpenURL

func (m model) authStatusLabel() string {
	switch m.authState {
	case authValid:
		return "Authenticated"
	case authInvalid:
		return "Not authenticated"
	case authPending:
		return "Waiting for browser authorization"
	case authExchanging:
		return "Fetching session"
	case authFailed:
		return "Authentication failed"
	default:
		return "No auth"
	}
}

func (m model) authStatusStyle() lipgloss.Style {
	switch m.authState {
	case authValid, authExchanging:
		return theme.SuccessStyle
	case authInvalid, authFailed:
		return theme.ErrorStyle
	default:
		return theme.MutedStyle
	}
}

// openAuth starts (or restarts) a fresh auth flow. A pending token is always
// cleared first so a stale token is never silently reused.
func (m model) openAuth() (tea.Model, tea.Cmd) {
	m.authState = authUnknown
	m.authToken = ""
	m.authError = nil
	m.authReturn = m.stage
	m.stage = stageAuth
	m.modeChoice = authModeChoice
	m.err = nil
	return m, m.requestAuthTokenCmd()
}

// reauthenticate begins a fresh flow from an invalid-session state. It never
// tries to repair the existing key; it requests a brand new token.
func (m model) reauthenticate() (tea.Model, tea.Cmd) {
	m.authState = authInvalid
	return m.openAuth()
}

func (m model) requestAuthTokenCmd() tea.Cmd {
	client := m.authClientFor()
	ctx := m.sessionContext()
	return func() tea.Msg {
		token, err := client.GetAuthToken(ctx)
		return authTokenMsg{token: token, err: err}
	}
}

func (m model) exchangeAuthSessionCmd(token string) tea.Cmd {
	client := m.authClientFor()
	ctx := m.sessionContext()
	return func() tea.Msg {
		session, err := client.GetSession(ctx, token)
		return authSessionMsg{session: session, err: err}
	}
}

func (m model) updateAuth(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		return m.leaveAuth()
	case "q", "ctrl+c":
		return m, tea.Quit
	case "enter":
		switch m.authState {
		case authPending:
			// The user has granted permission in the browser; exchange now.
			m.authState = authExchanging
			return m, m.exchangeAuthSessionCmd(m.authToken)
		case authFailed, authInvalid:
			// START AGAIN: a fresh token, never a retry of the dead one.
			m.authState = authUnknown
			m.authToken = ""
			m.authError = nil
			return m, m.requestAuthTokenCmd()
		case authValid:
			return m.leaveAuth()
		}
	case "o":
		if m.authState == authPending && m.authToken != "" {
			if err := openBrowserURL(m.client.AuthURL(m.authToken)); err != nil {
				m.authError = err
			}
		}
	}
	return m, nil
}

func (m model) updateAuthMsg(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case authTokenMsg:
		if msg.err != nil {
			m.authError = msg.err
			m.authState = authFailed
			return m, nil
		}
		m.authToken = msg.token
		m.authState = authPending
		// Open the authorization page automatically; the user can press o to
		// reopen it if the handler did not launch.
		if err := openBrowserURL(m.client.AuthURL(m.authToken)); err != nil {
			m.authError = err
		}
		return m, nil
	case authSessionMsg:
		if msg.err != nil {
			m.authError = msg.err
			m.authState = authFailed
			m.authToken = "" // require a fresh token on retry
			return m, nil
		}
		return m.applyAuthSession(msg.session)
	}
	return m, nil
}

// applyAuthSession persists the exchanged session and updates the live client
// so the change takes effect immediately, without restarting the app.
func (m model) applyAuthSession(session lastfm.Session) (tea.Model, tea.Cmd) {
	// Persist BEFORE setting m.cfg.SessionKey: PersistSessionKey is a no-op
	// when the in-memory key is already non-empty, so we must hand it the new
	// key while the old value is still present.
	if err := config.PersistSessionKey(m.cfg, session.Key); err != nil {
		m.authError = fmt.Errorf("save session key: %w", err)
		m.authState = authFailed
		return m, nil
	}
	// Rebuild the live client with the new session key. The username from the
	// session (if present) is used to keep the profile identity consistent.
	m.cfg.SessionKey = session.Key
	if strings.TrimSpace(session.Name) != "" {
		m.authUsername = session.Name
	}
	m.client = lastfm.New(m.cfg.APIKey, m.cfg.APISecret, m.cfg.Username, m.cfg.Password, m.cfg.SessionKey)
	m.authError = nil
	m.authState = authValid
	m.authToken = "" // clear the pending token after a successful exchange
	return m, m.saveAuthConfigCmd()
}

// saveAuthConfigCmd persists the rest of the config so a file/keychain-backed
// session key is written through the established persistence path.
func (m model) saveAuthConfigCmd() tea.Cmd {
	cfg := m.cfg
	return func() tea.Msg {
		if err := config.Save(cfg); err != nil {
			return authSessionMsg{err: fmt.Errorf("save config: %w", err)}
		}
		return nil
	}
}

func (m model) leaveAuth() (tea.Model, tea.Cmd) {
	returnStage := m.authReturn
	if returnStage == stageAuth {
		returnStage = stageInput
	}
	m.modeChoice = ""
	m.authReturn = stageInput
	if returnStage == stageConfig {
		return m.openSettingsSection(settingsAccount, settingsFocusContent)
	}
	if returnStage == stageScrobbling {
		m.scrobblePaused = false
		m.stage = stageScrobbling
		// Keep the actionable error until the user resumes; if auth is valid
		// the existing recovery semantics determine what remains.
		if m.authState != authValid {
			m.err = fmt.Errorf("Last.fm authentication expired or was revoked (press A to re-authenticate)")
		} else {
			m.err = nil
		}
		return m, nil
	}
	m.stage = returnStage
	if m.authState == authValid {
		m.err = nil
	}
	return m, nil
}

// markAuthInvalid records that the persisted session key is no longer valid.
// It is called when a signed operation returns Last.fm error 9.
func (m *model) markAuthInvalid() {
	m.authState = authInvalid
}

// scrobblePaused is set when an auth failure interrupts an active run, so the
// scrobble view can show the resume affordance instead of "finished".
func (m model) handleAuthInvalidDuringScrobble() (tea.Model, tea.Cmd) {
	m.markAuthInvalid()
	m.authReturn = stageScrobbling
	m.scrobblePaused = true
	m.err = fmt.Errorf("Last.fm authentication expired or was revoked (press A to re-authenticate)")
	m.authError = nil
	// Persist the pending queue so progress survives the auth detour.
	_ = m.store.SavePending(m.queueRecord("pending"))
	return m, nil
}

func (m model) renderAuthView() string {
	if m.authCompactLayout() {
		return m.renderAuthViewCompact()
	}
	return m.renderAuthViewFull()
}

// authAttachedCapsule renders a three-line capsule that interrupts the top
// border of a parent frame. Geometry is based on terminal display width; the
// middle row deliberately uses ┤/├ so the parent rule can join it cleanly.
func authAttachedCapsule(label, value string, totalWidth int, valueStyle lipgloss.Style) [3]string {
	border := theme.BorderStyle
	innerWidth := maxInt(1, totalWidth-2)
	content := theme.SummaryLabelStyle.Render(label+" ") + theme.SummaryArrowStyle.Render("❯") +
		" " + valueStyle.Render(value)
	return [3]string{
		border.Render("╭" + strings.Repeat("─", innerWidth) + "╮"),
		border.Render("┤") + centerText(fitStyled(content, innerWidth), innerWidth) + border.Render("├"),
		border.Render("╰" + strings.Repeat("─", innerWidth) + "╯"),
	}
}

// authStandaloneCapsule renders an ordinary detached information box. It must
// never use junction glyphs because there is no parent rule to connect to.
func authStandaloneCapsule(label, value string, totalWidth int, valueStyle lipgloss.Style) [3]string {
	border := theme.BorderStyle
	innerWidth := maxInt(1, totalWidth-2)
	content := theme.SummaryLabelStyle.Render(label+" ") + theme.SummaryArrowStyle.Render("❯") +
		" " + valueStyle.Render(value)
	return [3]string{
		border.Render("╭" + strings.Repeat("─", innerWidth) + "╮"),
		border.Render("│") + centerText(fitStyled(content, innerWidth), innerWidth) + border.Render("│"),
		border.Render("╰" + strings.Repeat("─", innerWidth) + "╯"),
	}
}

// authActionCapsule is a complete button box. Earlier versions tried to splice
// action buttons directly into the parent panel's bottom border; that created
// dangling joints and uneven rule lengths in real terminals. Buttons now sit
// cleanly *inside* the stable parent frame while RESULT/RETURN remain detached
// below it.
func authActionCapsule(label string, totalWidth int, hovered bool) [3]string {
	border := theme.BorderStyle
	labelStyle := theme.ModeStyle
	mnemonicStyle := theme.MnemonicStyle
	if hovered {
		border = theme.InnerBorderStyle
		labelStyle = theme.SelectedModeStyle
		mnemonicStyle = theme.SelectedMnemonicStyle
	}
	innerWidth := maxInt(1, totalWidth-2)
	labelText := labelStyle.Render(label)
	if label != "" {
		labelText = mnemonicStyle.Render(label[:1]) + labelStyle.Render(label[1:])
	}
	return [3]string{
		border.Render("╭" + strings.Repeat("─", innerWidth) + "╮"),
		border.Render("│") + centerText(fitStyled(labelText, innerWidth), innerWidth) + border.Render("│"),
		border.Render("╰" + strings.Repeat("─", innerWidth) + "╯"),
	}
}

// authStatusHeader attaches STATUS and ACCOUNT to one exact-width parent
// frame. The attachment positions are computed once and reused on all three
// structural rows, so the capsules cannot drift independently.
func authStatusHeader(m model) []string {
	totalWidth := m.panelWidth()
	statusLabel := map[authState]string{
		authPending:    "WAITING",
		authExchanging: "FETCHING",
		authValid:      "AUTHENTICATED",
		authFailed:     "FAILED",
		authInvalid:    "EXPIRED",
		authUnknown:    "IDLE",
	}[m.authState]
	statusValue := statusLabel
	if m.authState == authUnknown && strings.TrimSpace(m.cfg.SessionKey) != "" {
		statusValue = "AUTHENTICATED"
	}

	const gapWidth = 2
	available := maxInt(1, totalWidth-2-gapWidth)
	maxEach := maxInt(18, available/2)
	statusWidth := minInt(maxEach, maxInt(24, lipgloss.Width("STATUS ❯ "+statusValue)+4))
	accountWidth := minInt(maxEach, maxInt(24, lipgloss.Width("ACCOUNT ❯ "+m.authenticatedAs())+4))
	groupWidth := statusWidth + gapWidth + accountWidth
	if groupWidth > totalWidth-2 {
		overflow := groupWidth - (totalWidth - 2)
		if accountWidth >= statusWidth {
			accountWidth = maxInt(18, accountWidth-overflow)
		} else {
			statusWidth = maxInt(18, statusWidth-overflow)
		}
		groupWidth = statusWidth + gapWidth + accountWidth
	}

	status := authAttachedCapsule("STATUS", statusValue, statusWidth, m.authStatusStyle())
	account := authAttachedCapsule("ACCOUNT", m.authenticatedAs(), accountWidth, theme.PrimaryTextStyle)
	start := maxInt(1, (totalWidth-groupWidth)/2)
	rightInside := maxInt(0, totalWidth-start-groupWidth-1)
	leftInside := maxInt(0, start-1)
	border := theme.BorderStyle

	row0 := strings.Repeat(" ", start) + status[0] + strings.Repeat(" ", gapWidth) + account[0]
	row0 += strings.Repeat(" ", maxInt(0, totalWidth-lipgloss.Width(row0)))

	middleGroup := status[1] + border.Render(strings.Repeat("─", gapWidth)) + account[1]
	row1 := border.Render("╭"+strings.Repeat("─", leftInside)) + middleGroup +
		border.Render(strings.Repeat("─", rightInside)+"╮")

	bottomGroup := status[2] + strings.Repeat(" ", gapWidth) + account[2]
	row2 := border.Render("│") + strings.Repeat(" ", leftInside) + bottomGroup +
		strings.Repeat(" ", rightInside) + border.Render("│")

	return []string{row0, row1, row2}
}

// authActionRows places one or more complete action boxes inside the parent
// frame and centers them as a group. All three rows retain the parent's left
// and right border columns, eliminating the broken/dangling bottom joins seen
// in Ghostty.
func authActionRows(totalWidth int, actions [][3]string) []string {
	border := theme.BorderStyle
	if len(actions) == 0 {
		return nil
	}
	const gapWidth = 2
	groupWidth := 0
	for index, action := range actions {
		if index > 0 {
			groupWidth += gapWidth
		}
		groupWidth += lipgloss.Width(action[0])
	}
	innerWidth := maxInt(1, totalWidth-2)
	left := maxInt(0, (innerWidth-groupWidth)/2)
	right := maxInt(0, innerWidth-groupWidth-left)
	rows := make([]string, 0, 3)
	for row := 0; row < 3; row++ {
		parts := make([]string, 0, len(actions))
		for _, action := range actions {
			parts = append(parts, action[row])
		}
		group := strings.Join(parts, strings.Repeat(" ", gapWidth))
		rows = append(rows, border.Render("│")+strings.Repeat(" ", left)+group+strings.Repeat(" ", right)+border.Render("│"))
	}
	return rows
}

func (m model) authCurrentAction() string {
	switch m.authState {
	case authPending:
		return "Grant Last.fm permission in your browser, then return here."
	case authExchanging:
		return "Requesting a new Last.fm session key…"
	case authValid:
		return "The new session key is saved and already active."
	case authFailed:
		return "Last.fm did not authorize this token. Start a fresh authorization."
	case authInvalid:
		return "Your Last.fm session expired or was revoked. Authorize a new session to continue."
	default:
		if strings.TrimSpace(m.cfg.SessionKey) != "" {
			return "Your saved session looks healthy."
		}
		return "Authorize Last.fm to let this app scrobble on your behalf."
	}
}

// authResultLine renders the optional RESULT capsule. Empty when there is no
// result worth communicating — never an empty box.
func (m model) authResultLine() string {
	var text string
	style := theme.ErrorStyle
	switch m.authState {
	case authFailed:
		text = authErrorCodeString(m.authError)
	case authInvalid:
		text = "9 · Invalid session"
	case authValid:
		text = "Session key updated"
		style = theme.SuccessStyle
	default:
		return ""
	}
	capsule := authStandaloneCapsule("RESULT", text, maxInt(24, minInt(m.panelWidth()-4, 44)), style)
	return m.centerCapsule(capsule)
}

// centerCapsule pads a 3-line capsule to sit centered under/over the panel.
func (m model) centerCapsule(capsule [3]string) string {
	return centerToWidth(strings.Join(capsule[:], "\n"), m.panelWidth())
}

// authErrorCodeString summarizes a Last.fm API error as "N · short reason"
// without leaking tokens/secrets.
func authErrorCodeString(err error) string {
	code := lastfm.Code(err)
	switch code {
	case 14:
		return "14 · Unauthorized token"
	case 9:
		return "9 · Invalid session"
	}
	if code != 0 {
		return fmt.Sprintf("%d · API error", code)
	}
	if err != nil {
		return "request failed"
	}
	return "unknown error"
}

// authReturnTarget describes where leaveAuth will send the user.
func (m model) authReturnTarget() stage {
	target := m.authReturn
	if target == stageAuth || target == 0 {
		target = stageInput
	}
	return target
}

func (m model) authReturnAction() (label, region string) {
	switch m.authReturnTarget() {
	case stageScrobbling:
		return "RETURN TO SCROBBLING", "auth:return"
	case stageConfig:
		return "RETURN TO SETTINGS", "auth:return"
	default:
		return "RETURN TO DASHBOARD", "auth:return"
	}
}

// authPreservedProgress renders the RETURN context capsule for an interrupted
// scrobble run, e.g. "SCROBBLING • 7 / 12 preserved".
func (m model) authPreservedProgress() string {
	if m.authReturnTarget() != stageScrobbling || len(m.scrobbleQueue) == 0 {
		return ""
	}
	done := m.scrobbleIdx
	total := len(m.scrobbleQueue)
	text := fmt.Sprintf("SCROBBLING • %d / %d preserved", done, total)
	capsule := authStandaloneCapsule("RETURN", text, maxInt(28, minInt(m.panelWidth()-4, 48)), theme.PrimaryTextStyle)
	return m.centerCapsule(capsule)
}

func (m model) authStepRows() []string {
	const labelWidth = 14
	row := func(n int, name, state string, glyph string, glyphStyle lipgloss.Style) string {
		label := theme.SummaryLabelStyle.Render(padRight(fmt.Sprintf("%d %s", n, name), labelWidth+3))
		return label + glyphStyle.Render(glyph+" "+state)
	}
	switch m.authState {
	case authPending:
		return []string{
			row(1, "PERMISSION", "browser opened", "●", theme.AccentTextStyle),
			row(2, "SESSION KEY", "waiting", "○", theme.MutedStyle),
		}
	case authExchanging:
		return []string{
			row(1, "PERMISSION", "granted", "✓", theme.SuccessStyle),
			row(2, "SESSION KEY", "fetching", "●", theme.AccentTextStyle),
		}
	case authValid:
		return []string{
			row(1, "PERMISSION", "granted", "✓", theme.SuccessStyle),
			row(2, "SESSION KEY", "saved", "✓", theme.SuccessStyle),
		}
	case authFailed:
		return []string{
			row(1, "PERMISSION", "not authorized", "✗", theme.ErrorStyle),
			row(2, "SESSION KEY", "unavailable", "○", theme.MutedStyle),
		}
	case authInvalid:
		return []string{
			row(1, "PERMISSION", "required again", "○", theme.MutedStyle),
			row(2, "SESSION KEY", "invalid", "✗", theme.ErrorStyle),
		}
	default:
		return []string{
			row(1, "PERMISSION", "pending", "○", theme.MutedStyle),
			row(2, "SESSION KEY", "waiting", "○", theme.MutedStyle),
		}
	}
}

func (m model) renderAuthViewFull() string {
	totalWidth := m.panelWidth()
	border := theme.BorderStyle
	contentWidth := maxInt(8, totalWidth-4)

	lines := authStatusHeader(m)
	frame := func(content string) string {
		return border.Render("│") + " " + padRight(fitStyled(content, contentWidth), contentWidth) + " " + border.Render("│")
	}

	lines = append(lines,
		frame(theme.SummaryLabelStyle.Render("CURRENT ACTION")),
		frame(theme.MutedStyle.Render(strings.Repeat("─", contentWidth))),
		frame(theme.PrimaryTextStyle.Render(m.authCurrentAction())),
		frame(""),
	)
	for _, row := range m.authStepRows() {
		lines = append(lines, frame(row))
	}
	lines = append(lines, frame(""))

	var actions [][3]string
	switch m.authState {
	case authPending:
		actions = append(actions,
			authActionCapsule("OPEN LAST.FM ↗", 20, m.hoverRegion == "auth:open"),
			authActionCapsule("GET SESSION KEY", 22, m.hoverRegion == "auth:get-session"),
		)
	case authFailed, authInvalid:
		actions = append(actions, authActionCapsule("START AGAIN ↻", 18, m.hoverRegion == "auth:retry"))
	case authValid:
		label, _ := m.authReturnAction()
		actions = append(actions, authActionCapsule(label, maxInt(24, minInt(totalWidth-6, 30)), m.hoverRegion == "auth:return"))
	}
	lines = append(lines, authActionRows(totalWidth, actions)...)
	lines = append(lines, border.Render("╰"+strings.Repeat("─", maxInt(1, totalWidth-2))+"╯"))

	if result := m.authResultLine(); result != "" {
		lines = append(lines, "")
		lines = append(lines, strings.Split(result, "\n")...)
	}
	if progress := m.authPreservedProgress(); progress != "" {
		lines = append(lines, strings.Split(progress, "\n")...)
	}

	return m.centerToApp(strings.Join(lines, "\n"))
}

func (m model) authCompactLayout() bool {
	return m.panelWidth() < 56
}

func (m model) renderAuthViewTiny() string {
	return ""
}

func (m model) renderAuthViewCompact() string {
	title := theme.HeaderTextStyle.Render("LAST.FM AUTH")
	statusValue := m.authStatusLabel()
	rows := []string{
		title,
		theme.MutedStyle.Render(strings.Repeat("─", maxInt(8, m.panelWidth()-2))),
		theme.SummaryLabelStyle.Render(padRight("STATUS", 10)) + m.authStatusStyle().Render(statusValue),
		theme.SummaryLabelStyle.Render(padRight("ACCOUNT", 10)) + theme.PrimaryTextStyle.Render(m.authenticatedAs()),
		"",
		theme.PrimaryTextStyle.Render(fitStyled(m.authCurrentAction(), maxInt(8, m.panelWidth()-2))),
	}
	if result := m.authResultLine(); result != "" {
		rows = append(rows, "", fitStyled(stripANSI(result), maxInt(8, m.panelWidth())))
	}
	return m.centerToApp(strings.Join(rows, "\n"))
}

func (m model) authenticatedAs() string {
	if strings.TrimSpace(m.authUsername) != "" {
		return m.authUsername
	}
	if strings.TrimSpace(m.cfg.Username) != "" {
		return m.cfg.Username
	}
	return "Last.fm"
}

// authStatusOverview renders the Account > Auth Status value without ever
// exposing the secret, session key, or pending token.
func (m model) authStatusOverview() string {
	switch m.authState {
	case authValid:
		return "✓ " + m.authenticatedAs()
	case authInvalid:
		return "✗ not authenticated"
	case authPending:
		return "… awaiting browser grant"
	case authExchanging:
		return "… fetching session"
	case authFailed:
		return "✗ authentication failed"
	default:
		if strings.TrimSpace(m.cfg.SessionKey) != "" {
			return "✓ " + m.authenticatedAs()
		}
		return "✗ not authenticated"
	}
}

// isAuthSessionError reports whether err is a Last.fm API error with the given
// code (eg. 9 for invalid session key).
func isAuthSessionError(err error, code int) bool {
	return lastfm.Code(err) == code
}

func authSessionError(err error) (int, bool) {
	code := lastfm.Code(err)
	return code, code != 0
}
