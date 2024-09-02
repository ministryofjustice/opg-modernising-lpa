package scheduled

import (
	"slices"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWaiterReset(t *testing.T) {
	w := &waiter{retries: 1}
	w.Reset()
	assert.Equal(t, 0, w.retries)
}

func TestWaiterWait(t *testing.T) {
	var calledDur time.Duration
	w := &waiter{backoff: time.Second, sleep: func(dur time.Duration) { calledDur = dur }, maxRetries: 2}

	err := w.Wait()
	assert.Nil(t, err)
	assert.Equal(t, time.Second, calledDur)
}

func TestWaiterWaitWhenRetries(t *testing.T) {
	for range 5 {
		var calledDur time.Duration
		w := &waiter{backoff: time.Second, sleep: func(dur time.Duration) { calledDur = dur }, maxRetries: 3, retries: 2}

		err := w.Wait()
		assert.Nil(t, err)
		assert.True(t, slices.Contains([]time.Duration{time.Second, 2 * time.Second, 3 * time.Second}, calledDur))
	}
}

func TestWaiterWaitWhenRetriesExceedsMax(t *testing.T) {
	w := &waiter{backoff: time.Second, sleep: func(dur time.Duration) {}, maxRetries: 2, retries: 2}

	err := w.Wait()
	assert.ErrorContains(t, err, "waiter exceeded max retries")
}
