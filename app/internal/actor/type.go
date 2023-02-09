package actor

type Type int

const (
	TypeNone Type = iota
	TypeDonor
	TypeAttorney
	TypeReplacementAttorney
	TypeCertificateProvider
	TypePersonToNotify
)
