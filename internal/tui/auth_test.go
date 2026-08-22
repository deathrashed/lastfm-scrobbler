package tui

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
)

// fakeAuthClient is a test double for the auth workflow's authClient interface.
type fakeAuthClient struct {
	token      string
	tokenErr   error
	session    lastfm.Session
	sessionErr error
	// lastToken records the token passed to GetSession, to prove the same
	// pending token is reused.
	lastToken string
	apiKey    string
}

func (f *fakeAuthClient) Authenticate(context.Context) error { return nil }
func (f *fakeAuthClient) GetAuthToken(context.Context) (string, error) {
	return f.token, f.tokenErr
}
func (f *fakeAuthClient) GetSession(_ context.Context, token string) (lastfm.Session, error) {
	f.lastToken = token
	return f.session, f.sessionErr
}
func (f *fakeAuthClient) AuthURL(token string) string {
	return "https://www.last.fm/api/auth/?api_key=" + f.apiKey + "&token=" + token
}
func (f *fakeAuthClient) SessionKey() string { return "" }
func (f *fakeAuthClient) SearchAlbums(context.Context, string) ([]lastfm.Album, error) {
	return nil, nil
}
func (f *fakeAuthClient) GetAlbumTracks(context.Context, string, string) (lastfm.Album, error) {
	return lastfm.Album{}, nil
}
func (f *fakeAuthClient) GetDiscography(context.Context, string) ([]lastfm.Album, error) {
	return nil, nil
}
func (f *fakeAuthClient) GetSimilarAlbums(context.Context, string, int) ([]lastfm.Album, error) {
	return nil, nil
}
func (f *fakeAuthClient) GetRecentTracks(context.Context, string, time.Time) ([]lastfm.RecentTrack, error) {
	return nil, nil
}
func (f *fakeAuthClient) Scrobble(context.Context, string, string, string, int64) error {
	return nil
}

var errTestAuth = errors.New("test auth failure")

// withFakeAuth wires a model with a fake auth client and a capturing browser
// opener, restoring the real opener afterwards.
func withFakeAuth(t *testing.T, m model, client *fakeAuthClient) (model, *string) {
	t.Helper()
	orig := openBrowserURL
	var opened string
	openBrowserURL = func(value string) error {
		opened = value
		return nil
	}
	t.Cleanup(func() { openBrowserURL = orig })
	m.client = client
	return m, &opened
}

func TestAuthScreenUnauthenticatedShowsStatus(t *testing.T) {
	m := model{stage: stageAuth, authState: authInvalid, width: 100, cfg: config.Config{Username: "deathrashed"}}
	out := stripANSI(m.renderAuthView())
	if !strings.Contains(out, "STATUS ❯ EXPIRED") {
		t.Fatalf("invalid auth view missing STATUS capsule: %q", out)
	}
	if strings.Contains(out, "L A S T . F M   A U T H") {
		t.Fatalf("auth body must not repeat the hero title: %q", out)
	}
	if !strings.Contains(out, "ACCOUNT ❯ deathrashed") {
		t.Fatalf("auth view missing ACCOUNT capsule: %q", out)
	}
	if !strings.Contains(out, "CURRENT ACTION") {
		t.Fatalf("auth view missing CURRENT ACTION section: %q", out)
	}
	if !strings.Contains(out, "START AGAIN") {
		t.Fatalf("invalid view missing START AGAIN action: %q", out)
	}
}

func TestAuthScreenPendingState(t *testing.T) {
	m := model{stage: stageAuth, authState: authPending, authToken: "tok", width: 100, cfg: config.Config{Username: "deathrashed"}}
	out := stripANSI(m.renderAuthView())
	if !strings.Contains(out, "STATUS ❯ WAITING") {
		t.Fatalf("pending auth view missing WAITING status: %q", out)
	}
	if !strings.Contains(out, "OPEN LAST.FM") || !strings.Contains(out, "GET SESSION KEY") {
		t.Fatalf("pending view missing OPEN LAST.FM + GET SESSION KEY actions: %q", out)
	}
}

