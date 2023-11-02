package page

import (
	"fmt"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
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

func TestIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Lpa
		expected bool
	}{
		"set": {
			lpa: &Lpa{
				Donor:                 actor.Donor{FirstNames: "a", LastName: "b"},
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
				Donor:                 actor.Donor{FirstNames: "a", LastName: "b"},
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
			lpa := Lpa{Type: tc.LpaType}
			assert.Equal(t, tc.ExpectedLegalTerm, lpa.Type.LegalTermTransKey())
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

func TestCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Lpa
		url      string
		expected bool
	}{
		"empty path": {
			lpa:      &Lpa{},
			url:      "",
			expected: false,
		},
		"unexpected path": {
			lpa:      &Lpa{},
			url:      "/whatever",
			expected: true,
		},
		"getting help signing no certificate provider": {
			lpa: &Lpa{
				Type: LpaTypeHealthWelfare,
				Tasks: Tasks{
					YourDetails: actor.TaskCompleted,
				},
			},
			url:      Paths.GettingHelpSigning.Format("123"),
			expected: false,
		},
		"getting help signing": {
			lpa: &Lpa{
				Type: LpaTypeHealthWelfare,
				Tasks: Tasks{
					CertificateProvider: actor.TaskCompleted,
				},
			},
			url:      Paths.GettingHelpSigning.Format("123"),
			expected: true,
		},
		"check your lpa when unsure if can sign": {
			lpa: &Lpa{
				Type: LpaTypeHealthWelfare,
				Tasks: Tasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					LifeSustainingTreatment:    actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
					PeopleToNotify:             actor.TaskCompleted,
				},
			},
			url:      Paths.CheckYourLpa.Format("123"),
			expected: false,
		},
		"check your lpa when can sign": {
			lpa: &Lpa{
				Donor: actor.Donor{CanSign: form.Yes},
				Type:  LpaTypeHealthWelfare,
				Tasks: Tasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					LifeSustainingTreatment:    actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
					PeopleToNotify:             actor.TaskCompleted,
				},
			},
			url:      Paths.CheckYourLpa.Format("123"),
			expected: true,
		},
		"about payment without task": {
			lpa:      &Lpa{ID: "123"},
			url:      Paths.AboutPayment.Format("123"),
			expected: false,
		},
		"about payment with tasks": {
			lpa: &Lpa{
				Donor: actor.Donor{
					CanSign: form.Yes,
				},
				Type: LpaTypePropertyFinance,
				Tasks: Tasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					WhenCanTheLpaBeUsed:        actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
					PeopleToNotify:             actor.TaskCompleted,
					CheckYourLpa:               actor.TaskCompleted,
				},
			},
			url:      Paths.AboutPayment.Format("123"),
			expected: true,
		},
		"identity without task": {
			lpa:      &Lpa{},
			url:      Paths.IdentityWithOneLogin.Format("123"),
			expected: false,
		},
		"identity with tasks": {
			lpa: &Lpa{
				Donor: actor.Donor{
					CanSign: form.Yes,
				},
				Type: LpaTypeHealthWelfare,
				Tasks: Tasks{
					YourDetails:                actor.TaskCompleted,
					ChooseAttorneys:            actor.TaskCompleted,
					ChooseReplacementAttorneys: actor.TaskCompleted,
					LifeSustainingTreatment:    actor.TaskCompleted,
					Restrictions:               actor.TaskCompleted,
					CertificateProvider:        actor.TaskCompleted,
					PeopleToNotify:             actor.TaskCompleted,
					CheckYourLpa:               actor.TaskCompleted,
					PayForLpa:                  actor.PaymentTaskCompleted,
				},
			},
			url:      Paths.IdentityWithOneLogin.Format("123"),
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.CanGoTo(tc.url))
		})
	}
}

