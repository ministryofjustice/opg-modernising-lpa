package lpastore

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var testNow = time.Now()

func TestResolvingServiceGet(t *testing.T) {
	actorUID := actoruid.New()

	testcases := map[string]struct {
		donor    *donordata.Provided
		resolved *lpadata.Lpa
		error    error
		expected *lpadata.Lpa
	}{
		"online with all true": {
			donor: &donordata.Provided{
				SK:          dynamo.LpaOwnerKey(dynamo.OrganisationKey("S")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				SubmittedAt: time.Now(),
				CertificateProvider: donordata.CertificateProvider{
					FirstNames:   "Barry",
					Relationship: lpadata.Personally,
				},
				Tasks: donordata.Tasks{
					CheckYourLpa: task.StateCompleted,
					PayForLpa:    task.PaymentStateCompleted,
				},
				IdentityUserData: identity.UserData{
					Status:    identity.StatusConfirmed,
					CheckedAt: testNow,
				},
				Correspondent:       donordata.Correspondent{Email: "x"},
				AuthorisedSignatory: donordata.AuthorisedSignatory{UID: actorUID, FirstNames: "A", LastName: "S"},
				IndependentWitness:  donordata.IndependentWitness{UID: actorUID, FirstNames: "I", LastName: "W"},
				Voucher:             donordata.Voucher{Allowed: true, Email: "y"},
			},
			resolved: &lpadata.Lpa{
				LpaID: "1",
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "Paul",
				},
			},
			expected: &lpadata.Lpa{
				LpaOwnerKey:         dynamo.LpaOwnerKey(dynamo.OrganisationKey("S")),
				LpaID:               "1",
				LpaUID:              "M-1111",
				Drafted:             true,
				Submitted:           true,
				Paid:                true,
				IsOrganisationDonor: true,
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames:   "Paul",
					Relationship: lpadata.Personally,
				},
				Donor: lpadata.Donor{
					Channel: lpadata.ChannelOnline,
					IdentityCheck: &lpadata.IdentityCheck{
						Type:      "one-login",
						CheckedAt: testNow,
					},
				},
				Correspondent: lpadata.Correspondent{Email: "x"},
				Voucher:       lpadata.Voucher{Email: "y"},
			},
		},
		"online with no lpastore record": {
			donor: &donordata.Provided{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaUID: "M-1111",
				CertificateProvider: donordata.CertificateProvider{
					FirstNames:   "John",
					Relationship: lpadata.Personally,
				},
				Donor: donordata.Donor{Channel: lpadata.ChannelOnline},
				Attorneys: donordata.Attorneys{
					Attorneys:        []donordata.Attorney{{FirstNames: "a"}},
					TrustCorporation: donordata.TrustCorporation{Name: "b"},
				},
				ReplacementAttorneys: donordata.Attorneys{
					Attorneys:        []donordata.Attorney{{FirstNames: "c"}},
					TrustCorporation: donordata.TrustCorporation{Name: "d"},
				},
				IdentityUserData: identity.UserData{
					Status:    identity.StatusConfirmed,
					CheckedAt: time.Date(2020, time.January, 2, 12, 13, 14, 5, time.UTC),
				},
				Correspondent:       donordata.Correspondent{Email: "x"},
				AuthorisedSignatory: donordata.AuthorisedSignatory{UID: actorUID, FirstNames: "A", LastName: "S"},
				IndependentWitness:  donordata.IndependentWitness{UID: actorUID, FirstNames: "I", LastName: "W"},
				Voucher:             donordata.Voucher{Allowed: true, Email: "y"},
			},
			error: ErrNotFound,
			expected: &lpadata.Lpa{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaUID:      "M-1111",
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames:   "John",
					Relationship: lpadata.Personally,
				},
				Donor: lpadata.Donor{
					Channel: lpadata.ChannelOnline,
					IdentityCheck: &lpadata.IdentityCheck{
						CheckedAt: time.Date(2020, time.January, 2, 12, 13, 14, 5, time.UTC),
						Type:      "one-login",
					},
				},
				Attorneys: lpadata.Attorneys{
					Attorneys:        []lpadata.Attorney{{FirstNames: "a"}},
					TrustCorporation: lpadata.TrustCorporation{Name: "b"},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					Attorneys:        []lpadata.Attorney{{FirstNames: "c"}},
					TrustCorporation: lpadata.TrustCorporation{Name: "d"},
				},
				Correspondent: lpadata.Correspondent{Email: "x"},
				AuthorisedSignatory: lpadata.AuthorisedSignatory{
					UID:        actorUID,
					FirstNames: "A",
					LastName:   "S",
				},
				IndependentWitness: lpadata.IndependentWitness{
					UID:        actorUID,
					FirstNames: "I",
					LastName:   "W",
				},
				Voucher: lpadata.Voucher{Email: "y"},
			},
		},
		"online with all false": {
			donor: &donordata.Provided{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaID:  "1",
				LpaUID: "M-1111",
			},
			resolved: &lpadata.Lpa{LpaID: "1"},
			expected: &lpadata.Lpa{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				Donor:       lpadata.Donor{Channel: lpadata.ChannelOnline},
			},
		},
		"paper": {
			donor: &donordata.Provided{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:  "1",
				LpaUID: "M-1111",
			},
			resolved: &lpadata.Lpa{LpaID: "1"},
			expected: &lpadata.Lpa{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				Submitted:   true,
				Drafted:     true,
				Paid:        true,
				CertificateProvider: lpadata.CertificateProvider{
					Relationship: lpadata.Professionally,
				},
				Donor: lpadata.Donor{Channel: lpadata.ChannelPaper},
			},
		},
		"voucher not allowed": {
			donor: &donordata.Provided{
				SK:      dynamo.LpaOwnerKey(dynamo.OrganisationKey("S")),
				LpaID:   "1",
				LpaUID:  "M-1111",
				Voucher: donordata.Voucher{Email: "y"},
			},
			resolved: &lpadata.Lpa{
				LpaID: "1",
			},
			expected: &lpadata.Lpa{
				LpaOwnerKey:         dynamo.LpaOwnerKey(dynamo.OrganisationKey("S")),
				LpaID:               "1",
				LpaUID:              "M-1111",
				IsOrganisationDonor: true,
				Donor:               lpadata.Donor{Channel: lpadata.ChannelOnline},
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
		Return(&donordata.Provided{LpaID: "1"}, nil)

	service := NewResolvingService(donorStore, nil)
	lpa, err := service.Get(ctx)

	assert.Equal(t, &lpadata.Lpa{
		LpaID: "1",
		Donor: lpadata.Donor{Channel: lpadata.ChannelOnline},
	}, lpa)
	assert.Nil(t, err)
}

func TestResolvingServiceGetWhenNotFound(t *testing.T) {
	ctx := context.Background()

	donorStore := newMockDonorStore(t)
	donorStore.EXPECT().
		GetAny(ctx).
		Return(&donordata.Provided{LpaID: "1", LpaUID: "M-1111"}, nil)

	lpaClient := newMockLpaClient(t)
	lpaClient.EXPECT().
		Lpa(ctx, mock.Anything).
		Return(nil, ErrNotFound)

	service := NewResolvingService(donorStore, lpaClient)
	lpa, err := service.Get(ctx)

	assert.Equal(t, &lpadata.Lpa{
		LpaID:  "1",
		LpaUID: "M-1111",
		Donor:  lpadata.Donor{Channel: lpadata.ChannelOnline},
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
		Return(&donordata.Provided{LpaUID: "M-1111"}, nil)

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
		donors   []*donordata.Provided
		uids     []string
		resolved []*lpadata.Lpa
		error    error
		expected []*lpadata.Lpa
	}{
		"online with all true": {
			donors: []*donordata.Provided{{
				SK:          dynamo.LpaOwnerKey(dynamo.OrganisationKey("S")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				SubmittedAt: time.Now(),
				CertificateProvider: donordata.CertificateProvider{
					FirstNames:   "Barry",
					Relationship: lpadata.Personally,
				},
				Tasks: donordata.Tasks{
					CheckYourLpa: task.StateCompleted,
					PayForLpa:    task.PaymentStateCompleted,
				},
				IdentityUserData: identity.UserData{
					Status:    identity.StatusConfirmed,
					CheckedAt: testNow,
				},
			}},
			uids: []string{"M-1111"},
			resolved: []*lpadata.Lpa{{
				LpaID:  "1",
				LpaUID: "M-1111",
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames: "Paul",
				},
			}},
			expected: []*lpadata.Lpa{{
				LpaOwnerKey:         dynamo.LpaOwnerKey(dynamo.OrganisationKey("S")),
				LpaID:               "1",
				LpaUID:              "M-1111",
				Drafted:             true,
				Submitted:           true,
				Paid:                true,
				IsOrganisationDonor: true,
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames:   "Paul",
					Relationship: lpadata.Personally,
				},
				Donor: lpadata.Donor{
					Channel: lpadata.ChannelOnline,
					IdentityCheck: &lpadata.IdentityCheck{
						CheckedAt: testNow,
						Type:      "one-login",
					},
				},
			}},
		},
		"online with no lpastore record": {
			donors: []*donordata.Provided{{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaUID: "M-1111",
				CertificateProvider: donordata.CertificateProvider{
					FirstNames:   "John",
					Relationship: lpadata.Personally,
				},
				Donor: donordata.Donor{Channel: lpadata.ChannelOnline},
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
			expected: []*lpadata.Lpa{{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaUID:      "M-1111",
				CertificateProvider: lpadata.CertificateProvider{
					FirstNames:   "John",
					Relationship: lpadata.Personally,
				},
				Donor: lpadata.Donor{Channel: lpadata.ChannelOnline},
				Attorneys: lpadata.Attorneys{
					Attorneys:        []lpadata.Attorney{{FirstNames: "a"}},
					TrustCorporation: lpadata.TrustCorporation{Name: "b"},
				},
				ReplacementAttorneys: lpadata.Attorneys{
					Attorneys:        []lpadata.Attorney{{FirstNames: "c"}},
					TrustCorporation: lpadata.TrustCorporation{Name: "d"},
				},
			}},
		},
		"online with all false": {
			donors: []*donordata.Provided{{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaID:  "1",
				LpaUID: "M-1111",
			}},
			uids:     []string{"M-1111"},
			resolved: []*lpadata.Lpa{{LpaID: "1", LpaUID: "M-1111"}},
			expected: []*lpadata.Lpa{{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("S")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				Donor:       lpadata.Donor{Channel: lpadata.ChannelOnline},
			}},
		},
		"paper": {
			donors: []*donordata.Provided{{
				SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:  "1",
				LpaUID: "M-1111",
			}},
			uids:     []string{"M-1111"},
			resolved: []*lpadata.Lpa{{LpaID: "1", LpaUID: "M-1111"}},
			expected: []*lpadata.Lpa{{
				LpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
				LpaID:       "1",
				LpaUID:      "M-1111",
				Drafted:     true,
				Submitted:   true,
				Paid:        true,
				CertificateProvider: lpadata.CertificateProvider{
					Relationship: lpadata.Professionally,
				},
				Donor: lpadata.Donor{Channel: lpadata.ChannelPaper},
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
