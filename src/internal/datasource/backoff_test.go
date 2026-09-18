package datasource

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBackoffRetry(t *testing.T) {
	t.Run("succeeds after transient failures", func(t *testing.T) {
		attempts := 0
		err := BackoffRetry(context.Background(), 3, 0, func() error {
			attempts++
			if attempts < 3 {
				return errors.New("temporary error")
			}
			return nil
		})

		if err != nil {
			t.Fatalf("BackoffRetry() error = %v", err)
		}
		if attempts != 3 {
			t.Fatalf("attempts = %d, want 3", attempts)
		}
	})

	t.Run("returns last error", func(t *testing.T) {
		operationErr := errors.New("database unavailable")
		err := BackoffRetry(context.Background(), 2, 0, func() error {
			return operationErr
		})

		if !errors.Is(err, operationErr) {
			t.Fatalf("BackoffRetry() error = %v, want wrapped %v", err, operationErr)
		}
	})

	t.Run("rejects invalid attempts", func(t *testing.T) {
		if err := BackoffRetry(context.Background(), 0, 0, func() error { return nil }); err == nil {
			t.Fatal("BackoffRetry() error = nil, want validation error")
		}
	})

	t.Run("stops when context is canceled", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		operationStarted := make(chan struct{})
		operationErr := errors.New("temporary error")

		go func() {
			<-operationStarted
			cancel()
		}()

		err := BackoffRetry(ctx, 2, time.Hour, func() error {
			close(operationStarted)
			return operationErr
		})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("BackoffRetry() error = %v, want %v", err, context.Canceled)
		}
	})
}