func TestLpaProgress(t *testing.T) {
	lpaSignedAt := time.Now()

	testCases := map[string]struct {
		lpa                 *Lpa
		certificateProvider *actor.CertificateProviderProvidedDetails
		attorneys           []*actor.AttorneyProvidedDetails
		expectedProgress    Progress
	}{
		"initial state": {
			lpa:                 &Lpa{},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			expectedProgress: Progress{
				DonorSigned:               actor.TaskInProgress,
				CertificateProviderSigned: actor.TaskNotStarted,
				AttorneysSigned:           actor.TaskNotStarted,
				LpaSubmitted:              actor.TaskNotStarted,
				StatutoryWaitingPeriod:    actor.TaskNotStarted,
				LpaRegistered:             actor.TaskNotStarted,
			},
		},
		"lpa signed": {
			lpa:                 &Lpa{SignedAt: lpaSignedAt},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			expectedProgress: Progress{
				DonorSigned:               actor.TaskCompleted,
				CertificateProviderSigned: actor.TaskInProgress,
				AttorneysSigned:           actor.TaskNotStarted,
				LpaSubmitted:              actor.TaskNotStarted,
				StatutoryWaitingPeriod:    actor.TaskNotStarted,
				LpaRegistered:             actor.TaskNotStarted,
			},
		},
		"certificate provider signed": {
			lpa:                 &Lpa{SignedAt: lpaSignedAt},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			expectedProgress: Progress{
				DonorSigned:               actor.TaskCompleted,
				CertificateProviderSigned: actor.TaskCompleted,
				AttorneysSigned:           actor.TaskInProgress,
				LpaSubmitted:              actor.TaskNotStarted,
				StatutoryWaitingPeriod:    actor.TaskNotStarted,
				LpaRegistered:             actor.TaskNotStarted,
			},
		},
		"attorneys signed": {
			lpa: &Lpa{
				SignedAt:  lpaSignedAt,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "a1"}, {ID: "a2"}}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
				{ID: "a2", LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: Progress{
				DonorSigned:               actor.TaskCompleted,
				CertificateProviderSigned: actor.TaskCompleted,
				AttorneysSigned:           actor.TaskCompleted,
				LpaSubmitted:              actor.TaskInProgress,
				StatutoryWaitingPeriod:    actor.TaskNotStarted,
				LpaRegistered:             actor.TaskNotStarted,
			},
		},
		"submitted": {
			lpa: &Lpa{
				SignedAt:    lpaSignedAt,
				SubmittedAt: lpaSignedAt.Add(time.Hour),
				Attorneys:   actor.Attorneys{Attorneys: []actor.Attorney{{ID: "a1"}, {ID: "a2"}}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
				{ID: "a2", LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: Progress{
				DonorSigned:               actor.TaskCompleted,
				CertificateProviderSigned: actor.TaskCompleted,
				AttorneysSigned:           actor.TaskCompleted,
				LpaSubmitted:              actor.TaskCompleted,
				StatutoryWaitingPeriod:    actor.TaskInProgress,
				LpaRegistered:             actor.TaskNotStarted,
			},
		},
		"registered": {
			lpa: &Lpa{
				SignedAt:     lpaSignedAt,
				SubmittedAt:  lpaSignedAt.Add(time.Hour),
				RegisteredAt: lpaSignedAt.Add(2 * time.Hour),
				Attorneys:    actor.Attorneys{Attorneys: []actor.Attorney{{ID: "a1"}, {ID: "a2"}}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
				{ID: "a2", LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: Progress{
				DonorSigned:               actor.TaskCompleted,
				CertificateProviderSigned: actor.TaskCompleted,
				AttorneysSigned:           actor.TaskCompleted,
				LpaSubmitted:              actor.TaskCompleted,
				StatutoryWaitingPeriod:    actor.TaskCompleted,
				LpaRegistered:             actor.TaskCompleted,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedProgress, tc.lpa.Progress(tc.certificateProvider, tc.attorneys))
		})
	}
}

func TestAllAttorneysSigned(t *testing.T) {
	lpaSignedAt := time.Now()
	otherLpaSignedAt := lpaSignedAt.Add(time.Minute)
	attorneySigned := lpaSignedAt.Add(time.Second)

	testcases := map[string]struct {
		lpa       *Lpa
		attorneys []*actor.AttorneyProvidedDetails
		expected  bool
	}{
		"no attorneys": {
			expected: false,
		},
		"need attorney to sign": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{Attorneys: []actor.Attorney{{ID: "a1"}, {ID: "a2"}}},
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "r1"}}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "a3", LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
				{ID: "r1", IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"need replacement attorney to sign": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{Attorneys: []actor.Attorney{{ID: "a1"}}},
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "r1"}, {ID: "r2"}}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "r1", IsReplacement: true},
				{ID: "r2", IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"all attorneys signed": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{Attorneys: []actor.Attorney{{ID: "a1"}, {ID: "a2"}}},
				ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "r1"}}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "a2", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "r1", IsReplacement: true, LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: true,
		},
		"more attorneys signed": {
			lpa: &Lpa{
				SignedAt:  lpaSignedAt,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "a1"}, {ID: "a2"}}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "a2", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
				{ID: "a3", LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
			},
			expected: true,
		},
		"waiting for attorney to re-sign": {
			lpa: &Lpa{
				SignedAt:  lpaSignedAt,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{ID: "a1"}, {ID: "a2"}}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{ID: "a1", LpaSignedAt: otherLpaSignedAt, Confirmed: attorneySigned},
				{ID: "a2", LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned},
			},
			expected: false,
		},
		"trust corporations not signed": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "a"}},
				ReplacementAttorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "r"}},
			},
			expected: false,
		},
		"replacement trust corporations not signed": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "a"}},
				ReplacementAttorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "r"}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]actor.TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.Yes,
					AuthorisedSignatories:    [2]actor.TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
			},
			expected: false,
		},
		"trust corporations signed": {
			lpa: &Lpa{
				SignedAt:             lpaSignedAt,
				Attorneys:            actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "a"}},
				ReplacementAttorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "r"}},
			},
			attorneys: []*actor.AttorneyProvidedDetails{
				{
					IsTrustCorporation:       true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]actor.TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
				},
				{
					IsTrustCorporation:       true,
					IsReplacement:            true,
					WouldLikeSecondSignatory: form.No,
					AuthorisedSignatories:    [2]actor.TrustCorporationSignatory{{LpaSignedAt: lpaSignedAt, Confirmed: attorneySigned}},
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
		Donor: actor.Donor{Address: place.Address{Line1: "1"}},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{Address: place.Address{Line1: "2"}},
			{Address: place.Address{Line1: "3"}},
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{Address: place.Address{Line1: "4"}},
			{Address: place.Address{Line1: "5"}},
		}},
		CertificateProvider: actor.CertificateProvider{Address: place.Address{Line1: "6"}},
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
		Donor: actor.Donor{FirstNames: "Donor", LastName: "Actor", Address: address},
		Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "Attorney One", LastName: "Actor", Address: address},
			{FirstNames: "Attorney Two", LastName: "Actor"},
		}},
		ReplacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{
			{FirstNames: "Replacement Attorney One", LastName: "Actor"},
			{FirstNames: "Replacement Attorney Two", LastName: "Actor", Address: address},
		}},
		CertificateProvider: actor.CertificateProvider{FirstNames: "Certificate Provider", LastName: "Actor"},
	}

	want := []place.Address{address}

	assert.Equal(t, want, lpa.ActorAddresses())
}

