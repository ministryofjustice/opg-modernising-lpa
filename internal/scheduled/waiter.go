package scheduled

import (
	"errors"
	"math/rand/v2"
	"time"
)

type waiter struct {
	backoff    time.Duration
	sleep      func(time.Duration)
	maxRetries int
	retries    int
}

func (w *waiter) Reset() {
	w.retries = 0
}

func (w *waiter) Wait() error {
	w.retries++
	count := rand.IntN(w.retries) + 1
	w.sleep(time.Duration(count) * w.backoff)

	if w.retries > w.maxRetries {
		return errors.New("waiter exceeded max retries")
	}

	return nil
}
