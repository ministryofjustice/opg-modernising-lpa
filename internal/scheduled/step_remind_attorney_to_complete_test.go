package scheduled

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	event "github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestRunnerRemindAttorneyToComplete(t *testing.T) {
	attorneyUID := actoruid.New()
	replacementAttorneyUID := actoruid.New()
	trustCorporationUID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()

	testcases := map[string]struct {
		attorneys     []*attorneydata.Provided
		attorneyError error
	}{
		"not started": {
			attorneyError: dynamo.NotFoundError{},
		},
		"started": {
			attorneys: []*attorneydata.Provided{{
				UID:                       attorneyUID,
				ContactLanguagePreference: localize.En,
			}, {
				UID:                       replacementAttorneyUID,
				IsReplacement:             true,
				ContactLanguagePreference: localize.En,
			}, {
				UID:                       trustCorporationUID,
				IsTrustCorporation:        true,
				ContactLanguagePreference: localize.En,
			}, {
				UID:                       replacementTrustCorporationUID,
				IsReplacement:             true,
				IsTrustCorporation:        true,
				ContactLanguagePreference: localize.En,
			}},
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
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{
						UID:                       attorneyUID,
						FirstNames:                "c",
						LastName:                  "d",
						ContactLanguagePreference: localize.En,
					}},
					TrustCorporation: lpadata.TrustCorporation{
						UID:                       trustCorporationUID,
						Name:                      "trusty",
						ContactLanguagePreference: localize.En,
					},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{
						UID:                       replacementAttorneyUID,
						FirstNames:                "e",
						LastName:                  "f",
						ContactLanguagePreference: localize.En,
					}},
					TrustCorporation: lpadata.TrustCorporation{
						UID:                       replacementTrustCorporationUID,
						Name:                      "untrusty",
						ContactLanguagePreference: localize.En,
					},
				},
				AttorneysInvitedAt: testNow.AddDate(0, -3, -1),
				SignedAt:           testNow.AddDate(0, -3, 0).Add(-time.Second),
			}

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				One(ctx, row.TargetLpaKey, row.TargetLpaOwnerKey).
				Return(donor, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				Resolve(ctx, donor).
				Return(lpa, nil)

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				All(ctx, "lpa-uid").
				Return(tc.attorneys, tc.attorneyError)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				EmailGreeting(lpa).
				Return("hey")
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[0]), "lpa-uid", notify.AdviseAttorneyToSignOrOptOutEmail{
					DonorFullName:           "a b",
					DonorFullNamePossessive: "a b’s",
					LpaType:                 "personal-welfare",
					AttorneyFullName:        "c d",
					InvitedDate:             "1 October 1999",
					DeadlineDate:            "2 April 2000",
					AttorneyStartPageURL:    "http://app/attorney-start",
				}).
				Return(nil).
				Once()
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaTrustCorporation(lpa.Attorneys.TrustCorporation), "lpa-uid", notify.AdviseAttorneyToSignOrOptOutEmail{
					DonorFullName:           "a b",
					DonorFullNamePossessive: "a b’s",
					LpaType:                 "personal-welfare",
					AttorneyFullName:        "trusty",
					InvitedDate:             "1 October 1999",
					DeadlineDate:            "2 April 2000",
					AttorneyStartPageURL:    "http://app/attorney-start",
				}).
				Return(nil).
				Once()
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaAttorney(lpa.ReplacementAttorneys.Attorneys[0]), "lpa-uid", notify.AdviseAttorneyToSignOrOptOutEmail{
					DonorFullName:           "a b",
					DonorFullNamePossessive: "a b’s",
					LpaType:                 "personal-welfare",
					AttorneyFullName:        "e f",
					InvitedDate:             "1 October 1999",
					DeadlineDate:            "2 April 2000",
					AttorneyStartPageURL:    "http://app/attorney-start",
				}).
				Return(nil).
				Once()
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaTrustCorporation(lpa.ReplacementAttorneys.TrustCorporation), "lpa-uid", notify.AdviseAttorneyToSignOrOptOutEmail{
					DonorFullName:           "a b",
					DonorFullNamePossessive: "a b’s",
					LpaType:                 "personal-welfare",
					AttorneyFullName:        "untrusty",
					InvitedDate:             "1 October 1999",
					DeadlineDate:            "2 April 2000",
					AttorneyStartPageURL:    "http://app/attorney-start",
				}).
				Return(nil).
				Once()
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorAttorneyHasNotActedEmail{
					Greeting:             "hey",
					AttorneyFullName:     "c d",
					LpaType:              "personal-welfare",
					InvitedDate:          "1 October 1999",
					DeadlineDate:         "2 April 2000",
					AttorneyStartPageURL: "http://app/attorney-start",
				}).
				Return(nil).
				Once()
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorAttorneyHasNotActedEmail{
					Greeting:             "hey",
					AttorneyFullName:     "trusty",
					LpaType:              "personal-welfare",
					InvitedDate:          "1 October 1999",
					DeadlineDate:         "2 April 2000",
					AttorneyStartPageURL: "http://app/attorney-start",
				}).
				Return(nil).
				Once()
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorAttorneyHasNotActedEmail{
					Greeting:             "hey",
					AttorneyFullName:     "e f",
					LpaType:              "personal-welfare",
					InvitedDate:          "1 October 1999",
					DeadlineDate:         "2 April 2000",
					AttorneyStartPageURL: "http://app/attorney-start",
				}).
				Return(nil).
				Once()
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorAttorneyHasNotActedEmail{
					Greeting:             "hey",
					AttorneyFullName:     "untrusty",
					LpaType:              "personal-welfare",
					InvitedDate:          "1 October 1999",
					DeadlineDate:         "2 April 2000",
					AttorneyStartPageURL: "http://app/attorney-start",
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
				attorneyStore:            attorneyStore,
				notifyClient:             notifyClient,
				bundle:                   bundle,
				now:                      testNowFn,
				appPublicURL:             "http://app",
			}

			err := runner.stepRemindAttorneyToComplete(ctx, row)
			assert.Nil(t, err)
		})
	}
}

