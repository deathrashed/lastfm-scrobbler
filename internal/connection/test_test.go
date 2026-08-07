package connection

import (
	"context"
	"testing"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/config"
	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
)

type fakeClient struct{ authErr, searchErr error }

func (f fakeClient) Authenticate(context.Context) error { return f.authErr }
func (f fakeClient) SearchAlbums(context.Context, string) ([]lastfm.Album, error) {
	return []lastfm.Album{{Artist: "A", Title: "B"}}, f.searchErr
}
func (f fakeClient) GetAlbumTracks(context.Context, string, string) (lastfm.Album, error) {
	return lastfm.Album{}, nil
}
func (f fakeClient) GetDiscography(context.Context, string) ([]lastfm.Album, error) { return nil, nil }
func (f fakeClient) GetSimilarAlbums(context.Context, string, int) ([]lastfm.Album, error) {
	return nil, nil
}
func (f fakeClient) GetRecentTracks(context.Context, string, time.Time) ([]lastfm.RecentTrack, error) {
	return nil, nil
}
func (f fakeClient) Scrobble(context.Context, string, string, string, int64) error { return nil }

func TestConnectionReport(t *testing.T) {
	cfg := config.Config{APIKey: "key", APISecret: "secret", Username: "user", Password: "pass"}
	report := Test(context.Background(), cfg, fakeClient{})
	if !report.OK() {
		t.Fatalf("expected success: %#v", report)
	}
}
