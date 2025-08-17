package circuitbreaker

import (
	"context"
	"sync"
	"time"
)

// DebounceFirst returns a circuit breaker that only allows the first call to execute
// within the specified duration. Subsequent calls within that duration will return
// the result of the first call without executing the circuit again.
func DebounceFirst(circuit Circuit, d time.Duration) Circuit {
	var threshold time.Time
	var result string
	var err error
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		m.Lock()

		defer func() {
			threshold = time.Now().Add(d)
			m.Unlock()
		}()

		// If the threshold has not been reached, return the previous result
		if time.Now().Before(threshold) {
			return result, err
		}

		// Execute the circuit and set the threshold for future calls
		result, err = circuit(ctx)

		return result, err
	}
}

// DebounceLast returns a circuit breaker that allows the last call to execute
// after a specified duration. It will wait for the duration to pass before executing
// the circuit again, and it will return the result of the last call made within that duration.
// If the context is done before the duration, it will return the context error.
func DebounceLast(circuit Circuit, d time.Duration) Circuit {
	var threshold time.Time = time.Now()
	var ticker *time.Ticker
	var result string
	var err error
	var once sync.Once
	var m sync.Mutex

	return func(ctx context.Context) (string, error) {
		m.Lock()
		defer m.Unlock()

		threshold = time.Now().Add(d)

		once.Do(func() {
			ticker = time.NewTicker(time.Millisecond * 100)

			go func() {
				defer func() {
					m.Lock()
					ticker.Stop()
					once = sync.Once{}
					m.Unlock()
				}()

				for {
					select {
					case <-ticker.C:
						m.Lock()
						if time.Now().After(threshold) {
							result, err = circuit(ctx)
							m.Unlock()
							return
						}
						m.Unlock()
					case <-ctx.Done():
						m.Lock()
						result, err = "", ctx.Err()
						m.Unlock()
						return
					}
				}
			}()
		})
		return result, err
	}
}
