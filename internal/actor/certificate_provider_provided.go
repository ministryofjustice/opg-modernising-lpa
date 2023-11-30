package actor

import (
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// CertificateProviderProvidedDetails contains details about the certificate provider, provided by the certificate provider
type CertificateProviderProvidedDetails struct {
	PK, SK string
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
	// Details of the certificate provided to the applicant
	Certificate Certificate
	// Tasks the certificate provider will complete
	Tasks CertificateProviderTasks
	// ContactLanguagePreference is the language the certificate provider prefers to receive notifications in
	ContactLanguagePreference localize.Lang
}

func (c CertificateProviderProvidedDetails) Signed(after time.Time) bool {
	return c.Certificate.Agreed.After(after)
}

func (c *CertificateProviderProvidedDetails) CertificateProviderIdentityConfirmed(firstNames, lastName string) bool {
	return c.IdentityUserData.OK &&
		c.IdentityUserData.MatchName(firstNames, lastName) &&
		c.IdentityUserData.DateOfBirth.Equals(c.DateOfBirth)
}

type Certificate struct {
	// Confirmation the certificate provider agrees to the 'provide a certificate' statement and that ticking box is a legal signature
	AgreeToStatement bool
	// Date and time the certificate provider provided the certificate
	Agreed time.Time
}

type CertificateProviderTasks struct {
	ConfirmYourDetails    TaskState
	ConfirmYourIdentity   TaskState
	ReadTheLpa            TaskState
	ProvideTheCertificate TaskState
}
