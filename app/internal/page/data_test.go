package page

import (
	"fmt"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
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
	values := map[CanBeUsedWhen]string{CanBeUsedWhenCapacityLost: "when-capacity-lost", CanBeUsedWhenRegistered: "when-registered"}

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
				DonorIdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin, FirstNames: "a", LastName: "b"},
			},
			expected: true,
		},
		"missing provider": {
			lpa:      &Lpa{DonorIdentityUserData: identity.UserData{OK: true}},
			expected: false,
		},
		"not ok": {
			lpa:      &Lpa{DonorIdentityUserData: identity.UserData{Provider: identity.OneLogin}},
			expected: false,
		},
		"no match": {
			lpa: &Lpa{
				Donor:                 actor.Donor{FirstNames: "a", LastName: "b"},
				DonorIdentityUserData: identity.UserData{Provider: identity.OneLogin},
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
			LpaType:           "not-a-type",
			ExpectedLegalTerm: "",
		},
		"empty": {
			LpaType:           "",
			ExpectedLegalTerm: "",
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			lpa := Lpa{Type: tc.LpaType}
			assert.Equal(t, tc.ExpectedLegalTerm, lpa.TypeLegalTermTransKey())
		})
	}
}

func TestAttorneysSigningDeadline(t *testing.T) {
	lpa := Lpa{
		Submitted: time.Date(2020, time.January, 2, 3, 4, 5, 6, time.UTC),
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
		"about payment without task": {
			lpa:      &Lpa{ID: "123"},
			url:      Paths.AboutPayment.Format("123"),
			expected: false,
		},
		"about payment with tasks": {
			lpa: &Lpa{
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
		"select your identity options without task": {
			lpa:      &Lpa{},
			url:      Paths.SelectYourIdentityOptions.Format("123"),
			expected: false,
		},
		"select your identity options with tasks": {
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
					CheckYourLpa:               actor.TaskCompleted,
					PayForLpa:                  actor.TaskCompleted,
				},
			},
			url:      Paths.SelectYourIdentityOptions.Format("123"),
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
	testCases := map[string]struct {
		lpa              *Lpa
		cp               *actor.CertificateProviderProvidedDetails
		expectedProgress Progress
	}{
		"initial state": {
			lpa: &Lpa{},
			cp:  &actor.CertificateProviderProvidedDetails{},
			expectedProgress: Progress{
				LpaSigned:                   actor.TaskInProgress,
				CertificateProviderDeclared: actor.TaskNotStarted,
				AttorneysDeclared:           actor.TaskNotStarted,
				LpaSubmitted:                actor.TaskNotStarted,
				StatutoryWaitingPeriod:      actor.TaskNotStarted,
				LpaRegistered:               actor.TaskNotStarted,
			},
		},
		"lpa signed": {
			lpa: &Lpa{Submitted: time.Now()},
			cp:  &actor.CertificateProviderProvidedDetails{},
			expectedProgress: Progress{
				LpaSigned:                   actor.TaskCompleted,
				CertificateProviderDeclared: actor.TaskInProgress,
				AttorneysDeclared:           actor.TaskNotStarted,
				LpaSubmitted:                actor.TaskNotStarted,
				StatutoryWaitingPeriod:      actor.TaskNotStarted,
				LpaRegistered:               actor.TaskNotStarted,
			},
		},
		"certificate provider declared": {
			lpa: &Lpa{Submitted: time.Now()},
			cp:  &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: time.Now()}},
			expectedProgress: Progress{
				LpaSigned:                   actor.TaskCompleted,
				CertificateProviderDeclared: actor.TaskCompleted,
				AttorneysDeclared:           actor.TaskInProgress,
				LpaSubmitted:                actor.TaskNotStarted,
				StatutoryWaitingPeriod:      actor.TaskNotStarted,
				LpaRegistered:               actor.TaskNotStarted,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expectedProgress, tc.lpa.Progress(tc.cp))
		})
	}

}

