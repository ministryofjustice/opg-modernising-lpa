// Package task provides types to describe the different states a tasklist task
// can be in.
package task

//go:generate enumerator -type State -linecomment -trimprefix
type State uint8

const (
	StateNotStarted State = iota // notStarted
	StateInProgress              // inProgress
	StateCompleted               // completed
)

//go:generate enumerator -type PaymentState -trimprefix
type PaymentState uint8

const (
	PaymentStateNotStarted PaymentState = iota
	PaymentStateInProgress
	PaymentStatePending
	PaymentStateApproved
	PaymentStateDenied
	PaymentStateMoreEvidenceRequired
	PaymentStateCompleted
)

//go:generate enumerator -type IdentityState -trimprefix
type IdentityState uint8

const (
	IdentityStateNotStarted IdentityState = iota
	IdentityStateInProgress
	IdentityStatePending
	IdentityStateProblem
	IdentityStateCompleted
)
