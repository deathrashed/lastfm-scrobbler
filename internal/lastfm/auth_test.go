package lastfm

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
)

// authTestClient builds a *client with a fixed key/secret for signature tests.
func authTestClient() *client {
	return &client{
		apiKey:     "test_api_key",
		apiSecret:  "test_api_secret",
		username:   "tester",
		password:   "pw",
		sessionKey: "",
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}
}

func TestGetAuthTokenRequestsCorrectMethodAndSigns(t *testing.T) {
	var gotMethod string
	var gotParams url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotMethod = r.Method
		_ = r.ParseForm()
		gotParams = r.Form
		_ = json.NewEncoder(w).Encode(map[string]any{
			"token": "fresh_request_token",
		})
	}))
	defer srv.Close()

	c := authTestClient()
	c.baseURL = srv.URL // point the signed request at the test server
	tok, err := c.GetAuthToken(context.Background())
	if err != nil {
		t.Fatalf("GetAuthToken error: %v", err)
	}
	if tok != "fresh_request_token" {
		t.Fatalf("token = %q, want fresh_request_token", tok)
	}
	if gotMethod != http.MethodGet && gotMethod != http.MethodPost {
		t.Fatalf("unexpected method %q", gotMethod)
	}
	if gotParams.Get("method") != "auth.getToken" {
		t.Fatalf("method param = %q, want auth.getToken", gotParams.Get("method"))
	}
	if gotParams.Get("api_key") != "test_api_key" {
		t.Fatalf("api_key param = %q, want test_api_key", gotParams.Get("api_key"))
	}
	if gotParams.Get("api_sig") == "" {
		t.Fatal("request was not signed (missing api_sig)")
	}
	// The signature is computed by doRequest BEFORE the format param is
	// appended, so compare against params minus format.
	flat := map[string]string{}
	for k := range gotParams {
		if k == "format" || k == "api_sig" {
			continue
		}
		flat[k] = gotParams.Get(k)
	}
	expected := sign(flat, "test_api_secret")
	if gotParams.Get("api_sig") != expected {
		t.Fatalf("api_sig = %q, want %q", gotParams.Get("api_sig"), expected)
	}
}

func TestGetSessionIncludesTokenAndSigns(t *testing.T) {
	var gotParams url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		gotParams = r.Form
		_ = json.NewEncoder(w).Encode(map[string]any{
			"session": map[string]any{
				"name":       "deathrashed",
				"key":        "session_key_abc123",
				"subscriber": 1,
			},
		})
	}))
	defer srv.Close()

	c := authTestClient()
	c.baseURL = srv.URL
	sess, err := c.GetSession(context.Background(), "the_pending_token")
	if err != nil {
		t.Fatalf("GetSession error: %v", err)
	}
	if sess.Name != "deathrashed" {
		t.Fatalf("session name = %q, want deathrashed", sess.Name)
	}
	if sess.Key != "session_key_abc123" {
		t.Fatalf("session key = %q, want session_key_abc123", sess.Key)
	}
	if gotParams.Get("method") != "auth.getSession" {
		t.Fatalf("method param = %q, want auth.getSession", gotParams.Get("method"))
	}
	if gotParams.Get("token") != "the_pending_token" {
		t.Fatalf("token param = %q, want the_pending_token", gotParams.Get("token"))
	}
	if gotParams.Get("api_sig") == "" {
		t.Fatal("request was not signed (missing api_sig)")
	}
}

func TestGetSessionEmptyTokenIsRejected(t *testing.T) {
	c := authTestClient()
	if _, err := c.GetSession(context.Background(), "   "); err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestAuthURLContainsKeyAndToken(t *testing.T) {
	c := authTestClient()
	u := c.AuthURL("pending_token_xyz")
	if !strings.Contains(u, "https://www.last.fm/api/auth/") {
		t.Fatalf("auth URL = %q, missing Last.fm auth endpoint", u)
	}
	parsed, err := url.Parse(u)
	if err != nil {
		t.Fatalf("auth URL did not parse: %v", err)
	}
	q := parsed.Query()
	if q.Get("api_key") != "test_api_key" {
		t.Fatalf("api_key = %q, want test_api_key", q.Get("api_key"))
	}
	if q.Get("token") != "pending_token_xyz" {
		t.Fatalf("token = %q, want pending_token_xyz", q.Get("token"))
	}
	// Tokens are safe in a URL; verify proper escaping of reserved chars.
	if c.AuthURL("a b/c") != "https://www.last.fm/api/auth/?api_key=test_api_key&token=a+b%2Fc" {
		t.Fatalf("auth URL escaping wrong: %q", c.AuthURL("a b/c"))
	}
}

func TestAuthURLEmptyTokenIsEmpty(t *testing.T) {
	c := authTestClient()
	if got := c.AuthURL(""); got != "" {
		t.Fatalf("empty token should yield empty URL, got %q", got)
	}
}

func TestAPIError9BecomesAuthErrorWithCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   9,
			"message": "Invalid session key - Please re-authenticate",
		})
	}))
	defer srv.Close()

	c := authTestClient()
	c.baseURL = srv.URL
	_, err := c.GetSession(context.Background(), "token")
	if err == nil {
		t.Fatal("expected an error for API error 9")
	}
	if Code(err) != 9 {
		t.Fatalf("Code(err) = %d, want 9", Code(err))
	}
	var authErr *AuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("error %v is not an *AuthError", err)
	}
	if authErr.Code != 9 {
		t.Fatalf("authErr.Code = %d, want 9", authErr.Code)
	}
	if !strings.Contains(err.Error(), "Invalid session key") {
		t.Fatalf("error message unexpected: %v", err)
	}
}

func TestNonLastFMErrorDoesNotReportCode9(t *testing.T) {
	plain := context.Canceled
	if Code(plain) != 0 {
		t.Fatalf("plain error Code = %d, want 0", Code(plain))
	}
	net := &url.Error{Op: "Get", URL: "x", Err: http.ErrNotSupported}
	if Code(net) != 0 {
		t.Fatalf("url.Error Code = %d, want 0", Code(net))
	}
}
