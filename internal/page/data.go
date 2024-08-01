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
//   - [actor.PersonToNotify] - details about a person to notify, provided by the applicant
package page

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
)

var SessionDataFromContext = appcontext.SessionDataFromContext
var ContextWithSessionData = appcontext.ContextWithSessionData

func ChooseAttorneysState(attorneys donordata.Attorneys, decisions donordata.AttorneyDecisions) actor.TaskState {
	if attorneys.Len() == 0 {
		return actor.TaskNotStarted
	}

	if !attorneys.Complete() {
		return actor.TaskInProgress
	}

	if attorneys.Len() > 1 && !decisions.IsComplete() {
		return actor.TaskInProgress
	}

	return actor.TaskCompleted
}

func ChooseReplacementAttorneysState(donor *donordata.DonorProvidedDetails) actor.TaskState {
	if donor.WantReplacementAttorneys == form.No {
		return actor.TaskCompleted
	}

	if donor.ReplacementAttorneys.Len() == 0 {
		if donor.WantReplacementAttorneys.IsUnknown() {
			return actor.TaskNotStarted
		}

		return actor.TaskInProgress
	}

	if !donor.ReplacementAttorneys.Complete() {
		return actor.TaskInProgress
	}

	if donor.ReplacementAttorneys.Len() > 1 &&
		(donor.Attorneys.Len() == 1 || donor.AttorneyDecisions.How.IsJointly()) &&
		!donor.ReplacementAttorneyDecisions.IsComplete() {
		return actor.TaskInProgress
	}

	if donor.AttorneyDecisions.How.IsJointlyAndSeverally() {
		if donor.HowShouldReplacementAttorneysStepIn.Empty() {
			return actor.TaskInProgress
		}

		if donor.ReplacementAttorneys.Len() > 1 &&
			donor.HowShouldReplacementAttorneysStepIn.IsWhenAllCanNoLongerAct() &&
			!donor.ReplacementAttorneyDecisions.IsComplete() {
			return actor.TaskInProgress
		}
	}

	return actor.TaskCompleted
}
