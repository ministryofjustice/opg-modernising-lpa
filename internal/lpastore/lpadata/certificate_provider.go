package lpadata

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type CertificateProvider struct {
	UID        actoruid.UID  `json:"uid"`
	FirstNames string        `json:"firstNames"`
	LastName   string        `json:"lastName"`
	Email      string        `json:"email,omitempty"`
	Phone      string        `json:"phone,omitempty"`
	Address    place.Address `json:"address"`
	Channel    Channel       `json:"channel"`

	// These are given by the certificate provider, so will not be present when
	// creating.
	SignedAt                  *time.Time     `json:"signedAt,omitempty"`
	ContactLanguagePreference localize.Lang  `json:"contactLanguagePreference,omitempty"`
	IdentityCheck             *IdentityCheck `json:"identityCheck,omitempty"`

	// Relationship is not stored in the lpa-store so is defaulted to
	// Professional. We require it to determine whether to show the home address
	// page to a certificate provider.
	Relationship CertificateProviderRelationship `json:"-"`
}

func (c CertificateProvider) FullName() string {
	return c.FirstNames + " " + c.LastName
}
