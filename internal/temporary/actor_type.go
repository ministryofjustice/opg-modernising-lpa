package temporary

type ActorType uint8

const (
	ActorTypeNone ActorType = iota
	ActorTypeDonor
	ActorTypeAttorney
	ActorTypeReplacementAttorney
	ActorTypeCertificateProvider
	ActorTypePersonToNotify
	ActorTypeAuthorisedSignatory
	ActorTypeIndependentWitness
	ActorTypeTrustCorporation
	ActorTypeReplacementTrustCorporation
	ActorTypeVoucher
)

func (t ActorType) String() string {
	switch t {
	case ActorTypeDonor:
		return "donor"
	case ActorTypeAttorney:
		return "attorney"
	case ActorTypeReplacementAttorney:
		return "replacementAttorney"
	case ActorTypeCertificateProvider:
		return "certificateProvider"
	case ActorTypePersonToNotify:
		return "personToNotify"
	case ActorTypeAuthorisedSignatory:
		return "signatory"
	case ActorTypeIndependentWitness:
		return "independentWitness"
	case ActorTypeTrustCorporation:
		return "trustCorporation"
	case ActorTypeReplacementTrustCorporation:
		return "replacementTrustCorporation"
	case ActorTypeVoucher:
		return "voucher"
	default:
		return ""
	}
}

type Types struct {
	None                        ActorType
	Donor                       ActorType
	Attorney                    ActorType
	ReplacementAttorney         ActorType
	TrustCorporation            ActorType
	ReplacementTrustCorporation ActorType
	CertificateProvider         ActorType
	PersonToNotify              ActorType
	AuthorisedSignatory         ActorType
	IndependentWitness          ActorType
	Voucher                     ActorType
}

var ActorTypes = Types{
	None:                        ActorTypeNone,
	Donor:                       ActorTypeDonor,
	Attorney:                    ActorTypeAttorney,
	ReplacementAttorney:         ActorTypeReplacementAttorney,
	TrustCorporation:            ActorTypeTrustCorporation,
	ReplacementTrustCorporation: ActorTypeReplacementTrustCorporation,
	CertificateProvider:         ActorTypeCertificateProvider,
	PersonToNotify:              ActorTypePersonToNotify,
	AuthorisedSignatory:         ActorTypeAuthorisedSignatory,
	IndependentWitness:          ActorTypeIndependentWitness,
	Voucher:                     ActorTypeVoucher,
}
