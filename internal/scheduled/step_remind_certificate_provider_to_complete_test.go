package scheduled

import (
	"context"
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
	donorUID := actoruid.New()
	correspondentUID := actoruid.New()

	nilNotifyClient := func(*testing.T, context.Context, *lpadata.Lpa) *mockNotifyClient { return nil }
	nilEventClient := func(*testing.T, context.Context, *lpadata.Lpa) *mockEventClient { return nil }

	signedAt := testNow.AddDate(0, -3, -1)
	invitedAt := testNow.AddDate(0, -3, -1)

	testcases := map[string]struct {
		lpa                      *lpadata.Lpa
		certificateProvider      *certificateproviderdata.Provided
		certificateProviderError error
		notifyClient             func(*testing.T, context.Context, *lpadata.Lpa) *mockNotifyClient
		eventClient              func(*testing.T, context.Context, *lpadata.Lpa) *mockEventClient
		localizer                func(*testing.T) *mockLocalizer
	}{
		"online donor online certificate provider not started": {
			lpa: &lpadata.Lpa{
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
				SignedAt:                     signedAt,
				CertificateProviderInvitedAt: invitedAt,
			},
			certificateProviderError: dynamo.NotFoundError{},
			notifyClient: func(t *testing.T, ctx context.Context, lpa *lpadata.Lpa) *mockNotifyClient {
				notifyClient := newMockNotifyClient(t)
				notifyClient.EXPECT().
					EmailGreeting(lpa).
					Return("hey")
				notifyClient.EXPECT().
					SendActorEmail(ctx, notify.ToLpaCertificateProvider(nil, lpa), "lpa-uid", notify.AdviseCertificateProviderToSignOrOptOutEmail{
						DonorFullName:                   "a b",
						DonorFullNamePossessive:         "a b’s",
						LpaType:                         "Personal welfare",
						CertificateProviderFullName:     "c d",
						InvitedDate:                     "1 March 2000",
						DeadlineDate:                    "1 April 2000",
						CertificateProviderStartPageURL: "http://example.com/certificate-provider",
						CertificateProviderOptOutURL:    "http://example.com/certificate-provider-opt-out",
					}).
					Return(nil).
					Once()
				notifyClient.EXPECT().
					SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorCertificateProviderHasNotActedEmail{
						Greeting:                        "hey",
						CertificateProviderFullName:     "c d",
						LpaType:                         "Personal welfare",
						LpaReferenceNumber:              "lpa-uid",
						InvitedDate:                     "1 March 2000",
						DeadlineDate:                    "1 April 2000",
						CertificateProviderStartPageURL: "http://example.com/certificate-provider",
					}).
					Return(nil).
					Once()
				return notifyClient
			},
			eventClient: nilEventClient,
			localizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Possessive("a b").
					Return("a b’s")
				l.EXPECT().
					T("personal-welfare").
					Return("Personal welfare").
					Twice()
				l.EXPECT().
					FormatDate(signedAt.AddDate(0, 6, 0)).
					Return("1 April 2000")
				l.EXPECT().
					FormatDate(invitedAt).
					Return("1 March 2000")
				return l
			},
		},
		"online donor online certificate provider started": {
			lpa: &lpadata.Lpa{
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
				SignedAt:                     signedAt,
				CertificateProviderInvitedAt: invitedAt,
			},
			certificateProvider: &certificateproviderdata.Provided{
				ContactLanguagePreference: localize.En,
			},
			notifyClient: func(t *testing.T, ctx context.Context, lpa *lpadata.Lpa) *mockNotifyClient {
				notifyClient := newMockNotifyClient(t)
				notifyClient.EXPECT().
					EmailGreeting(lpa).
					Return("hey")
				notifyClient.EXPECT().
					SendActorEmail(ctx, notify.ToLpaCertificateProvider(nil, lpa), "lpa-uid", notify.AdviseCertificateProviderToSignOrOptOutEmailAccessCodeUsed{
						DonorFullName:                   "a b",
						DonorFullNamePossessive:         "a b’s",
						LpaType:                         "Personal welfare",
						CertificateProviderFullName:     "c d",
						InvitedDate:                     "1 March 2000",
						DeadlineDate:                    "1 April 2000",
						CertificateProviderStartPageURL: "http://example.com/certificate-provider",
						CertificateProviderOptOutURL:    "http://example.com/certificate-provider-opt-out",
					}).
					Return(nil).
					Once()
				notifyClient.EXPECT().
					SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorCertificateProviderHasNotActedEmail{
						Greeting:                        "hey",
						CertificateProviderFullName:     "c d",
						LpaType:                         "Personal welfare",
						LpaReferenceNumber:              "lpa-uid",
						InvitedDate:                     "1 March 2000",
						DeadlineDate:                    "1 April 2000",
						CertificateProviderStartPageURL: "http://example.com/certificate-provider",
					}).
					Return(nil).
					Once()
				return notifyClient
			},
			eventClient: nilEventClient,
			localizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Possessive("a b").
					Return("a b’s")
				l.EXPECT().
					T("personal-welfare").
					Return("Personal welfare").
					Twice()
				l.EXPECT().
					FormatDate(signedAt.AddDate(0, 6, 0)).
					Return("1 April 2000")
				l.EXPECT().
					FormatDate(invitedAt).
					Return("1 March 2000")
				return l
			},
		},
		"online donor paper certificate provider": {
			lpa: &lpadata.Lpa{
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
					Channel:                   lpadata.ChannelPaper,
				},
				SignedAt:                     signedAt,
				CertificateProviderInvitedAt: invitedAt,
			},
			certificateProviderError: dynamo.NotFoundError{},
			notifyClient: func(t *testing.T, ctx context.Context, lpa *lpadata.Lpa) *mockNotifyClient {
				notifyClient := newMockNotifyClient(t)
				notifyClient.EXPECT().
					EmailGreeting(lpa).
					Return("hey")
				notifyClient.EXPECT().
					SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorPaperCertificateProviderHasNotActedEmail{
						Greeting:                    "hey",
						CertificateProviderFullName: "c d",
						LpaType:                     "Personal welfare",
						PostedDate:                  "1 March 2000",
						DeadlineDate:                "1 April 2000",
					}).
					Return(nil).
					Once()
				return notifyClient
			},
			eventClient: func(t *testing.T, ctx context.Context, lpa *lpadata.Lpa) *mockEventClient {
				eventClient := newMockEventClient(t)
				eventClient.EXPECT().
					SendLetterRequested(ctx, event.LetterRequested{
						UID:        "lpa-uid",
						LetterType: "ADVISE_CERTIFICATE_PROVIDER_TO_SIGN_OR_OPT_OUT",
						ActorType:  actor.TypeCertificateProvider,
						ActorUID:   lpa.CertificateProvider.UID,
					}).
					Return(nil)
				return eventClient
			},
			localizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					T("personal-welfare").
					Return("Personal welfare")
				l.EXPECT().
					FormatDate(signedAt.AddDate(0, 6, 0)).
					Return("1 April 2000")
				l.EXPECT().
					FormatDate(invitedAt).
					Return("1 March 2000")
				return l
			},
		},
		"paper donor online certificate provider": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor: lpadata.Donor{
					UID:                       donorUID,
					FirstNames:                "a",
					LastName:                  "b",
					ContactLanguagePreference: localize.En,
					Channel:                   lpadata.ChannelPaper,
				},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames:                "c",
					LastName:                  "d",
					ContactLanguagePreference: localize.En,
				},
				SignedAt:                     signedAt,
				CertificateProviderInvitedAt: invitedAt,
			},
			certificateProviderError: dynamo.NotFoundError{},
			notifyClient: func(t *testing.T, ctx context.Context, lpa *lpadata.Lpa) *mockNotifyClient {
				notifyClient := newMockNotifyClient(t)
				notifyClient.EXPECT().
					SendActorEmail(ctx, notify.ToLpaCertificateProvider(nil, lpa), "lpa-uid", notify.AdviseCertificateProviderToSignOrOptOutEmail{
						DonorFullName:                   "a b",
						DonorFullNamePossessive:         "a b’s",
						LpaType:                         "Personal welfare",
						CertificateProviderFullName:     "c d",
						InvitedDate:                     "1 March 2000",
						DeadlineDate:                    "1 April 2000",
						CertificateProviderStartPageURL: "http://example.com/certificate-provider",
						CertificateProviderOptOutURL:    "http://example.com/certificate-provider-opt-out",
					}).
					Return(nil).
					Once()
				return notifyClient
			},
			eventClient: func(t *testing.T, ctx context.Context, lpa *lpadata.Lpa) *mockEventClient {
				eventClient := newMockEventClient(t)
				eventClient.EXPECT().
					SendLetterRequested(ctx, event.LetterRequested{
						UID:        "lpa-uid",
						LetterType: "INFORM_DONOR_CERTIFICATE_PROVIDER_HAS_NOT_ACTED",
						ActorType:  actor.TypeDonor,
						ActorUID:   donorUID,
					}).
					Return(nil)
				return eventClient
			},
			localizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Possessive("a b").
					Return("a b’s")
				l.EXPECT().
					T("personal-welfare").
					Return("Personal welfare")
				l.EXPECT().
					FormatDate(signedAt.AddDate(0, 6, 0)).
					Return("1 April 2000")
				l.EXPECT().
					FormatDate(invitedAt).
					Return("1 March 2000")
				return l
			},
		},
		"paper donor paper certificate provider": {
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
				SignedAt: signedAt,
			},
			certificateProviderError: dynamo.NotFoundError{},
			notifyClient:             nilNotifyClient,
			eventClient: func(t *testing.T, ctx context.Context, lpa *lpadata.Lpa) *mockEventClient {
				eventClient := newMockEventClient(t)
				eventClient.EXPECT().
					SendLetterRequested(ctx, event.LetterRequested{
						UID:        "lpa-uid",
						LetterType: "ADVISE_CERTIFICATE_PROVIDER_TO_SIGN_OR_OPT_OUT",
						ActorType:  actor.TypeCertificateProvider,
						ActorUID:   lpa.CertificateProvider.UID,
					}).
					Return(nil)
				eventClient.EXPECT().
					SendLetterRequested(ctx, event.LetterRequested{
						UID:        "lpa-uid",
						LetterType: "INFORM_DONOR_CERTIFICATE_PROVIDER_HAS_NOT_ACTED",
						ActorType:  actor.TypeDonor,
						ActorUID:   donorUID,
					}).
					Return(nil)
				return eventClient
			},
			localizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Possessive("a b").
					Return("a b’s")
				l.EXPECT().
					T("personal-welfare").
					Return("Personal welfare")
				l.EXPECT().
					FormatDate(signedAt.AddDate(0, 6, 0)).
					Return("1 April 2000")
				l.EXPECT().
					FormatDate(invitedAt).
					Return("1 March 2000")
				return l
			},
		},
		"paper correspondent paper certificate provider": {
			lpa: &lpadata.Lpa{
				LpaUID: "lpa-uid",
				Type:   lpadata.LpaTypePersonalWelfare,
				Donor: lpadata.Donor{
					FirstNames:                "a",
					LastName:                  "b",
					Channel:                   lpadata.ChannelPaper,
					ContactLanguagePreference: localize.En,
				},
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames:                "c",
					LastName:                  "d",
					ContactLanguagePreference: localize.En,
					Channel:                   lpadata.ChannelPaper,
				},
				Correspondent: lpadata.Correspondent{
					UID:     correspondentUID,
					Address: place.Address{Line1: "123"},
				},
				SignedAt: signedAt,
			},
			certificateProviderError: dynamo.NotFoundError{},
			notifyClient:             nilNotifyClient,
			eventClient: func(t *testing.T, ctx context.Context, lpa *lpadata.Lpa) *mockEventClient {
				eventClient := newMockEventClient(t)
				eventClient.EXPECT().
					SendLetterRequested(ctx, event.LetterRequested{
						UID:        "lpa-uid",
						LetterType: "ADVISE_CERTIFICATE_PROVIDER_TO_SIGN_OR_OPT_OUT",
						ActorType:  actor.TypeCertificateProvider,
						ActorUID:   lpa.CertificateProvider.UID,
					}).
					Return(nil)
				eventClient.EXPECT().
					SendLetterRequested(ctx, event.LetterRequested{
						UID:        "lpa-uid",
						LetterType: "INFORM_DONOR_CERTIFICATE_PROVIDER_HAS_NOT_ACTED",
						ActorType:  actor.TypeCorrespondent,
						ActorUID:   correspondentUID,
					}).
					Return(nil)
				return eventClient
			},
			localizer: func(t *testing.T) *mockLocalizer {
				l := newMockLocalizer(t)
				l.EXPECT().
					Possessive("a b").
					Return("a b’s")
				l.EXPECT().
					T("personal-welfare").
					Return("Personal welfare")
				l.EXPECT().
					FormatDate(signedAt.AddDate(0, 6, 0)).
					Return("1 April 2000")
				l.EXPECT().
					FormatDate(invitedAt).
					Return("1 March 2000")
				return l
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

			certificateProviderStore := newMockCertificateProviderStore(t)
			certificateProviderStore.EXPECT().
				One(ctx, row.TargetLpaKey).
				Return(tc.certificateProvider, tc.certificateProviderError)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Resolve(ctx, donor).
				Return(tc.lpa, nil)

			notifyClient := tc.notifyClient(t, ctx, tc.lpa)
			eventClient := tc.eventClient(t, ctx, tc.lpa)

			bundle := newMockBundle(t)
			if notifyClient != nil {
				bundle.EXPECT().
					For(localize.En).
					Return(tc.localizer(t))
			}

			runner := &Runner{
				donorStore:                   donorStore,
				lpaStoreResolvingService:     lpaStoreResolvingService,
				certificateProviderStore:     certificateProviderStore,
				notifyClient:                 notifyClient,
				eventClient:                  eventClient,
				bundle:                       bundle,
				now:                          testNowFn,
				certificateProviderStartURL:  "http://example.com/certificate-provider",
				certificateProviderOptOutURL: "http://example.com/certificate-provider-opt-out",
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

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				Possessive(mock.Anything).
				Return("a b’s")
			localizer.EXPECT().
				T(mock.Anything).
				Return("Personal welfare")
			localizer.EXPECT().
				FormatDate(mock.Anything).
				Return("1 April 2000")
			localizer.EXPECT().
				FormatDate(mock.Anything).
				Return("1 March 2000")

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
