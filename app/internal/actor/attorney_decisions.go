package actor

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
	JointlyForSomeSeverallyForOthers // mixed
)

// AttorneyDecisions contains details about how an attorney or replacement attorney should act, provided by the applicant
type AttorneyDecisions struct {
	// How attorneys should make decisions
	How AttorneysAct
	// Details on how attorneys should make decisions if acting jointly for some decisions, and jointly and severally for other decisions
	Details string
	// Confirmation the applicant is happy with all attorneys being unable to act if one cannot act
	HappyIfOneCannotActNoneCan YesNo
	// Confirmation the applicant is happy with any remaining attorneys being able to act if one cannot act
	HappyIfRemainingCanContinueToAct YesNo
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

func (d AttorneyDecisions) RequiresHappiness(attorneyCount int) bool {
	return attorneyCount > 1 && (d.How == Jointly || d.How == JointlyForSomeSeverallyForOthers)
}

func (d AttorneyDecisions) IsComplete(attorneyCount int) bool {
	return !d.How.Empty() &&
		(!d.RequiresHappiness(attorneyCount) ||
			d.HappyIfOneCannotActNoneCan == Yes ||
			d.HappyIfOneCannotActNoneCan == No && !d.HappyIfRemainingCanContinueToAct.Empty())
}
