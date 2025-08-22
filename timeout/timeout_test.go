package timeout

import (
	"context"
	"testing"
	"time"
)

func TestTimeout_FastFunctionSuccess(t *testing.T) {
	fastFunction := func(string) (string, error) {
		return "result", nil
	}

	ctxt, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	timeout := Timeout(fastFunction)
	res, err := timeout(ctxt, "some input")

	time.Sleep(1 * time.Second)

	if err != nil {
		t.Errorf("expected no error. got %v", err)
	}
	if res != "result" {
		t.Errorf("received incorrect result")
	}
}

func TestTimeout_SlowFunctionSuccess(t *testing.T) {
	slowFunction := func(string) (string, error) {
		time.Sleep(100 * time.Second)
		return "result", nil
	}

	ctxt, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	timeout := Timeout(slowFunction)
	res, err := timeout(ctxt, "some input")

	time.Sleep(1 * time.Second)

	if err == nil {
		t.Error("expected error. got none")
	}
	if err.Error() != context.DeadlineExceeded.Error() {
		t.Errorf("expected context deadline exceeded error. got %v", err)
	}
	if res != "" {
		t.Errorf("expected empty result. got %s", res)
	}
}
