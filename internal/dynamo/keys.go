package dynamo

import (
	"encoding/base64"
	"errors"
	"strings"
)

const (
	lpaPrefix                      = "LPA"
	donorPrefix                    = "DONOR"
	subPrefix                      = "SUB"
	attorneyPrefix                 = "ATTORNEY"
	certificateProviderPrefix      = "CERTIFICATE_PROVIDER"
	documentPrefix                 = "DOCUMENT"
	evidenceReceivedPrefix         = "EVIDENCE_RECEIVED"
	organisationPrefix             = "ORGANISATION"
	memberPrefix                   = "MEMBER"
	memberInvitePrefix             = "MEMBERINVITE"
	memberIDPrefix                 = "MEMBERID"
	metadataPrefix                 = "METADATA"
	donorSharePrefix               = "DONORSHARE"
	donorInvitePrefix              = "DONORINVITE"
	certificateProviderSharePrefix = "CERTIFICATEPROVIDERSHARE"
	attorneySharePrefix            = "ATTORNEYSHARE"
)

func readKey(s string) (any, error) {
	prefix, _, ok := strings.Cut(s, "#")
	if !ok {
		return nil, errors.New("malformed key")
	}

	switch prefix {
	case lpaPrefix:
		return LpaKeyType(s), nil
	case donorSharePrefix:
		return DonorShareKeyType(s), nil
	case certificateProviderSharePrefix:
		return CertificateProviderShareKeyType(s), nil
	case attorneySharePrefix:
		return AttorneyShareKeyType(s), nil
	case donorPrefix:
		return DonorKeyType(s), nil
	case subPrefix:
		return SubKeyType(s), nil
	case attorneyPrefix:
		return AttorneyKeyType(s), nil
	case certificateProviderPrefix:
		return CertificateProviderKeyType(s), nil
	case documentPrefix:
		return DocumentKeyType(s), nil
	case evidenceReceivedPrefix:
		return EvidenceReceivedKeyType(s), nil
	case organisationPrefix:
		return OrganisationKeyType(s), nil
	case memberPrefix:
		return MemberKeyType(s), nil
	case memberInvitePrefix:
		return MemberInviteKeyType(s), nil
	case memberIDPrefix:
		return MemberIDKeyType(s), nil
	case metadataPrefix:
		return MetadataKeyType(s), nil
	case donorInvitePrefix:
		return DonorInviteKeyType(s), nil
	default:
		return nil, errors.New("unknown key prefix")
	}
}

type PK interface{ PK() string }

type SK interface{ SK() string }

type LpaKeyType string

func (t LpaKeyType) PK() string { return string(t) }

// LpaKey is used as the PK for all Lpa related information.
func LpaKey(s string) LpaKeyType {
	return LpaKeyType(lpaPrefix + "#" + s)
}

type DonorKeyType string

func (t DonorKeyType) SK() string { return string(t) }
func (t DonorKeyType) lpaOwner()  {}

// DonorKey is used as the SK (with LpaKey as PK) for donor entered
// information. It is set to PAPER when the donor information has been provided
// from paper forms.
func DonorKey(s string) DonorKeyType {
	return DonorKeyType(donorPrefix + "#" + s)
}

type SubKeyType string

func (t SubKeyType) SK() string { return string(t) }

// SubKey is used as the SK (with LpaKey as PK) to allow queries on a OneLogin
// sub against all Lpas an actor may have provided information on.
func SubKey(s string) SubKeyType {
	return SubKeyType(subPrefix + "#" + s)
}

type AttorneyKeyType string

func (t AttorneyKeyType) SK() string { return string(t) }

// AttorneyKey is used as the SK (with LpaKey as PK) for attorney entered
// information.
func AttorneyKey(s string) AttorneyKeyType {
	return AttorneyKeyType(attorneyPrefix + "#" + s)
}

type CertificateProviderKeyType string

func (t CertificateProviderKeyType) SK() string { return string(t) }

// CertificateProviderKey is used as the SK (with LpaKey as PK) for certificate
// provider entered information.
func CertificateProviderKey(s string) CertificateProviderKeyType {
	return CertificateProviderKeyType(certificateProviderPrefix + "#" + s)
}

type DocumentKeyType string

func (t DocumentKeyType) SK() string { return string(t) }

