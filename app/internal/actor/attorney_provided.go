package actor

import (
	"time"
)

// AttorneyProvidedDetails contains details about an attorney or replacement attorney, provided by the attorney or replacement attorney
type AttorneyProvidedDetails struct {
	// The identifier of the attorney or replacement attorney being edited
	ID string
	// The identifier of the LPA the attorney or replacement attorney is named in
	LpaID string
	// Tracking when AttorneyProvidedDetails is updated
	UpdatedAt time.Time
	// Determines if the details relate to an attorney or replacement attorney
	IsReplacement bool
	// Mobile number of the attorney or replacement attorney
	Mobile string
	// Whether the name of the attorney or replacement attorney provided by the applicant is correct
	IsNameCorrect string
	// The corrected name of the attorney or replacement attorney. Only applies if IsNameCorrect = "no"
	CorrectedName string
	// Confirming the attorney or replacement attorney agrees to responsibilities and confirms the tick box is a legal signature
	Confirmed bool
	// Used to show attorney task list
	Tasks AttorneyTasks
}

type AttorneyTasks struct {
	ConfirmYourDetails TaskState
	ReadTheLpa         TaskState
	SignTheLpa         TaskState
}
