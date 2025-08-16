package circuitbreaker

import (
	"context"
	"errors"
	"sync"
	"time"
)

type Circuit func(context.Context) (string, error)

func Breaker(circuit Circuit, failureThreshold uint) Circuit {
	var consecutiveFailures int = 0
	var lastAttempt = time.Now()
	var m sync.RWMutex

	return func(ctx context.Context) (string, error) {
		m.RLock() // Read lock for reading attempt count

		attemptCount := consecutiveFailures - int(failureThreshold)
		if attemptCount >= 0 {
			shouldRetryAt := lastAttempt.Add(time.Second * 2 << attemptCount)
			if !time.Now().After(shouldRetryAt) {
				m.RUnlock()
				return "", errors.New("service unreachable")
			}
		}

		m.RUnlock()

		response, err := circuit(ctx) // Invoke the request

		m.Lock() // Read-write lock
		defer m.Unlock()

		lastAttempt = time.Now()

		if err != nil {
			consecutiveFailures++
			return response, err
		}

		consecutiveFailures = 0 // Reset consecutive failures if request was successful

		return response, nil
	}
}
