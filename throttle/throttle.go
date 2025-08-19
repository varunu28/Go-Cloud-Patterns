package throttle

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type Effector func(context.Context) (string, error)

// Throttle is used to limit number of calls in a specified duration
func Throttle(e Effector, max uint, refill uint, d time.Duration) Effector {
	var tokens = max
	var once sync.Once
	var mu sync.Mutex

	return func(ctx context.Context) (string, error) {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}

		once.Do(func() {
			ticker := time.NewTicker(d)

			go func() {
				defer ticker.Stop()

				for {
					select {
					case <-ctx.Done():
						return

					// Gets executed at every d duration at which we refil the token bucket
					case <-ticker.C:
						mu.Lock()
						t := min(tokens+refill, max)
						tokens = t
						mu.Unlock()
					}
				}
			}()
		})

		mu.Lock()
		defer mu.Unlock()

		// Check if we have tokens available
		if tokens <= 0 {
			return "", fmt.Errorf("too many calls")
		}

		tokens--

		return e(ctx)
	}
}
