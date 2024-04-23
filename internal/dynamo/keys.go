package dynamo

import (
	"encoding/base64"
	"fmt"
)

// LpaKey is used as the PK for all Lpa related information.
func LpaKey(s string) string {
	return "LPA#" + s
}

// DonorKey is used as the SK (with LpaKey as PK) for donor entered
// information. It is set to PAPER when the donor information has been provided
// from paper forms.
func DonorKey(s string) string {
	return "DONOR#" + s
}

// SubKey is used as the SK (with LpaKey as PK) to allow queries on a OneLogin
// sub against all Lpas an actor may have provided information on.
func SubKey(s string) string {
	return "SUB#" + s
}

// AttorneyKey is used as the SK (with LpaKey as PK) for attorney entered
// information.
func AttorneyKey(s string) string {
	return "ATTORNEY#" + s
}

// CertificateProviderKey is used as the SK (with LpaKey as PK) for certificate
// provider entered information.
func CertificateProviderKey(s string) string {
	return "CERTIFICATE_PROVIDER#" + s
}

// DocumentKey is used as the SK (with LpaKey as PK) for any documents uploaded
// as evidence for reduced fees.
func DocumentKey(s3Key string) string {
	return "DOCUMENT#" + s3Key
}

// EvidenceReceivedKey is used as the SK (with LpaKey as PK) to show that paper
// evidence has been submitted for an Lpa.
func EvidenceReceivedKey() string {
	return "EVIDENCE_RECEIVED"
}

// OrganisationKey is used as the PK to group organisation data; or as the SK
// (with OrganisationKey as PK) for the organisation itself; or as the SK (with
// LpaKey as PK) for the donor information entered by a member of an
// organisation.
func OrganisationKey(organisationID string) string {
	return "ORGANISATION#" + organisationID
}

// MemberKey is used as the SK (with OrganisationKey as PK) for a member of an
// organisation.
func MemberKey(sessionID string) string {
	return "MEMBER#" + sessionID
}

// MemberInviteKey is used as the SK (with OrganisationKey as PK) for a member
// invite.
func MemberInviteKey(email string) string {
	return fmt.Sprintf("MEMBERINVITE#%s", base64.StdEncoding.EncodeToString([]byte(email)))
}

// MemberIDKey is used as the SK (with OrganisationKey as PK) to allow
// retrieving a member using their ID instead of their OneLogin sub.
func MemberIDKey(memberID string) string {
	return "MEMBERID#" + memberID
}

// MetadataKey is used as the SK when the value of the SK is not used for any purpose.
func MetadataKey(s string) string {
	return "METADATA#" + s
}

// DonorShareKey is used as the PK for sharing an Lpa with a donor.
func DonorShareKey(code string) string {
	return "DONORSHARE#" + code
}

// DonorInviteKey is used as the SK (with DonorShareKey as PK) for an invitation
// to a donor to link an Lpa being created by a member of an organisation.
func DonorInviteKey(organisationID, lpaID string) string {
	return "DONORINVITE#" + organisationID + "#" + lpaID
}

// CertificateProviderShareKey is used as the PK for sharing an Lpa with a donor.
func CertificateProviderShareKey(code string) string {
	return "CERTIFICATEPROVIDERSHARE#" + code
}

// AttorneyShareKey is used as the PK for sharing an Lpa with a donor.
func AttorneyShareKey(code string) string {
	return "ATTORNEYSHARE#" + code
}
