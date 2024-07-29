package donordata

//go:generate enumerator -type AttorneysAct -linecomment -empty
type AttorneysAct uint8

const (
	// Jointly indicates attorneys or replacement attorneys should act jointly
	Jointly AttorneysAct = iota + 1 // jointly

	// JointlyAndSeverally indicates attorneys or replacement attorneys should act
	// jointly and severally
	JointlyAndSeverally // jointly-and-severally

	// JointlyForSomeSeverallyForOthers indicates attorneys or replacement
	// attorneys should act jointly for some decisions, and jointly and severally
	// for other decisions
	JointlyForSomeSeverallyForOthers // jointly-for-some-severally-for-others
)

// AttorneyDecisions contains details about how an attorney or replacement attorney should act, provided by the applicant
type AttorneyDecisions struct {
	// How attorneys should make decisions
	How AttorneysAct
	// Details on how attorneys should make decisions if acting jointly for some decisions, and jointly and severally for other decisions
	Details string
}

func MakeAttorneyDecisions(existing AttorneyDecisions, how AttorneysAct, details string) AttorneyDecisions {
	if existing.How == how {
		if how == JointlyForSomeSeverallyForOthers {
			existing.Details = details
		}

		return existing
	}

	if how != JointlyForSomeSeverallyForOthers {
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