func TestRunnerRemindAttorneyToCompleteWhenOnPaper(t *testing.T) {
	donorUID := actoruid.New()
	correspondentUID := actoruid.New()
	attorneyUID := actoruid.New()
	replacementAttorneyUID := actoruid.New()
	trustCorporationUID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()

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
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{
						UID:                       attorneyUID,
						FirstNames:                "c",
						LastName:                  "d",
						ContactLanguagePreference: localize.En,
						Channel:                   lpadata.ChannelPaper,
					}},
					TrustCorporation: lpadata.TrustCorporation{
						UID:                       trustCorporationUID,
						Name:                      "trusty",
						ContactLanguagePreference: localize.En,
						Channel:                   lpadata.ChannelPaper,
					},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{
						UID:                       replacementAttorneyUID,
						FirstNames:                "e",
						LastName:                  "f",
						ContactLanguagePreference: localize.En,
						Channel:                   lpadata.ChannelPaper,
					}},
					TrustCorporation: lpadata.TrustCorporation{
						UID:                       replacementTrustCorporationUID,
						Name:                      "untrusty",
						ContactLanguagePreference: localize.En,
						Channel:                   lpadata.ChannelPaper,
					},
				},
				SignedAt: testNow.AddDate(0, -3, 0).Add(-time.Second),
			},
			donorLetterRequest: event.LetterRequested{
				UID:        "lpa-uid",
				LetterType: "INFORM_DONOR_ATTORNEY_HAS_NOT_ACTED",
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
				Attorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{
						UID:                       attorneyUID,
						FirstNames:                "c",
						LastName:                  "d",
						ContactLanguagePreference: localize.En,
						Channel:                   lpadata.ChannelPaper,
					}},
					TrustCorporation: lpadata.TrustCorporation{
						UID:                       trustCorporationUID,
						Name:                      "trusty",
						ContactLanguagePreference: localize.En,
						Channel:                   lpadata.ChannelPaper,
					},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					Attorneys: []lpadata.Attorney{{
						UID:                       replacementAttorneyUID,
						FirstNames:                "e",
						LastName:                  "f",
						ContactLanguagePreference: localize.En,
						Channel:                   lpadata.ChannelPaper,
					}},
					TrustCorporation: lpadata.TrustCorporation{
						UID:                       replacementTrustCorporationUID,
						Name:                      "untrusty",
						ContactLanguagePreference: localize.En,
						Channel:                   lpadata.ChannelPaper,
					},
				},
				Correspondent: lpadata.Correspondent{
					UID:     correspondentUID,
					Address: place.Address{Line1: "123"},
				},
				SignedAt: testNow.AddDate(0, -3, 0).Add(-time.Second),
			},
			donorLetterRequest: event.LetterRequested{
				UID:        "lpa-uid",
				LetterType: "INFORM_DONOR_ATTORNEY_HAS_NOT_ACTED",
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

			attorneyStore := newMockAttorneyStore(t)
			attorneyStore.EXPECT().
				All(ctx, "lpa-uid").
				Return(nil, dynamo.NotFoundError{})

			eventClient := newMockEventClient(t)
			eventClient.EXPECT().
				SendLetterRequested(ctx, event.LetterRequested{
					UID:        "lpa-uid",
					LetterType: "ADVISE_ATTORNEY_TO_SIGN_OR_OPT_OUT",
					ActorType:  actor.TypeAttorney,
					ActorUID:   attorneyUID,
				}).
				Return(nil)
			eventClient.EXPECT().
				SendLetterRequested(ctx, event.LetterRequested{
					UID:        "lpa-uid",
					LetterType: "ADVISE_ATTORNEY_TO_SIGN_OR_OPT_OUT",
					ActorType:  actor.TypeTrustCorporation,
					ActorUID:   trustCorporationUID,
				}).
				Return(nil)
			eventClient.EXPECT().
				SendLetterRequested(ctx, event.LetterRequested{
					UID:        "lpa-uid",
					LetterType: "ADVISE_ATTORNEY_TO_SIGN_OR_OPT_OUT",
					ActorType:  actor.TypeReplacementAttorney,
					ActorUID:   replacementAttorneyUID,
				}).
				Return(nil)
			eventClient.EXPECT().
				SendLetterRequested(ctx, event.LetterRequested{
					UID:        "lpa-uid",
					LetterType: "ADVISE_ATTORNEY_TO_SIGN_OR_OPT_OUT",
					ActorType:  actor.TypeReplacementTrustCorporation,
					ActorUID:   replacementTrustCorporationUID,
				}).
				Return(nil)
			eventClient.EXPECT().
				SendLetterRequested(ctx, tc.donorLetterRequest).
				Return(nil)

			runner := &Runner{
				donorStore:               donorStore,
				lpaStoreResolvingService: lpaStoreResolvingService,
				attorneyStore:            attorneyStore,
				eventClient:              eventClient,
				now:                      testNowFn,
			}

			err := runner.stepRemindAttorneyToComplete(ctx, row)
			assert.Nil(t, err)
		})
	}
}

