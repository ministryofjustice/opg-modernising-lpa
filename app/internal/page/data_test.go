package page

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
)

var validAttorney = actor.Attorney{
	ID:          "123",
	Address:     address,
	FirstNames:  "Joan",
	LastName:    "Jones",
	DateOfBirth: date.New("2000", "1", "2"),
}

var address = place.Address{
	Line1:      "a",
	Line2:      "b",
	Line3:      "c",
	TownOrCity: "d",
	Postcode:   "e",
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
		LpaType           string
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
			lpa:      &Lpa{},
			url:      Paths.AboutPayment,
			expected: false,
		},
		"about payment with tasks": {
			lpa: &Lpa{Tasks: Tasks{
				YourDetails:                TaskCompleted,
				ChooseAttorneys:            TaskCompleted,
				ChooseReplacementAttorneys: TaskCompleted,
				WhenCanTheLpaBeUsed:        TaskCompleted,
				Restrictions:               TaskCompleted,
				CertificateProvider:        TaskCompleted,
				PeopleToNotify:             TaskCompleted,
				CheckYourLpa:               TaskCompleted,
			}},
			url:      Paths.AboutPayment,
			expected: true,
		},
		"select your identity options without task": {
			lpa:      &Lpa{},
			url:      Paths.SelectYourIdentityOptions,
			expected: false,
		},
		"select your identity options with task": {
			lpa:      &Lpa{Tasks: Tasks{PayForLpa: TaskCompleted}},
			url:      Paths.SelectYourIdentityOptions,
			expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.CanGoTo(tc.url))
		})
	}
}

func TestTaskStateString(t *testing.T) {
	testCases := []struct {
		State    TaskState
		Expected string
	}{
		{
			State:    TaskNotStarted,
			Expected: "notStarted",
		},
		{
			State:    TaskInProgress,
			Expected: "inProgress",
		},
		{
			State:    TaskCompleted,
			Expected: "completed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Expected, func(t *testing.T) {
			assert.Equal(t, tc.Expected, tc.State.String())
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
				LpaSigned:                   TaskInProgress,
				CertificateProviderDeclared: TaskNotStarted,
				AttorneysDeclared:           TaskNotStarted,
				LpaSubmitted:                TaskNotStarted,
				StatutoryWaitingPeriod:      TaskNotStarted,
				LpaRegistered:               TaskNotStarted,
			},
		},
		"lpa signed": {
			lpa: &Lpa{Submitted: time.Now()},
			cp:  &actor.CertificateProviderProvidedDetails{},
			expectedProgress: Progress{
				LpaSigned:                   TaskCompleted,
				CertificateProviderDeclared: TaskInProgress,
				AttorneysDeclared:           TaskNotStarted,
				LpaSubmitted:                TaskNotStarted,
				StatutoryWaitingPeriod:      TaskNotStarted,
				LpaRegistered:               TaskNotStarted,
			},
		},
		"certificate provider declared": {
			lpa: &Lpa{Submitted: time.Now()},
			cp:  &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: time.Now()}},
			expectedProgress: Progress{
				LpaSigned:                   TaskCompleted,
				CertificateProviderDeclared: TaskCompleted,
				AttorneysDeclared:           TaskInProgress,
				LpaSubmitted:                TaskNotStarted,
				StatutoryWaitingPeriod:      TaskNotStarted,
				LpaRegistered:               TaskNotStarted,
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
		Donor: actor.Donor{FirstNames: "Donor", LastName: "Actor", Address: address},
		Attorneys: []actor.Attorney{
			{FirstNames: "Attorney One", LastName: "Actor", Address: address},
			{FirstNames: "Attorney Two", LastName: "Actor", Address: address},
		},
		ReplacementAttorneys: []actor.Attorney{
			{FirstNames: "Replacement Attorney One", LastName: "Actor", Address: address},
			{FirstNames: "Replacement Attorney Two", LastName: "Actor", Address: address},
		},
		CertificateProviderDetails: actor.CertificateProvider{FirstNames: "Certificate Provider", LastName: "Actor", Address: address},
	}

	want := []AddressDetail{
		{Name: "Donor Actor", Role: actor.TypeDonor, Address: address},
		{Name: "Certificate Provider Actor", Role: actor.TypeCertificateProvider, Address: address},
		{Name: "Attorney One Actor", Role: actor.TypeAttorney, Address: address},
		{Name: "Attorney Two Actor", Role: actor.TypeAttorney, Address: address},
		{Name: "Replacement Attorney One Actor", Role: actor.TypeReplacementAttorney, Address: address},
		{Name: "Replacement Attorney Two Actor", Role: actor.TypeReplacementAttorney, Address: address},
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
		CertificateProviderDetails: actor.CertificateProvider{FirstNames: "Certificate Provider", LastName: "Actor"},
	}

	want := []AddressDetail{
		{Name: "Donor Actor", Role: actor.TypeDonor, Address: address},
		{Name: "Attorney One Actor", Role: actor.TypeAttorney, Address: address},
		{Name: "Replacement Attorney Two Actor", Role: actor.TypeReplacementAttorney, Address: address},
	}

	assert.Equal(t, want, lpa.ActorAddresses())
}