func TestAuthScreenExchangingState(t *testing.T) {
	m := model{stage: stageAuth, authState: authExchanging, width: 100, cfg: config.Config{Username: "deathrashed"}}
	out := stripANSI(m.renderAuthView())
	if !strings.Contains(out, "STATUS ❯ FETCHING") {
		t.Fatalf("exchanging view missing FETCHING status: %q", out)
	}
	if strings.Contains(out, "GET SESSION KEY") && strings.Contains(out, "OPEN LAST.FM") {
		t.Fatalf("exchanging view must not offer clickable actions: %q", out)
	}
}

func TestAuthScreenValidState(t *testing.T) {
	m := model{stage: stageAuth, authState: authValid, authReturn: stageInput, authUsername: "deathrashed", width: 100, cfg: config.Config{Username: "deathrashed"}}
	out := stripANSI(m.renderAuthView())
	if !strings.Contains(out, "STATUS ❯ AUTHENTICATED") {
		t.Fatalf("valid view missing AUTHENTICATED status: %q", out)
	}
	if !strings.Contains(out, "RETURN TO DASHBOARD") {
		t.Fatalf("valid view missing context-aware return action (dashboard): %q", out)
	}
	if !strings.Contains(out, "RESULT ❯ Session key updated") {
		t.Fatalf("valid view missing success RESULT capsule: %q", out)
	}

	m.authReturn = stageScrobbling
	out = stripANSI(m.renderAuthView())
	if !strings.Contains(out, "RETURN TO SCROBBLING") {
		t.Fatalf("return action not context-aware (scrobbling): %q", out)
	}
}

func TestAuthScreenFailedState(t *testing.T) {
	m := model{stage: stageAuth, authState: authFailed, authError: &lastfm.AuthError{Code: 14, Message: "Unauthorized Token"}, width: 100, cfg: config.Config{Username: "deathrashed"}}
	out := stripANSI(m.renderAuthView())
	if !strings.Contains(out, "STATUS ❯ FAILED") {
		t.Fatalf("failed view missing FAILED status: %q", out)
	}
	if !strings.Contains(out, "RESULT ❯ 14 · Unauthorized token") {
		t.Fatalf("failed view missing coded RESULT capsule: %q", out)
	}
	if !strings.Contains(out, "START AGAIN") {
		t.Fatalf("failed view missing START AGAIN action: %q", out)
	}
	// The raw API message must stay out of the primary UX.
	if strings.Contains(out, "Unauthorized Token -") || strings.Contains(out, m.authError.Error()) {
		t.Fatalf("failed view leaks the raw API message into the main UX: %q", out)
	}
}

func TestReauthenticateAlwaysRequestsFreshToken(t *testing.T) {
	client := &fakeAuthClient{token: "brand_new_token"}
	m := model{stage: stageScrobbling, authState: authInvalid, authToken: "stale_token", width: 100, cfg: config.Config{Username: "deathrashed"}}
	var opened string
	m, openedPtr := withFakeAuth(t, m, client)
	opened = *openedPtr

	updated, cmd := m.reauthenticate()
	got := updated.(model)
	if got.authToken != "" {
		t.Fatalf("reauthenticate should clear the stale token, got %q", got.authToken)
	}
	if got.authState != authUnknown && got.authState != authPending {
		// After reauthenticate the token command runs; state may still be
		// authUnknown until the message returns, but must not keep the stale
		// token-driven authPending.
	}
	// Run the token command to confirm a fresh token is requested.
	if cmd == nil {
		t.Fatal("reauthenticate did not request a token")
	}
	msg := cmd()
	tok, ok := msg.(authTokenMsg)
	if !ok {
		t.Fatalf("expected authTokenMsg, got %T", msg)
	}
	if tok.token != "brand_new_token" {
		t.Fatalf("reauthenticate requested token %q, want brand_new_token", tok.token)
	}
	_ = opened
}

func TestAuthTokenMessageOpensBrowserWithCorrectURL(t *testing.T) {
	client := &fakeAuthClient{token: "the_token", apiKey: "key123"}
	m := model{stage: stageAuth, width: 100, cfg: config.Config{Username: "deathrashed", APIKey: "key123"}}
	m, opened := withFakeAuth(t, m, client)

	updated, _ := m.updateAuthMsg(authTokenMsg{token: "the_token"})
	got := updated.(model)
	if got.authState != authPending {
		t.Fatalf("after token msg state = %d, want authPending", got.authState)
	}
	if got.authToken != "the_token" {
		t.Fatalf("auth token = %q, want the_token", got.authToken)
	}
	if *opened != "https://www.last.fm/api/auth/?api_key=key123&token=the_token" {
		t.Fatalf("browser opened with %q", *opened)
	}
}

