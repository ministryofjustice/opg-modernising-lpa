package scheduled

//go:generate enumerator -type Action -trimprefix
type Action uint8

const (
	// ActionExpireDonorIdentity will check that the target donor has not signed
	// their LPA, and if so remove their identity data and notify them of the
	// change.
	ActionExpireDonorIdentity Action = iota + 1

	// ActionRemindCertificateProviderToComplete will check that the target
	// certificate provider has neither provided the certificate nor opted-out,
	// and if so send them a reminder email or letter, plus another to the donor
	// (or correspondent, if set).
	ActionRemindCertificateProviderToComplete

	// ActionRemindCertificateProviderToConfirmIdentity will check that the target
	// certificate provider has not confirmed their identity, and if so send them
	// a reminder email or letter, plus another to the donor (or correspondent, if
	// set).
	ActionRemindCertificateProviderToConfirmIdentity

	// ActionRemindAttorneyToComplete will check that the target attorney has
	// neither signed nor opted-out, and if so send them a reminder email or
	// letter, plus another to the donor (or correspondent, if set).
	ActionRemindAttorneyToComplete
)
