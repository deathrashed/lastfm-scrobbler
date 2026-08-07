package scrobbler

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
)

type runnerClient struct {
	remainingFailures int
	calls             int
}

func (c *runnerClient) Authenticate(context.Context) error { return nil }
func (c *runnerClient) SearchAlbums(context.Context, string) ([]lastfm.Album, error) {
	return nil, nil
}
func (c *runnerClient) GetAlbumTracks(context.Context, string, string) (lastfm.Album, error) {
	return lastfm.Album{}, nil
}
func (c *runnerClient) GetDiscography(context.Context, string) ([]lastfm.Album, error) {
	return nil, nil
}
func (c *runnerClient) GetSimilarAlbums(context.Context, string, int) ([]lastfm.Album, error) {
	return nil, nil
}
func (c *runnerClient) GetRecentTracks(context.Context, string, time.Time) ([]lastfm.RecentTrack, error) {
	return nil, nil
}
func (c *runnerClient) Scrobble(context.Context, string, string, string, int64) error {
	c.calls++
	if c.remainingFailures > 0 {
		c.remainingFailures--
		return errors.New("temporary failure")
	}
	return nil
}

func TestRunOneRetriesAndReportsAttempts(t *testing.T) {
	client := &runnerClient{remainingFailures: 2}
	attempts, err := RunOne(context.Background(), client, Track{Artist: "A", Album: "B", Title: "C"}, Options{Retries: 2})
	if err != nil || attempts != 3 || client.calls != 3 {
		t.Fatalf("attempts = %d, calls = %d, err = %v", attempts, client.calls, err)
	}
}
