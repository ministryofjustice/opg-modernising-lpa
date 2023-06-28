package actor

import "fmt"

type AttorneysAct string

const (
	// Jointly indicates attorneys or replacement attorneys should act jointly
	Jointly = AttorneysAct("jointly")
	// JointlyAndSeverally indicates attorneys or replacement attorneys should act
	// jointly and severally
	JointlyAndSeverally = AttorneysAct("jointly-and-severally")
	// JointlyForSomeSeverallyForOthers indicates attorneys or replacement
	// attorneys should act jointly for some decisions, and jointly and severally
	// for other decisions
	JointlyForSomeSeverallyForOthers = AttorneysAct("mixed")
)

func ParseAttorneysAct(s string) (AttorneysAct, error) {
	switch s {
	case "jointly":
		return Jointly, nil
	case "jointly-and-severally":
		return JointlyAndSeverally, nil
	case "mixed":
		return JointlyForSomeSeverallyForOthers, nil
	default:
		return AttorneysAct(""), fmt.Errorf("invalid AttorneysAct '%s'", s)
	}
}

func (e AttorneysAct) IsJointly() bool {
	return e == Jointly
}

func (e AttorneysAct) IsJointlyAndSeverally() bool {
	return e == JointlyAndSeverally
}

func (e AttorneysAct) IsJointlyForSomeSeverallyForOthers() bool {
	return e == JointlyForSomeSeverallyForOthers
}

func (e AttorneysAct) String() string {
	return string(e)
}

type AttorneysActOptions struct {
	Jointly                          AttorneysAct
	JointlyAndSeverally              AttorneysAct
	JointlyForSomeSeverallyForOthers AttorneysAct
}

var AttorneysActValues = AttorneysActOptions{
	Jointly:                          Jointly,
	JointlyAndSeverally:              JointlyAndSeverally,
	JointlyForSomeSeverallyForOthers: JointlyForSomeSeverallyForOthers,
}

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
	return d.How.String() != "" &&
		(!d.RequiresHappiness(attorneyCount) ||
			d.HappyIfOneCannotActNoneCan == Yes ||
			d.HappyIfOneCannotActNoneCan == No && (d.HappyIfRemainingCanContinueToAct == Yes || d.HappyIfRemainingCanContinueToAct == No))
}
