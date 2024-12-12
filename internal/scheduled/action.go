package scheduled

import "time"

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
)

// ExpireDonorIdentityAt gives the time to run ActionExpireDonorIdentity, which
// is 6 months after the donor has checked the LPA.
func ExpireDonorIdentityAt(donorCheckedAt time.Time) time.Time {
	return donorCheckedAt.AddDate(0, 6, 0)
}

// RemindCertificateProviderAt gives the time to run
// ActionRemindCertificateProviderToConfirmIdentity and
// ActionRemindCertificateProviderToComplete, which is the latest of
//
//	3 months after the certificate provider invite is sent
//	3 months until the LPA expires
func RemindCertificateProviderAt(inviteSentAt, donorSignedAt time.Time) time.Time {
	afterInvite := inviteSentAt.AddDate(0, 3, 0)
	beforeExpiry := donorSignedAt.AddDate(0, 21, 0)

	if afterInvite.After(beforeExpiry) {
		return afterInvite
	}

	return beforeExpiry
}
