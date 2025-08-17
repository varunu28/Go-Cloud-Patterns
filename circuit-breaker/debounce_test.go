package circuitbreaker

import (
	"context"
	"strconv"
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

	result2, err := debouncedCircuit(context.Background())
	if err != nil {
		t.Errorf("Expected result, got error %v", err)
	}
	secondCallTime, err := strconv.ParseInt(result2, 10, 64)
	if err != nil {
		t.Errorf("Expected a valid time string, got %s", result2)
	}

	// The second call should return the same result as the first call
	if firstCallTime != secondCallTime {
		t.Errorf("Expected second call time to be the same as first call time, got %v and %v", firstCallTime, secondCallTime)
	}

	// Sleep in order to ensure the debounce period has passed
	time.Sleep(1001 * time.Millisecond)

	// Call again. This time it should execute the circuit again
	result3, err := debouncedCircuit(context.Background())
	if err != nil {
		t.Errorf("Expected result, got error %v", err)
	}
	thirdCallTime, err := strconv.ParseInt(result3, 10, 64)
	if err != nil {
		t.Errorf("Expected a valid time string, got %s", result3)
	}

	// This time we should have a different time in the result
	if secondCallTime == thirdCallTime {
		t.Errorf("Expected third call time to be different from last call time, got %v", thirdCallTime)
	}
}
