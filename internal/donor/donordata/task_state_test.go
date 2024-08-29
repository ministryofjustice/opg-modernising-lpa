package donordata

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
)

var testAddress = place.Address{Line1: "1"}

func TestChooseAttorneysState(t *testing.T) {
	testcases := map[string]struct {
		attorneys Attorneys
		decisions AttorneyDecisions
		taskState task.State
	}{
		"empty": {
			taskState: task.StateNotStarted,
		},
		"trust corporation": {
			attorneys: Attorneys{TrustCorporation: TrustCorporation{
				Name:    "a",
				Address: place.Address{Line1: "a"},
			}},
			taskState: task.StateCompleted,
		},
		"trust corporation incomplete": {
			attorneys: Attorneys{TrustCorporation: TrustCorporation{
				Name: "a",
			}},
			taskState: task.StateInProgress,
		},
		"single with email": {
			attorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Email:      "a",
			}}},
			taskState: task.StateInProgress,
		},
		"single with address": {
			attorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    place.Address{Line1: "a"},
			}}},
			taskState: task.StateCompleted,
		},
		"single incomplete": {
			attorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
			}}},
			taskState: task.StateInProgress,
		},
		"multiple without decisions": {
			attorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			taskState: task.StateInProgress,
		},
		"multiple with decisions": {
			attorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			decisions: AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			taskState: task.StateCompleted,
		},
		"multiple incomplete with decisions": {
			attorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			decisions: AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			taskState: task.StateInProgress,
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
		replacementAttorneys         Attorneys
		attorneyDecisions            AttorneyDecisions
		howReplacementsStepIn        lpadata.ReplacementAttorneysStepIn
		replacementAttorneyDecisions AttorneyDecisions
		taskState                    task.State
	}{
		"empty": {
			taskState: task.StateNotStarted,
		},
		"do not want": {
			want:      form.No,
			taskState: task.StateCompleted,
		},
		"do want": {
			want:      form.Yes,
			taskState: task.StateInProgress,
		},
		"single with email": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Email:      "a",
			}}},
			taskState: task.StateInProgress,
		},
		"single with address": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    place.Address{Line1: "a"},
			}}},
			taskState: task.StateCompleted,
		},
		"single incomplete": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
			}}},
			taskState: task.StateInProgress,
		},
		"multiple without decisions": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			taskState: task.StateCompleted,
		},
		"multiple jointly and severally": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			replacementAttorneyDecisions: AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			taskState:                    task.StateCompleted,
		},
		"multiple jointly": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			replacementAttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
			taskState:                    task.StateCompleted,
		},
		"multiple jointly for some severally for others": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			replacementAttorneyDecisions: AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers},
			taskState:                    task.StateCompleted,
		},
		"jointly and severally attorneys single": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}}},
			attorneyDecisions: AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			taskState:         task.StateInProgress,
		},
		"jointly and severally attorneys single with step in": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}}},
			attorneyDecisions:     AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			howReplacementsStepIn: lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			taskState:             task.StateCompleted,
		},
		"jointly attorneys single": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}}},
			attorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
			taskState:         task.StateCompleted,
		},
		"jointly for some severally for others attorneys single": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}}},
			attorneyDecisions: AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers},
			taskState:         task.StateCompleted,
		},
		"jointly for some severally for others attorneys multiple": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			attorneyDecisions: AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers},
			taskState:         task.StateCompleted,
		},
		"jointly and severally attorneys multiple": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			attorneyDecisions: AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			taskState:         task.StateInProgress,
		},
		"jointly and severally attorneys multiple with step in": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			attorneyDecisions:     AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			howReplacementsStepIn: lpadata.ReplacementAttorneysStepInWhenOneCanNoLongerAct,
			taskState:             task.StateCompleted,
		},
		"jointly and severally attorneys multiple with step in when none can act": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			attorneyDecisions:     AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			howReplacementsStepIn: lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			taskState:             task.StateInProgress,
		},
		"jointly and severally attorneys multiple with step in when none can act jointly": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			attorneyDecisions:            AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			howReplacementsStepIn:        lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			replacementAttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
			taskState:                    task.StateCompleted,
		},
		"jointly and severally attorneys multiple with step in when none can act jointly for some severally for others": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			attorneyDecisions:            AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			howReplacementsStepIn:        lpadata.ReplacementAttorneysStepInWhenAllCanNoLongerAct,
			replacementAttorneyDecisions: AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers},
			taskState:                    task.StateCompleted,
		},
		"jointly attorneys multiple without decisions": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			attorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
			taskState:         task.StateInProgress,
		},
		"jointly attorneys multiple jointly and severally": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			attorneyDecisions:            AttorneyDecisions{How: lpadata.Jointly},
			replacementAttorneyDecisions: AttorneyDecisions{How: lpadata.JointlyAndSeverally},
			taskState:                    task.StateCompleted,
		},
		"jointly attorneys multiple with jointly": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			attorneyDecisions:            AttorneyDecisions{How: lpadata.Jointly},
			replacementAttorneyDecisions: AttorneyDecisions{How: lpadata.Jointly},
			taskState:                    task.StateCompleted,
		},
		"jointly attorneys multiple jointly for some severally for others": {
			want: form.Yes,
			replacementAttorneys: Attorneys{Attorneys: []Attorney{{
				FirstNames: "a",
				Address:    testAddress,
			}, {
				FirstNames: "b",
				Address:    testAddress,
			}}},
			attorneyDecisions:            AttorneyDecisions{How: lpadata.Jointly},
			replacementAttorneyDecisions: AttorneyDecisions{How: lpadata.JointlyForSomeSeverallyForOthers},
			taskState:                    task.StateCompleted,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.taskState, ChooseReplacementAttorneysState(&Provided{
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
		donor    *Provided
		expected int
	}{
		"denied": {
			donor:    &Provided{FeeType: pay.HalfFee, Tasks: task.DonorTasks{PayForLpa: task.PaymentStateDenied}},
			expected: 8200,
		},
		"half": {
			donor:    &Provided{FeeType: pay.HalfFee},
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
		Donor        *Provided
		ExpectedCost pay.AmountPence
	}{
		"not paid": {
			Donor:        &Provided{FeeType: pay.HalfFee},
			ExpectedCost: pay.AmountPence(4100),
		},
		"fully paid": {
			Donor:        &Provided{FeeType: pay.HalfFee, PaymentDetails: []Payment{{Amount: 4100}}},
			ExpectedCost: pay.AmountPence(0),
		},
		"denied partially paid": {
			Donor:        &Provided{FeeType: pay.HalfFee, PaymentDetails: []Payment{{Amount: 4100}}, Tasks: task.DonorTasks{PayForLpa: task.PaymentStateDenied}},
			ExpectedCost: pay.AmountPence(4100),
		},
		"denied fully paid": {
			Donor:        &Provided{FeeType: pay.HalfFee, PaymentDetails: []Payment{{Amount: 4100}, {Amount: 4100}}, Tasks: task.DonorTasks{PayForLpa: task.PaymentStateDenied}},
			ExpectedCost: pay.AmountPence(0),
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
			donor := &Provided{
				Donor:               Donor{LastName: tc.donor},
				CertificateProvider: CertificateProvider{LastName: tc.certificateProvider, Address: place.Address{Line1: "x"}},
			}

			for _, a := range tc.attorneys {
				donor.Attorneys.Attorneys = append(donor.Attorneys.Attorneys, Attorney{LastName: a})
			}

			for _, a := range tc.replacementAttorneys {
				donor.ReplacementAttorneys.Attorneys = append(donor.ReplacementAttorneys.Attorneys, Attorney{LastName: a})
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
			donor := &Provided{
				Donor:               Donor{Address: tc.donor},
				CertificateProvider: CertificateProvider{LastName: "x", Address: tc.certificateProvider},
			}

			for _, attorney := range tc.attorneys {
				donor.Attorneys.Attorneys = append(donor.Attorneys.Attorneys, Attorney{Address: attorney})
			}

			for _, attorney := range tc.replacementAttorneys {
				donor.ReplacementAttorneys.Attorneys = append(donor.ReplacementAttorneys.Attorneys, Attorney{Address: attorney})
			}

			assert.Equal(t, tc.expected, donor.CertificateProviderSharesDetails())
		})
	}
}