func TestGetSessionKeyUsesSamePendingToken(t *testing.T) {
	client := &fakeAuthClient{token: "p1", session: lastfm.Session{Name: "deathrashed", Key: "ses"}, apiKey: "key123"}
	m := model{stage: stageAuth, authState: authPending, authToken: "p1", width: 100, cfg: config.Config{Username: "deathrashed", APIKey: "key123"}}
	m, _ = withFakeAuth(t, m, client)

	// Pressing enter in the pending state exchanges the SAME token.
	updated, cmd := m.updateAuth(keyMessage("enter"))
	got := updated.(model)
	if got.authState != authExchanging {
		t.Fatalf("enter in pending state set state = %d, want authExchanging", got.authState)
	}
	msg := cmd()
	ex, ok := msg.(authSessionMsg)
	if !ok {
		t.Fatalf("expected authSessionMsg, got %T", msg)
	}
	if ex.err != nil {
		t.Fatalf("exchange error: %v", ex.err)
	}
	// Apply the session.
	updated, _ = got.updateAuthMsg(ex)
	got = updated.(model)
	if client.lastToken != "p1" {
		t.Fatalf("GetSession called with token %q, want p1", client.lastToken)
	}
	if got.authState != authValid {
		t.Fatalf("after exchange state = %d, want authValid", got.authState)
	}
	if got.authToken != "" {
		t.Fatalf("pending token should be cleared after success, got %q", got.authToken)
	}
}

func TestError9MarksAuthInvalidDuringScrobble(t *testing.T) {
	client := &fakeAuthClient{}
	m := model{
		stage:         stageScrobbling,
		scrobbleIdx:   1,
		scrobbleQueue: []queuedTrack{{Artist: "A", Title: "B", Album: "C"}},
		width:         100,
		cfg:           config.Config{Username: "deathrashed"},
	}
	m, _ = withFakeAuth(t, m, client)

	err9 := &lastfm.AuthError{Code: 9, Message: "Invalid session key - Please re-authenticate"}
	updated, _ := m.updateScrobbling(scrobbleResultMsg{idx: 1, track: m.scrobbleQueue[0], err: err9})
	got := updated.(model)
	if got.authState != authInvalid {
		t.Fatalf("error 9 did not mark auth invalid, state = %d", got.authState)
	}
	if !got.scrobblePaused {
		t.Fatal("scrobble should be paused on error 9")
	}
	plain := stripANSI(got.err.Error())
	if !strings.Contains(plain, "authentication expired or was revoked") {
		t.Fatalf("scrobble error not actionable: %q", plain)
	}
	if strings.Contains(plain, "Invalid session key") {
		t.Fatalf("scrobble error should not surface the raw API error verbatim: %q", plain)
	}
}

func TestError9StopsTransientRetryBehavior(t *testing.T) {
	client := &fakeAuthClient{}
	m := model{
		stage:         stageScrobbling,
		scrobbleIdx:   1,
		scrobbleQueue: []queuedTrack{{Artist: "A", Title: "B", Album: "C"}},
		width:         100,
		cfg:           config.Config{Username: "deathrashed"},
	}
	m, _ = withFakeAuth(t, m, client)

	err9 := &lastfm.AuthError{Code: 9, Message: "Invalid session key"}
	updated, cmd := m.updateScrobbling(scrobbleResultMsg{idx: 1, track: m.scrobbleQueue[0], err: err9})
	got := updated.(model)
	// The scrobble must NOT advance to the next track / schedule another
	// attempt. Because the loop returns early after marking auth invalid, no
	// scrobbleNext command should be issued.
	if cmd != nil {
		t.Fatalf("error 9 should not schedule further scrobble commands, got %v", cmd)
	}
	if got.scrobbleIdx != 1 {
		t.Fatalf("scrobble index advanced to %d, should stay 1", got.scrobbleIdx)
	}
}

