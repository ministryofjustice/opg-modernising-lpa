package scheduled

import (
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
