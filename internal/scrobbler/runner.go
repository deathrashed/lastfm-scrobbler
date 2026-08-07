package scrobbler

import (
	"context"
	"time"

	"github.com/deathrashed/lastfm-scrobbler/internal/lastfm"
)

type Track struct {
	Artist string
	Title  string
	Album  string
}

type Options struct {
	Retries    int
	RetryDelay time.Duration
}

func RunOne(ctx context.Context, client lastfm.Client, track Track, options Options) (int, error) {
	attempts := options.Retries + 1
	if attempts < 1 {
		attempts = 1
	}
	timestamp := time.Now().Unix()
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return attempt - 1, err
		}
		callCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		err = client.Scrobble(callCtx, track.Artist, track.Title, track.Album, timestamp)
		cancel()
		if ctxErr := ctx.Err(); ctxErr != nil {
			return attempt, ctxErr
		}
		if err == nil {
			return attempt, nil
		}
		if attempt < attempts && options.RetryDelay > 0 {
			timer := time.NewTimer(options.RetryDelay)
			select {
			case <-ctx.Done():
				timer.Stop()
				return attempt, ctx.Err()
			case <-timer.C:
			}
		}
	}
	return attempts, err
}

func Wait(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return ctx.Err()
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
