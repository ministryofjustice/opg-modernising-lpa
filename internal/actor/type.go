package actor

type Type uint8

const (
	TypeNone Type = iota
	TypeDonor
	TypeAttorney
	TypeReplacementAttorney
	TypeCertificateProvider
	TypePersonToNotify
	TypeAuthorisedSignatory
	TypeIndependentWitness
	TypeTrustCorporation
	TypeReplacementTrustCorporation
	TypeVoucher
)

func (t Type) IsVoucher() bool {
	return t == TypeVoucher
}

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
	case TypeTrustCorporation:
		return "trustCorporation"
	case TypeReplacementTrustCorporation:
		return "replacementTrustCorporation"
	case TypeVoucher:
		return "voucher"
	default:
		return ""
	}
}

type Types struct {
	None                        Type
	Donor                       Type
	Attorney                    Type
	ReplacementAttorney         Type
	TrustCorporation            Type
	ReplacementTrustCorporation Type
	CertificateProvider         Type
	PersonToNotify              Type
	AuthorisedSignatory         Type
	IndependentWitness          Type
	Voucher                     Type
}

var ActorTypes = Types{
	None:                        TypeNone,
	Donor:                       TypeDonor,
	Attorney:                    TypeAttorney,
	ReplacementAttorney:         TypeReplacementAttorney,
	TrustCorporation:            TypeTrustCorporation,
	ReplacementTrustCorporation: TypeReplacementTrustCorporation,
	CertificateProvider:         TypeCertificateProvider,
	PersonToNotify:              TypePersonToNotify,
	AuthorisedSignatory:         TypeAuthorisedSignatory,
	IndependentWitness:          TypeIndependentWitness,
	Voucher:                     TypeVoucher,
}
