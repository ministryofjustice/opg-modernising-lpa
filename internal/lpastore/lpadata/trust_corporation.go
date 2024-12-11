package lpadata

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type TrustCorporation struct {
	UID             actoruid.UID    `json:"uid"`
	Name            string          `json:"name"`
	CompanyNumber   string          `json:"companyNumber"`
	Email           string          `json:"email,omitempty"`
	Address         place.Address   `json:"address"`
	Channel         Channel         `json:"channel"`
	Status          AttorneyStatus  `json:"status"`
	AppointmentType AppointmentType `json:"appointmentType"`

	// Mobile may be given by the trust corporation, or a paper donor
	Mobile string `json:"mobile,omitempty"`

	// These are given by the trust corporation, so will not be present on
	// creation.
	ContactLanguagePreference localize.Lang               `json:"contactLanguagePreference,omitempty"`
	Signatories               []TrustCorporationSignatory `json:"signatories,omitempty"`

	Removed bool `json:"-"`
}
