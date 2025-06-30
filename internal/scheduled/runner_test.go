package scheduled

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx               = context.WithValue(context.Background(), (*string)(nil), "value")
	expectedError     = errors.New("hey")
	testNow           = time.Date(2000, time.January, 2, 12, 13, 14, 15, time.UTC)
	testNowFn         = func() time.Time { return testNow }
	testSinceDuration = time.Millisecond * 5
	testSinceFn       = func(t time.Time) time.Duration { return testSinceDuration }
	testEvent         = &Event{
		Action:            99,
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
		LpaUID:            "lpa-uid",
	}
)

func (m *mockScheduledStore) ExpectPops(returns ...any) {
	for i := 0; i < len(returns); i += 2 {
		var ev *Event
		if returns[i] != nil {
			ev = returns[i].(*Event)
		}

		var err error
		if returns[i+1] != nil {
			err = returns[i+1].(error)
		}

		m.EXPECT().Pop(mock.Anything, mock.Anything).Return(ev, err).Once()
	}
}

func TestNewRunner(t *testing.T) {
	logger := newMockLogger(t)
	store := newMockScheduledStore(t)
	donorStore := newMockDonorStore(t)
	certificateProviderStore := newMockCertificateProviderStore(t)
	attorneyStore := newMockAttorneyStore(t)
	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	notifyClient := newMockNotifyClient(t)
	eventClient := newMockEventClient(t)
	metricsClient := newMockMetricsClient(t)
	bundle := newMockBundle(t)

	runner := NewRunner(logger, store, donorStore, certificateProviderStore, attorneyStore, lpaStoreResolvingService, notifyClient, eventClient, bundle, metricsClient, true, "certificateProviderStartURL", "attorneyStartURL", "appPublicURL")

	assert.Equal(t, logger, runner.logger)
	assert.Equal(t, store, runner.store)
	assert.Equal(t, donorStore, runner.donorStore)
	assert.Equal(t, certificateProviderStore, runner.certificateProviderStore)
	assert.Equal(t, attorneyStore, runner.attorneyStore)
	assert.Equal(t, lpaStoreResolvingService, runner.lpaStoreResolvingService)
	assert.Equal(t, notifyClient, runner.notifyClient)
	assert.Equal(t, metricsClient, runner.metricsClient)
	assert.Equal(t, true, runner.metricsEnabled)
	assert.Equal(t, "certificateProviderStartURL", runner.certificateProviderStartURL)
	assert.Equal(t, "attorneyStartURL", runner.attorneyStartURL)
	assert.Equal(t, "appPublicURL"+page.PathCertificateProviderEnterAccessCodeOptOut.Format(), runner.certificateProviderOptOutURL)
	assert.Equal(t, "appPublicURL"+page.PathAttorneyEnterAccessCodeOptOut.Format(), runner.attorneyOptOutURL)
}

