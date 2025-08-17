package retry

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

func TestRetry(t *testing.T) {
	var count int
	var m sync.Mutex
	effector := func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()

		count++
		if count < 3 {
			return "", errors.New("forced error")
		}
		return "response", nil
	}

	r := Retry(effector, 5, 1*time.Millisecond)

	res, err := r(context.Background())

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if res != "response" {
		t.Errorf("expected response 'response', got %s", res)
	}

	m.Lock()
	defer m.Unlock()

	if count != 3 {
		t.Errorf("expected 3 retries, got %d", count)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	var count int
	var m sync.Mutex
	effector := func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()

		count++
		if count < 3 {
			return "", errors.New("forced error")
		}
		return "response", nil
	}

	r := Retry(effector, 5, 1*time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := r(ctx)

	if err == nil {
		t.Error("expected error due to context cancellation, got nil")
	}

	m.Lock()
	defer m.Unlock()
	if count != 1 {
		t.Errorf("expected only one attempt before context cancellation, got %d", count)
	}
}
