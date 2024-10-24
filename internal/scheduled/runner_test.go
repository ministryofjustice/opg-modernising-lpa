package scheduled

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	ctx           = context.WithValue(context.Background(), (*string)(nil), "value")
	expectedError = errors.New("hey")
	testNow       = time.Now()
	testNowFn     = func() time.Time { return testNow }
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
	notifyClient := newMockNotifyClient(t)

	runner := NewRunner(logger, store, donorStore, notifyClient)

	assert.Equal(t, logger, runner.logger)
	assert.Equal(t, store, runner.store)
	assert.Equal(t, donorStore, runner.donorStore)
	assert.Equal(t, notifyClient, runner.notifyClient)
}

func TestRunnerRun(t *testing.T) {
	event := &Event{
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
	logger.EXPECT().
		InfoContext(ctx, "no scheduled tasks to process")

	store := newMockScheduledStore(t)
	store.EXPECT().
		Pop(ctx, testNow).
		Return(event, nil).
		Once()
	store.EXPECT().
		Pop(ctx, testNow).
		Return(nil, dynamo.NotFoundError{}).
		Once()

	waiter := newMockWaiter(t)
	waiter.EXPECT().Reset()

	actionFunc := newMockActionFunc(t)
	actionFunc.EXPECT().
		Execute(ctx, event).
		Return(nil)

	runner := &Runner{
		now:    testNowFn,
		logger: logger,
		store:  store,
		waiter: waiter,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
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
	event := &Event{
		Action:            99,
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

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
		Return(event, nil).
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

	runner := &Runner{
		now:    testNowFn,
		logger: logger,
		store:  store,
		waiter: waiter,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
	}
	err := runner.Run(ctx)
	assert.Nil(t, err)
}

func TestRunnerRunWhenActionErrors(t *testing.T) {
	event := &Event{
		Action:            99,
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

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
		Return(event, nil).
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

	runner := &Runner{
		now:    testNowFn,
		logger: logger,
		store:  store,
		waiter: waiter,
		actions: map[Action]ActionFunc{
			99: actionFunc.Execute,
		},
	}
	err := runner.Run(ctx)
	assert.Nil(t, err)
}

func TestRunnerRunWhenWaitingError(t *testing.T) {
	testcases := []error{
		dynamo.ConditionalCheckFailedError{},
		expectedError,
	}

	for _, waitingError := range testcases {
		t.Run(waitingError.Error(), func(t *testing.T) {
			event := &Event{
				Action:            99,
				TargetLpaKey:      dynamo.LpaKey("an-lpa"),
				TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
			}

			logger := newMockLogger(t)
			logger.EXPECT().
				InfoContext(ctx, "runner action", slog.String("action", "Action(99)"))
			logger.EXPECT().
				ErrorContext(ctx, "error getting scheduled task", slog.Any("err", waitingError))
			logger.EXPECT().
				InfoContext(ctx, "runner action success",
					slog.String("action", "Action(99)"),
					slog.String("target_pk", "LPA#an-lpa"),
					slog.String("target_sk", "DONOR#a-donor"))
			logger.EXPECT().
				InfoContext(ctx, "no scheduled tasks to process")

			store := newMockScheduledStore(t)
			store.ExpectPops(
				nil, waitingError,
				event, nil,
				nil, dynamo.NotFoundError{})

			waiter := newMockWaiter(t)
			waiter.EXPECT().Reset().Twice()
			waiter.EXPECT().Wait().Return(nil).Once()

			actionFunc := newMockActionFunc(t)
			actionFunc.EXPECT().
				Execute(mock.Anything, mock.Anything).
				Return(nil)

			runner := &Runner{
				now:    testNowFn,
				logger: logger,
				store:  store,
				waiter: waiter,
				actions: map[Action]ActionFunc{
					99: actionFunc.Execute,
				},
			}
			err := runner.Run(ctx)
			assert.Nil(t, err)
		})
	}
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

func TestRunnerCancelDonorIdentity(t *testing.T) {
	lpaKey := dynamo.LpaKey("an-lpa")
	donorKey := dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor"))
	event := &Event{
		TargetLpaKey:      lpaKey,
		TargetLpaOwnerKey: donorKey,
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(ctx, lpaKey, donorKey).
		Return(&donordata.Provided{
			LpaUID:           "lpa-uid",
			Donor:            donordata.Donor{Email: "donor@example.com", ContactLanguagePreference: localize.Cy},
			IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
		}, nil)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			LpaUID:           "lpa-uid",
			Donor:            donordata.Donor{Email: "donor@example.com", ContactLanguagePreference: localize.Cy},
			IdentityUserData: identity.UserData{Status: identity.StatusExpired},
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, localize.Cy, "donor@example.com", "lpa-uid", notify.DonorIdentityCheckExpiredEmail{}).
		Return(nil)

	runner := &Runner{
		donorStore:   donorStore,
		notifyClient: notifyClient,
	}
	err := runner.stepCancelDonorIdentity(ctx, event)

	assert.Nil(t, err)
}

func TestRunnerCancelDonorIdentityWhenDonorStoreErrors(t *testing.T) {
	event := &Event{
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
	err := runner.stepCancelDonorIdentity(ctx, event)

	assert.ErrorContains(t, err, "error retrieving donor: hey")
}

func TestRunnerCancelDonorIdentityWhenStepIgnored(t *testing.T) {
	testcases := map[string]*donordata.Provided{
		"identity not confirmed": {
			IdentityUserData: identity.UserData{Status: identity.StatusFailed},
		},
		"already signed": {
			IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			SignedAt:         time.Now(),
		},
	}

	for name, provided := range testcases {
		t.Run(name, func(t *testing.T) {
			lpaKey := dynamo.LpaKey("an-lpa")
			donorKey := dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor"))
			event := &Event{
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
			err := runner.stepCancelDonorIdentity(ctx, event)

			assert.Equal(t, errStepIgnored, err)
		})
	}
}

func TestRunnerCancelDonorIdentityWhenNotifySendErrors(t *testing.T) {
	event := &Event{
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything).
		Return(&donordata.Provided{
			LpaUID:           "lpa-uid",
			Donor:            donordata.Donor{Email: "donor@example.com"},
			IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
		}, nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	runner := &Runner{
		donorStore:   donorStore,
		notifyClient: notifyClient,
	}
	err := runner.stepCancelDonorIdentity(ctx, event)

	assert.ErrorIs(t, err, expectedError)
}

func TestRunnerCancelDonorIdentityWhenDonorStorePutErrors(t *testing.T) {
	event := &Event{
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything).
		Return(&donordata.Provided{
			LpaUID:           "lpa-uid",
			Donor:            donordata.Donor{Email: "donor@example.com"},
			IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
		}, nil)
	donorStore.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	runner := &Runner{
		donorStore:   donorStore,
		notifyClient: notifyClient,
	}
	err := runner.stepCancelDonorIdentity(ctx, event)

	assert.ErrorIs(t, err, expectedError)
}
