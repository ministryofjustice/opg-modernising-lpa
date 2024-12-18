package scheduled

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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
				EmailGreeting(lpa).
				Return("hey")
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaCertificateProvider(nil, lpa), "lpa-uid", notify.AdviseCertificateProviderToSignOrOptOutEmail{
					DonorFullName:                   "a b",
					DonorFullNamePossessive:         "a b’s",
					LpaType:                         "personal-welfare",
					CertificateProviderFullName:     "c d",
					InvitedDate:                     "1 October 1999",
					DeadlineDate:                    "2 April 2000",
					CertificateProviderStartPageURL: "http://app/certificate-provider-start",
				}).
				Return(nil).
				Once()
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorCertificateProviderHasNotActedEmail{
					Greeting:                        "hey",
					CertificateProviderFullName:     "c d",
					LpaType:                         "personal-welfare",
					InvitedDate:                     "1 October 1999",
					DeadlineDate:                    "2 April 2000",
					CertificateProviderStartPageURL: "http://app/certificate-provider-start",
				}).
				Return(nil).
				Once()

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
				appPublicURL:             "http://app",
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
		"expiry almost in 3 months": {
			CertificateProviderInvitedAt: testNow.AddDate(0, -3, 0),
			SignedAt:                     testNow.AddDate(0, -3, 1),
		},
		"submitted expiry almost in 3 months": {
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

func TestRunnerRemindCertificateProviderToCompleteWhenCertificateProviderStoreErrors(t *testing.T) {
	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		One(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	runner := &Runner{
		certificateProviderStore: certificateProviderStore,
		now:                      testNowFn,
	}

	err := runner.stepRemindCertificateProviderToComplete(ctx, &Event{})
	assert.ErrorIs(t, err, expectedError)
}

func TestRunnerRemindCertificateProviderToCompleteWhenDonorStoreErrors(t *testing.T) {
	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		One(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, expectedError)

	runner := &Runner{
		donorStore:               donorStore,
		certificateProviderStore: certificateProviderStore,
		now:                      testNowFn,
	}

	err := runner.stepRemindCertificateProviderToComplete(ctx, &Event{})
	assert.ErrorIs(t, err, expectedError)
}

func TestRunnerRemindCertificateProviderToCompleteWhenLpaStoreResolvingServiceErrors(t *testing.T) {
	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		One(mock.Anything, mock.Anything).
		Return(nil, dynamo.NotFoundError{})

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything).
		Return(&donordata.Provided{}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Resolve(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	runner := &Runner{
		donorStore:               donorStore,
		certificateProviderStore: certificateProviderStore,
		lpaStoreResolvingService: lpaStoreResolvingService,
		now:                      testNowFn,
	}

	err := runner.stepRemindCertificateProviderToComplete(ctx, &Event{})
	assert.ErrorIs(t, err, expectedError)
}

func TestRunnerRemindCertificateProviderToCompleteWhenNotifyClientErrors(t *testing.T) {
	testcases := map[string]func(*mockNotifyClient){
		"first": func(m *mockNotifyClient) {
			m.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError).
				Once()
		},
		"second": func(m *mockNotifyClient) {
			m.EXPECT().
				EmailGreeting(mock.Anything).
				Return("hey")
			m.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil).
				Once()
			m.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError).
				Once()
		},
	}

	for name, setupNotifyClient := range testcases {
		t.Run(name, func(t *testing.T) {
			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				One(mock.Anything, mock.Anything).
				Return(nil, dynamo.NotFoundError{})

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				One(mock.Anything, mock.Anything, mock.Anything).
				Return(&donordata.Provided{}, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Resolve(mock.Anything, mock.Anything).
				Return(&lpadata.Lpa{}, nil)

			notifyClient := newMockNotifyClient(t)
			setupNotifyClient(notifyClient)

			localizer := &localize.Localizer{}

			bundle := newMockBundle(t)
			bundle.EXPECT().
				For(mock.Anything).
				Return(localizer)

			runner := &Runner{
				donorStore:               donorStore,
				certificateProviderStore: certificateProviderStore,
				lpaStoreResolvingService: lpaStoreResolvingService,
				notifyClient:             notifyClient,
				bundle:                   bundle,
				now:                      testNowFn,
			}

			err := runner.stepRemindCertificateProviderToComplete(ctx, &Event{})
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestRunnerRemindCertificateProviderToCompleteWhenEventClientErrors(t *testing.T) {
	testcases := map[string]func(*mockEventClient){
		"first": func(m *mockEventClient) {
			m.EXPECT().
				SendLetterRequested(mock.Anything, mock.Anything).
				Return(expectedError).
				Once()
		},
		"second": func(m *mockEventClient) {
			m.EXPECT().
				SendLetterRequested(mock.Anything, mock.Anything).
				Return(nil).
				Once()
			m.EXPECT().
				SendLetterRequested(mock.Anything, mock.Anything).
				Return(expectedError).
				Once()
		},
	}

	for name, setupEventClient := range testcases {
		t.Run(name, func(t *testing.T) {
			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				One(mock.Anything, mock.Anything).
				Return(nil, dynamo.NotFoundError{})

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				One(mock.Anything, mock.Anything, mock.Anything).
				Return(&donordata.Provided{}, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Resolve(mock.Anything, mock.Anything).
				Return(&lpadata.Lpa{
					Donor:               lpadata.Donor{Channel: lpadata.ChannelPaper},
					CertificateProvider: lpadata.CertificateProvider{Channel: lpadata.ChannelPaper},
				}, nil)

			eventClient := newMockEventClient(t)
			setupEventClient(eventClient)

			runner := &Runner{
				donorStore:               donorStore,
				certificateProviderStore: certificateProviderStore,
				lpaStoreResolvingService: lpaStoreResolvingService,
				eventClient:              eventClient,
				now:                      testNowFn,
			}

			err := runner.stepRemindCertificateProviderToComplete(ctx, &Event{})
			assert.ErrorIs(t, err, expectedError)
		})
	}
}
