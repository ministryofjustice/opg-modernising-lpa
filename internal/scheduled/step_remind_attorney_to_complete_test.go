package scheduled

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

			attorneysInvitedAt := testNow.AddDate(0, -3, -1)
			signedAt := testNow.AddDate(0, -3, 0).Add(-time.Second)

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
				AttorneysInvitedAt: attorneysInvitedAt,
				SignedAt:           signedAt,
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
				All(ctx, row.TargetLpaKey).
				Return(tc.attorneys, tc.attorneyError)

			notifyClient := newMockNotifyClient(t)
			notifyClient.EXPECT().
				EmailGreeting(lpa).
				Return("hey")
			notifyClient.EXPECT().
				SendActorEmail(ctx, notify.ToLpaAttorney(lpa.Attorneys.Attorneys[0]), "lpa-uid", notify.AdviseAttorneyToSignOrOptOutEmail{
					DonorFullName:           "a b",
					DonorFullNamePossessive: "a b’s",
					LpaType:                 "Personal welfare",
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
					LpaType:                 "Personal welfare",
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
					LpaType:                 "Personal welfare",
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
					LpaType:                 "Personal welfare",
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
					LpaType:              "Personal welfare",
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
					LpaType:              "Personal welfare",
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
					LpaType:              "Personal welfare",
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
					LpaType:              "Personal welfare",
					InvitedDate:          "1 October 1999",
					DeadlineDate:         "2 April 2000",
					AttorneyStartPageURL: "http://app/attorney-start",
				}).
				Return(nil).
				Once()

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				Possessive("a b").
				Return("a b’s")
			localizer.EXPECT().
				T("personal-welfare").
				Return("Personal welfare").
				Times(8)
			localizer.EXPECT().
				FormatDate(attorneysInvitedAt).
				Return("1 October 1999")
			localizer.EXPECT().
				FormatDate(signedAt.AddDate(0, 6, 0)).
				Return("2 April 2000")

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
				All(ctx, row.TargetLpaKey).
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

func TestRunnerRemindAttorneyToCompleteWhenAttorneysOnPaper(t *testing.T) {
	donorUID := actoruid.New()
	attorneyUID := actoruid.New()
	replacementAttorneyUID := actoruid.New()
	trustCorporationUID := actoruid.New()
	replacementTrustCorporationUID := actoruid.New()

	signedAt := testNow.AddDate(0, -3, 0).Add(-time.Second)
	attorneysInvitedAt := testNow.AddDate(0, -3, -1).Add(-time.Second)

	lpa := &lpadata.Lpa{
		LpaUID: "lpa-uid",
		Type:   lpadata.LpaTypePersonalWelfare,
		Donor: lpadata.Donor{
			UID:                       donorUID,
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
		SignedAt:           signedAt,
		AttorneysInvitedAt: attorneysInvitedAt,
	}

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
		Return(lpa, nil)

	attorneyStore := newMockAttorneyStore(t)
	attorneyStore.EXPECT().
		All(ctx, row.TargetLpaKey).
		Return(nil, dynamo.NotFoundError{})

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(lpa).
		Return("hey")
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorPaperAttorneyHasNotActedEmail{
			Greeting:         "hey",
			AttorneyFullName: "c d",
			LpaType:          "Personal welfare",
			PostedDate:       "1 October 1999",
			DeadlineDate:     "2 April 2000",
		}).
		Return(nil).
		Once()
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorPaperAttorneyHasNotActedEmail{
			Greeting:         "hey",
			AttorneyFullName: "trusty",
			LpaType:          "Personal welfare",
			PostedDate:       "1 October 1999",
			DeadlineDate:     "2 April 2000",
		}).
		Return(nil).
		Once()
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorPaperAttorneyHasNotActedEmail{
			Greeting:         "hey",
			AttorneyFullName: "e f",
			LpaType:          "Personal welfare",
			PostedDate:       "1 October 1999",
			DeadlineDate:     "2 April 2000",
		}).
		Return(nil).
		Once()
	notifyClient.EXPECT().
		SendActorEmail(ctx, notify.ToLpaDonor(lpa), "lpa-uid", notify.InformDonorPaperAttorneyHasNotActedEmail{
			Greeting:         "hey",
			AttorneyFullName: "untrusty",
			LpaType:          "Personal welfare",
			PostedDate:       "1 October 1999",
			DeadlineDate:     "2 April 2000",
		}).
		Return(nil).
		Once()

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

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T("personal-welfare").
		Return("Personal welfare").
		Times(4)
	localizer.EXPECT().
		FormatDate(attorneysInvitedAt).
		Return("1 October 1999")
	localizer.EXPECT().
		FormatDate(signedAt.AddDate(0, 6, 0)).
		Return("2 April 2000")

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(localize.En).
		Return(localizer)

	runner := &Runner{
		donorStore:               donorStore,
		lpaStoreResolvingService: lpaStoreResolvingService,
		attorneyStore:            attorneyStore,
		eventClient:              eventClient,
		notifyClient:             notifyClient,
		bundle:                   bundle,
		now:                      testNowFn,
	}

	err := runner.stepRemindAttorneyToComplete(ctx, row)
	assert.Nil(t, err)
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

