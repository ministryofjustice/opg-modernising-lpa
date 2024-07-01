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
	PaymentTaskNotStarted PaymentTask = iota
	PaymentTaskInProgress
	PaymentTaskPending
	PaymentTaskApproved
	PaymentTaskDenied
	PaymentTaskMoreEvidenceRequired
	PaymentTaskCompleted
)

//go:generate enumerator -type IdentityTask -trimprefix
type IdentityTask uint8

const (
	IdentityTaskNotStarted IdentityTask = iota
	IdentityTaskInProgress
	IdentityTaskProblem
	IdentityTaskCompleted
)
