package datasource

import (
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"time"
)

const (
	maxAttempts = 5
	baseDelay   = 200 * time.Millisecond
)

func BackoffRetry(
	ctx context.Context,
	maxAttempts int,
	baseDelay time.Duration,
	operation func() error,
) error {
	if maxAttempts < 1 {
		return errors.New("max attempts must be greater than zero")
	}
	if baseDelay < 0 {
		return errors.New("base delay must not be negative")
	}

	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return err
		}

		lastErr = operation()
		if lastErr == nil {
			return nil
		}
		if attempt == maxAttempts-1 {
			break
		}

		delay := time.Duration(1<<attempt) * baseDelay
		halfDelay := delay / 2
		jitter := time.Duration(rand.Int64N(int64(halfDelay) + 1))
		delay = halfDelay + jitter

		timer := time.NewTimer(delay)
		select {
		case <-timer.C:
		case <-ctx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return ctx.Err()
		}
	}

	return fmt.Errorf("all %d attempts failed, last error: %w", maxAttempts, lastErr)
}