func TestActorAddresses(t *testing.T) {
	lpa := &Lpa{
		Donor: actor.Donor{Address: place.Address{Line1: "1"}},
		Attorneys: []actor.Attorney{
			{Address: place.Address{Line1: "2"}},
			{Address: place.Address{Line1: "3"}},
		},
		ReplacementAttorneys: []actor.Attorney{
			{Address: place.Address{Line1: "4"}},
			{Address: place.Address{Line1: "5"}},
		},
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
		Attorneys: []actor.Attorney{
			{FirstNames: "Attorney One", LastName: "Actor", Address: address},
			{FirstNames: "Attorney Two", LastName: "Actor"},
		},
		ReplacementAttorneys: []actor.Attorney{
			{FirstNames: "Replacement Attorney One", LastName: "Actor"},
			{FirstNames: "Replacement Attorney Two", LastName: "Actor", Address: address},
		},
		CertificateProvider: actor.CertificateProvider{FirstNames: "Certificate Provider", LastName: "Actor"},
	}

	want := []place.Address{address}

	assert.Equal(t, want, lpa.ActorAddresses())
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
		"single with email": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			taskState: actor.TaskCompleted,
		},
		"single with address": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Address:    place.Address{Line1: "a"},
			}},
			taskState: actor.TaskCompleted,
		},
		"single incomplete": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
			}},
			taskState: actor.TaskInProgress,
		},
		"multiple without decisions": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			taskState: actor.TaskInProgress,
		},
		"multiple with decisions": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			decisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState: actor.TaskCompleted,
		},
		"multiple incomplete with decisions": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			decisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState: actor.TaskInProgress,
		},
		"multiple with happy decisions": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			decisions: actor.AttorneyDecisions{How: actor.Jointly, HappyIfOneCannotActNoneCan: actor.Yes},
			taskState: actor.TaskCompleted,
		},
		"multiple with unhappy decisions": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			decisions: actor.AttorneyDecisions{How: actor.Jointly, HappyIfOneCannotActNoneCan: actor.No},
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
		want                         actor.YesNo
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
			want:      actor.No,
			taskState: actor.TaskCompleted,
		},
		"do want": {
			want:      actor.Yes,
			taskState: actor.TaskInProgress,
		},
		"single with email": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			taskState: actor.TaskCompleted,
		},
		"single with address": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Address:    place.Address{Line1: "a"},
			}},
			taskState: actor.TaskCompleted,
		},
		"single incomplete": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
			}},
			taskState: actor.TaskInProgress,
		},
		"multiple without decisions": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			taskState: actor.TaskInProgress,
		},
		"multiple jointly and severally": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:                    actor.TaskCompleted,
		},
		"multiple jointly": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:                    actor.TaskInProgress,
		},
		"multiple mixed": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:                    actor.TaskInProgress,
		},
		"multiple jointly happily": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly, HappyIfOneCannotActNoneCan: actor.Yes},
			taskState:                    actor.TaskCompleted,
		},
		"multiple mixed happily": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers, HappyIfOneCannotActNoneCan: actor.Yes},
			taskState:                    actor.TaskCompleted,
		},
		"jointly and severally attorneys single": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:         actor.TaskInProgress,
		},
		"jointly and severally attorneys single with step in": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			attorneyDecisions:     actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn: "somehow",
			taskState:             actor.TaskCompleted,
		},
		"jointly attorneys single": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:         actor.TaskCompleted,
		},
		"mixed attorneys single": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:         actor.TaskCompleted,
		},

		"jointly and severally attorneys multiple": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:         actor.TaskInProgress,
		},
		"jointly and severally attorneys multiple with step in": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:     actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn: ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			taskState:             actor.TaskCompleted,
		},
		"jointly and severally attorneys multiple with step in when none can act": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:     actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn: ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			taskState:             actor.TaskInProgress,
		},
		"jointly and severally attorneys multiple with step in when none can act jointly": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:                    actor.TaskInProgress,
		},
		"jointly and severally attorneys multiple with step in when none can act jointly happily": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly, HappyIfOneCannotActNoneCan: actor.Yes},
			taskState:                    actor.TaskCompleted,
		},
		"jointly and severally attorneys multiple with step in when none can act mixed": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:                    actor.TaskInProgress,
		},
		"jointly and severally attorneys multiple with step in when none can act mixed happily": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers, HappyIfOneCannotActNoneCan: actor.Yes},
			taskState:                    actor.TaskCompleted,
		},
		"jointly attorneys multiple without decisions": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:         actor.TaskInProgress,
		},
		"jointly attorneys multiple jointly and severally": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:                    actor.TaskCompleted,
		},
		"jointly attorneys multiple jointly": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:                    actor.TaskInProgress,
		},
		"jointly attorneys multiple with jointly happily": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly, HappyIfOneCannotActNoneCan: actor.Yes},
			taskState:                    actor.TaskCompleted,
		},
		"jointly attorneys multiple mixed": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:                    actor.TaskInProgress,
		},
		"jointly attorneys multiple with mixed happily": {
			want: actor.Yes,
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers, HappyIfOneCannotActNoneCan: actor.Yes},
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

func TestIsHealthAndWelfareLpa(t *testing.T) {
	assert.True(t, (&Lpa{Type: LpaTypeHealthWelfare}).IsHealthAndWelfareLpa())
	assert.False(t, (&Lpa{Type: LpaTypePropertyFinance}).IsHealthAndWelfareLpa())
}