func TestAllLayAttorneysFirstNames(t *testing.T) {
	lpa := &Lpa{
		Attorneys: actor.Attorneys{
			Attorneys: []actor.Attorney{
				{FirstNames: "John", LastName: "Smith"},
				{FirstNames: "Barry", LastName: "Smith"},
			},
		},
		ReplacementAttorneys: actor.Attorneys{
			Attorneys: []actor.Attorney{
				{FirstNames: "John2", LastName: "Smithe"},
				{FirstNames: "Barry2", LastName: "Smithe"},
			},
		},
	}

	assert.Equal(t, []string{"John", "Barry", "John2", "Barry2"}, lpa.AllLayAttorneysFirstNames())
}

func TestAllLayAttorneysFullNames(t *testing.T) {
	lpa := &Lpa{
		Attorneys: actor.Attorneys{
			Attorneys: []actor.Attorney{
				{FirstNames: "John", LastName: "Smith"},
				{FirstNames: "Barry", LastName: "Smith"},
			},
		},
		ReplacementAttorneys: actor.Attorneys{
			Attorneys: []actor.Attorney{
				{FirstNames: "John2", LastName: "Smithe"},
				{FirstNames: "Barry2", LastName: "Smithe"},
			},
		},
	}

	assert.Equal(t, []string{"John Smith", "Barry Smith", "John2 Smithe", "Barry2 Smithe"}, lpa.AllLayAttorneysFullNames())
}

