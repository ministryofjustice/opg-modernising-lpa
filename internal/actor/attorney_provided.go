package actor

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
)

// AttorneyProvidedDetails contains details about an attorney or replacement
// attorney, provided by the attorney or replacement attorney
type AttorneyProvidedDetails = attorneydata.Provided

type AttorneyTasks = attorneydata.Tasks

// TrustCorporationSignatory contains the details of a person who signed the LPA on behalf of a trust corporation
type TrustCorporationSignatory = attorneydata.TrustCorporationSignatory
