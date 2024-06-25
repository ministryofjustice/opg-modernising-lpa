package actor

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// CertificateProviderProvidedDetails contains details about the certificate provider, provided by the certificate provider
type CertificateProviderProvidedDetails struct {
	PK dynamo.LpaKeyType
	SK dynamo.CertificateProviderKeyType
	// UID of the actor
	UID actoruid.UID
	// The identifier of the LPA the certificate provider is providing a certificate for
	LpaID string
	// Tracking when CertificateProviderProvidedDetails is updated
	UpdatedAt time.Time
	// Date of birth of the certificate provider
	DateOfBirth date.Date
	// HomeAddress is the personal address of the certificate provider
	HomeAddress place.Address
	// Data returned from an identity check
	IdentityUserData identity.UserData
	// SignedAt is when the certificate provider submitted their signature
	SignedAt time.Time
	// Tasks the certificate provider will complete
	Tasks CertificateProviderTasks
	// ContactLanguagePreference is the language the certificate provider prefers to receive notifications in
	ContactLanguagePreference localize.Lang
	// Email is the email address returned from OneLogin when the certificate provider logged in
	Email string
}

func (c *CertificateProviderProvidedDetails) CertificateProviderIdentityConfirmed(firstNames, lastName string) bool {
	return c.IdentityUserData.Status.IsConfirmed() &&
		c.IdentityUserData.MatchName(firstNames, lastName) &&
		c.IdentityUserData.DateOfBirth.Equals(c.DateOfBirth)
}

type CertificateProviderTasks struct {
	ConfirmYourDetails    TaskState
	ConfirmYourIdentity   TaskState
	ReadTheLpa            TaskState
	ProvideTheCertificate TaskState
}
