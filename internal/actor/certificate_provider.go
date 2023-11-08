package actor

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

//go:generate enumerator -type CertificateProviderRelationship -linecomment -empty
type CertificateProviderRelationship uint8

const (
	Personally     CertificateProviderRelationship = iota + 1 // personally
	Professionally                                            // professionally
)

//go:generate enumerator -type CertificateProviderCarryOutBy -linecomment -empty
type CertificateProviderCarryOutBy uint8

const (
	Paper  CertificateProviderCarryOutBy = iota + 1 // paper
	Online                                          // online
)

//go:generate enumerator -type CertificateProviderRelationshipLength -linecomment -empty
type CertificateProviderRelationshipLength uint8

const (
	LessThanTwoYears           CertificateProviderRelationshipLength = iota + 1 // lt-2-years
	GreaterThanEqualToTwoYears                                                  // gte-2-years
	RelationshipLengthUnknown                                                   // unknown
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
	// HasNonUKMobile indicates whether the value of Mobile is a non-UK mobile number
	HasNonUKMobile bool
	// Email of the certificate provider
	Email string
	// How the certificate provider wants to perform their role (paper or online)
	CarryOutBy CertificateProviderCarryOutBy
	// The certificate provider's relationship to the applicant
	Relationship CertificateProviderRelationship
	// Amount of time Relationship has been in place if Personally
	RelationshipLength CertificateProviderRelationshipLength
}

func (c CertificateProvider) FullName() string {
	return fmt.Sprintf("%s %s", c.FirstNames, c.LastName)
}
