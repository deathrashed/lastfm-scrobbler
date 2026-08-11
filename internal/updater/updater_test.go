package updater

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestCheckGitHubResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v1.1.0","html_url":"https://example.invalid/release","body":"notes"}`))
	}))
	defer server.Close()

	result, err := (Checker{}).Check(context.Background(), "v1.0.0", server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Available || result.Latest != "v1.1.0" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestCheckRequiresSource(t *testing.T) {
	_, err := (Checker{}).Check(context.Background(), "v1.0.0", "", "")
	if err == nil {
		t.Fatal("expected missing update source error")
	}
}

func TestCheckUsesOfficialRepositoryByDefault(t *testing.T) {
	var endpoint string
	checker := Checker{Client: &http.Client{Transport: roundTripFunc(func(request *http.Request) (*http.Response, error) {
		endpoint = request.URL.String()
		return &http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"tag_name":"v1.1.0"}`)),
			Header:     make(http.Header),
		}, nil
	})}}
	_, err := checker.Check(context.Background(), "v1.0.0", "", "deathrashed/lastfm-scrobbler")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://api.github.com/repos/deathrashed/lastfm-scrobbler/releases/latest"
	if endpoint != want {
		t.Fatalf("default endpoint = %q, want %q", endpoint, want)
	}
}

func TestCheckRejectsMalformedCustomURL(t *testing.T) {
	_, err := (Checker{}).Check(context.Background(), "v1.0.0", "://bad", "deathrashed/lastfm-scrobbler")
	if err == nil || !strings.Contains(err.Error(), "missing protocol") {
		t.Fatalf("malformed custom URL error = %v", err)
	}
}
