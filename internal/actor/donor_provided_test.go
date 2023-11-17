package actor

import (
	"fmt"
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

func TestLpaType(t *testing.T) {
	values := map[LpaType]string{LpaTypeHealthWelfare: "hw", LpaTypePropertyFinance: "pfa"}

	for value, s := range values {
		t.Run(fmt.Sprintf("parse %s", s), func(t *testing.T) {
			parsed, err := ParseLpaType(s)
			assert.Nil(t, err)
			assert.Equal(t, value, parsed)
		})

		t.Run(fmt.Sprintf("string %s", s), func(t *testing.T) {
			assert.Equal(t, s, value.String())
		})
	}

	t.Run("parse invalid", func(t *testing.T) {
		_, err := ParseLpaType("invalid")
		assert.NotNil(t, err)
	})

	t.Run("IsHealthWelfare", func(t *testing.T) {
		assert.True(t, LpaTypeHealthWelfare.IsHealthWelfare())
		assert.False(t, LpaTypePropertyFinance.IsHealthWelfare())
	})

	t.Run("IsPropertyFinance", func(t *testing.T) {
		assert.True(t, LpaTypePropertyFinance.IsPropertyFinance())
		assert.False(t, LpaTypeHealthWelfare.IsPropertyFinance())
	})
}

func TestTypeLegalTermTransKey(t *testing.T) {
	testCases := map[string]struct {
		LpaType           LpaType
		ExpectedLegalTerm string
	}{
		"PFA": {
			LpaType:           LpaTypePropertyFinance,
			ExpectedLegalTerm: "pfaLegalTerm",
		},
		"HW": {
			LpaType:           LpaTypeHealthWelfare,
			ExpectedLegalTerm: "hwLegalTerm",
		},
		"unexpected": {
			LpaType:           LpaType(5),
			ExpectedLegalTerm: "",
		},
		"empty": {
			ExpectedLegalTerm: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedLegalTerm, tc.LpaType.LegalTermTransKey())
		})
	}
}

func TestCanBeUsedWhen(t *testing.T) {
	values := map[CanBeUsedWhen]string{CanBeUsedWhenCapacityLost: "when-capacity-lost", CanBeUsedWhenHasCapacity: "when-has-capacity"}

	for value, s := range values {
		t.Run(fmt.Sprintf("parse(%s)", s), func(t *testing.T) {
			parsed, err := ParseCanBeUsedWhen(s)
			assert.Nil(t, err)
			assert.Equal(t, value, parsed)
		})

		t.Run(fmt.Sprintf("string(%s)", s), func(t *testing.T) {
			assert.Equal(t, s, value.String())
		})
	}

	t.Run("parse invalid", func(t *testing.T) {
		_, err := ParseCanBeUsedWhen("invalid")
		assert.NotNil(t, err)
	})
}

func TestLifeSustainingTreatment(t *testing.T) {
	values := map[LifeSustainingTreatment]string{LifeSustainingTreatmentOptionA: "option-a", LifeSustainingTreatmentOptionB: "option-b"}

	for value, s := range values {
		t.Run(fmt.Sprintf("parse(%s)", s), func(t *testing.T) {
			parsed, err := ParseLifeSustainingTreatment(s)
			assert.Nil(t, err)
			assert.Equal(t, value, parsed)
		})

		t.Run(fmt.Sprintf("string(%s)", s), func(t *testing.T) {
			assert.Equal(t, s, value.String())
		})
	}

	t.Run("parse invalid", func(t *testing.T) {
		_, err := ParseLifeSustainingTreatment("invalid")
		assert.NotNil(t, err)
	})

	t.Run("IsOptionA", func(t *testing.T) {
		assert.True(t, LifeSustainingTreatmentOptionA.IsOptionA())
		assert.False(t, LifeSustainingTreatmentOptionB.IsOptionA())
	})

	t.Run("IsOptionB", func(t *testing.T) {
		assert.True(t, LifeSustainingTreatmentOptionB.IsOptionB())
		assert.False(t, LifeSustainingTreatmentOptionA.IsOptionB())
	})
}

