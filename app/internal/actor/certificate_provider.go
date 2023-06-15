package actor

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/place"
)

// CertificateProvider contains details about the certificate provider, provided by the applicant
type CertificateProvider struct {
	// First names of the certificate provider
	FirstNames string
	// Last name of the certificate provider
	LastName string
	// Address of the certificate provider
	Address place.Address
	// Mobile number of the certificate provider, used to send witness codes
	Mobile string
	// Email of the certificate provider
	Email string
	// How the certificate provider wants to perform their role (paper or online)
	CarryOutBy string
	// The certificate provider's relationship to the applicant
	Relationship string
	// If CertificateProvider.Relationship="other", what that means
	RelationshipDescription string
	// Amount of time Relationship has been in place (does not apply to health professionals or legal professionals)
	RelationshipLength string
}

func (c CertificateProvider) FullName() string {
	return fmt.Sprintf("%s %s", c.FirstNames, c.LastName)
}