func (m *mockMetricsClient) assertPutMetrics(processed, ignored, errored float64, err error) {
	expected := &cloudwatch.PutMetricDataInput{
		Namespace: aws.String("schedule-runner"),
		MetricData: []types.MetricDatum{
			{
				MetricName: aws.String("TasksProcessed"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(processed),
			},
			{
				MetricName: aws.String("TasksIgnored"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(ignored),
			},
			{
				MetricName: aws.String("Errors"),
				Unit:       types.StandardUnitCount,
				Value:      aws.Float64(errored),
			},
			{
				MetricName: aws.String("ProcessingTime"),
				Unit:       types.StandardUnitMilliseconds,
				Value:      aws.Float64(float64(testSinceDuration.Milliseconds())),
			},
		},
	}

	m.EXPECT().
		PutMetrics(ctx, expected).
		Return(err)
}

func TestRunnerRun(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner action", slog.String("action", "Action(99)"))
	logger.EXPECT().
		InfoContext(ctx, "runner action success",
			slog.String("action", "Action(99)"),
			slog.String("target_pk", "LPA#an-lpa"),
			slog.String("target_sk", "DONOR#a-donor"))
	logger.EXPECT().
		InfoContext(ctx, "no scheduled tasks to process")

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(ctx, testNow).
		Return(testEvent, nil).
		Once()
	store.EXPECT().
		Pop(ctx, testNow).
		Return(nil, dynamo.NotFoundError{}).
		Once()

	waiter := newMockWaiter(t)
	waiter.EXPECT().Reset()

	actionFunc := newMockActionFunc(t)
	actionFunc.EXPECT().
		Execute(ctx, testEvent).
		Return(nil)

	metricsClient := newMockMetricsClient(t)
	metricsClient.assertPutMetrics(1, 0, 0, nil)

	runner := &Runner{
		now:    testNowFn,
		since:  testSinceFn,
		logger: logger,
		store:  store,
		waiter: waiter,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
		metricsClient:  metricsClient,
		metricsEnabled: true,
	}

	err := runner.Run(ctx)

	assert.Nil(t, err)
}

func TestRunnerRunWhenStepErrors(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(ctx, "error getting scheduled task", slog.Any("err", expectedError))

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(mock.Anything, testNow).
		Return(nil, expectedError).
		Once()

	waiter := newMockWaiter(t)
	waiter.EXPECT().Reset()
	waiter.EXPECT().
		Wait().
		Return(expectedError)

	runner := &Runner{
		now:    testNowFn,
		logger: logger,
		store:  store,
		waiter: waiter,
	}

	err := runner.Run(ctx)
	assert.Equal(t, expectedError, err)
}

func TestRunnerRunWhenActionIgnored(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner action", slog.String("action", "Action(99)"))
	logger.EXPECT().
		InfoContext(ctx, "runner action ignored",
			slog.String("action", "Action(99)"),
			slog.String("target_pk", "LPA#an-lpa"),
			slog.String("target_sk", "DONOR#a-donor"))
	logger.EXPECT().
		InfoContext(ctx, "no scheduled tasks to process")

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(ctx, testNow).
		Return(testEvent, nil).
		Once()
	store.EXPECT().
		Pop(ctx, testNow).
		Return(nil, dynamo.NotFoundError{}).
		Once()

	waiter := newMockWaiter(t)
	waiter.EXPECT().Reset()

	actionFunc := newMockActionFunc(t)
	actionFunc.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(errStepIgnored)

	metricsClient := newMockMetricsClient(t)
	metricsClient.assertPutMetrics(0, 1, 0, nil)

	runner := &Runner{
		now:    testNowFn,
		since:  testSinceFn,
		logger: logger,
		store:  store,
		waiter: waiter,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
		metricsClient:  metricsClient,
		metricsEnabled: true,
	}
	err := runner.Run(ctx)

	assert.Nil(t, err)
}

func TestRunnerRunWhenActionErrors(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner action", slog.String("action", "Action(99)"))
	logger.EXPECT().
		ErrorContext(ctx, "runner action error",
			slog.String("action", "Action(99)"),
			slog.String("target_pk", "LPA#an-lpa"),
			slog.String("target_sk", "DONOR#a-donor"),
			slog.Any("err", expectedError))
	logger.EXPECT().
		InfoContext(ctx, "no scheduled tasks to process")

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(ctx, testNow).
		Return(testEvent, nil).
		Once()
	store.EXPECT().
		Pop(ctx, testNow).
		Return(nil, dynamo.NotFoundError{}).
		Once()

	waiter := newMockWaiter(t)
	waiter.EXPECT().Reset()

	actionFunc := newMockActionFunc(t)
	actionFunc.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	metricsClient := newMockMetricsClient(t)
	metricsClient.assertPutMetrics(0, 0, 1, nil)

	runner := &Runner{
		now:    testNowFn,
		since:  testSinceFn,
		logger: logger,
		store:  store,
		waiter: waiter,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
		metricsClient:  metricsClient,
		metricsEnabled: true,
	}
	err := runner.Run(ctx)

	assert.Nil(t, err)
}

func TestRunnerRunWhenWaitingError(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner action", slog.String("action", "Action(99)"))
	logger.EXPECT().
		ErrorContext(ctx, "error getting scheduled task", slog.Any("err", expectedError))
	logger.EXPECT().
		InfoContext(ctx, "runner action success",
			slog.String("action", "Action(99)"),
			slog.String("target_pk", "LPA#an-lpa"),
			slog.String("target_sk", "DONOR#a-donor"))
	logger.EXPECT().
		InfoContext(ctx, "no scheduled tasks to process")

	store := newMockScheduledStore(t)
	store.ExpectPops(
		nil, expectedError,
		testEvent, nil,
		nil, dynamo.NotFoundError{})

	waiter := newMockWaiter(t)
	waiter.EXPECT().Reset().Twice()
	waiter.EXPECT().Wait().Return(nil).Once()

	actionFunc := newMockActionFunc(t)
	actionFunc.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(nil)

	metricsClient := newMockMetricsClient(t)
	metricsClient.assertPutMetrics(1, 0, 0, nil)

	runner := &Runner{
		now:    testNowFn,
		since:  testSinceFn,
		logger: logger,
		store:  store,
		waiter: waiter,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
		metricsClient:  metricsClient,
		metricsEnabled: true,
	}

	err := runner.Run(ctx)

	assert.Nil(t, err)
}

func TestRunnerRunWhenMetricsDisabled(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, mock.Anything, mock.Anything)
	logger.EXPECT().
		InfoContext(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	logger.EXPECT().
		InfoContext(ctx, mock.Anything)

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(ctx, mock.Anything).
		Return(testEvent, nil).
		Once()
	store.EXPECT().
		Pop(ctx, mock.Anything).
		Return(nil, dynamo.NotFoundError{}).
		Once()

	waiter := newMockWaiter(t)
	waiter.EXPECT().Reset()

	actionFunc := newMockActionFunc(t)
	actionFunc.EXPECT().
		Execute(ctx, mock.Anything).
		Return(nil)

	runner := &Runner{
		now:    testNowFn,
		since:  testSinceFn,
		logger: logger,
		store:  store,
		waiter: waiter,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
		metricsClient:  nil,
		metricsEnabled: false,
	}

	err := runner.Run(ctx)

	assert.Nil(t, err)
}

func TestRunnerRunWhenMetricsClientError(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, mock.Anything, mock.Anything)
	logger.EXPECT().
		InfoContext(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	logger.EXPECT().
		InfoContext(ctx, mock.Anything)
	logger.EXPECT().
		ErrorContext(ctx, "error putting metrics", slog.Any("err", expectedError))

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(ctx, mock.Anything).
		Return(testEvent, nil).
		Once()
	store.EXPECT().
		Pop(ctx, mock.Anything).
		Return(nil, dynamo.NotFoundError{}).
		Once()

	waiter := newMockWaiter(t)
	waiter.EXPECT().Reset()

	actionFunc := newMockActionFunc(t)
	actionFunc.EXPECT().
		Execute(ctx, mock.Anything).
		Return(nil)

	metricsClient := newMockMetricsClient(t)
	metricsClient.assertPutMetrics(1, 0, 0, expectedError)

	runner := &Runner{
		now:    testNowFn,
		since:  testSinceFn,
		logger: logger,
		store:  store,
		waiter: waiter,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
		metricsClient:  metricsClient,
		metricsEnabled: true,
	}

	err := runner.Run(ctx)

	assert.Equal(t, expectedError, err)
}

func TestRunnerRunWhenConditionalCheckFailsAndWaiterErrors(t *testing.T) {
	store := newMockScheduledStore(t)
	store.ExpectPops(
		nil, dynamo.ConditionalCheckFailedError{},
		nil, dynamo.ConditionalCheckFailedError{})

	waiter := newMockWaiter(t)
	waiter.EXPECT().Reset().Once()
	waiter.EXPECT().Wait().Return(nil).Once()
	waiter.EXPECT().Wait().Return(expectedError).Once()

	logger := newMockLogger(t)
	logger.EXPECT().
		ErrorContext(ctx, "error getting scheduled task", slog.Any("err", dynamo.ConditionalCheckFailedError{}))

	runner := &Runner{
		now:    testNowFn,
		store:  store,
		waiter: waiter,
		logger: logger,
	}

	err := runner.Run(ctx)
	assert.Equal(t, expectedError, err)
}
