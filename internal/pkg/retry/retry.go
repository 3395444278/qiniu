package retry

import (
	"context"
	"time"
)

func WithRetry(ctx context.Context, fn func() error, maxAttempts int, delay time.Duration) error {
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := fn(); err == nil {
				return nil
			} else {
				lastErr = err
				time.Sleep(delay * time.Duration(attempt+1))
			}
		}
	}
	return lastErr
}
