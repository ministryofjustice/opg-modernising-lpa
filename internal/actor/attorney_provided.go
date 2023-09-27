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
	// IsReplacement is true when the details relate to an attorney appointed as a replacement
	IsReplacement bool
	// IsTrustCorporation is true when the details relate to a trust corporation
	IsTrustCorporation bool
	// Mobile number of the attorney or replacement attorney
	Mobile string
	// Confirming the attorney or replacement attorney agrees to responsibilities and confirms the tick box is a legal signature
	Confirmed time.Time
	// WouldLikeSecondSignatory captures whether two signatories will be used for a trust corporation
	WouldLikeSecondSignatory form.YesNo
	// AuthorisedSignatories captures the details of the person who signed on behalf of a trust corporation, if one is acting as an attorney
	AuthorisedSignatories [2]TrustCorporationSignatory
	// Used to show attorney task list
	Tasks AttorneyTasks
}

func (d AttorneyProvidedDetails) Signed() bool {
	if d.IsTrustCorporation {
		switch d.WouldLikeSecondSignatory {
		case form.Yes:
			return !d.AuthorisedSignatories[0].Confirmed.IsZero() && !d.AuthorisedSignatories[1].Confirmed.IsZero()
		case form.No:
			return !d.AuthorisedSignatories[0].Confirmed.IsZero()
		default:
			return false
		}
	}

	return !d.Confirmed.IsZero()
}

type AttorneyTasks struct {
	ConfirmYourDetails TaskState
	ReadTheLpa         TaskState
	SignTheLpa         TaskState
	SignTheLpaSecond   TaskState
}

// TrustCorporationSignatory contains the details of a person who signed the LPA on behalf of a trust corporation
type TrustCorporationSignatory struct {
	FirstNames        string
	LastName          string
	ProfessionalTitle string
	Confirmed         time.Time
}
