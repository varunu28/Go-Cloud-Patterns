package main

import (
	"context"
	"errors"
	"fmt"
	"go-cloud-patterns/retry"
	"time"
)

var count int

func EmulateTransientError(ctx context.Context) (string, error) {
	count++

	if count < 3 {
		return "intentional failure", errors.New("error")
	}

	return "success", nil
}

func main() {
	r := retry.Retry(EmulateTransientError, 5, 1*time.Second)

	res, err := r(context.Background())

	fmt.Println(res, err)
}
