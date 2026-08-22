package lastfm

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
)

// AuthError is a typed Last.fm API error carrying the numeric code the API
// returns. It lets callers branch on specific failure modes (such as an
// invalid/revoked session key) without string matching.
type AuthError struct {
	Code    int
	Message string
}

func (e *AuthError) Error() string {
	return fmt.Sprintf("last.fm API error %d: %s", e.Code, e.Message)
}

// Code returns the numeric Last.fm error code if err is (or wraps) an
// AuthError, otherwise 0. Callers use this to detect eg. error 9 (invalid
// session key) without parsing the message text.
func Code(err error) int {
	var authErr *AuthError
	if errors.As(err, &authErr) {
		return authErr.Code
	}
	return 0
}

// Session is the result of a successful auth.getSession exchange.
type Session struct {
	Name string
	Key  string
}

// GetAuthToken requests an unauthorized request token from auth.getToken. The
// token is used to build the authorization URL the user grants in their
// browser before auth.getSession is called.
func (c *client) GetAuthToken(ctx context.Context) (string, error) {
	var response struct {
		Token string `json:"token"`
	}
	if err := c.doRequest(ctx, map[string]string{"method": "auth.getToken"}, &response); err != nil {
		return "", err
	}
	if strings.TrimSpace(response.Token) == "" {
		return "", fmt.Errorf("last.fm returned an empty auth token")
	}
	return response.Token, nil
}

// GetSession exchanges a previously authorized request token for a session via
// auth.getSession. The request is signed using the shared secret.
func (c *client) GetSession(ctx context.Context, token string) (Session, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return Session{}, fmt.Errorf("auth.getSession requires a non-empty token")
	}
	params := map[string]string{
		"method": "auth.getSession",
		"token":  token,
	}
	var response struct {
		Session struct {
			Name string `json:"name"`
			Key  string `json:"key"`
		} `json:"session"`
	}
	if err := c.doRequest(ctx, params, &response); err != nil {
		return Session{}, err
	}
	if strings.TrimSpace(response.Session.Key) == "" {
		return Session{}, fmt.Errorf("last.fm returned an empty session key")
	}
	return Session{Name: response.Session.Name, Key: response.Session.Key}, nil
}

// AuthURL returns the Last.fm desktop authorization URL for the given token.
// The token must come from GetAuthToken and the api_key is the application key
// configured on the client. No secrets appear in the URL.
func (c *client) AuthURL(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	values := url.Values{}
	values.Set("api_key", c.apiKey)
	values.Set("token", token)
	return "https://www.last.fm/api/auth/?" + values.Encode()
}