func TestRunnerRemindAttorneyToCompleteWhenNotValidTime(t *testing.T) {
	testcases := map[string]*lpadata.Lpa{
		"invite sent almost 3 months ago": {
			AttorneysInvitedAt: testNow.AddDate(0, -3, 1),
			SignedAt:           testNow.AddDate(0, -3, 0),
		},
		"expiry almost in 3 months": {
			AttorneysInvitedAt: testNow.AddDate(0, -3, 0),
			SignedAt:           testNow.AddDate(0, -3, 1),
		},
		"submitted expiry almost in 3 months": {
			Donor: lpadata.Donor{
				IdentityCheck: &lpadata.IdentityCheck{
					CheckedAt: testNow,
				},
			},
			AttorneysInvitedAt: testNow.AddDate(0, -3, 0),
			SignedAt:           testNow.AddDate(-2, 3, 1),
			Submitted:          true,
		},
	}

	for name, lpa := range testcases {
		t.Run(name, func(t *testing.T) {
			donor := &donordata.Provided{
				LpaUID: "lpa-uid",
			}

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
				now:                      testNowFn,
			}

			err := runner.stepRemindAttorneyToComplete(ctx, &Event{})
			assert.Equal(t, errStepIgnored, err)
		})
	}
}

func TestRunnerRemindAttorneyToCompleteWhenAttorneyAlreadyCompleted(t *testing.T) {
	attorneyUID := actoruid.New()
	replacementAttorneyUID := actoruid.New()
	trustCorporationUID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()

	row := &Event{
		TargetLpaKey:      dynamo.LpaKey("an-lpa"),
		TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}
	attorneys := []*attorneydata.Provided{{
		UID:      attorneyUID,
		SignedAt: time.Now(),
	}, {
		UID:                      trustCorporationUID,
		IsTrustCorporation:       true,
		WouldLikeSecondSignatory: form.No,
		AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{
			SignedAt: time.Now(),
		}},
	}, {
		UID:           replacementAttorneyUID,
		IsReplacement: true,
		SignedAt:      time.Now(),
	}, {
		UID:                      replacementTrustCorporationUID,
		IsTrustCorporation:       true,
		IsReplacement:            true,
		WouldLikeSecondSignatory: form.Yes,
		AuthorisedSignatories: [2]attorneydata.TrustCorporationSignatory{{
			SignedAt: time.Now(),
		}, {
			SignedAt: time.Now(),
		}},
	}}
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
		Attorneys: lpadata.Attorneys{
			Attorneys: []lpadata.Attorney{{
				UID:                       attorneyUID,
				FirstNames:                "c",
				LastName:                  "d",
				ContactLanguagePreference: localize.En,
			}},
			TrustCorporation: lpadata.TrustCorporation{
				UID:                       trustCorporationUID,
				Name:                      "trusty",
				ContactLanguagePreference: localize.En,
			},
		},
		ReplacementAttorneys: lpadata.Attorneys{
			Attorneys: []lpadata.Attorney{{
				UID:                       replacementAttorneyUID,
				FirstNames:                "e",
				LastName:                  "f",
				ContactLanguagePreference: localize.En,
			}},
			TrustCorporation: lpadata.TrustCorporation{
				UID:                       replacementTrustCorporationUID,
				Name:                      "untrusty",
				ContactLanguagePreference: localize.En,
			},
		},
		AttorneysInvitedAt: testNow.AddDate(0, -3, -1),
		SignedAt:           testNow.AddDate(0, -3, 0).Add(-time.Second),
	}

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything).
		Return(donor, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Resolve(mock.Anything, mock.Anything).
		Return(lpa, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		All(mock.Anything, mock.Anything).
		Return(attorneys, nil)

	runner := &Runner{
		donorStore:               donorStore,
		lpaStoreResolvingService: lpaStoreResolvingService,
		attorneyStore:            attorneyStore,
		now:                      testNowFn,
	}

	err := runner.stepRemindAttorneyToComplete(ctx, row)
	assert.Equal(t, errStepIgnored, err)
}

func TestRunnerRemindAttorneyToCompleteWhenDonorStoreErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything).
		Return(nil, expectedError)

	runner := &Runner{
		donorStore: donorStore,
		now:        testNowFn,
	}

	err := runner.stepRemindAttorneyToComplete(ctx, &Event{})
	assert.ErrorIs(t, err, expectedError)
}

