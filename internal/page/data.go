// Package page contains the core code and business logic of Make and Register a Lasting Power of Attorney (MRLPA)
//
// Useful links:
//   - [actor.Lpa] - details about the LPA being drafted
//   - [donordata.Donor] - details about the donor, provided by the applicant
//   - [donordata.CertificateProvider] - details about the certificate provider, provided by the applicant
//   - [certificateproviderdata.Provided] - details about the certificate provider, provided by the certificate provider
//   - [donordata.Attorney] - details about an attorney or replacement attorney, provided by the applicant
//   - [donordata.AttorneyDecisions] - details about how an attorney or replacement attorney should act, provided by the applicant
//   - [attorneydata.Provided] - details about an attorney or replacement attorney, provided by the attorney or replacement attorney
//   - [donordata.PersonToNotify] - details about a person to notify, provided by the applicant
package page

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

var SessionDataFromContext = appcontext.SessionDataFromContext
var ContextWithSessionData = appcontext.ContextWithSessionData

func ChooseAttorneysState(attorneys donordata.Attorneys, decisions donordata.AttorneyDecisions) task.State {
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

func ChooseReplacementAttorneysState(donor *donordata.Provided) task.State {
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
