package circuitbreaker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestBreaker_AllSuccess(t *testing.T) {
	circuit := func(ctx context.Context) (string, error) {
		return "OK", nil
	}

	breaker := Breaker(circuit, 3)

	for i := 0; i < 5; i++ {
		resp, err := breaker(context.Background())
		if err != nil || resp != "OK" {
			t.Fatalf("unexpected error: %v, resp: %s", err, resp)
		}
	}
}

func TestBreaker_CompleteFailure(t *testing.T) {
	failureCircuit := func(ctx context.Context) (string, error) {
		return "", errors.New("complete failure")
	}

	breaker := Breaker(failureCircuit, 3)

	for i := 0; i < 5; i++ {
		resp, err := breaker(context.Background())
		if err == nil || resp != "" {
			t.Fatalf("expected error but got: %v", resp)
		}
		if i >= 3 && err.Error() != "service unreachable" {
			t.Fatalf("expected 'service unreachable' error, got: %v", err)
		}
	}
}

func TestBreaker_PartialFailure(t *testing.T) {
	var attempts = 0
	partialFailCircuit := func(ctx context.Context) (string, error) {
		if attempts < 3 {
			attempts++
			return "", errors.New("partial failure")
		} else {
			return "OK", nil
		}
	}

	breaker := Breaker(partialFailCircuit, 3)

	// expected error for first 3 attempts
	for i := 0; i < 3; i++ {
		resp, err := breaker(context.Background())
		if i < 3 && (err == nil || resp != "") {
			t.Fatalf("expected error & empty response for first 3 attempts. Got %s", resp)
		}
	}

	// circuit breaker opens after 3 consecutive failures
	resp, err := breaker(context.Background())
	if err == nil || resp != "" || err.Error() != "service unreachable" {
		t.Fatalf("expected 'service unreachable' error on 4th attempt, got: %v, resp: %s", err, resp)

	}

	// sleep in order for circuit breaker to move in half-open state
	time.Sleep(time.Second * 3)
	resp, err = breaker(context.Background())
	if err != nil || resp != "OK" {
		t.Fatalf("unexpected error: %v, resp: %s", err, resp)
	}
}
