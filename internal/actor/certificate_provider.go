package actor

import (
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
)

type CertificateProviderRelationship = donordata.CertificateProviderRelationship

const (
	Personally     = donordata.Personally
	Professionally = donordata.Professionally
)

type CertificateProviderRelationshipLength = donordata.CertificateProviderRelationshipLength

const (
	RelationshipLengthUnknown  = donordata.RelationshipLengthUnknown
	LessThanTwoYears           = donordata.LessThanTwoYears
	GreaterThanEqualToTwoYears = donordata.GreaterThanEqualToTwoYears
)

type CertificateProvider = donordata.CertificateProvider
