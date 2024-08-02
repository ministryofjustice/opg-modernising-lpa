package donordata

import (
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

//go:generate enumerator -type CertificateProviderRelationshipLength -linecomment
type CertificateProviderRelationshipLength uint8

const (
	RelationshipLengthUnknown  CertificateProviderRelationshipLength = iota // unknown
	LessThanTwoYears                                                        // lt-2-years
	GreaterThanEqualToTwoYears                                              // gte-2-years
)

// CertificateProvider contains details about the certificate provider, provided by the applicant
type CertificateProvider struct {
	// UID for the actor
	UID actoruid.UID
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
	CarryOutBy lpadata.Channel
	// The certificate provider's relationship to the applicant
	Relationship lpadata.CertificateProviderRelationship
	// Amount of time Relationship has been in place if Personally
	RelationshipLength CertificateProviderRelationshipLength
}

func (c CertificateProvider) FullName() string {
	return fmt.Sprintf("%s %s", c.FirstNames, c.LastName)
}
