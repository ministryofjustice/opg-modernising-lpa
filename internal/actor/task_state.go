package actor

type TaskState uint8

const (
	TaskNotStarted TaskState = iota
	TaskInProgress
	TaskCompleted
)

func (t TaskState) NotStarted() bool { return t == TaskNotStarted }
func (t TaskState) InProgress() bool { return t == TaskInProgress }
func (t TaskState) Completed() bool  { return t == TaskCompleted }

func (t TaskState) String() string {
	switch t {
	case TaskNotStarted:
		return "notStarted"
	case TaskInProgress:
		return "inProgress"
	case TaskCompleted:
		return "completed"
	}
	return ""
}

//go:generate enumerator -type PaymentTask -trimprefix
type PaymentTask uint8

const (
	// PaymentTaskNotStarted -> PaymentTaskInProgress
	PaymentTaskNotStarted PaymentTask = iota
	// PaymentTaskInProgress -> PaymentTaskCompleted, if full fee
	// PaymentTaskInProgress -> PaymentTaskPending, otherwise
	PaymentTaskInProgress
	// PaymentTaskPending -> PaymentTaskCompleted, if approved and paid
	// PaymentTaskPending -> PaymentTaskApproved, if approved and payment required
	// PaymentTaskPending -> PaymentTaskDenied, if denied
	// PaymentTaskPending -> PaymentTaskMoreEvidenceRequired, if more evidence required
	PaymentTaskPending
	// PaymentTaskApproved -> PaymentTaskCompleted, when missing payment received
	PaymentTaskApproved
	// PaymentTaskDenied -> PaymentTaskCompleted, when missing payment received
	PaymentTaskDenied
	// PaymentTaskMoreEvidenceRequired -> PaymentTaskPending
	PaymentTaskMoreEvidenceRequired
	// (end state)
	PaymentTaskCompleted
)
