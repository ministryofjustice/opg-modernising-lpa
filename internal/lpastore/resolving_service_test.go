package lpastore

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestResolvingServiceGet(t *testing.T) {
	testcases := map[string]struct {
		donor    *actor.DonorProvidedDetails
		resolved *Lpa
		error    error
		expected *Lpa
	}{
		"online with all true": {
			donor: &actor.DonorProvidedDetails{
				SK:          dynamo.OrganisationKey("S"),
				LpaID:       "1",
				LpaUID:      "M-1111",
				SubmittedAt: time.Now(),
				CertificateProvider: actor.CertificateProvider{
					FirstNames:   "Barry",
					Relationship: actor.Personally,
				},
				Tasks: actor.DonorTasks{
					PayForLpa: actor.PaymentTaskCompleted,
				},
				DonorIdentityUserData: identity.UserData{
					OK: true,
				},
			},
			resolved: &Lpa{
				LpaID: "1",
				CertificateProvider: CertificateProvider{
					FirstNames: "Paul",
				},
			},
			expected: &Lpa{
				LpaID:                  "1",
				LpaUID:                 "M-1111",
				DonorIdentityConfirmed: true,
				Submitted:              true,
				Paid:                   true,
				IsOrganisationDonor:    true,
				CertificateProvider: CertificateProvider{
					FirstNames:   "Paul",
					Relationship: actor.Personally,
				},
				Donor: actor.Donor{Channel: actor.ChannelOnline},
			},
		},
		"online with no lpastore record": {
			donor: &actor.DonorProvidedDetails{
				SK:     dynamo.DonorKey("S"),
				LpaUID: "M-1111",
				CertificateProvider: actor.CertificateProvider{
					FirstNames:   "John",
					Relationship: actor.Personally,
				},
				Donor: actor.Donor{Channel: actor.ChannelOnline},
				Attorneys: actor.Attorneys{
					Attorneys:        []actor.Attorney{{FirstNames: "a"}},
					TrustCorporation: actor.TrustCorporation{Name: "b"},
				},
				ReplacementAttorneys: actor.Attorneys{
					Attorneys:        []actor.Attorney{{FirstNames: "c"}},
					TrustCorporation: actor.TrustCorporation{Name: "d"},
				},
			},
			error: ErrNotFound,
			expected: &Lpa{
				LpaUID: "M-1111",
				CertificateProvider: CertificateProvider{
					FirstNames:   "John",
					Relationship: actor.Personally,
				},
				Donor: actor.Donor{Channel: actor.ChannelOnline},
				Attorneys: Attorneys{
					Attorneys:        []Attorney{{FirstNames: "a"}},
					TrustCorporation: TrustCorporation{Name: "b"},
				},
				ReplacementAttorneys: Attorneys{
					Attorneys:        []Attorney{{FirstNames: "c"}},
					TrustCorporation: TrustCorporation{Name: "d"},
				},
			},
		},
		"online with all false": {
			donor: &actor.DonorProvidedDetails{
				SK:     dynamo.DonorKey("S"),
				LpaID:  "1",
				LpaUID: "M-1111",
			},
			resolved: &Lpa{LpaID: "1"},
			expected: &Lpa{
				LpaID:  "1",
				LpaUID: "M-1111",
				Donor:  actor.Donor{Channel: actor.ChannelOnline},
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
				LpaID:     "1",
				LpaUID:    "M-1111",
				Submitted: true,
				Paid:      true,
				CertificateProvider: CertificateProvider{
					Relationship: actor.Professionally,
				},
				Donor: actor.Donor{Channel: actor.ChannelPaper},
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
				Return(tc.resolved, tc.error)

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
		Donor:  actor.Donor{Channel: actor.ChannelOnline},
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
		Return(&actor.DonorProvidedDetails{LpaUID: "M-1111"}, nil)

	lpaClient := newMockLpaClient(t)
	lpaClient.EXPECT().
		Lpa(ctx, mock.Anything).
		Return(nil, expectedError)

	service := NewResolvingService(donorStore, lpaClient)
	_, err := service.Get(ctx)

	assert.Equal(t, expectedError, err)
}