func TestTrustCorporationOriginal(t *testing.T) {
	lpa := &Lpa{
		Attorneys:            actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "Corp"}},
		ReplacementAttorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{Name: "Trust"}},
	}

	assert.Equal(t, []string{"Corp", "Trust"}, lpa.TrustCorporationsNames())
}

func TestChooseAttorneysState(t *testing.T) {
	testcases := map[string]struct {
		attorneys actor.Attorneys
		decisions actor.AttorneyDecisions
		taskState actor.TaskState
	}{
		"empty": {
			taskState: actor.TaskNotStarted,
		},
		"trust corporation": {
			attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
				Name:    "a",
				Address: place.Address{Line1: "a"},
			}},
			taskState: actor.TaskCompleted,
		},
		"trust corporation incomplete": {
			attorneys: actor.Attorneys{TrustCorporation: actor.TrustCorporation{
				Name: "a",
			}},
			taskState: actor.TaskInProgress,
		},
		"single with email": {
			attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}}},
			taskState: actor.TaskCompleted,
		},
		"single with address": {
			attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Address:    place.Address{Line1: "a"},
			}}},
			taskState: actor.TaskCompleted,
		},
		"single incomplete": {
			attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
			}}},
			taskState: actor.TaskInProgress,
		},
		"multiple without decisions": {
			attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			taskState: actor.TaskInProgress,
		},
		"multiple with decisions": {
			attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			decisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState: actor.TaskCompleted,
		},
		"multiple incomplete with decisions": {
			attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			decisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState: actor.TaskInProgress,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.taskState, ChooseAttorneysState(tc.attorneys, tc.decisions))
		})
	}
}

