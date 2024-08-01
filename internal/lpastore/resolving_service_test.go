package lpastore

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	donordata "github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
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
				SK:          dynamo.LpaOwnerKey(dynamo.OrganisationKey("S")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				SubmittedAt: time.Now(),
				CertificateProvider: donordata.CertificateProvider{
					FirstNames:   "Barry",
					Relationship: donordata.Personally,
				},
				Tasks: actor.DonorTasks{
					CheckYourLpa: actor.TaskCompleted,
					PayForLpa:    actor.PaymentTaskCompleted,
				},
				DonorIdentityUserData: identity.UserData{
					Status:      identity.StatusConfirmed,
					RetrievedAt: time.Now(),
				},
			},
			resolved: &Lpa{
				LpaID: "1",
				CertificateProvider: CertificateProvider{
					FirstNames: "Paul",
				},
			},
			expected: &Lpa{
				LpaOwnerKey:         dynamo.LpaOwnerKey(dynamo.OrganisationKey("S")),
				LpaID:               "1",
				LpaUID:              "M-1111",
				Drafted:             true,
				Submitted:           true,
				Paid:                true,
				IsOrganisationDonor: true,
				CertificateProvider: CertificateProvider{
					FirstNames:   "Paul",
					Relationship: donordata.Personally,
				},
				Donor: Donor{Channel: donordata.ChannelOnline},
			},
		},
		"online with no lpastore record": {
			donor: &actor.DonorProvidedDetails{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaUID: "M-1111",
				CertificateProvider: donordata.CertificateProvider{
					FirstNames:   "John",
					Relationship: donordata.Personally,
				},
				Donor: actor.Donor{Channel: donordata.ChannelOnline},
				Attorneys: donordata.Attorneys{
					Attorneys:        []donordata.Attorney{{FirstNames: "a"}},
					TrustCorporation: donordata.TrustCorporation{Name: "b"},
				},
				ReplacementAttorneys: donordata.Attorneys{
					Attorneys:        []donordata.Attorney{{FirstNames: "c"}},
					TrustCorporation: donordata.TrustCorporation{Name: "d"},
				},
				DonorIdentityUserData: identity.UserData{
					Status:      identity.StatusConfirmed,
					RetrievedAt: time.Date(2020, time.January, 2, 12, 13, 14, 5, time.UTC),
				},
			},
			error: ErrNotFound,
			expected: &Lpa{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaUID:      "M-1111",
				CertificateProvider: CertificateProvider{
					FirstNames:   "John",
					Relationship: donordata.Personally,
				},
				Donor: Donor{
					Channel: donordata.ChannelOnline,
					IdentityCheck: IdentityCheck{
						CheckedAt: time.Date(2020, time.January, 2, 12, 13, 14, 5, time.UTC),
						Type:      "one-login",
					},
				},
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
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaID:  "1",
				LpaUID: "M-1111",
			},
			resolved: &Lpa{LpaID: "1"},
			expected: &Lpa{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				Donor:       Donor{Channel: donordata.ChannelOnline},
			},
		},
		"paper": {
			donor: &actor.DonorProvidedDetails{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:  "1",
				LpaUID: "M-1111",
			},
			resolved: &Lpa{LpaID: "1"},
			expected: &Lpa{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				Submitted:   true,
				Drafted:     true,
				Paid:        true,
				CertificateProvider: CertificateProvider{
					Relationship: donordata.Professionally,
				},
				Donor: Donor{Channel: donordata.ChannelPaper},
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

func TestResolvingServiceGetWhenNoUID(t *testing.T) {
	ctx := context.Background()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(ctx).
		Return(&actor.DonorProvidedDetails{LpaID: "1"}, nil)

	service := NewResolvingService(donorStore, nil)
	lpa, err := service.Get(ctx)

	assert.Equal(t, &Lpa{
		LpaID: "1",
		Donor: Donor{Channel: donordata.ChannelOnline},
	}, lpa)
	assert.Nil(t, err)
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
		Donor:  Donor{Channel: donordata.ChannelOnline},
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

func TestResolvingServiceResolveList(t *testing.T) {
	testcases := map[string]struct {
		donors   []*actor.DonorProvidedDetails
		uids     []string
		resolved []*Lpa
		error    error
		expected []*Lpa
	}{
		"online with all true": {
			donors: []*actor.DonorProvidedDetails{{
				SK:          dynamo.LpaOwnerKey(dynamo.OrganisationKey("S")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				SubmittedAt: time.Now(),
				CertificateProvider: donordata.CertificateProvider{
					FirstNames:   "Barry",
					Relationship: donordata.Personally,
				},
				Tasks: actor.DonorTasks{
					CheckYourLpa: actor.TaskCompleted,
					PayForLpa:    actor.PaymentTaskCompleted,
				},
				DonorIdentityUserData: identity.UserData{
					Status: identity.StatusConfirmed,
				},
			}},
			uids: []string{"M-1111"},
			resolved: []*Lpa{{
				LpaID:  "1",
				LpaUID: "M-1111",
				CertificateProvider: CertificateProvider{
					FirstNames: "Paul",
				},
			}},
			expected: []*Lpa{{
				LpaOwnerKey:         dynamo.LpaOwnerKey(dynamo.OrganisationKey("S")),
				LpaID:               "1",
				LpaUID:              "M-1111",
				Drafted:             true,
				Submitted:           true,
				Paid:                true,
				IsOrganisationDonor: true,
				CertificateProvider: CertificateProvider{
					FirstNames:   "Paul",
					Relationship: donordata.Personally,
				},
				Donor: Donor{Channel: donordata.ChannelOnline},
			}},
		},
		"online with no lpastore record": {
			donors: []*actor.DonorProvidedDetails{{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaUID: "M-1111",
				CertificateProvider: donordata.CertificateProvider{
					FirstNames:   "John",
					Relationship: donordata.Personally,
				},
				Donor: actor.Donor{Channel: donordata.ChannelOnline},
				Attorneys: donordata.Attorneys{
					Attorneys:        []donordata.Attorney{{FirstNames: "a"}},
					TrustCorporation: donordata.TrustCorporation{Name: "b"},
				},
				ReplacementAttorneys: donordata.Attorneys{
					Attorneys:        []donordata.Attorney{{FirstNames: "c"}},
					TrustCorporation: donordata.TrustCorporation{Name: "d"},
				},
			}},
			uids: []string{"M-1111"},
			expected: []*Lpa{{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaUID:      "M-1111",
				CertificateProvider: CertificateProvider{
					FirstNames:   "John",
					Relationship: donordata.Personally,
				},
				Donor: Donor{Channel: donordata.ChannelOnline},
				Attorneys: Attorneys{
					Attorneys:        []Attorney{{FirstNames: "a"}},
					TrustCorporation: TrustCorporation{Name: "b"},
				},
				ReplacementAttorneys: Attorneys{
					Attorneys:        []Attorney{{FirstNames: "c"}},
					TrustCorporation: TrustCorporation{Name: "d"},
				},
			}},
		},
		"online with all false": {
			donors: []*actor.DonorProvidedDetails{{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaID:  "1",
				LpaUID: "M-1111",
			}},
			uids:     []string{"M-1111"},
			resolved: []*Lpa{{LpaID: "1", LpaUID: "M-1111"}},
			expected: []*Lpa{{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				Donor:       Donor{Channel: donordata.ChannelOnline},
			}},
		},
		"paper": {
			donors: []*actor.DonorProvidedDetails{{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:  "1",
				LpaUID: "M-1111",
			}},
			uids:     []string{"M-1111"},
			resolved: []*Lpa{{LpaID: "1", LpaUID: "M-1111"}},
			expected: []*Lpa{{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				Drafted:     true,
				Submitted:   true,
				Paid:        true,
				CertificateProvider: CertificateProvider{
					Relationship: donordata.Professionally,
				},
				Donor: Donor{Channel: donordata.ChannelPaper},
			}},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()

			lpaClient := newMockLpaClient(t)
			lpaClient.EXPECT().
				Lpas(ctx, tc.uids).
				Return(tc.resolved, tc.error)

			service := NewResolvingService(nil, lpaClient)
			lpas, err := service.ResolveList(ctx, tc.donors)

			assert.Nil(t, err)
			assert.Equal(t, tc.expected, lpas)
		})
	}
}