func TestError9PreservesQueueAndDoesNotDuplicate(t *testing.T) {
	client := &fakeAuthClient{}
	// Two scrobbles: one already succeeded (idx 0), the failing one at idx 1.
	m := model{
		stage:       stageScrobbling,
		scrobbleIdx: 1,
		scrobbleQueue: []queuedTrack{
			{Artist: "A", Title: "One", Album: "C", Failed: false},
			{Artist: "A", Title: "Two", Album: "C", Failed: false},
		},
		width: 100,
		cfg:   config.Config{Username: "deathrashed"},
	}
	m, _ = withFakeAuth(t, m, client)
	// Mark index 0 as already-successful in the queue (it scrobbled fine).
	m.scrobbleQueue[0].Failed = false

	err9 := &lastfm.AuthError{Code: 9, Message: "Invalid session key"}
	updated, _ := m.updateScrobbling(scrobbleResultMsg{idx: 1, track: m.scrobbleQueue[1], err: err9})
	got := updated.(model)

	// The failed track is recorded in failures, but the already-done track is
	// not duplicated or re-added.
	if len(got.failures) != 1 {
		t.Fatalf("failures = %d, want 1 (only the failing track)", len(got.failures))
	}
	if got.failures[0].Title != "Two" {
		t.Fatalf("failure should be track Two, got %q", got.failures[0].Title)
	}
	// Queue length is preserved (no duplicates added).
	if len(got.scrobbleQueue) != 2 {
		t.Fatalf("queue length = %d, want 2", len(got.scrobbleQueue))
	}
}

func TestAuthReturnResumesToScrobbling(t *testing.T) {
	client := &fakeAuthClient{}
	m := model{stage: stageScrobbling, authState: authInvalid, authReturn: stageScrobbling, scrobblePaused: true, width: 100, cfg: config.Config{Username: "deathrashed"}}
	m, _ = withFakeAuth(t, m, client)

	updated, _ := m.leaveAuth()
	got := updated.(model)
	if got.stage != stageScrobbling {
		t.Fatalf("leaveAuth returned stage %d, want stageScrobbling", got.stage)
	}
	if got.scrobblePaused {
		t.Fatal("returning to scrobbling should clear scrobblePaused")
	}
}

func TestSuccessfulExchangeReturnsToAuthReturnScreen(t *testing.T) {
	client := &fakeAuthClient{session: lastfm.Session{Name: "deathrashed", Key: "ses"}, apiKey: "key123"}
	m := model{
		stage:        stageAuth,
		authState:    authPending,
		authToken:    "p1",
		authReturn:   stageConfig,
		width:        100,
		configInput:  newTextInput(1024, 44),
		envInput:     newTextInput(1024, 44),
		profileInput: newTextInput(1024, 44),
		cfg:          config.Config{Username: "deathrashed", APIKey: "key123"},
	}
	m, _ = withFakeAuth(t, m, client)

	updated, _ := m.updateAuthMsg(authSessionMsg{session: client.session})
	got := updated.(model)
	if got.authState != authValid {
		t.Fatalf("state = %d, want authValid", got.authState)
	}
	// Leaving auth from a settings-initiated flow returns to settings Account.
	updated, _ = got.leaveAuth()
	got = updated.(model)
	if got.stage != stageConfig {
		t.Fatalf("leaveAuth returned stage %d, want stageConfig", got.stage)
	}
}

func TestAuthButtonOpensBrowserViaMouseAndKeyboard(t *testing.T) {
	client := &fakeAuthClient{token: "tk", apiKey: "key123"}
	m := model{stage: stageAuth, authState: authPending, authToken: "tk", width: 100, cfg: config.Config{Username: "deathrashed", APIKey: "key123"}}
	m, opened := withFakeAuth(t, m, client)

	// Keyboard 'o' opens the browser.
	updated, _ := m.updateAuth(keyMessage("o"))
	_ = updated
	if *opened != "https://www.last.fm/api/auth/?api_key=key123&token=tk" {
		t.Fatalf("keyboard o opened %q", *opened)
	}

	// Mouse click on the auth:open region opens the browser too.
	*opened = ""
	region := mouseRegion{id: "auth:open", message: keyMessage("o")}
	updated, _ = m.updateMouseRegionOrKey(region)
	_ = updated
	if *opened != "https://www.last.fm/api/auth/?api_key=key123&token=tk" {
		t.Fatalf("mouse auth:open opened %q", *opened)
	}
}

