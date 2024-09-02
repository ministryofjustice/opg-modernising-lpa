package identity

//go:generate enumerator -type Status --linecomment --trimprefix
type Status uint8

const (
	StatusUnknown              Status = iota // unknown
	StatusConfirmed                          // confirmed
	StatusFailed                             // failed
	StatusInsufficientEvidence               // insufficient-evidence
	StatusExpired                            // expired
)