func TestReplacementAttorneysStepIn(t *testing.T) {
	values := map[ReplacementAttorneysStepIn]string{
		ReplacementAttorneysStepInWhenAllCanNoLongerAct: "all",
		ReplacementAttorneysStepInWhenOneCanNoLongerAct: "one",
		ReplacementAttorneysStepInAnotherWay:            "other",
	}

	for value, s := range values {
		t.Run(fmt.Sprintf("parse(%s)", s), func(t *testing.T) {
			parsed, err := ParseReplacementAttorneysStepIn(s)
			assert.Nil(t, err)
			assert.Equal(t, value, parsed)
		})

		t.Run(fmt.Sprintf("string(%s)", s), func(t *testing.T) {
			assert.Equal(t, s, value.String())
		})
	}

	t.Run("parse invalid", func(t *testing.T) {
		_, err := ParseReplacementAttorneysStepIn("invalid")
		assert.NotNil(t, err)
	})

	t.Run("IsWhenAllCanNoLongerAct", func(t *testing.T) {
		assert.True(t, ReplacementAttorneysStepInWhenAllCanNoLongerAct.IsWhenAllCanNoLongerAct())
		assert.False(t, ReplacementAttorneysStepInWhenOneCanNoLongerAct.IsWhenAllCanNoLongerAct())
		assert.False(t, ReplacementAttorneysStepInAnotherWay.IsWhenAllCanNoLongerAct())
	})

	t.Run("IsWhenOneCanNoLongerAct", func(t *testing.T) {
		assert.True(t, ReplacementAttorneysStepInWhenOneCanNoLongerAct.IsWhenOneCanNoLongerAct())
		assert.False(t, ReplacementAttorneysStepInWhenAllCanNoLongerAct.IsWhenOneCanNoLongerAct())
		assert.False(t, ReplacementAttorneysStepInAnotherWay.IsWhenOneCanNoLongerAct())
	})

	t.Run("IsAnotherWay", func(t *testing.T) {
		assert.True(t, ReplacementAttorneysStepInAnotherWay.IsAnotherWay())
		assert.False(t, ReplacementAttorneysStepInWhenAllCanNoLongerAct.IsAnotherWay())
		assert.False(t, ReplacementAttorneysStepInWhenOneCanNoLongerAct.IsAnotherWay())
	})
}

func TestGenerateHash(t *testing.T) {
	lpa := &Lpa{}
	hash, err := lpa.GenerateHash()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0x4b37df4a36f24c8), hash)

	lpa.ID = "1"
	hash, err = lpa.GenerateHash()
	assert.Nil(t, err)
	assert.Equal(t, uint64(0xb058317f6e9a325b), hash)
}

func TestIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Lpa
		expected bool
	}{
		"set": {
			lpa: &Lpa{
				Donor:                 Donor{FirstNames: "a", LastName: "b"},
				DonorIdentityUserData: identity.UserData{OK: true, FirstNames: "a", LastName: "b"},
			},
			expected: true,
		},
		"not ok": {
			lpa:      &Lpa{DonorIdentityUserData: identity.UserData{}},
			expected: false,
		},
		"no match": {
			lpa: &Lpa{
				Donor:                 Donor{FirstNames: "a", LastName: "b"},
				DonorIdentityUserData: identity.UserData{},
			},
			expected: false,
		},
		"none": {
			lpa:      &Lpa{},
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
	lpa := Lpa{
		SignedAt: time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC),
	}

	expected := time.Date(2020, time.January, 30, 3, 4, 5, 6, time.UTC)
	assert.Equal(t, expected, lpa.AttorneysAndCpSigningDeadline())
}

func TestAllAttorneysSigned(t *testing.T) {
	lpaSignedAt := time.Now()
	otherLpaSignedAt := lpaSignedAt.Add(time.Minute)
	attorneySigned := lpaSignedAt.Add(time.Second)

	testcases := map[string]struct {
		lpa       *Lpa
		attorneys []*AttorneyProvidedDetails
		expected  bool
	}{
		"no attorneys": {
			expected: false,
		},
		"need attorney to sign": {
			lpa: &Lpa{
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
			lpa: &Lpa{
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
			lpa: &Lpa{
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
			lpa: &Lpa{
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
			lpa: &Lpa{
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
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "a"}},
				ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "r"}},
			},
			expected: false,
		},
		"replacement trust corporations not signed": {
			lpa: &Lpa{
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
			lpa: &Lpa{
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
	lpa := &Lpa{
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

	assert.Equal(t, want, lpa.ActorAddresses())
}

func TestActorAddressesActorWithNoAddressIgnored(t *testing.T) {
	lpa := &Lpa{
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

	assert.Equal(t, want, lpa.ActorAddresses())
}

func TestAllLayAttorneysFirstNames(t *testing.T) {
	lpa := &Lpa{
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

	assert.Equal(t, []string{"John", "Barry", "John2", "Barry2"}, lpa.AllLayAttorneysFirstNames())
}

func TestAllLayAttorneysFullNames(t *testing.T) {
	lpa := &Lpa{
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

	assert.Equal(t, []string{"John Smith", "Barry Smith", "John2 Smithe", "Barry2 Smithe"}, lpa.AllLayAttorneysFullNames())
}

func TestTrustCorporationOriginal(t *testing.T) {
	lpa := &Lpa{
		Attorneys:            Attorneys{TrustCorporation: TrustCorporation{Name: "Corp"}},
		ReplacementAttorneys: Attorneys{TrustCorporation: TrustCorporation{Name: "Trust"}},
	}

	assert.Equal(t, []string{"Corp", "Trust"}, lpa.TrustCorporationsNames())
}
