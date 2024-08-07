package donordata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

func ChooseAttorneysState(attorneys Attorneys, decisions AttorneyDecisions) task.State {
	if attorneys.Len() == 0 {
		return task.StateNotStarted
	}

	if !attorneys.Complete() {
		return task.StateInProgress
	}

	if attorneys.Len() > 1 && !decisions.IsComplete() {
		return task.StateInProgress
	}

	return task.StateCompleted
}

func ChooseReplacementAttorneysState(donor *Provided) task.State {
	if donor.WantReplacementAttorneys == form.No {
		return task.StateCompleted
	}

	if donor.ReplacementAttorneys.Len() == 0 {
		if donor.WantReplacementAttorneys.IsUnknown() {
			return task.StateNotStarted
		}

		return task.StateInProgress
	}

	if !donor.ReplacementAttorneys.Complete() {
		return task.StateInProgress
	}

	if donor.ReplacementAttorneys.Len() > 1 &&
		(donor.Attorneys.Len() == 1 || donor.AttorneyDecisions.How.IsJointly()) &&
		!donor.ReplacementAttorneyDecisions.IsComplete() {
		return task.StateInProgress
	}

	if donor.AttorneyDecisions.How.IsJointlyAndSeverally() {
		if donor.HowShouldReplacementAttorneysStepIn.Empty() {
			return task.StateInProgress
		}

		if donor.ReplacementAttorneys.Len() > 1 &&
			donor.HowShouldReplacementAttorneysStepIn.IsWhenAllCanNoLongerAct() &&
			!donor.ReplacementAttorneyDecisions.IsComplete() {
			return task.StateInProgress
		}
	}

	return task.StateCompleted
}
