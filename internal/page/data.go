// Package page contains the core code and business logic of Make and Register a Lasting Power of Attorney (MRLPA)
//
// Useful links:
//   - [actor.Lpa] - details about the LPA being drafted
//   - [actor.Donor] - details about the donor, provided by the applicant
//   - [actor.CertificateProvider] - details about the certificate provider, provided by the applicant
//   - [actor.CertificateProviderProvidedDetails] - details about the certificate provider, provided by the certificate provider
//   - [actor.Attorney] - details about an attorney or replacement attorney, provided by the applicant
//   - [actor.AttorneyDecisions] - details about how an attorney or replacement attorney should act, provided by the applicant
//   - [actor.AttorneyProvidedDetails] - details about an attorney or replacement attorney, provided by the attorney or replacement attorney
//   - [actor.PersonToNotify] - details about a person to notify, provided by the applicant
package page

import (
	"context"
	"strings"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
)

type SessionData struct {
	SessionID string
	LpaID     string
}

type SessionMissingError struct{}

func (s SessionMissingError) Error() string {
	return "Session data not set"
}

func SessionDataFromContext(ctx context.Context) (*SessionData, error) {
	data, ok := ctx.Value((*SessionData)(nil)).(*SessionData)

	if !ok {
		return nil, SessionMissingError{}
	}

	return data, nil
}

func ContextWithSessionData(ctx context.Context, data *SessionData) context.Context {
	return context.WithValue(ctx, (*SessionData)(nil), data)
}

func CanGoTo(donor *actor.DonorProvidedDetails, url string) bool {
	path, _, _ := strings.Cut(url, "?")
	if path == "" {
		return false
	}

	if strings.HasPrefix(path, "/lpa/") {
		_, lpaPath, _ := strings.Cut(strings.TrimPrefix(path, "/lpa/"), "/")
		return canGoToLpaPath(donor, "/"+lpaPath)
	}

	return true
}

func ChooseAttorneysState(attorneys actor.Attorneys, decisions actor.AttorneyDecisions) actor.TaskState {
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

func ChooseReplacementAttorneysState(donor *actor.DonorProvidedDetails) actor.TaskState {
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
		(donor.Attorneys.Len() == 1 || donor.AttorneyDecisions.How.IsJointly() || donor.AttorneyDecisions.How.IsJointlyForSomeSeverallyForOthers()) &&
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
