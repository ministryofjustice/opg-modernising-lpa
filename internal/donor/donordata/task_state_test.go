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
			donor:    &Provided{FeeType: pay.HalfFee, Tasks: Tasks{PayForLpa: task.PaymentStateDenied}},
			expected: pay.FeeFull,
		},
		"half": {
			donor:    &Provided{FeeType: pay.HalfFee},
			expected: pay.FeeHalf,
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
			ExpectedCost: pay.AmountPence(pay.FeeHalf),
		},
		"fully paid": {
			Donor:        &Provided{FeeType: pay.HalfFee, PaymentDetails: []Payment{{Amount: pay.FeeHalf}}},
			ExpectedCost: pay.AmountPence(0),
		},
		"denied partially paid": {
			Donor:        &Provided{FeeType: pay.HalfFee, PaymentDetails: []Payment{{Amount: pay.FeeHalf}}, Tasks: Tasks{PayForLpa: task.PaymentStateDenied}},
			ExpectedCost: pay.AmountPence(pay.FeeHalf),
		},
		"denied fully paid": {
			Donor:        &Provided{FeeType: pay.HalfFee, PaymentDetails: []Payment{{Amount: pay.FeeHalf}, {Amount: pay.FeeHalf}}, Tasks: Tasks{PayForLpa: task.PaymentStateDenied}},
			ExpectedCost: pay.AmountPence(0),
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.ExpectedCost, tc.Donor.FeeAmount())
		})
	}

}