func TestChooseAttorneysState(t *testing.T) {
	testcases := map[string]struct {
		attorneys actor.Attorneys
		decisions actor.AttorneyDecisions
		taskState TaskState
	}{
		"empty": {
			taskState: TaskNotStarted,
		},
		"single with email": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			taskState: TaskCompleted,
		},
		"single with address": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Address:    place.Address{Line1: "a"},
			}},
			taskState: TaskCompleted,
		},
		"single incomplete": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
			}},
			taskState: TaskInProgress,
		},
		"multiple without decisions": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			taskState: TaskInProgress,
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
			taskState: TaskCompleted,
		},
		"multiple incomplete with decisions": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			decisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState: TaskInProgress,
		},
		"multiple with happy decisions": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			decisions: actor.AttorneyDecisions{How: actor.Jointly, HappyIfOneCannotActNoneCan: "yes"},
			taskState: TaskCompleted,
		},
		"multiple with unhappy decisions": {
			attorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			decisions: actor.AttorneyDecisions{How: actor.Jointly, HappyIfOneCannotActNoneCan: "no"},
			taskState: TaskInProgress,
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
		want                         string
		replacementAttorneys         actor.Attorneys
		attorneyDecisions            actor.AttorneyDecisions
		howReplacementsStepIn        string
		replacementAttorneyDecisions actor.AttorneyDecisions
		taskState                    TaskState
	}{
		"empty": {
			taskState: TaskNotStarted,
		},
		"do not want": {
			want:      "no",
			taskState: TaskCompleted,
		},
		"do want": {
			want:      "yes",
			taskState: TaskInProgress,
		},
		"single with email": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			taskState: TaskCompleted,
		},
		"single with address": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Address:    place.Address{Line1: "a"},
			}},
			taskState: TaskCompleted,
		},
		"single incomplete": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
			}},
			taskState: TaskInProgress,
		},
		"multiple without decisions": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			taskState: TaskInProgress,
		},
		"multiple jointly and severally": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:                    TaskCompleted,
		},
		"multiple jointly": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:                    TaskInProgress,
		},
		"multiple mixed": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:                    TaskInProgress,
		},
		"multiple jointly happily": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly, HappyIfOneCannotActNoneCan: "yes"},
			taskState:                    TaskCompleted,
		},
		"multiple mixed happily": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers, HappyIfOneCannotActNoneCan: "yes"},
			taskState:                    TaskCompleted,
		},
		"jointly and severally attorneys single": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:         TaskInProgress,
		},
		"jointly and severally attorneys single with step in": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			attorneyDecisions:     actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn: "somehow",
			taskState:             TaskCompleted,
		},
		"jointly attorneys single": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:         TaskCompleted,
		},
		"mixed attorneys single": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:         TaskCompleted,
		},

		"jointly and severally attorneys multiple": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:         TaskInProgress,
		},
		"jointly and severally attorneys multiple with step in": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:     actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn: OneCanNoLongerAct,
			taskState:             TaskCompleted,
		},
		"jointly and severally attorneys multiple with step in when none can act": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:     actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn: AllCanNoLongerAct,
			taskState:             TaskInProgress,
		},
		"jointly and severally attorneys multiple with step in when none can act jointly": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        AllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:                    TaskInProgress,
		},
		"jointly and severally attorneys multiple with step in when none can act jointly happily": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        AllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly, HappyIfOneCannotActNoneCan: "yes"},
			taskState:                    TaskCompleted,
		},
		"jointly and severally attorneys multiple with step in when none can act mixed": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        AllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:                    TaskInProgress,
		},
		"jointly and severally attorneys multiple with step in when none can act mixed happily": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        AllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers, HappyIfOneCannotActNoneCan: "yes"},
			taskState:                    TaskCompleted,
		},
		"jointly attorneys multiple without decisions": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:         TaskInProgress,
		},
		"jointly attorneys multiple jointly and severally": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			taskState:                    TaskCompleted,
		},
		"jointly attorneys multiple jointly": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:                    TaskInProgress,
		},
		"jointly attorneys multiple with jointly happily": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly, HappyIfOneCannotActNoneCan: "yes"},
			taskState:                    TaskCompleted,
		},
		"jointly attorneys multiple mixed": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers},
			taskState:                    TaskInProgress,
		},
		"jointly attorneys multiple with mixed happily": {
			want: "yes",
			replacementAttorneys: actor.Attorneys{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.Jointly},
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.JointlyForSomeSeverallyForOthers, HappyIfOneCannotActNoneCan: "yes"},
			taskState:                    TaskCompleted,
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