func TestChooseReplacementAttorneysState(t *testing.T) {
	testcases := map[string]struct {
		want                         form.YesNo
		replacementAttorneys         actor.Attorneys
		attorneyDecisions            actor.AttorneyDecisions
		howReplacementsStepIn        ReplacementAttorneysStepIn
		replacementAttorneyDecisions actor.AttorneyDecisions
		taskState                    actor.TaskState
	}{
		"empty": {
			taskState: actor.TaskNotStarted,
		},
		"do not want": {
			want:      form.No,
			taskState: actor.TaskCompleted,
		},
		"do want": {
			want:      form.Yes,
			taskState: actor.TaskInProgress,
		},
		"single with email": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}}},
			taskState: actor.TaskCompleted,
		},
		"single with address": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Address:    place.Address{Line1: "a"},
			}}},
			taskState: actor.TaskCompleted,
		},
		"single incomplete": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
			}}},
			taskState: actor.TaskInProgress,
		},
		"multiple without decisions": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			taskState: actor.TaskInProgress,
		},
		"multiple jointly and severally": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:                    actor.TaskCompleted,
		},
		"multiple jointly": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:                    actor.TaskCompleted,
		},
		"multiple mixed": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:                    actor.TaskCompleted,
		},
		"jointly and severally attorneys single": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:         actor.TaskInProgress,
		},
		"jointly and severally attorneys single with step in": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}}},
			attorneyDecisions:     actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn: ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			taskState:             actor.TaskCompleted,
		},
		"jointly attorneys single": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:         actor.TaskCompleted,
		},
		"mixed attorneys single": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:         actor.TaskCompleted,
		},

		"jointly and severally attorneys multiple": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:         actor.TaskInProgress,
		},
		"jointly and severally attorneys multiple with step in": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			attorneyDecisions:     actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn: ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			taskState:             actor.TaskCompleted,
		},
		"jointly and severally attorneys multiple with step in when none can act": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			attorneyDecisions:     actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn: ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			taskState:             actor.TaskInProgress,
		},
		"jointly and severally attorneys multiple with step in when none can act jointly": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:                    actor.TaskCompleted,
		},
		"jointly and severally attorneys multiple with step in when none can act mixed": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:                    actor.TaskCompleted,
		},
		"jointly attorneys multiple without decisions": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:         actor.TaskInProgress,
		},
		"jointly attorneys multiple jointly and severally": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:                    actor.TaskCompleted,
		},
		"jointly attorneys multiple with jointly": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:                    actor.TaskCompleted,
		},
		"jointly attorneys multiple mixed": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:                    actor.TaskCompleted,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.taskState, ChooseReplacementAttorneysState(&Lpa{
				WantReplacementAttorneys:            tc.want,
				AttorneyDecisions:                   tc.attorneyDecisions,
				ReplacementAttorneys:                tc.replacementAttorneys,
				ReplacementAttorneyDecisions:        tc.replacementAttorneyDecisions,
				HowShouldReplacementAttorneysStepIn: tc.howReplacementsStepIn,
			}))
		})
	}
}

func TestLpaCost(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Lpa
		expected int
	}{
		"denied": {
			lpa:      &Lpa{FeeType: pay.HalfFee, Tasks: Tasks{PayForLpa: actor.PaymentTaskDenied}},
			expected: 8200,
		},
		"half": {
			lpa:      &Lpa{FeeType: pay.HalfFee},
			expected: 4100,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.Cost())
		})
	}
}

func TestFeeAmount(t *testing.T) {
	testCases := map[string]struct {
		Lpa          *Lpa
		ExpectedCost int
	}{
		"not paid": {
			Lpa:          &Lpa{FeeType: pay.HalfFee},
			ExpectedCost: 4100,
		},
		"fully paid": {
			Lpa:          &Lpa{FeeType: pay.HalfFee, PaymentDetails: []Payment{{Amount: 4100}}},
			ExpectedCost: 0,
		},
		"denied partially paid": {
			Lpa:          &Lpa{FeeType: pay.HalfFee, PaymentDetails: []Payment{{Amount: 4100}}, Tasks: Tasks{PayForLpa: actor.PaymentTaskDenied}},
			ExpectedCost: 4100,
		},
		"denied fully paid": {
			Lpa:          &Lpa{FeeType: pay.HalfFee, PaymentDetails: []Payment{{Amount: 4100}, {Amount: 4100}}, Tasks: Tasks{PayForLpa: actor.PaymentTaskDenied}},
			ExpectedCost: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedCost, tc.Lpa.FeeAmount())
		})
	}

}

