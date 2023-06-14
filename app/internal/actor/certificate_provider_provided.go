package actor

import (
	"fmt"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

// CertificateProviderProvidedDetails are details provided by the certificate provider
type CertificateProviderProvidedDetails struct {
	// The identifier of the LPA the certificate provider is providing a certificate for
	LpaID string
	// Tracking when CertificateProviderProvidedDetails is updated
	UpdatedAt time.Time
	// First names of the certificate provider
	FirstNames string
	// Last name of the certificate provider
	LastName string
	// Email of the certificate provider
	Email string
	// Address of the certificate provider
	Address place.Address
	// Mobile number of the certificate provider
	Mobile string
	// Date of birth of the certificate provider
	DateOfBirth date.Date
	// The full name provided by the certificate provider. Only requested if the certificate provider indicates their name as provided by the applicant is incorrect
	DeclaredFullName string
	// The method by which the certificate provider will complete identity checks
	IdentityOption identity.Option
	// Data returned from an identity check service
	IdentityUserData identity.UserData
	// Details of the certificate provided to the applicant
	Certificate Certificate
}

type Certificate struct {
	// Confirmation the certificate provider agrees to the 'provide a certificate' statement and that ticking box is a legal signature
	AgreeToStatement bool
	// Date and time the certificate provider provided the certificate
	Agreed time.Time
}

func (c CertificateProviderProvidedDetails) FullName() string {
	return fmt.Sprintf("%s %s", c.FirstNames, c.LastName)
}

func (c *CertificateProviderProvidedDetails) CertificateProviderIdentityConfirmed() bool {
	return c.IdentityUserData.OK && c.IdentityUserData.Provider != identity.UnknownOption &&
		c.IdentityUserData.MatchName(c.FirstNames, c.LastName) &&
		c.IdentityUserData.DateOfBirth.Equals(c.DateOfBirth)
}
