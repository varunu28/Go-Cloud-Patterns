package throttle

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestThrottle(t *testing.T) {
	var count int
	var m sync.Mutex
	serviceFunc := func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()
		count++

		return "service response", nil
	}

	throttle := Throttle(serviceFunc, 3, 3, 3*time.Second)

	for i := 0; i < 5; i++ {
		resp, err := throttle(context.Background())

		if i < 3 && err != nil {
			t.Errorf("expected no error. got %v", err)
		}
		if i < 3 && resp != "service response" {
			t.Errorf("got incorrect response: %s", resp)
		}

		if i >= 3 && err == nil {
			t.Errorf("expected error for too many requests. got nil")
		}
		if i >= 3 && resp != "" {
			t.Errorf("expected empty response. Got %s", resp)
		}
	}

	m.Lock()
	defer m.Unlock()
	if count != 3 {
		t.Errorf("expected service to be invoked 3 times. Got %d invocations", count)
	}
}

func TestThrottle_WithRefil(t *testing.T) {
	var count int
	var m sync.Mutex
	serviceFunc := func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()
		count++

		return "service response", nil
	}

	throttle := Throttle(serviceFunc, 3, 3, 2*time.Second)
	for i := 0; i < 4; i++ {
		resp, err := throttle(context.Background())
		if i < 3 {
			if err != nil {
				t.Errorf("expected no error. Got %v", err)
			}
			if resp != "service response" {
				t.Errorf("expected service response. got %s", resp)
			}
		} else {
			if err == nil {
				t.Error("expected error. got nil")
			}
			if resp != "" {
				t.Errorf("expected empty response. got %s", resp)
			}
		}
	}

	// Sleep for period after which bucket is refilled
	time.Sleep(3 * time.Second)

	resp, err := throttle(context.Background())
	if err != nil {
		t.Errorf("expected no error. Got %v", err)
	}
	if resp != "service response" {
		t.Errorf("expected service response. got %s", resp)
	}

	m.Lock()
	defer m.Unlock()
	// 3 with initial bucket + 1 with refilled bucket
	if count != 4 {
		t.Errorf("expected 4 invocations. got %d", count)
	}
}

func TestThrottle_ErrorContext(t *testing.T) {
	var count int
	var m sync.Mutex
	serviceFunc := func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()
		count++

		return "service response", nil
	}

	throttle := Throttle(serviceFunc, 3, 3, 2*time.Second)

	ctx, cancel := context.WithCancelCause(context.Background())
	cancel(fmt.Errorf("custom error context"))

	for i := 0; i < 3; i++ {
		resp, err := throttle(ctx)
		if err == nil || err.Error() != context.Canceled.Error() {
			t.Errorf("expected error with cancellation. got %v", err)
		}
		if resp != "" {
			t.Errorf("expected empty response. got %s", resp)
		}
	}

	m.Lock()
	defer m.Unlock()

	if count != 0 {
		t.Errorf("expected no invocation for service function. got %d", count)
	}
}

func TestThrottle_ExpiredContext(t *testing.T) {
	var count int
	var m sync.Mutex
	serviceFunc := func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()
		count++

		return "service response", nil
	}

	throttle := Throttle(serviceFunc, 3, 3, 2*time.Second)

	// Context that expires after 1 second
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	resp, err := throttle(ctx)
	if err != nil {
		t.Errorf("expected no error. got %v", err)
	}
	if resp != "service response" {
		t.Errorf("expected service response. got %s", resp)
	}

	// sleep in order for context to expire
	time.Sleep(2 * time.Second)

	resp2, err := throttle(ctx)
	if err == nil || err.Error() != context.DeadlineExceeded.Error() {
		t.Errorf("expected error. got %v", err)
	}
	if resp2 != "" {
		t.Errorf("expected empty response. got %s", resp2)
	}

	// We expect only 1 service invocation. After the context is cancelled the throttle will not invoke the service
	m.Lock()
	defer m.Unlock()

	if count != 1 {
		t.Errorf("expected only 1 invocation. got %d", count)
	}
}
