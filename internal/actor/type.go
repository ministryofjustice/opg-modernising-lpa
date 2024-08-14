package actor

//go:generate enumerator -type Type -linecomment -trimprefix
type Type uint8

const (
	TypeNone                        Type = iota
	TypeDonor                            // donor
	TypeAttorney                         // attorney
	TypeReplacementAttorney              // replacementAttorney
	TypeCertificateProvider              // certificateProvider
	TypePersonToNotify                   // personToNotify
	TypeAuthorisedSignatory              // signatory
	TypeIndependentWitness               // independentWitness
	TypeTrustCorporation                 // trustCorporation
	TypeReplacementTrustCorporation      // replacementTrustCorporation
	TypeVoucher                          // voucher
)
