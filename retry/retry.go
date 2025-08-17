package retry

import (
	"context"
	"log"
	"time"
)

type Effector func(context.Context) (string, error)

func Retry(effector Effector, retries int, delay time.Duration) Effector {
	return func(ctx context.Context) (string, error) {
		for r := 0; ; r++ {
			response, err := effector(ctx)
			if err == nil || r > retries {
				return response, err
			}

			log.Printf("Attempt %d failed. Retrying in %v", r+1, delay)

			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return "", err
			}
		}
	}
}
