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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	certificateproviderdata "github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
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
	notifyClient := newMockNotifyClient(t)
	metricsClient := newMockMetricsClient(t)
	bundle := newMockBundle(t)

	runner := NewRunner(logger, store, donorStore, certificateProviderStore, notifyClient, bundle, metricsClient, true)

	assert.Equal(t, logger, runner.logger)
	assert.Equal(t, store, runner.store)
	assert.Equal(t, donorStore, runner.donorStore)
	assert.Equal(t, notifyClient, runner.notifyClient)
	assert.Equal(t, metricsClient, runner.metricsClient)
	assert.Equal(t, true, runner.metricsEnabled)
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

func TestRunnerCancelDonorIdentity(t *testing.T) {
	lpaKey := dynamo.LpaKey("an-lpa")
	donorKey := dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor"))
	event := &Event{
		TargetLpaKey:      lpaKey,
		TargetLpaOwnerKey: donorKey,
	}

	provided := &donordata.Provided{
		LpaUID:           "lpa-uid",
		Donor:            donordata.Donor{Email: "donor@example.com", ContactLanguagePreference: localize.Cy},
		IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(ctx, lpaKey, donorKey).
		Return(provided, nil)
	donorStore.EXPECT().
		Put(ctx, &donordata.Provided{
			LpaUID:           "lpa-uid",
			Donor:            donordata.Donor{Email: "donor@example.com", ContactLanguagePreference: localize.Cy},
			IdentityUserData: identity.UserData{Status: identity.StatusExpired},
		}).
		Return(nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToDonor(provided), "lpa-uid", notify.DonorIdentityCheckExpiredEmail{}).
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
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
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
		SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	runner := &Runner{
		donorStore:   donorStore,
		notifyClient: notifyClient,
	}
	err := runner.stepCancelDonorIdentity(ctx, event)

	assert.ErrorIs(t, err, expectedError)
}

func TestRunnerRemindCertificateProviderToComplete(t *testing.T) {
	testcases := map[string]struct {
		certificateProvider      *certificateproviderdata.Provided
		certificateProviderError error
	}{
		"certificate provider not started": {
			certificateProviderError: dynamo.NotFoundError{},
		},
		"certificate provider started": {
			certificateProvider: &certificateproviderdata.Provided{
				ContactLanguagePreference: localize.En,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			row := &Event{
				TargetLpaKey:      dynamo.LpaKey("an-lpa"),
				TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
			}
			donor := &donordata.Provided{
				LpaUID: "lpa-uid",
			}
			lpa := &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor: lpadata.Donor{
					FirstNames:                "a",
					LastName:                  "b",
					ContactLanguagePreference: localize.En,
				},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames:                "c",
					LastName:                  "d",
					ContactLanguagePreference: localize.En,
				},
				CertificateProviderInvitedAt: testNow.AddDate(0, -3, -1),
				SignedAt:                     testNow.AddDate(0, -3, 0).Add(-time.Second),
			}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				One(ctx, row.TargetLpaKey, row.TargetLpaOwnerKey).
				Return(donor, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				One(ctx, row.TargetLpaKey).
				Return(tc.certificateProvider, tc.certificateProviderError)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Resolve(ctx, donor).
				Return(lpa, nil)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaCertificateProvider(nil, lpa), "lpa-uid", notify.AdviseCertificateProviderToSignOrOptOutEmail{
					DonorFullName:               "a b",
					LpaType:                     "personal-welfare",
					CertificateProviderFullName: "c d",
					InvitedDate:                 "1 October 1999",
					DeadlineDate:                "2 April 2000",
				}).
				Return(nil)
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorCertificateProviderHasNotActedEmail{
					CertificateProviderFullName: "c d",
					LpaType:                     "personal-welfare",
					DonorFullName:               "a b",
					InvitedDate:                 "1 October 1999",
					DeadlineDate:                "2 April 2000",
				}).
				Return(nil)

			localizer := &localize.Localizer{}

			bundle := newMockBundle(t)
			bundle.EXPECT().
				For(localize.En).
				Return(localizer)

			runner := &Runner{
				donorStore:               donorStore,
				lpaStoreResolvingService: lpaStoreResolvingService,
				certificateProviderStore: certificateProviderStore,
				notifyClient:             notifyClient,
				bundle:                   bundle,
				now:                      testNowFn,
			}

			err := runner.stepRemindCertificateProviderToComplete(ctx, row)
			assert.Nil(t, err)
		})
	}
}

