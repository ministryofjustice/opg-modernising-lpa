package actor

type Type int

const (
	TypeNone Type = iota
	TypeDonor
	TypeAttorney
	TypeReplacementAttorney
	TypeCertificateProvider
	TypePersonToNotify
	TypeAuthorisedSignatory
	TypeIndependentWitness
	TypeSupporter
)

func (t Type) String() string {
	switch t {
	case TypeDonor:
		return "donor"
	case TypeAttorney:
		return "attorney"
	case TypeReplacementAttorney:
		return "replacementAttorney"
	case TypeCertificateProvider:
		return "certificateProvider"
	case TypePersonToNotify:
		return "personToNotify"
	case TypeAuthorisedSignatory:
		return "signatory"
	case TypeIndependentWitness:
		return "independentWitness"
	case TypeSupporter:
		return "supporter"
	default:
		return ""
	}
}

type Types struct {
	None                Type
	Donor               Type
	Attorney            Type
	ReplacementAttorney Type
	CertificateProvider Type
	PersonToNotify      Type
	AuthorisedSignatory Type
	IndependentWitness  Type
	Supporter           Type
}

var ActorTypes = Types{
	None:                TypeNone,
	Donor:               TypeDonor,
	Attorney:            TypeAttorney,
	ReplacementAttorney: TypeReplacementAttorney,
	CertificateProvider: TypeCertificateProvider,
	PersonToNotify:      TypePersonToNotify,
	AuthorisedSignatory: TypeAuthorisedSignatory,
	IndependentWitness:  TypeIndependentWitness,
	Supporter:           TypeSupporter,
}
