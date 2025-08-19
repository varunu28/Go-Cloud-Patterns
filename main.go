package main

import (
	"context"
	"fmt"
	"go-cloud-patterns/throttle"
	"time"
)

func main() {
	serviceFunc := func(ctx context.Context) (string, error) {
		return "service response", nil
	}

	t := throttle.Throttle(serviceFunc, 3, 3, 5*time.Second)

	for i := 0; i < 5; i++ {
		response, err := t(context.Background())
		if err != nil {
			fmt.Printf("Received error: %v\n", err.Error())
		} else {
			fmt.Printf("response: %s\n", response)
		}
	}
}