func TestRunnerRemindCertificateProviderToCompleteWhenOnPaper(t *testing.T) {
	donorUID := actoruid.New()
	correspondentUID := actoruid.New()

	testcases := map[string]struct {
		lpa                *lpadata.Lpa
		donorLetterRequest event.LetterRequested
	}{
		"to donor": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor: lpadata.Donor{
					UID:        donorUID,
					FirstNames: "a",
					LastName:   "b",
					Channel:    lpadata.ChannelPaper,
				},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					Channel:    lpadata.ChannelPaper,
				},
				SignedAt: testNow.AddDate(0, -3, 0).Add(-time.Second),
			},
			donorLetterRequest: event.LetterRequested{
				UID:        "lpa-uid",
				LetterType: "INFORM_DONOR_CERTIFICATE_PROVIDER_HAS_NOT_ACTED",
				ActorType:  actor.TypeDonor,
				ActorUID:   donorUID,
			},
		},
		"to correspondent": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor: lpadata.Donor{
					FirstNames: "a",
					LastName:   "b",
					Channel:    lpadata.ChannelPaper,
				},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "c",
					LastName:   "d",
					Channel:    lpadata.ChannelPaper,
				},
				Correspondent: lpadata.Correspondent{
					UID:     correspondentUID,
					Address: place.Address{Line1: "123"},
				},
				SignedAt: testNow.AddDate(0, -3, 0).Add(-time.Second),
			},
			donorLetterRequest: event.LetterRequested{
				UID:        "lpa-uid",
				LetterType: "INFORM_DONOR_CERTIFICATE_PROVIDER_HAS_NOT_ACTED",
				ActorType:  actor.TypeCorrespondent,
				ActorUID:   correspondentUID,
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			row := &Event{
				TargetLpaKey:      dynamo.LpaKey("an-lpa"),
				TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
			}
			donor := &donordata.Provided{
				LpaUID: "lpa-uid",
			}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				One(ctx, row.TargetLpaKey, row.TargetLpaOwnerKey).
				Return(donor, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Resolve(ctx, donor).
				Return(tc.lpa, nil)

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				One(ctx, row.TargetLpaKey).
				Return(nil, dynamo.NotFoundError{})

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendLetterRequested(ctx, event.LetterRequested{
					UID:        "lpa-uid",
					LetterType: "ADVISE_CERTIFICATE_PROVIDER_TO_SIGN_OR_OPT_OUT",
					ActorType:  actor.TypeCertificateProvider,
					ActorUID:   tc.lpa.CertificateProvider.UID,
				}).
				Return(nil)
			eventClient.EXPECT().
				SendLetterRequested(ctx, tc.donorLetterRequest).
				Return(nil)

			runner := &Runner{
				donorStore:               donorStore,
				lpaStoreResolvingService: lpaStoreResolvingService,
				certificateProviderStore: certificateProviderStore,
				eventClient:              eventClient,
				now:                      testNowFn,
			}

			err := runner.stepRemindCertificateProviderToComplete(ctx, row)
			assert.Nil(t, err)
		})
	}
}

func TestRunnerRemindCertificateProviderToCompleteWhenCertificateProviderAlreadyCompleted(t *testing.T) {
	row := &Event{
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}
	certificateProvider := &certificateproviderdata.Provided{
		Tasks: certificateproviderdata.Tasks{ProvideTheCertificate: task.StateCompleted},
	}

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		One(ctx, row.TargetLpaKey).
		Return(certificateProvider, nil)

	runner := &Runner{
		certificateProviderStore: certificateProviderStore,
		now:                      testNowFn,
	}

	err := runner.stepRemindCertificateProviderToComplete(ctx, row)
	assert.Equal(t, errStepIgnored, err)
}

func TestRunnerRemindCertificateProviderToCompleteWhenNotValidTime(t *testing.T) {
	testcases := map[string]*lpadata.Lpa{
		"invite sent almost 3 months ago": {
			CertificateProviderInvitedAt: testNow.AddDate(0, -3, 1),
			SignedAt:                     testNow.AddDate(0, -3, 0),
		},
		"expiry almost 3 months ago": {
			CertificateProviderInvitedAt: testNow.AddDate(0, -3, 0),
			SignedAt:                     testNow.AddDate(0, -3, 1),
		},
		"submitted expiry almost 3 months ago": {
			Donor: lpadata.Donor{
				IdentityCheck: &lpadata.IdentityCheck{
					CheckedAt: testNow,
				},
			},
			CertificateProviderInvitedAt: testNow.AddDate(0, -3, 0),
			SignedAt:                     testNow.AddDate(-2, 3, 1),
			Submitted:                    true,
		},
	}

	for name, lpa := range testcases {
		t.Run(name, func(t *testing.T) {
			donor := &donordata.Provided{
				LpaUID: "lpa-uid",
			}

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				One(mock.Anything, mock.Anything).
				Return(nil, dynamo.NotFoundError{})

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				One(mock.Anything, mock.Anything, mock.Anything).
				Return(donor, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Resolve(mock.Anything, mock.Anything).
				Return(lpa, nil)

			runner := &Runner{
				lpaStoreResolvingService: lpaStoreResolvingService,
				donorStore:               donorStore,
				certificateProviderStore: certificateProviderStore,
				now:                      testNowFn,
			}

			err := runner.stepRemindCertificateProviderToComplete(ctx, &Event{})
			assert.Equal(t, errStepIgnored, err)
		})
	}
}
