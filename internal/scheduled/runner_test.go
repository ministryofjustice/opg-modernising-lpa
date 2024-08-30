package scheduled

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	donordata "github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx           = context.WithValue(context.Background(), (*string)(nil), "value")
	expectedError = errors.New("hey")
	testNow       = time.Now()
	testNowFn     = func() time.Time { return testNow }

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
			Return(returns[i].(Event), err).
			Once()
	}
}

func TestNewRunner(t *testing.T) {
	logger := newMockLogger(t)
	store := newMockScheduledStore(t)
	donorStore := newMockDonorStore(t)
	notifyClient := newMockNotifyClient(t)

	runner := NewRunner(logger, store, donorStore, notifyClient)

	assert.Equal(t, logger, runner.logger)
	assert.Equal(t, store, runner.store)
	assert.Equal(t, donorStore, runner.donorStore)
	assert.Equal(t, notifyClient, runner.notifyClient)
	assert.Equal(t, time.Hour, runner.period)
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
		Return(Event{}, dynamo.NotFoundError{}).
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
	row := Event{
		Action:            99,
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner step started", mock.Anything)
	logger.EXPECT().
		InfoContext(ctx, "runner step finished", mock.Anything)
	logger.EXPECT().
		InfoContext(mock.Anything, "runner action", mock.Anything)
	logger.EXPECT().
		InfoContext(mock.Anything, "runner action success", mock.Anything, mock.Anything, mock.Anything)

	store := newMockScheduledStore(t)
	store.ExpectPops(
		row, nil,
		Event{}, dynamo.NotFoundError{},
		row, nil,
		Event{}, dynamo.NotFoundError{},
		row, nil)

	var runTimes []time.Time
	runner := &Runner{
		now:    time.Now,
		period: period,
		logger: logger,
		store:  store,
		actions: map[Action]ActionFunc{
			Action(99): func(_ context.Context, _ Event) error {
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

func TestRunnerRunWhenStepErrors(t *testing.T) {
	ctx, _ := context.WithTimeout(ctx, period)

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner step started", mock.Anything)
	logger.EXPECT().
		InfoContext(ctx, "runner step finished", mock.Anything)
	logger.EXPECT().
		ErrorContext(ctx, "runner step error", slog.Any("err", expectedError))

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(mock.Anything, testNow).
		Return(Event{}, expectedError).
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

func TestRunnerStep(t *testing.T) {
	row := Event{
		Action:            99,
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner action", slog.String("action", "Action(99)"))
	logger.EXPECT().
		InfoContext(ctx, "runner action success",
			slog.String("action", "Action(99)"),
			slog.String("target_pk", "LPA#an-lpa"),
			slog.String("target_sk", "DONOR#a-donor"))

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(ctx, testNow).
		Return(row, nil).
		Once()
	store.EXPECT().
		Pop(ctx, testNow).
		Return(Event{}, dynamo.NotFoundError{}).
		Once()

	actionFunc := newMockActionFunc(t)
	actionFunc.EXPECT().
		Execute(ctx, row).
		Return(nil)

	runner := &Runner{
		now:    testNowFn,
		logger: logger,
		store:  store,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
	}
	err := runner.step(ctx)
	assert.Nil(t, err)
}

func TestRunnerStepWhenActionErrors(t *testing.T) {
	logger := newMockLogger(t)
	logger.EXPECT().
		InfoContext(ctx, "runner action", slog.String("action", "Action(99)"))
	logger.EXPECT().
		ErrorContext(ctx, "runner action error", slog.String("action", "Action(99)"), slog.Any("err", expectedError))

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(ctx, testNow).
		Return(Event{Action: 99}, nil).
		Once()
	store.EXPECT().
		Pop(ctx, testNow).
		Return(Event{}, dynamo.NotFoundError{}).
		Once()

	actionFunc := newMockActionFunc(t)
	actionFunc.EXPECT().
		Execute(mock.Anything, mock.Anything).
		Return(expectedError)

	runner := &Runner{
		now:    testNowFn,
		logger: logger,
		store:  store,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
	}
	err := runner.step(ctx)
	assert.Nil(t, err)
}

func TestRunnerStepCancelDonorIdentity(t *testing.T) {
	lpaKey := dynamo.LpaKey("an-lpa")
	donorKey := dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor"))
	row := Event{
		TargetLpaKey:      lpaKey,
		TargetLpaOwnerKey: donorKey,
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(ctx, lpaKey, donorKey).
		Return(&donordata.Provided{
			LpaUID:                "lpa-uid",
			Donor:                 donordata.Donor{Email: "donor@example.com"},
			DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
		}, nil)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			LpaUID:                "lpa-uid",
			Donor:                 donordata.Donor{Email: "donor@example.com"},
			DonorIdentityUserData: identity.UserData{Status: identity.StatusExpired},
			Tasks:                 donordata.Tasks{ConfirmYourIdentityAndSign: task.IdentityStateNotStarted},
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, "donor@example.com", "lpa-uid", notify.DonorIdentityCheckExpiredEmail{}).
		Return(nil)

	runner := &Runner{
		donorStore:   donorStore,
		notifyClient: notifyClient,
	}
	err := runner.stepCancelDonorIdentity(ctx, row)

	assert.Nil(t, err)
}

func TestRunnerStepCancelDonorIdentityWhenDonorStoreErrors(t *testing.T) {
	row := Event{
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, expectedError)

	runner := &Runner{
		donorStore: donorStore,
	}
	err := runner.stepCancelDonorIdentity(ctx, row)

	assert.ErrorContains(t, err, "error retrieving donor: hey")
}

func TestRunnerStepCancelDonorIdentityWhenStepIgnored(t *testing.T) {
	testcases := map[string]*donordata.Provided{
		"identity not confirmed": &donordata.Provided{
			DonorIdentityUserData: identity.UserData{Status: identity.StatusFailed},
		},
		"already signed": &donordata.Provided{
			DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			SignedAt:              time.Now(),
		},
	}

	for name, provided := range testcases {
		t.Run(name, func(t *testing.T) {
			lpaKey := dynamo.LpaKey("an-lpa")
			donorKey := dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor"))
			row := Event{
				TargetLpaKey:      lpaKey,
				TargetLpaOwnerKey: donorKey,
			}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				One(ctx, lpaKey, donorKey).
				Return(provided, nil)

			runner := &Runner{
				donorStore: donorStore,
			}
			err := runner.stepCancelDonorIdentity(ctx, row)

			assert.Equal(t, errStepIgnored, err)
		})
	}
}

func TestRunnerStepCancelDonorIdentityWhenNotifySendErrors(t *testing.T) {
	row := Event{
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything).
		Return(&donordata.Provided{
			LpaUID:                "lpa-uid",
			Donor:                 donordata.Donor{Email: "donor@example.com"},
			DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
		}, nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	runner := &Runner{
		donorStore:   donorStore,
		notifyClient: notifyClient,
	}
	err := runner.stepCancelDonorIdentity(ctx, row)

	assert.ErrorIs(t, err, expectedError)
}

func TestRunnerStepCancelDonorIdentityWhenDonorStorePutErrors(t *testing.T) {
	row := Event{
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything).
		Return(&donordata.Provided{
			LpaUID:                "lpa-uid",
			Donor:                 donordata.Donor{Email: "donor@example.com"},
			DonorIdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
		}, nil)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	runner := &Runner{
		donorStore:   donorStore,
		notifyClient: notifyClient,
	}
	err := runner.stepCancelDonorIdentity(ctx, row)

	assert.ErrorIs(t, err, expectedError)
}
