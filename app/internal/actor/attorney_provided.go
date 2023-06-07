package actor

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type AttorneyProvidedDetails struct {
	ID            string
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
