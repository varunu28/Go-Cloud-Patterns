package main

import (
	"context"
	"errors"
	"fmt"
	circuitbreaker "go-cloud-patterns/circuit-breaker"
	"io"
	"net/http"
	"time"
)

// An implementation of Circuit interface which invokes an endpoint that always returns an error
func httpCircuit(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://localhost:8081/api/v1/verify", nil)
	if err != nil {
		return "", err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("bad response from sever")
	}

	body, _ := io.ReadAll(resp.Body)
	return string(body), nil
}

func main() {
	circuitBreakerRequest := circuitbreaker.Breaker(httpCircuit, 3)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	for i := 0; i < 10; i++ {
		response, err := circuitBreakerRequest(ctx)
		if err != nil {
			fmt.Printf("Attempt %d failed: %v\n", i+1, err)
		} else {
			fmt.Printf("Attempt %d succeeded: %s\n", i+1, response)
			break
		}

		time.Sleep(time.Second)
	}
}
