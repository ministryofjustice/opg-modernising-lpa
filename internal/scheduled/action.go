package scheduled

//go:generate enumerator -type Action -trimprefix
type Action uint8

const (
	// ActionExpireDonorIdentity will check that the target donor has not signed
	// their LPA, and if so remove their identity data and notify them of the
	// change.
	ActionExpireDonorIdentity Action = iota + 1
)
