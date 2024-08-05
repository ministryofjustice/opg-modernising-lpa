// Package attorneydata provides types that describe the data entered by an
// attorney or trust corporation.
package attorneydata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

// Provided contains details about an attorney or replacement
// attorney, provided by the attorney or replacement attorney
type Provided struct {
	PK dynamo.LpaKeyType
	SK dynamo.AttorneyKeyType
	// The identifier of the attorney or replacement attorney being edited
	UID actoruid.UID
	// The identifier of the LPA the attorney or replacement attorney is named in
	LpaID string
	// Tracking when AttorneyProvidedDetails is updated
	UpdatedAt time.Time
	// IsReplacement is true when the details relate to an attorney appointed as a
	// replacement
	IsReplacement bool
	// IsTrustCorporation is true when the details relate to a trust corporation
	IsTrustCorporation bool
	// Mobile number of the attorney or replacement attorney
	Mobile string
	// SignedAt is when the attorney or replacement attorney submitted their
	// signature
	SignedAt time.Time
	// WouldLikeSecondSignatory captures whether two signatories will be used for
	// a trust corporation
	WouldLikeSecondSignatory form.YesNo
	// AuthorisedSignatories captures the details of the person who signed on
	// behalf of a trust corporation, if one is acting as an attorney
	AuthorisedSignatories [2]TrustCorporationSignatory
	// Used to show attorney task list
	Tasks Tasks
	// ContactLanguagePreference is the language the attorney or replacement
	// attorney prefers to receive notifications in
	ContactLanguagePreference localize.Lang
	// Email is the email address returned from OneLogin when the attorney logged in
	Email string
}

// Signed checks whether the attorney has confirmed and if that confirmation is
// still valid by checking that it was made for the donor's current signature.
func (d Provided) Signed() bool {
	if d.IsTrustCorporation {
		switch d.WouldLikeSecondSignatory {
		case form.Yes:
			return !d.AuthorisedSignatories[0].SignedAt.IsZero() &&
				!d.AuthorisedSignatories[1].SignedAt.IsZero()
		case form.No:
			return !d.AuthorisedSignatories[0].SignedAt.IsZero()
		default:
			return false
		}
	}

	return !d.SignedAt.IsZero()
}

type Tasks struct {
	ConfirmYourDetails task.State
	ReadTheLpa         task.State
	SignTheLpa         task.State
	SignTheLpaSecond   task.State
}

// TrustCorporationSignatory contains the details of a person who signed the LPA on behalf of a trust corporation
type TrustCorporationSignatory struct {
	FirstNames        string
	LastName          string
	ProfessionalTitle string
	SignedAt          time.Time
}
