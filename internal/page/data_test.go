package page

import (
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/stretchr/testify/assert"
)

func TestCanGoTo(t *testing.T) {
	testCases := map[string]struct {
		donor    *actor.DonorProvidedDetails
		url      string
		expected bool
	}{
		"empty path": {
			donor:    &actor.DonorProvidedDetails{},
			url:      "",
			expected: false,
		},
		"unexpected path": {
			donor:    &actor.DonorProvidedDetails{},
			url:      "/whatever",
			expected: true,
		},
		"getting help signing no certificate provider": {
			donor: &actor.DonorProvidedDetails{
				Type: actor.LpaTypePersonalWelfare,
				Tasks: actor.DonorTasks{
					YourDetails: actor.TaskCompleted,
				},
			},
			url:      Paths.GettingHelpSigning.Format("123"),
			expected: false,
		},
		"getting help signing": {
			donor: &actor.DonorProvidedDetails{
				Type: actor.LpaTypePersonalWelfare,
				Tasks: actor.DonorTasks{
					CertificateProvider: actor.TaskCompleted,
				},
			},
			url:      Paths.GettingHelpSigning.Format("123"),
			expected: true,
		},
		"check your lpa when unsure if can sign": {
			donor: &actor.DonorProvidedDetails{
				Type: actor.LpaTypePersonalWelfare,
				Tasks: actor.DonorTasks{
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
			donor: &actor.DonorProvidedDetails{
				Donor: actor.Donor{CanSign: form.Yes},
				Type:  actor.LpaTypePersonalWelfare,
				Tasks: actor.DonorTasks{
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
			donor:    &actor.DonorProvidedDetails{LpaID: "123"},
			url:      Paths.AboutPayment.Format("123"),
			expected: false,
		},
		"about payment with tasks": {
			donor: &actor.DonorProvidedDetails{
				Donor: actor.Donor{
					CanSign: form.Yes,
				},
				Type: actor.LpaTypePropertyAndAffairs,
				Tasks: actor.DonorTasks{
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
			donor:    &actor.DonorProvidedDetails{},
			url:      Paths.IdentityWithOneLogin.Format("123"),
			expected: false,
		},
		"identity with tasks": {
			donor: &actor.DonorProvidedDetails{
				Donor: actor.Donor{
					CanSign: form.Yes,
				},
				Type: actor.LpaTypePersonalWelfare,
				Tasks: actor.DonorTasks{
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
			assert.Equal(t, tc.expected, CanGoTo(tc.donor, tc.url))
		})
	}
}

func TestLpaProgress(t *testing.T) {
	lpaSignedAt := time.Now()
	uid1 := actoruid.New()
	uid2 := actoruid.New()

	testCases := map[string]struct {
		donor               *actor.DonorProvidedDetails
		certificateProvider *actor.CertificateProviderProvidedDetails
		attorneys           []*actor.AttorneyProvidedDetails
		expectedProgress    actor.Progress
	}{
		"initial state": {
			donor:               &actor.DonorProvidedDetails{},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			expectedProgress: actor.Progress{
				DonorSigned:               actor.TaskInProgress,
				CertificateProviderSigned: actor.TaskNotStarted,
				AttorneysSigned:           actor.TaskNotStarted,
				LpaSubmitted:              actor.TaskNotStarted,
				StatutoryWaitingPeriod:    actor.TaskNotStarted,
				LpaRegistered:             actor.TaskNotStarted,
			},
		},
		"lpa signed": {
			donor:               &actor.DonorProvidedDetails{SignedAt: lpaSignedAt},
			certificateProvider: &actor.CertificateProviderProvidedDetails{},
			expectedProgress: actor.Progress{
				DonorSigned:               actor.TaskCompleted,
				CertificateProviderSigned: actor.TaskInProgress,
				AttorneysSigned:           actor.TaskNotStarted,
				LpaSubmitted:              actor.TaskNotStarted,
				StatutoryWaitingPeriod:    actor.TaskNotStarted,
				LpaRegistered:             actor.TaskNotStarted,
			},
		},
		"certificate provider signed": {
			donor:               &actor.DonorProvidedDetails{SignedAt: lpaSignedAt},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			expectedProgress: actor.Progress{
				DonorSigned:               actor.TaskCompleted,
				CertificateProviderSigned: actor.TaskCompleted,
				AttorneysSigned:           actor.TaskInProgress,
				LpaSubmitted:              actor.TaskNotStarted,
				StatutoryWaitingPeriod:    actor.TaskNotStarted,
				LpaRegistered:             actor.TaskNotStarted,
			},
		},
		"attorneys signed": {
			donor: &actor.DonorProvidedDetails{
				SignedAt:  lpaSignedAt,
				Attorneys: actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}, {UID: uid2}}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: actor.Progress{
				DonorSigned:               actor.TaskCompleted,
				CertificateProviderSigned: actor.TaskCompleted,
				AttorneysSigned:           actor.TaskCompleted,
				LpaSubmitted:              actor.TaskInProgress,
				StatutoryWaitingPeriod:    actor.TaskNotStarted,
				LpaRegistered:             actor.TaskNotStarted,
			},
		},
		"submitted": {
			donor: &actor.DonorProvidedDetails{
				SignedAt:    lpaSignedAt,
				SubmittedAt: lpaSignedAt.Add(time.Hour),
				Attorneys:   actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}, {UID: uid2}}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: actor.Progress{
				DonorSigned:               actor.TaskCompleted,
				CertificateProviderSigned: actor.TaskCompleted,
				AttorneysSigned:           actor.TaskCompleted,
				LpaSubmitted:              actor.TaskCompleted,
				StatutoryWaitingPeriod:    actor.TaskInProgress,
				LpaRegistered:             actor.TaskNotStarted,
			},
		},
		"registered": {
			donor: &actor.DonorProvidedDetails{
				SignedAt:     lpaSignedAt,
				SubmittedAt:  lpaSignedAt.Add(time.Hour),
				RegisteredAt: lpaSignedAt.Add(2 * time.Hour),
				Attorneys:    actor.Attorneys{Attorneys: []actor.Attorney{{UID: uid1}, {UID: uid2}}},
			},
			certificateProvider: &actor.CertificateProviderProvidedDetails{Certificate: actor.Certificate{Agreed: lpaSignedAt.Add(time.Second)}},
			attorneys: []*actor.AttorneyProvidedDetails{
				{UID: uid1, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
				{UID: uid2, LpaSignedAt: lpaSignedAt, Confirmed: lpaSignedAt.Add(time.Minute)},
			},
			expectedProgress: actor.Progress{
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
			assert.Equal(t, tc.expectedProgress, tc.donor.Progress(tc.certificateProvider, tc.attorneys))
		})
	}
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
		howReplacementsStepIn        actor.ReplacementAttorneysStepIn
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
			taskState: actor.TaskCompleted,
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
		"multiple jointly for some severally for others": {
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
			howReplacementsStepIn: actor.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
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
		"jointly for some severally for others attorneys single": {
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
			howReplacementsStepIn: actor.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
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
			howReplacementsStepIn: actor.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
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
			howReplacementsStepIn:        actor.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			replacementAttorneyDecisions: actor.AttorneyDecisions{How: actor.Jointly},
			taskState:                    actor.TaskCompleted,
		},
		"jointly and severally attorneys multiple with step in when none can act jointly for some severally for others": {
			want: form.Yes,
			replacementAttorneys: actor.Attorneys{Attorneys: []actor.Attorney{{
				FirstNames: "a",
				Email:      "a",
			}, {
				FirstNames: "b",
				Email:      "b",
			}}},
			attorneyDecisions:            actor.AttorneyDecisions{How: actor.JointlyAndSeverally},
			howReplacementsStepIn:        actor.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
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
		"jointly attorneys multiple jointly for some severally for others": {
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
			assert.Equal(t, tc.taskState, ChooseReplacementAttorneysState(&actor.DonorProvidedDetails{
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
		donor    *actor.DonorProvidedDetails
		expected int
	}{
		"denied": {
			donor:    &actor.DonorProvidedDetails{FeeType: pay.HalfFee, Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}},
			expected: 8200,
		},
		"half": {
			donor:    &actor.DonorProvidedDetails{FeeType: pay.HalfFee},
			expected: 4100,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.donor.Cost())
		})
	}
}

func TestFeeAmount(t *testing.T) {
	testCases := map[string]struct {
		Donor        *actor.DonorProvidedDetails
		ExpectedCost int
	}{
		"not paid": {
			Donor:        &actor.DonorProvidedDetails{FeeType: pay.HalfFee},
			ExpectedCost: 4100,
		},
		"fully paid": {
			Donor:        &actor.DonorProvidedDetails{FeeType: pay.HalfFee, PaymentDetails: []actor.Payment{{Amount: 4100}}},
			ExpectedCost: 0,
		},
		"denied partially paid": {
			Donor:        &actor.DonorProvidedDetails{FeeType: pay.HalfFee, PaymentDetails: []actor.Payment{{Amount: 4100}}, Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}},
			ExpectedCost: 4100,
		},
		"denied fully paid": {
			Donor:        &actor.DonorProvidedDetails{FeeType: pay.HalfFee, PaymentDetails: []actor.Payment{{Amount: 4100}, {Amount: 4100}}, Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}},
			ExpectedCost: 0,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedCost, tc.Donor.FeeAmount())
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
			donor := &actor.DonorProvidedDetails{
				Donor:               actor.Donor{LastName: tc.donor},
				CertificateProvider: actor.CertificateProvider{LastName: tc.certificateProvider, Address: place.Address{Line1: "x"}},
			}

			for _, a := range tc.attorneys {
				donor.Attorneys.Attorneys = append(donor.Attorneys.Attorneys, actor.Attorney{LastName: a})
			}

			for _, a := range tc.replacementAttorneys {
				donor.ReplacementAttorneys.Attorneys = append(donor.ReplacementAttorneys.Attorneys, actor.Attorney{LastName: a})
			}

			assert.Equal(t, tc.expected, donor.CertificateProviderSharesDetails())
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
			donor := &actor.DonorProvidedDetails{
				Donor:               actor.Donor{Address: tc.donor},
				CertificateProvider: actor.CertificateProvider{LastName: "x", Address: tc.certificateProvider},
			}

			for _, attorney := range tc.attorneys {
				donor.Attorneys.Attorneys = append(donor.Attorneys.Attorneys, actor.Attorney{Address: attorney})
			}

			for _, attorney := range tc.replacementAttorneys {
				donor.ReplacementAttorneys.Attorneys = append(donor.ReplacementAttorneys.Attorneys, actor.Attorney{Address: attorney})
			}

			assert.Equal(t, tc.expected, donor.CertificateProviderSharesDetails())
		})
	}
}
