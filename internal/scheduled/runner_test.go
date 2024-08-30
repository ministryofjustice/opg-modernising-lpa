package scheduled

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx       = context.WithValue(context.Background(), (*string)(nil), "value")
	testNow   = time.Now()
	testNowFn = func() time.Time { return testNow }

	// set resolution lower to make tests more accurate, but the clock won't be
	// perfect so 2ms seems a reasonable trade-off
	resolution = 2 * time.Millisecond
	// set period higher to make tests more accurate, but that will make them
	// slower
	period = 20 * resolution
)

func (m *mockScheduledStore) ExpectPops(returns ...any) {
	for i := 0; i < len(returns); i += 2 {
		var err error
		if returns[i+1] != nil {
			err = returns[i+1].(error)
		}

		m.EXPECT().
			Pop(mock.Anything, mock.Anything).
			Return(returns[i].(Row), err).
			Once()
	}
}

func TestRunnerRun(t *testing.T) {
	ctx, _ := context.WithTimeout(ctx, period)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner step started", mock.Anything)
	logger.EXPECT().
		InfoContext(ctx, "runner step finished", mock.Anything)

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(mock.Anything, testNow).
		Return(Row{}, dynamo.NotFoundError{}).
		Once()

	runner := &Runner{
		now:    testNowFn,
		period: time.Hour,
		logger: logger,
		store:  store,
	}

	err := runner.Run(ctx)
	assert.Nil(t, err)
}

func TestRunnerRunWhenPeriodElapses(t *testing.T) {
	ctx, cancel := context.WithTimeout(ctx, 3*period)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner step started", mock.Anything)
	logger.EXPECT().
		InfoContext(ctx, "runner step finished", mock.Anything)
	logger.EXPECT().
		InfoContext(mock.Anything, "runner action", mock.Anything).
		Times(3)

	store := newMockScheduledStore(t)
	store.ExpectPops(
		Row{Action: 99}, nil,
		Row{}, dynamo.NotFoundError{},
		Row{Action: 99}, nil,
		Row{}, dynamo.NotFoundError{},
		Row{Action: 99}, nil)

	var runTimes []time.Time
	runner := &Runner{
		now:    time.Now,
		period: period,
		logger: logger,
		store:  store,
		actions: map[Action]func(context.Context, Row) error{
			Action(99): func(_ context.Context, _ Row) error {
				if runTimes = append(runTimes, time.Now()); len(runTimes) == 3 {
					cancel()
				}
				return nil
			},
		},
	}

	err := runner.Run(ctx)
	assert.Nil(t, err)
	assert.Len(t, runTimes, 3)
	assert.InDelta(t, period, runTimes[1].Sub(runTimes[0]), float64(resolution))
	assert.InDelta(t, period, runTimes[2].Sub(runTimes[1]), float64(resolution))
}
