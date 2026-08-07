package updater

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestCheckGitHubResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v10.1.0","html_url":"https://example.invalid/release","body":"notes"}`))
	}))
	defer server.Close()

	result, err := (Checker{}).Check(context.Background(), "v10.0.0", server.URL, "")
	if err != nil {
		t.Fatal(err)
	}
	if !result.Available || result.Latest != "v10.1.0" {
		t.Fatalf("unexpected result: %#v", result)
	}
}

func TestCheckRequiresSource(t *testing.T) {
	_, err := (Checker{}).Check(context.Background(), "v10.0.0", "", "")
	if err == nil {
		t.Fatal("expected missing update source error")
	}
}
