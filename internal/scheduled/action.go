package scheduled

//go:generate enumerator -type Action -trimprefix
type Action uint8

const (
	// ActionCancelDonorIdentity will check that the target donor has not signed
	// their LPA, and if so remove their identity data and notify them of the
	// change.
	ActionCancelDonorIdentity Action = iota
)
