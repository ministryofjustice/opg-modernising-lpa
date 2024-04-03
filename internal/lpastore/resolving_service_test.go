package lpastore

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestResolvingServiceGet(t *testing.T) {
	testcases := map[string]struct {
		donor    *actor.DonorProvidedDetails
		resolved *Lpa
		expected *Lpa
	}{
		"digital with all true": {
			donor: &actor.DonorProvidedDetails{
				SK:          dynamo.OrganisationKey("S"),
				LpaID:       "1",
				LpaUID:      "M-1111",
				SubmittedAt: time.Now(),
				CertificateProvider: actor.CertificateProvider{
					Relationship: actor.Personally,
				},
				Tasks: actor.DonorTasks{
					PayForLpa: actor.PaymentTaskCompleted,
				},
				DonorIdentityUserData: identity.UserData{
					OK: true,
				},
			},
			resolved: &Lpa{LpaID: "1"},
			expected: &Lpa{
				LpaID:                  "1",
				LpaUID:                 "M-1111",
				DonorIdentityConfirmed: true,
				Submitted:              true,
				Paid:                   true,
				IsOrganisationDonor:    true,
				CertificateProvider: actor.CertificateProvider{
					Relationship: actor.Personally,
				},
			},
		},
		"digital with all false": {
			donor: &actor.DonorProvidedDetails{
				SK:     dynamo.DonorKey("S"),
				LpaID:  "1",
				LpaUID: "M-1111",
			},
			resolved: &Lpa{LpaID: "1"},
			expected: &Lpa{
				LpaID:  "1",
				LpaUID: "M-1111",
			},
		},
		"paper": {
			donor: &actor.DonorProvidedDetails{
				SK:     dynamo.DonorKey("PAPER"),
				LpaID:  "1",
				LpaUID: "M-1111",
			},
			resolved: &Lpa{LpaID: "1"},
			expected: &Lpa{
				LpaID:                  "1",
				LpaUID:                 "M-1111",
				DonorIdentityConfirmed: true,
				Submitted:              true,
				Paid:                   true,
				CertificateProvider: actor.CertificateProvider{
					Relationship: actor.Professionally,
				},
			},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			donorStore := newMockDonorStore(t)
			donorStore.EXPECT().
				GetAny(ctx).
				Return(tc.donor, nil)

			lpaClient := newMockLpaClient(t)
			lpaClient.EXPECT().
				Lpa(ctx, "M-1111").
				Return(tc.resolved, nil)

			service := NewResolvingService(donorStore, lpaClient)
			lpa, err := service.Get(ctx)

			assert.Nil(t, err)
			assert.Equal(t, tc.expected, lpa)
		})
	}
}

func TestResolvingServiceGetWhenNotFound(t *testing.T) {
	ctx := context.Background()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(ctx).
		Return(&actor.DonorProvidedDetails{LpaID: "1", LpaUID: "M-1111"}, nil)

	lpaClient := newMockLpaClient(t)
	lpaClient.EXPECT().
		Lpa(ctx, mock.Anything).
		Return(nil, ErrNotFound)

	service := NewResolvingService(donorStore, lpaClient)
	lpa, err := service.Get(ctx)

	assert.Equal(t, &Lpa{
		LpaID:  "1",
		LpaUID: "M-1111",
	}, lpa)
	assert.Nil(t, err)
}

func TestResolvingServiceGetWhenDonorStoreErrors(t *testing.T) {
	ctx := context.Background()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(ctx).
		Return(nil, expectedError)

	service := NewResolvingService(donorStore, nil)
	_, err := service.Get(ctx)

	assert.Equal(t, expectedError, err)
}

func TestResolvingServiceGetWhenLpaClientErrors(t *testing.T) {
	ctx := context.Background()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(ctx).
		Return(&actor.DonorProvidedDetails{}, nil)

	lpaClient := newMockLpaClient(t)
	lpaClient.EXPECT().
		Lpa(ctx, mock.Anything).
		Return(nil, expectedError)

	service := NewResolvingService(donorStore, lpaClient)
	_, err := service.Get(ctx)

	assert.Equal(t, expectedError, err)
}
