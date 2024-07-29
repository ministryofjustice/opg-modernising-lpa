package task

type State uint8

const (
	StateNotStarted State = iota
	StateInProgress
	StateCompleted
)

func (t State) NotStarted() bool { return t == StateNotStarted }
func (t State) InProgress() bool { return t == StateInProgress }
func (t State) Completed() bool  { return t == StateCompleted }

func (t State) String() string {
	switch t {
	case StateNotStarted:
		return "notStarted"
	case StateInProgress:
		return "inProgress"
	case StateCompleted:
		return "completed"
	}
	return ""
}

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