func TestRunnerRemindAttorneyToCompleteWhenNotifyClientErrors(t *testing.T) {
	actorUID := actoruid.New()

	notifyCases := map[string]func(*mockNotifyClient){
		"email to attorney": func(notifyClient *mockNotifyClient) {
			notifyClient.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError).
				Once()
		},
		"email to donor": func(notifyClient *mockNotifyClient) {
			notifyClient.EXPECT().
				EmailGreeting(mock.Anything).
				Return("hey")
			notifyClient.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(nil).
				Once()
			notifyClient.EXPECT().
				SendActorEmail(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
				Return(expectedError).
				Once()
		},
	}

	attorneysInvitedAt := testNow.AddDate(0, -3, -1)
	signedAt := testNow.AddDate(0, -3, 0).Add(-time.Second)

	lpaCases := map[string]*lpadata.Lpa{
		"attorney": {
			LpaUID: "lpa-uid",
			Type:   lpadata.LpaTypePersonalWelfare,
			Donor: lpadata.Donor{
				FirstNames:                "a",
				LastName:                  "b",
				ContactLanguagePreference: localize.En,
			},
			Attorneys: lpadata.Attorneys{
				Attorneys: []lpadata.Attorney{{
					UID:                       actorUID,
					FirstNames:                "c",
					LastName:                  "d",
					ContactLanguagePreference: localize.En,
				}},
			},
			AttorneysInvitedAt: attorneysInvitedAt,
			SignedAt:           signedAt,
		},
		"replacement attorney": {
			LpaUID: "lpa-uid",
			Type:   lpadata.LpaTypePersonalWelfare,
			Donor: lpadata.Donor{
				FirstNames:                "a",
				LastName:                  "b",
				ContactLanguagePreference: localize.En,
			},
			ReplacementAttorneys: lpadata.Attorneys{
				Attorneys: []lpadata.Attorney{{
					UID:                       actorUID,
					FirstNames:                "e",
					LastName:                  "f",
					ContactLanguagePreference: localize.En,
				}},
			},
			AttorneysInvitedAt: attorneysInvitedAt,
			SignedAt:           signedAt,
		},
		"trust corporation": {
			LpaUID: "lpa-uid",
			Type:   lpadata.LpaTypePersonalWelfare,
			Donor: lpadata.Donor{
				FirstNames:                "a",
				LastName:                  "b",
				ContactLanguagePreference: localize.En,
			},
			Attorneys: lpadata.Attorneys{
				TrustCorporation: lpadata.TrustCorporation{
					UID:                       actorUID,
					Name:                      "trusty",
					ContactLanguagePreference: localize.En,
				},
			},
			AttorneysInvitedAt: attorneysInvitedAt,
			SignedAt:           signedAt,
		},
		"replacement trust corporation": {
			LpaUID: "lpa-uid",
			Type:   lpadata.LpaTypePersonalWelfare,
			Donor: lpadata.Donor{
				FirstNames:                "a",
				LastName:                  "b",
				ContactLanguagePreference: localize.En,
			},
			ReplacementAttorneys: lpadata.Attorneys{
				TrustCorporation: lpadata.TrustCorporation{
					UID:                       actorUID,
					Name:                      "untrusty",
					ContactLanguagePreference: localize.En,
				},
			},
			AttorneysInvitedAt: attorneysInvitedAt,
			SignedAt:           signedAt,
		},
	}

	for setupName, setupNotify := range notifyCases {
		for name, lpa := range lpaCases {
			t.Run(name+" "+setupName, func(t *testing.T) {
				row := &Event{
					TargetLpaKey:      dynamo.LpaKey("an-lpa"),
					TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
				}
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

				attorneyStore := newMockAttorneyStore(t)
				attorneyStore.EXPECT().
					All(mock.Anything, mock.Anything).
					Return([]*attorneydata.Provided{{}}, nil)

				notifyClient := newMockNotifyClient(t)
				setupNotify(notifyClient)

				localizer := newMockLocalizer(t)
				localizer.EXPECT().
					Possessive(mock.Anything).
					Return("a b’s")
				localizer.EXPECT().
					T(mock.Anything).
					Return("Personal welfare")
				localizer.EXPECT().
					FormatDate(mock.Anything).
					Return("1 October 1999")
				localizer.EXPECT().
					FormatDate(mock.Anything).
					Return("2 April 2000")

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
				assert.ErrorIs(t, err, expectedError)
			})
		}
	}
}

