package timeout

import "context"

// SlowFunction can be any network call, I/O operation etc
type SlowFunction func(string) (string, error)

// WithContext adds a context to the SlowFunction
type WithContext func(context.Context, string) (string, error)

// Timeout consumes a SlowFunction & returns a WithContext
func Timeout(f SlowFunction) WithContext {
	return func(ctx context.Context, arg string) (string, error) {
		chres := make(chan string)
		cherr := make(chan error)

		go func() {
			res, err := f(arg)
			chres <- res
			cherr <- err
		}()

		select {
		case res := <-chres:
			return res, <-cherr
		case <-ctx.Done():
			return "", ctx.Err()
		}
	}
}
