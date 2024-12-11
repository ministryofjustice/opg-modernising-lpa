package lpadata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type Attorney struct {
	UID             actoruid.UID    `json:"uid"`
	FirstNames      string          `json:"firstNames"`
	LastName        string          `json:"lastName"`
	DateOfBirth     date.Date       `json:"dateOfBirth"`
	Email           string          `json:"email,omitempty"`
	Address         place.Address   `json:"address"`
	Channel         Channel         `json:"channel"`
	Status          AttorneyStatus  `json:"status"`
	AppointmentType AppointmentType `json:"appointmentType"`

	// Mobile may be given by the attorney, or a paper donor
	Mobile string `json:"mobile,omitempty"`

	// These are given by the attorney, so will not be present on creation.
	SignedAt                  *time.Time    `json:"signedAt,omitempty"`
	ContactLanguagePreference localize.Lang `json:"contactLanguagePreference,omitempty"`

	Removed bool `json:"-"`
}

func (a Attorney) FullName() string {
	return a.FirstNames + " " + a.LastName
}