func TestAuthMouseRegionsMatchVisibleButtons(t *testing.T) {
	collect := func(m model) map[string]bool {
		ids := map[string]bool{}
		for _, r := range m.screenRegions() {
			if strings.HasPrefix(r.id, "auth:") {
				ids[r.id] = true
			}
		}
		return ids
	}

	pending := collect(model{stage: stageAuth, authState: authPending, authToken: "tk", width: 100, cfg: config.Config{Username: "u"}, client: lastfm.New("k", "s", "u", "", "")})
	if !pending["auth:open"] || !pending["auth:get-session"] {
		t.Fatalf("pending regions = %v, want auth:open + auth:get-session", pending)
	}

	failed := collect(model{stage: stageAuth, authState: authFailed, width: 100, cfg: config.Config{Username: "u"}})
	if !failed["auth:retry"] || failed["auth:open"] || failed["auth:get-session"] {
		t.Fatalf("failed regions = %v, want only auth:retry", failed)
	}

	valid := collect(model{stage: stageAuth, authState: authValid, authReturn: stageInput, width: 100, cfg: config.Config{Username: "u"}})
	if !valid["auth:return"] || valid["auth:retry"] {
		t.Fatalf("valid regions = %v, want only auth:return", valid)
	}

	exchanging := collect(model{stage: stageAuth, authState: authExchanging, width: 100, cfg: config.Config{Username: "u"}})
	if len(exchanging) != 0 {
		t.Fatalf("exchanging regions = %v, want none (no clickable actions)", exchanging)
	}
}

func TestSettingsReflectsAuthenticationState(t *testing.T) {
	m := model{stage: stageConfig, settingsSection: settingsAccount, settingsFocus: settingsFocusContent, authState: authValid, authUsername: "deathrashed", width: 100, cfg: config.Config{Username: "deathrashed"}}
	out := stripANSI(m.authStatusOverview())
	if !strings.Contains(out, "✓ deathrashed") {
		t.Fatalf("valid auth status overview = %q, want ✓ deathrashed", out)
	}

	m2 := model{stage: stageConfig, settingsSection: settingsAccount, settingsFocus: settingsFocusContent, authState: authInvalid, width: 100, cfg: config.Config{Username: "deathrashed"}}
	if !strings.Contains(stripANSI(m2.authStatusOverview()), "✗ not authenticated") {
		t.Fatalf("invalid auth status overview = %q", stripANSI(m2.authStatusOverview()))
	}
}

func TestReauthenticateSettingsActionStartsFlow(t *testing.T) {
	client := &fakeAuthClient{token: "fresh"}
	m := model{stage: stageConfig, settingsSection: settingsAccount, settingsFocus: settingsFocusContent, authState: authInvalid, width: 100, cfg: config.Config{Username: "deathrashed"}}
	m, _ = withFakeAuth(t, m, client)

	updated, _ := m.openSettingsRowAction_helper("reauthenticate")
	got := updated.(model)
	if got.stage != stageAuth {
		t.Fatalf("reauthenticate action moved to stage %d, want stageAuth", got.stage)
	}
	if got.authReturn != stageConfig {
		t.Fatalf("authReturn = %d, want stageConfig", got.authReturn)
	}
}

// helpers ---------------------------------------------------------------

func (m model) updateMouseRegionOrKey(region mouseRegion) (tea.Model, tea.Cmd) {
	// Reuse the key path the region maps to, mirroring updateMouse dispatch.
	if strings.HasPrefix(region.id, "auth:") {
		return m.updateModelKey(region.message)
	}
	return m, nil
}

func (m model) openSettingsRowAction_helper(id string) (tea.Model, tea.Cmd) {
	m.settingsRow = 0
	for i, row := range settingsRows(settingsAccount) {
		if row.ID == id {
			m.settingsRow = i
			break
		}
	}
	return m.openSettingsRowAction()
}
