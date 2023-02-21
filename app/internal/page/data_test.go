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
		"yoti": {
			lpa:      &Lpa{YotiUserData: identity.UserData{OK: true}},
			expected: true,
		},
		"one login": {
			lpa:      &Lpa{OneLoginUserData: identity.UserData{OK: true}},
			expected: true,
		},
		"none": {
			lpa:      &Lpa{},
			expected: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.lpa.IdentityConfirmed())
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
		"Combined": {
			LpaType:           LpaTypeCombined,
			ExpectedLegalTerm: "combinedLegalTerm",
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

func TestWitnessCodeExpired(t *testing.T) {
	now := time.Now()

	testCases := map[string]struct {
		Duration string
		Expected bool
	}{
		"now": {
			Duration: "0s",
			Expected: false,
		},
		"29m59s ago": {
			Duration: "-29m59s",
			Expected: false,
		},
		"30m ago": {
			Duration: "-30m",
			Expected: true,
		},
		"30m01s ago": {
			Duration: "-30m01s",
			Expected: true,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			duration, _ := time.ParseDuration(tc.Duration)

			lpa := Lpa{WitnessCode: WitnessCode{
				Created: now.Add(duration),
			}}

			assert.Equal(t, tc.Expected, lpa.WitnessCode.HasExpired())
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
		expectedProgress Progress
	}{
		"initial state": {
			lpa: &Lpa{},
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
			lpa: &Lpa{Submitted: time.Now(), Certificate: Certificate{Agreed: time.Now()}},
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
			assert.Equal(t, tc.expectedProgress, tc.lpa.Progress())
		})
	}

}
