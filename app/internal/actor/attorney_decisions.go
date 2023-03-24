package actor

const (
	Jointly                          = "jointly"
	JointlyAndSeverally              = "jointly-and-severally"
	JointlyForSomeSeverallyForOthers = "mixed"
)

type AttorneyDecisions struct {
	How                              string
	Details                          string
	HappyIfOneCannotActNoneCan       string
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
