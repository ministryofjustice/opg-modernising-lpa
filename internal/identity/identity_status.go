package identity

//go:generate enumerator -type IdentityStatus --linecomment --trimprefix
type IdentityStatus uint8

const (
	IdentityStatusUnknown              IdentityStatus = iota // unknown
	IdentityStatusConfirmed                                  // confirmed
	IdentityStatusFailed                                     // failed
	IdentityStatusInsufficientEvidence                       // insufficient-evidence
)