func TestCertificateProviderSharesDetailsNames(t *testing.T) {
	testcases := map[string]struct {
		certificateProvider  string
		donor                string
		attorneys            []string
		replacementAttorneys []string
		expected             bool
	}{
		"no match": {
			certificateProvider:  "a",
			attorneys:            []string{"b"},
			replacementAttorneys: []string{"c"},
		},
		"match donor": {
			certificateProvider: "a",
			donor:               "a",
			expected:            true,
		},
		"match attorney": {
			certificateProvider: "a",
			attorneys:           []string{"b", "a"},
			expected:            true,
		},
		"match replacement attorney": {
			certificateProvider:  "a",
			replacementAttorneys: []string{"b", "a"},
			expected:             true,
		},
		"half-start on certificate provider match donor": {
			certificateProvider: "a-c",
			donor:               "a",
			expected:            true,
		},
		"half-end on certificate provider match donor": {
			certificateProvider: "c-a",
			donor:               "a",
			expected:            true,
		},
		"half-start on donor match donor": {
			certificateProvider: "a",
			donor:               "a-c",
			expected:            true,
		},
		"half-end on donor match donor": {
			certificateProvider: "a",
			donor:               "c-a",
			expected:            true,
		},
		"half-start on certificate provider match attorney": {
			certificateProvider: "a-c",
			attorneys:           []string{"b", "a"},
			expected:            true,
		},
		"half-end on certificate provider match attorney": {
			certificateProvider: "c-a",
			attorneys:           []string{"b", "a"},
			expected:            true,
		},
		"half-start on attorney match attorney": {
			certificateProvider: "a",
			attorneys:           []string{"b", "a-c"},
			expected:            true,
		},
		"half-end on attorney match attorney": {
			certificateProvider: "a",
			attorneys:           []string{"b", "c-a"},
			expected:            true,
		},
		"half-start on certificate provider match replacement attorney": {
			certificateProvider:  "a-c",
			replacementAttorneys: []string{"b", "a"},
			expected:             true,
		},
		"half-end on certificate provider match replacement attorney": {
			certificateProvider:  "c-a",
			replacementAttorneys: []string{"b", "a"},
			expected:             true,
		},
		"half-start on replacement attorney match replacement attorney": {
			certificateProvider:  "a",
			replacementAttorneys: []string{"b", "a-c"},
			expected:             true,
		},
		"half-end on replacement attorney match replacement attorney": {
			certificateProvider:  "a",
			replacementAttorneys: []string{"b", "c-a"},
			expected:             true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			lpa := &Lpa{
				Donor:               actor.Donor{LastName: tc.donor},
				CertificateProvider: actor.CertificateProvider{LastName: tc.certificateProvider, Address: place.Address{Line1: "x"}},
			}

			for _, a := range tc.attorneys {
				lpa.Attorneys.Attorneys = append(lpa.Attorneys.Attorneys, actor.Attorney{LastName: a})
			}

			for _, a := range tc.replacementAttorneys {
				lpa.ReplacementAttorneys.Attorneys = append(lpa.ReplacementAttorneys.Attorneys, actor.Attorney{LastName: a})
			}

			assert.Equal(t, tc.expected, lpa.CertificateProviderSharesDetails())
		})
	}
}

func TestCertificateProviderSharesDetailsAddresses(t *testing.T) {
	a := place.Address{Line1: "a", Postcode: "a"}
	b := place.Address{Line1: "b", Postcode: "a"}
	c := place.Address{Line1: "a", Postcode: "b"}

	testcases := map[string]struct {
		certificateProvider  place.Address
		donor                place.Address
		attorneys            []place.Address
		replacementAttorneys []place.Address
		expected             bool
	}{
		"no match": {
			certificateProvider:  a,
			attorneys:            []place.Address{b},
			replacementAttorneys: []place.Address{c},
		},
		"match donor": {
			certificateProvider: a,
			donor:               a,
			expected:            true,
		},
		"match attorney": {
			certificateProvider: a,
			attorneys:           []place.Address{b, a},
			expected:            true,
		},
		"match replacement attorney": {
			certificateProvider:  a,
			replacementAttorneys: []place.Address{b, a},
			expected:             true,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			lpa := &Lpa{
				Donor:               actor.Donor{Address: tc.donor},
				CertificateProvider: actor.CertificateProvider{LastName: "x", Address: tc.certificateProvider},
			}

			for _, attorney := range tc.attorneys {
				lpa.Attorneys.Attorneys = append(lpa.Attorneys.Attorneys, actor.Attorney{Address: attorney})
			}

			for _, attorney := range tc.replacementAttorneys {
				lpa.ReplacementAttorneys.Attorneys = append(lpa.ReplacementAttorneys.Attorneys, actor.Attorney{Address: attorney})
			}

			assert.Equal(t, tc.expected, lpa.CertificateProviderSharesDetails())
		})
	}
}
