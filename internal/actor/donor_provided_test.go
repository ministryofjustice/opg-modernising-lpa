package actor

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
)

var address = place.Address{
	Line1:      "a",
	Line2:      "b",
	Line3:      "c",
	TownOrCity: "d",
	Postcode:   "e",
}

func TestGenerateHash(t *testing.T) {
	donor := &DonorProvidedDetails{}
	hash, err := donor.GenerateHash()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0xfc0fd048c2122462), hash)

	donor.LpaID = "1"
	hash, err = donor.GenerateHash()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x5ebbfe89cc2fbb8), hash)
}

func TestIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		lpa      *DonorProvidedDetails
		expected bool
	}{
		"set": {
			lpa: &DonorProvidedDetails{
				Donor:                 Donor{FirstNames: "a", LastName: "b"},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b"},
			},
			expected: true,
		},
		"not ok": {
			lpa:      &DonorProvidedDetails{DonorIdentityUserData: identity.UserData{}},
			expected: false,
		},
		"no match": {
			lpa: &DonorProvidedDetails{
				Donor:                 Donor{FirstNames: "a", LastName: "b"},
				DonorIdentityUserData: identity.UserData{},
			},
			expected: false,
		},
		"none": {
			lpa:      &DonorProvidedDetails{},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.DonorIdentityConfirmed())
		})
	}
}

func TestAttorneysSigningDeadline(t *testing.T) {
	donor := DonorProvidedDetails{
		SignedAt: time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC),
	}

	expected := time.Date(2020, time.January, 30, 3, 4, 5, 6, time.UTC)
	assert.Equal(t, expected, donor.AttorneysAndCpSigningDeadline())
}

func TestAllAttorneysSigned(t *testing.T) {
	lpaSignedAt := time.Now()
	otherLpaSignedAt := lpaSignedAt.Add(time.Minute)
	attorneySigned := lpaSignedAt.Add(time.Second)

	testcases := map[string]struct {
		lpa       *DonorProvidedDetails
		attorneys []*AttorneyProvidedDetails
		expected  bool
	}{
		"no attorneys": {
			expected: false,
		},
		"need attorney to sign": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{Attorneys: []Attorney{{ID: "a1"}, {ID: "a2"}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{ID: "r1"}}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "a3", LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
				{ID: "r1", IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"need replacement attorney to sign": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{Attorneys: []Attorney{{ID: "a1"}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{ID: "r1"}, {ID: "r2"}}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "r1", IsReplacement: true},
				{ID: "r2", IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"all attorneys signed": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{Attorneys: []Attorney{{ID: "a1"}, {ID: "a2"}}},
				ReplacementAttorneys: Attorneys{Attorneys: []Attorney{{ID: "r1"}}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "a2", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "r1", IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: true,
		},
		"more attorneys signed": {
			lpa: &DonorProvidedDetails{
				SignedAt:  lpaSignedAt,
				Attorneys: Attorneys{Attorneys: []Attorney{{ID: "a1"}, {ID: "a2"}}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "a2", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "a3", LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
			},
			expected: true,
		},
		"waiting for attorney to re-sign": {
			lpa: &DonorProvidedDetails{
				SignedAt:  lpaSignedAt,
				Attorneys: Attorneys{Attorneys: []Attorney{{ID: "a1"}, {ID: "a2"}}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
				{ID: "a2", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"trust corporations not signed": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r"}},
			},
			expected: false,
		},
		"replacement trust corporations not signed": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r"}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.Yes,
					AuthorisedSignatories:    [2]TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
			},
			expected: false,
		},
		"trust corporations signed": {
			lpa: &DonorProvidedDetails{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r"}},
			},
			attorneys: []*AttorneyProvidedDetails{
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
				{
					IsTrustCorporation:       true,
					IsReplacement:            true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
			},
			expected: true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.AllAttorneysSigned(tc.attorneys))
		})
	}
}

func TestActorAddresses(t *testing.T) {
	donor := &DonorProvidedDetails{
		Donor: Donor{Address: place.Address{Line1: "1"}},
		Attorneys: Attorneys{Attorneys: []Attorney{
			{Address: place.Address{Line1: "2"}},
			{Address: place.Address{Line1: "3"}},
		}},
		ReplacementAttorneys: Attorneys{Attorneys: []Attorney{
			{Address: place.Address{Line1: "4"}},
			{Address: place.Address{Line1: "5"}},
		}},
		CertificateProvider: CertificateProvider{Address: place.Address{Line1: "6"}},
	}

	want := []place.Address{
		{Line1: "1"},
		{Line1: "6"},
		{Line1: "2"},
		{Line1: "3"},
		{Line1: "4"},
		{Line1: "5"},
	}

	assert.Equal(t, want, donor.ActorAddresses())
}

func TestActorAddressesActorWithNoAddressIgnored(t *testing.T) {
	donor := &DonorProvidedDetails{
		Donor: Donor{FirstNames: "Donor", LastName: "Actor", Address: address},
		Attorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "Attorney One", LastName: "Actor", Address: address},
			{FirstNames: "Attorney Two", LastName: "Actor"},
		}},
		ReplacementAttorneys: Attorneys{Attorneys: []Attorney{
			{FirstNames: "Replacement Attorney One", LastName: "Actor"},
			{FirstNames: "Replacement Attorney Two", LastName: "Actor", Address: address},
		}},
		CertificateProvider: CertificateProvider{FirstNames: "Certificate Provider", LastName: "Actor"},
	}

	want := []place.Address{address}

	assert.Equal(t, want, donor.ActorAddresses())
}

func TestAllLayAttorneysFirstNames(t *testing.T) {
	donor := &DonorProvidedDetails{
		Attorneys: Attorneys{
			Attorneys: []Attorney{
				{FirstNames: "John", LastName: "Smith"},
				{FirstNames: "Barry", LastName: "Smith"},
			},
		},
		ReplacementAttorneys: Attorneys{
			Attorneys: []Attorney{
				{FirstNames: "John2", LastName: "Smithe"},
				{FirstNames: "Barry2", LastName: "Smithe"},
			},
		},
	}

	assert.Equal(t, []string{"John", "Barry", "John2", "Barry2"}, donor.AllLayAttorneysFirstNames())
}

func TestAllLayAttorneysFullNames(t *testing.T) {
	donor := &DonorProvidedDetails{
		Attorneys: Attorneys{
			Attorneys: []Attorney{
				{FirstNames: "John", LastName: "Smith"},
				{FirstNames: "Barry", LastName: "Smith"},
			},
		},
		ReplacementAttorneys: Attorneys{
			Attorneys: []Attorney{
				{FirstNames: "John2", LastName: "Smithe"},
				{FirstNames: "Barry2", LastName: "Smithe"},
			},
		},
	}

	assert.Equal(t, []string{"John Smith", "Barry Smith", "John2 Smithe", "Barry2 Smithe"}, donor.AllLayAttorneysFullNames())
}

func TestTrustCorporationOriginal(t *testing.T) {
	donor := &DonorProvidedDetails{
		Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "Corp"}},
		ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "Trust"}},
	}

	assert.Equal(t, []string{"Corp", "Trust"}, donor.TrustCorporationsNames())
}
