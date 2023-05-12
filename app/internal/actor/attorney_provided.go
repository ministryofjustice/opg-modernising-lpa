package actor

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type AttorneyProvidedDetails struct {
	LpaID         string
	UpdatedAt     time.Time
	IsReplacement bool

	DateOfBirth   date.Date
	Mobile        string
	Address       place.Address
	IsNameCorrect string
	CorrectedName string
	Confirmed     bool
	Tasks         AttorneyTasks
}

type AttorneyTasks struct {
	ConfirmYourDetails TaskState
	ReadTheLpa         TaskState
	SignTheLpa         TaskState
}

// TODO: move this somewhere, just copied now to prevent an import cycle
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
