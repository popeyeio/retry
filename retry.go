package retry

import (
	"context"
	"errors"
	"fmt"
	"time"
)

var (
	ErrInvalidTimes   = errors.New("[retry] invalid times")
	ErrInvalidTimeout = errors.New("[retry] invalid timeout")
	ErrRetryTimeout   = errors.New("[retry] retry timeout")
)

type RetryableFunc func() error

func Do(info string, times int, timeout time.Duration, run RetryableFunc, stopErrors ...error) (err error) {
	if times < 1 {
		return ErrInvalidTimes
	}
	if timeout <= 0 {
		return ErrInvalidTimeout
	}
	if run == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	timesChan := make(chan int, 1)
	defer close(timesChan)

	errChan := make(chan error, 1)
	go func() {
		for i := range timesChan {
			if i > 0 {
				fmt.Printf("[retry] retry %s for %d times\n", info, i)
			}
			errChan <- run()
		}
	}()

	for i := 0; i < times; i++ {
		timesChan <- i

		select {
		case <-ctx.Done():
			return ErrRetryTimeout
		case err = <-errChan:
			if err == nil || contains(stopErrors, err) {
				return
			}
		}
	}
	return
}

func contains(errs []error, err error) bool {
	for _, e := range errs {
		if e == err {
			return true
		}
	}
	return false
}
