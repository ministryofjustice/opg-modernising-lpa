package actor

import "github.com/ministryofjustice/opg-modernising-lpa/internal/temporary"

type Type = temporary.ActorType

const (
	TypeNone                        = temporary.ActorTypeNone
	TypeDonor                       = temporary.ActorTypeDonor
	TypeAttorney                    = temporary.ActorTypeAttorney
	TypeReplacementAttorney         = temporary.ActorTypeReplacementAttorney
	TypeCertificateProvider         = temporary.ActorTypeCertificateProvider
	TypePersonToNotify              = temporary.ActorTypePersonToNotify
	TypeAuthorisedSignatory         = temporary.ActorTypeAuthorisedSignatory
	TypeIndependentWitness          = temporary.ActorTypeIndependentWitness
	TypeTrustCorporation            = temporary.ActorTypeTrustCorporation
	TypeReplacementTrustCorporation = temporary.ActorTypeReplacementTrustCorporation
	TypeVoucher                     = temporary.ActorTypeVoucher
)

type Types = temporary.Types

var ActorTypes = temporary.ActorTypes
