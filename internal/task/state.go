package task

type State uint8

const (
	NotStarted State = iota
	InProgress
	Completed
)

func (t State) NotStarted() bool { return t == NotStarted }
func (t State) InProgress() bool { return t == InProgress }
func (t State) Completed() bool  { return t == Completed }

func (t State) String() string {
	switch t {
	case NotStarted:
		return "notStarted"
	case InProgress:
		return "inProgress"
	case Completed:
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
