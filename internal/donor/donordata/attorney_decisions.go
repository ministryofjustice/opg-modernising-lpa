package donordata

import "github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"

// AttorneyDecisions contains details about how an attorney or replacement attorney should act, provided by the applicant
type AttorneyDecisions struct {
	// How attorneys should make decisions
	How lpadata.AttorneysAct
	// Details on how attorneys should make decisions if acting jointly for some decisions, and jointly and severally for other decisions
	Details string
}

func MakeAttorneyDecisions(existing AttorneyDecisions, how lpadata.AttorneysAct, details string) AttorneyDecisions {
	if existing.How == how {
		if how == lpadata.JointlyForSomeSeverallyForOthers {
			existing.Details = details
		}

		return existing
	}

	if how != lpadata.JointlyForSomeSeverallyForOthers {
		return AttorneyDecisions{How: how}
	}

	return AttorneyDecisions{
		How:     how,
		Details: details,
	}
}

func (d AttorneyDecisions) IsComplete() bool {
	return !d.How.Empty()
}