func TestRunnerRemindAttorneyToCompleteWhenEventClientErrors(t *testing.T) {
	actorUID := actoruid.New()

	eventCases := map[string]func(*mockEventClient){
		"email to attorney": func(eventClient *mockEventClient) {
			eventClient.EXPECT().
				SendLetterRequested(mock.Anything, mock.Anything).
				Return(expectedError).
				Once()
		},
		"email to donor": func(eventClient *mockEventClient) {
			eventClient.EXPECT().
				SendLetterRequested(mock.Anything, mock.Anything).
				Return(nil).
				Once()
			eventClient.EXPECT().
				SendLetterRequested(mock.Anything, mock.Anything).
				Return(expectedError).
				Once()
		},
	}

	lpaCases := map[string]*lpadata.Lpa{
		"attorney": {
			LpaUID: "lpa-uid",
			Type:   lpadata.LpaTypePersonalWelfare,
			Donor: lpadata.Donor{
				FirstNames:                "a",
				LastName:                  "b",
				ContactLanguagePreference: localize.En,
				Channel:                   lpadata.ChannelPaper,
			},
			Attorneys: lpadata.Attorneys{
				Attorneys: []lpadata.Attorney{{
					UID:                       actorUID,
					FirstNames:                "c",
					LastName:                  "d",
					ContactLanguagePreference: localize.En,
					Channel:                   lpadata.ChannelPaper,
				}},
			},
			AttorneysInvitedAt: testNow.AddDate(0, -3, -1),
			SignedAt:           testNow.AddDate(0, -3, 0).Add(-time.Second),
		},
		"replacement attorney": {
			LpaUID: "lpa-uid",
			Type:   lpadata.LpaTypePersonalWelfare,
			Donor: lpadata.Donor{
				FirstNames:                "a",
				LastName:                  "b",
				ContactLanguagePreference: localize.En,
				Channel:                   lpadata.ChannelPaper,
			},
			ReplacementAttorneys: lpadata.Attorneys{
				Attorneys: []lpadata.Attorney{{
					UID:                       actorUID,
					FirstNames:                "e",
					LastName:                  "f",
					ContactLanguagePreference: localize.En,
					Channel:                   lpadata.ChannelPaper,
				}},
			},
			AttorneysInvitedAt: testNow.AddDate(0, -3, -1),
			SignedAt:           testNow.AddDate(0, -3, 0).Add(-time.Second),
		},
		"trust corporation": {
			LpaUID: "lpa-uid",
			Type:   lpadata.LpaTypePersonalWelfare,
			Donor: lpadata.Donor{
				FirstNames:                "a",
				LastName:                  "b",
				ContactLanguagePreference: localize.En,
				Channel:                   lpadata.ChannelPaper,
			},
			Attorneys: lpadata.Attorneys{
				TrustCorporation: lpadata.TrustCorporation{
					UID:                       actorUID,
					Name:                      "trusty",
					ContactLanguagePreference: localize.En,
					Channel:                   lpadata.ChannelPaper,
				},
			},
			AttorneysInvitedAt: testNow.AddDate(0, -3, -1),
			SignedAt:           testNow.AddDate(0, -3, 0).Add(-time.Second),
		},
		"replacement trust corporation": {
			LpaUID: "lpa-uid",
			Type:   lpadata.LpaTypePersonalWelfare,
			Donor: lpadata.Donor{
				FirstNames:                "a",
				LastName:                  "b",
				ContactLanguagePreference: localize.En,
				Channel:                   lpadata.ChannelPaper,
			},
			ReplacementAttorneys: lpadata.Attorneys{
				TrustCorporation: lpadata.TrustCorporation{
					UID:                       actorUID,
					Name:                      "untrusty",
					ContactLanguagePreference: localize.En,
					Channel:                   lpadata.ChannelPaper,
				},
			},
			AttorneysInvitedAt: testNow.AddDate(0, -3, -1),
			SignedAt:           testNow.AddDate(0, -3, 0).Add(-time.Second),
		},
	}

	for setupName, setupEvent := range eventCases {
		for name, lpa := range lpaCases {
			t.Run(name+" "+setupName, func(t *testing.T) {
				row := &Event{
					TargetLpaKey:      dynamo.LpaKey("an-lpa"),
					TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
				}
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

				attorneyStore := newMockAttorneyStore(t)
				attorneyStore.EXPECT().
					All(mock.Anything, mock.Anything).
					Return([]*attorneydata.Provided{{}}, nil)

				eventClient := newMockEventClient(t)
				setupEvent(eventClient)

				runner := &Runner{
					donorStore:               donorStore,
					lpaStoreResolvingService: lpaStoreResolvingService,
					attorneyStore:            attorneyStore,
					eventClient:              eventClient,
					now:                      testNowFn,
					appPublicURL:             "http://app",
				}

				err := runner.stepRemindAttorneyToComplete(ctx, row)
				assert.ErrorIs(t, err, expectedError)
			})
		}
	}
}
