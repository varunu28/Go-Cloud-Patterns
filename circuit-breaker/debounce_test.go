package circuitbreaker

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"
)

func TestDebounceFirst(t *testing.T) {
	circuit := func(ctx context.Context) (string, error) {
		return strconv.FormatInt(time.Now().UnixMilli(), 10), nil
	}

	debouncedCircuit := DebounceFirst(circuit, 1000*time.Millisecond)

	result, err := debouncedCircuit(context.Background())
	if err != nil {
		t.Errorf("Expected result, got error %v", err)
	}
	firstCallTime, err := strconv.ParseInt(result, 10, 64)
	if err != nil {
		t.Errorf("Expected a valid time string, got %s", result)
	}

	// Sleep within the debounce period
	time.Sleep(800 * time.Millisecond)

	result2, err := debouncedCircuit(context.Background())
	if err != nil {
		t.Errorf("Expected result, got error %v", err)
	}
	secondCallTime, err := strconv.ParseInt(result2, 10, 64)
	if err != nil {
		t.Errorf("Expected a valid time string, got %s", result2)
	}

	// The second call should return the same result as the first call as it is within the debounce period
	if firstCallTime != secondCallTime {
		t.Errorf("Expected second call time to be the same as first call time, got %v and %v", firstCallTime, secondCallTime)
	}

	// Sleep in order to ensure the debounce period has passed
	time.Sleep(300 * time.Millisecond)

	// Call again. This time it should execute the circuit again
	result3, err := debouncedCircuit(context.Background())
	if err != nil {
		t.Errorf("Expected result, got error %v", err)
	}
	thirdCallTime, err := strconv.ParseInt(result3, 10, 64)
	if err != nil {
		t.Errorf("Expected a valid time string, got %s", result3)
	}

	// This time we should have a different time in the result as the debounce period has passed
	if firstCallTime == thirdCallTime {
		t.Errorf("Expected third call time to be different from last call time, got %v", thirdCallTime)
	}
}

func TestDebounceLast(t *testing.T) {
	var executionCount int
	var mu sync.Mutex
	circuit := func(ctx context.Context) (string, error) {
		mu.Lock()
		executionCount++
		mu.Unlock()
		return strconv.FormatInt(time.Now().UnixMilli(), 10), nil
	}

	debouncedCircuit := DebounceLast(circuit, 200*time.Millisecond)

	// Each call should return the empty result & not trigger the circuit execution
	for i := 0; i < 5; i++ {
		result, err := debouncedCircuit(context.Background())
		if err != nil {
			t.Errorf("Expected result, got error %v", err)
		}
		if result != "" {
			t.Errorf("Expected empty result, got %s", result)
		}
		time.Sleep(50 * time.Millisecond) // Keep resetting the debounce timer
	}

	// Verify that the circuit has not executed yet
	mu.Lock()
	count := executionCount
	mu.Unlock()
	if count != 0 {
		t.Errorf("Expected circuit to not execute yet, executed %d times", count)
	}

	// Wait for the debounce period to pass
	time.Sleep(300 * time.Millisecond)

	// Now the circuit should execute
	mu.Lock()
	count = executionCount
	mu.Unlock()
	if count != 1 {
		t.Errorf("Expected circuit to execute once after debounce period, executed %d times", count)
	}

	// Now all calls should return the result of the last execution
	result, err := debouncedCircuit(context.Background())
	if err != nil {
		t.Errorf("Expected result, got error %v", err)
	}
	if result == "" {
		t.Errorf("Expected non-empty result, got %s", result)
	}
}

func TestDebounceLast_ContextCancellation(t *testing.T) {
	var executionCount int
	var mu sync.Mutex
	circuit := func(ctx context.Context) (string, error) {
		mu.Lock()
		executionCount++
		mu.Unlock()
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
			return strconv.FormatInt(time.Now().UnixMilli(), 10), nil
		}
	}

	debouncedCircuit := DebounceLast(circuit, 300*time.Millisecond)

	// Start a context with a timeout that is less than the debounce period
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	// With this call, the threshold will become 300 + 300ms = 600ms
	result, err := debouncedCircuit(ctx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if result != "" {
		t.Errorf("Expected empty result, got %s", result)
	}

	// Wait for the context to timeout but less than the debounce period
	time.Sleep(250 * time.Millisecond)

	_, err2 := debouncedCircuit(ctx)
	if err2 == nil || err2 != context.DeadlineExceeded {
		t.Errorf("Expected context deadline exceeded error, got %v", err2)
	}

	// Verify that the circuit has not executed yet
	mu.Lock()
	count := executionCount
	mu.Unlock()
	if count != 0 {
		t.Errorf("Expected circuit to not execute yet, executed %d times", count)
	}
}
