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

func TestCertificateProviderIdentityConfirmed(t *testing.T) {
	testCases := map[string]struct {
		lpa      *Lpa
		expected bool
	}{
		"set": {
			lpa: &Lpa{
				CertificateProvider:                 actor.CertificateProvider{FirstNames: "a", LastName: "b"},
				CertificateProviderIdentityUserData: identity.UserData{OK: true, Provider: identity.OneLogin, FirstNames: "a", LastName: "b"},
			},
			expected: true,
		},
		"missing provider": {
			lpa:      &Lpa{CertificateProviderIdentityUserData: identity.UserData{OK: true}},
			expected: false,
		},
		"not ok": {
			lpa:      &Lpa{CertificateProviderIdentityUserData: identity.UserData{Provider: identity.OneLogin}},
			expected: false,
		},
		"no match": {
			lpa: &Lpa{
				CertificateProvider:                 actor.CertificateProvider{FirstNames: "a", LastName: "b"},
				CertificateProviderIdentityUserData: identity.UserData{Provider: identity.OneLogin},
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
			assert.Equal(t, tc.expected, tc.lpa.CertificateProviderIdentityConfirmed())
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
		CertificateProvider: actor.CertificateProvider{FirstNames: "Certificate Provider", LastName: "Actor", Address: address},
	}

	want := []AddressDetail{
		{Name: "Donor Actor", Role: "Donor", Address: address},
		{Name: "Certificate Provider Actor", Role: "Certificate Provider", Address: address},
		{Name: "Attorney One Actor", Role: "Attorney", Address: address},
		{Name: "Attorney Two Actor", Role: "Attorney", Address: address},
		{Name: "Replacement Attorney One Actor", Role: "Replacement Attorney", Address: address},
		{Name: "Replacement Attorney Two Actor", Role: "Replacement Attorney", Address: address},
	}

	got := lpa.ActorAddresses()
	assert.Equal(t, want, got)
}
