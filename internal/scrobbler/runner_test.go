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
	blockScrobble     bool
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
func (c *runnerClient) Scrobble(ctx context.Context, _ string, _ string, _ string, _ int64) error {
	c.calls++
	if c.blockScrobble {
		<-ctx.Done()
		return ctx.Err()
	}
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

func TestRunOneCancelsDuringRequest(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &runnerClient{blockScrobble: true}
	done := make(chan error, 1)
	go func() {
		_, err := RunOne(ctx, client, Track{Artist: "A", Album: "B", Title: "C"}, Options{Retries: 3})
		done <- err
	}()
	time.Sleep(5 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("RunOne error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("RunOne did not stop after cancellation")
	}
	if client.calls != 1 {
		t.Fatalf("calls = %d, want 1", client.calls)
	}
}

func TestRunOneCancelsDuringRetryDelay(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	client := &runnerClient{remainingFailures: 3}
	done := make(chan error, 1)
	go func() {
		_, err := RunOne(ctx, client, Track{Artist: "A", Album: "B", Title: "C"}, Options{Retries: 3, RetryDelay: time.Second})
		done <- err
	}()
	time.Sleep(5 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("RunOne error = %v, want context.Canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("retry delay did not stop after cancellation")
	}
	if client.calls != 1 {
		t.Fatalf("calls = %d, want no post-cancel retry", client.calls)
	}
}

func TestWaitCancelsDuringInterval(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- Wait(ctx, time.Second) }()
	time.Sleep(5 * time.Millisecond)
	cancel()
	if err := <-done; !errors.Is(err, context.Canceled) {
		t.Fatalf("Wait error = %v, want context.Canceled", err)
	}
}
