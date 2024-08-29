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
)

func TestRunnerRun(t *testing.T) {
	ctx, cancel := context.WithTimeout(ctx, time.Second)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner next step scheduled", mock.Anything)
	logger.EXPECT().
		InfoContext(ctx, "runner step started", mock.Anything)
	logger.EXPECT().
		InfoContext(ctx, "runner step finished", mock.Anything)

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(ctx, testNow).
		Return(Row{}, dynamo.NotFoundError{})

	// it runs on start
	runner := &Runner{
		now:      testNowFn,
		lastStep: testNow.Add(-time.Hour),
		logger:   logger,
		store:    store,
		actions: map[Action]func(context.Context, Row) error{
			Action(99): func(_ context.Context, _ Row) error {
				cancel()
				return nil
			},
		},
	}

	go func() {
		err := runner.Run(ctx)
		assert.Nil(t, err)
	}()

	// it runs after period
}
