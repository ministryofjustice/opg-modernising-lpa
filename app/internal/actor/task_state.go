package actor

type TaskState int

const (
	TaskNotStarted TaskState = iota
	TaskInProgress
	TaskCompleted
)

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