func TestRunnerRemindAttorneyToCompleteWhenLpaStoreResolvingServiceErrors(t *testing.T) {
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
		lpaStoreResolvingService: lpaStoreResolvingService,
		now:                      testNowFn,
	}

	err := runner.stepRemindAttorneyToComplete(ctx, &Event{})
	assert.ErrorIs(t, err, expectedError)
}

func TestRunnerRemindAttorneyToCompleteWhenAttorneyStoreErrors(t *testing.T) {
	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything).
		Return(&donordata.Provided{}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		Resolve(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		All(mock.Anything, mock.Anything).
		Return(nil, expectedError)

	runner := &Runner{
		donorStore:               donorStore,
		lpaStoreResolvingService: lpaStoreResolvingService,
		attorneyStore:            attorneyStore,
		now:                      testNowFn,
	}

	err := runner.stepRemindAttorneyToComplete(ctx, &Event{})
	assert.ErrorIs(t, err, expectedError)
}

// func TestRunnerRemindAttorneyToCompleteWhenNotifyClientErrors(t *testing.T) {
//	testcases := map[string]func(*mockNotifyClient){
//		"first": func(m *mockNotifyClient) {
//			m.EXPECT().
//				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//				Return(expectedError).
//				Once()
//		},
//		"second": func(m *mockNotifyClient) {
//			m.EXPECT().
//				EmailGreeting(mock.Anything).
//				Return("hey")
//			m.EXPECT().
//				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//				Return(nil).
//				Once()
//			m.EXPECT().
//				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
//				Return(expectedError).
//				Once()
//		},
//	}

//	for name, setupNotifyClient := range testcases {
//		t.Run(name, func(t *testing.T) {
//			attorneyStore := newMockAttorneyStore(t)
//			attorneyStore.EXPECT().
//				One(mock.Anything, mock.Anything).
//				Return(nil, dynamo.NotFoundError{})

//			donorStore := newMockDonorStore(t)
//			donorStore.EXPECT().
//				One(mock.Anything, mock.Anything, mock.Anything).
//				Return(&donordata.Provided{}, nil)

//			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
//			lpaStoreResolvingService.EXPECT().
//				Resolve(mock.Anything, mock.Anything).
//				Return(&lpadata.Lpa{}, nil)

//			notifyClient := newMockNotifyClient(t)
//			setupNotifyClient(notifyClient)

//			localizer := &localize.Localizer{}

//			bundle := newMockBundle(t)
//			bundle.EXPECT().
//				For(mock.Anything).
//				Return(localizer)

//			runner := &Runner{
//				donorStore:               donorStore,
//				attorneyStore:            attorneyStore,
//				lpaStoreResolvingService: lpaStoreResolvingService,
//				notifyClient:             notifyClient,
//				bundle:                   bundle,
//				now:                      testNowFn,
//			}

//			err := runner.stepRemindAttorneyToComplete(ctx, &Event{})
//			assert.ErrorIs(t, err, expectedError)
//		})
//	}
// }

// func TestRunnerRemindAttorneyToCompleteWhenEventClientErrors(t *testing.T) {
//	testcases := map[string]func(*mockEventClient){
//		"first": func(m *mockEventClient) {
//			m.EXPECT().
//				SendLetterRequested(mock.Anything, mock.Anything).
//				Return(expectedError).
//				Once()
//		},
//		"second": func(m *mockEventClient) {
//			m.EXPECT().
//				SendLetterRequested(mock.Anything, mock.Anything).
//				Return(nil).
//				Once()
//			m.EXPECT().
//				SendLetterRequested(mock.Anything, mock.Anything).
//				Return(expectedError).
//				Once()
//		},
//	}

//	for name, setupEventClient := range testcases {
//		t.Run(name, func(t *testing.T) {
//			attorneyStore := newMockAttorneyStore(t)
//			attorneyStore.EXPECT().
//				One(mock.Anything, mock.Anything).
//				Return(nil, dynamo.NotFoundError{})

//			donorStore := newMockDonorStore(t)
//			donorStore.EXPECT().
//				One(mock.Anything, mock.Anything, mock.Anything).
//				Return(&donordata.Provided{}, nil)

//			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
//			lpaStoreResolvingService.EXPECT().
//				Resolve(mock.Anything, mock.Anything).
//				Return(&lpadata.Lpa{
//					Donor:    lpadata.Donor{Channel: lpadata.ChannelPaper},
//					Attorney: lpadata.Attorney{Channel: lpadata.ChannelPaper},
//				}, nil)

//			eventClient := newMockEventClient(t)
//			setupEventClient(eventClient)

//			runner := &Runner{
//				donorStore:               donorStore,
//				attorneyStore:            attorneyStore,
//				lpaStoreResolvingService: lpaStoreResolvingService,
//				eventClient:              eventClient,
//				now:                      testNowFn,
//			}

//			err := runner.stepRemindAttorneyToComplete(ctx, &Event{})
//			assert.ErrorIs(t, err, expectedError)
//		})
//	}
// }
