package actor

const (
	Jointly                          = "jointly"
	JointlyAndSeverally              = "jointly-and-severally"
	JointlyForSomeSeverallyForOthers = "mixed"
)

type AttorneyDecisions struct {
	// How attorneys should make decisions
	How string
	// Details on how attorneys should make decisions if acting jointly for some decisions, and jointly and severally for other decisions
	Details string
	// Confirmation the applicant is happy with all attorneys being unable to act if one cannot act
	HappyIfOneCannotActNoneCan string
	// Confirmation the applicant is happy with any remaining attorneys being able to act if one cannot act
	HappyIfRemainingCanContinueToAct string
}

func MakeAttorneyDecisions(existing AttorneyDecisions, how, details string) AttorneyDecisions {
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
	return d.How != "" &&
		(!d.RequiresHappiness(attorneyCount) ||
			d.HappyIfOneCannotActNoneCan == "yes" ||
			d.HappyIfOneCannotActNoneCan == "no" && d.HappyIfRemainingCanContinueToAct != "")
}
