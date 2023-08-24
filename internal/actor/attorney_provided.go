package actor

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
)

// AttorneyProvidedDetails contains details about an attorney or replacement attorney, provided by the attorney or replacement attorney
type AttorneyProvidedDetails struct {
	PK, SK string
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
	// Confirming the attorney or replacement attorney agrees to responsibilities and confirms the tick box is a legal signature
	Confirmed time.Time
	// WouldLikeSecondSignatory captures whether two signatories will be used for a trust corporation
	WouldLikeSecondSignatory form.YesNo
	// AuthorisedSignatories captures the details of the person who signed on behalf of a trust corporation, if one is acting as an attorney
	AuthorisedSignatories [2]AuthorisedSignatory
	// Used to show attorney task list
	Tasks AttorneyTasks
}

type AttorneyTasks struct {
	ConfirmYourDetails TaskState
	ReadTheLpa         TaskState
	SignTheLpa         TaskState
	SignTheLpaSecond   TaskState
}

type AuthorisedSignatory struct {
	FirstNames        string
	LastName          string
	ProfessionalTitle string
	Confirmed         time.Time
}