// DocumentKey is used as the SK (with LpaKey as PK) for any documents uploaded
// as evidence for reduced fees.
func DocumentKey(s3Key string) DocumentKeyType {
	return DocumentKeyType(documentPrefix + "#" + s3Key)
}

type EvidenceReceivedKeyType string

func (t EvidenceReceivedKeyType) SK() string { return string(t) }

// EvidenceReceivedKey is used as the SK (with LpaKey as PK) to show that paper
// evidence has been submitted for an Lpa.
func EvidenceReceivedKey() EvidenceReceivedKeyType {
	return EvidenceReceivedKeyType(evidenceReceivedPrefix + "#")
}

type OrganisationKeyType string

func (t OrganisationKeyType) PK() string { return string(t) }
func (t OrganisationKeyType) SK() string { return string(t) }
func (t OrganisationKeyType) lpaOwner()  {}

// OrganisationKey is used as the PK to group organisation data; or as the SK
// (with OrganisationKey as PK) for the organisation itself; or as the SK (with
// LpaKey as PK) for the donor information entered by a member of an
// organisation.
func OrganisationKey(organisationID string) OrganisationKeyType {
	return OrganisationKeyType(organisationPrefix + "#" + organisationID)
}

type MemberKeyType string

func (t MemberKeyType) SK() string { return string(t) }

// MemberKey is used as the SK (with OrganisationKey as PK) for a member of an
// organisation.
func MemberKey(sessionID string) MemberKeyType {
	return MemberKeyType(memberPrefix + "#" + sessionID)
}

type MemberInviteKeyType string

func (t MemberInviteKeyType) SK() string { return string(t) }

// MemberInviteKey is used as the SK (with OrganisationKey as PK) for a member
// invite.
func MemberInviteKey(email string) MemberInviteKeyType {
	return MemberInviteKeyType(memberInvitePrefix + "#" + base64.StdEncoding.EncodeToString([]byte(email)))
}

type MemberIDKeyType string

func (t MemberIDKeyType) SK() string { return string(t) }

// MemberIDKey is used as the SK (with OrganisationKey as PK) to allow
// retrieving a member using their ID instead of their OneLogin sub.
func MemberIDKey(memberID string) MemberIDKeyType {
	return MemberIDKeyType(memberIDPrefix + "#" + memberID)
}

type MetadataKeyType string

func (t MetadataKeyType) SK() string { return string(t) }
func (t MetadataKeyType) shareSK()   {}

// MetadataKey is used as the SK when the value of the SK is not used for any purpose.
func MetadataKey(s string) MetadataKeyType {
	return MetadataKeyType(metadataPrefix + "#" + s)
}

type DonorShareKeyType string

func (t DonorShareKeyType) PK() string { return string(t) }
func (t DonorShareKeyType) share()     {}

// DonorShareKey is used as the PK for sharing an Lpa with a donor.
func DonorShareKey(code string) DonorShareKeyType {
	return DonorShareKeyType(donorSharePrefix + "#" + code)
}

type DonorInviteKeyType string

func (t DonorInviteKeyType) SK() string { return string(t) }
func (t DonorInviteKeyType) shareSK()   {}

// DonorInviteKey is used as the SK (with DonorShareKey as PK) for an invitation
// to a donor to link an Lpa being created by a member of an organisation.
func DonorInviteKey(organisationID, lpaID string) DonorInviteKeyType {
	return DonorInviteKeyType(donorInvitePrefix + "#" + organisationID + "#" + lpaID)
}

type CertificateProviderShareKeyType string

func (t CertificateProviderShareKeyType) PK() string { return string(t) }
func (t CertificateProviderShareKeyType) share()     {}

// CertificateProviderShareKey is used as the PK for sharing an Lpa with a donor.
func CertificateProviderShareKey(code string) CertificateProviderShareKeyType {
	return CertificateProviderShareKeyType(certificateProviderSharePrefix + "#" + code)
}

type AttorneyShareKeyType string

func (t AttorneyShareKeyType) PK() string { return string(t) }
func (t AttorneyShareKeyType) share()     {}

// AttorneyShareKey is used as the PK for sharing an Lpa with a donor.
func AttorneyShareKey(code string) AttorneyShareKeyType {
	return AttorneyShareKeyType(attorneySharePrefix + "#" + code)
}
