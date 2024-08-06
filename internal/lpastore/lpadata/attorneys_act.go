package lpadata

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
