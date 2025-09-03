package dynamo

import (
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	lpaPrefix                       = "LPA"
	donorPrefix                     = "DONOR"
	subPrefix                       = "SUB"
	attorneyPrefix                  = "ATTORNEY"
	certificateProviderPrefix       = "CERTIFICATE_PROVIDER"
	documentPrefix                  = "DOCUMENT"
	evidenceReceivedPrefix          = "EVIDENCE_RECEIVED"
	organisationPrefix              = "ORGANISATION"
	memberPrefix                    = "MEMBER"
	memberInvitePrefix              = "MEMBERINVITE"
	memberIDPrefix                  = "MEMBERID"
	voucherPrefix                   = "VOUCHER"
	metadataPrefix                  = "METADATA"
	voucherAccessSortPrefix         = "VOUCHERACCESSSORT"
	donorAccessPrefix               = "DONORACCESS"
	donorInvitePrefix               = "DONORINVITE"
	certificateProviderAccessPrefix = "CERTIFICATEPROVIDERACCESS"
	attorneyAccessPrefix            = "ATTORNEYACCESS"
	voucherAccessPrefix             = "VOUCHERACCESS"
	scheduledDayPrefix              = "SCHEDULEDDAY"
	scheduledPrefix                 = "SCHEDULED"
	reservedPrefix                  = "RESERVED"
	uidPrefix                       = "UID"
	sessionPrefix                   = "SESSION"
	reusePrefix                     = "REUSE"
	actorAccessPrefix               = "ACTORACCESS"
	accessLimiterPrefix             = "ACCESSRATE"
)

func readKey(s string) (any, error) {
	prefix, _, ok := strings.Cut(s, "#")
	if !ok {
		return nil, fmt.Errorf("malformed key '%s'", s)
	}

	switch prefix {
	case lpaPrefix:
		return LpaKeyType(s), nil
	case donorAccessPrefix:
		return DonorAccessKeyType(s), nil
	case certificateProviderAccessPrefix:
		return CertificateProviderAccessKeyType(s), nil
	case attorneyAccessPrefix:
		return AttorneyAccessKeyType(s), nil
	case voucherAccessPrefix:
		return VoucherAccessKeyType(s), nil
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
	case voucherAccessSortPrefix:
		return VoucherAccessSortKeyType(s), nil
	case donorInvitePrefix:
		return DonorInviteKeyType(s), nil
	case voucherPrefix:
		return VoucherKeyType(s), nil
	case scheduledDayPrefix:
		return ScheduledDayKeyType(s), nil
	case scheduledPrefix:
		return ScheduledKeyType(s), nil
	case reservedPrefix:
		return ReservedKeyType(s), nil
	case uidPrefix:
		return UIDKeyType(s), nil
	case sessionPrefix:
		return SessionKeyType(s), nil
	case reusePrefix:
		return ReuseKeyType(s), nil
	case actorAccessPrefix:
		return ActorAccessKeyType(s), nil
	default:
		return nil, errors.New("unknown key prefix")
	}
}

type PK interface{ PK() string }

type SK interface{ SK() string }

type LpaKeyType string

func (t LpaKeyType) PK() string { return string(t) }
func (t LpaKeyType) ID() string { return t.PK()[len(lpaPrefix)+1:] }

// LpaKey is used as the PK for all Lpa related information.
func LpaKey(s string) LpaKeyType {
	return LpaKeyType(lpaPrefix + "#" + s)
}

type DonorKeyType string

func (t DonorKeyType) SK() string { return string(t) }
func (t DonorKeyType) lpaOwner()  {} // mark as usable with LpaOwnerKey

func (t DonorKeyType) ToSub() SubKeyType {
	_, after, _ := strings.Cut(t.SK(), "#")
	return SubKey(after)
}

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
func (t CertificateProviderKeyType) Sub() string {
	return t.SK()[len(certificateProviderPrefix)+1:]
}

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
func (t OrganisationKeyType) ID() string { return t.PK()[len(organisationPrefix)+1:] }
func (t OrganisationKeyType) lpaOwner()  {} // mark as usable with LpaOwnerKey

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

type VoucherKeyType string

func (t VoucherKeyType) SK() string { return string(t) }

// VoucherKey is used as the SK (with LpaKey as PK) for voucher entered
// information.
func VoucherKey(s string) VoucherKeyType {
	return VoucherKeyType(voucherPrefix + "#" + s)
}

type MetadataKeyType string

func (t MetadataKeyType) SK() string  { return string(t) }
func (t MetadataKeyType) accessSort() {} // mark as usable with AccessSortKey

// Metadata is used as the SK when the value of the SK is not used for any
// purpose.
func MetadataKey(s string) MetadataKeyType {
	return MetadataKeyType(metadataPrefix + "#" + s)
}

type VoucherAccessSortKeyType string

func (t VoucherAccessSortKeyType) SK() string  { return string(t) }
func (t VoucherAccessSortKeyType) accessSort() {} // mark as usable with AccessSortKey

// VoucherAccessSortKey is used as the SK (with AccessKey as PK) for sharing an Lpa
// with an actor.
func VoucherAccessSortKey(lpa LpaKeyType) VoucherAccessSortKeyType {
	return VoucherAccessSortKeyType(voucherAccessSortPrefix + "#" + lpa.ID())
}

type DonorAccessKeyType string

func (t DonorAccessKeyType) PK() string { return string(t) }
func (t DonorAccessKeyType) access()    {} // mark as usable with AccessKey

// DonorAccessKey is used as the PK for sharing an Lpa with a donor.
func DonorAccessKey(code string) DonorAccessKeyType {
	return DonorAccessKeyType(donorAccessPrefix + "#" + code)
}

type DonorInviteKeyType string

func (t DonorInviteKeyType) SK() string  { return string(t) }
func (t DonorInviteKeyType) accessSort() {} // mark as usable with AccessSortKey

// DonorInviteKey is used as the SK (with DonorAccessKey as PK) for an invitation
// to a donor to link an Lpa being created by a member of an organisation.
func DonorInviteKey(organisation OrganisationKeyType, lpa LpaKeyType) DonorInviteKeyType {
	return DonorInviteKeyType(donorInvitePrefix + "#" + organisation.ID() + "#" + lpa.ID())
}

type CertificateProviderAccessKeyType string

func (t CertificateProviderAccessKeyType) PK() string { return string(t) }
func (t CertificateProviderAccessKeyType) access()    {} // mark as usable with AccessKey

// CertificateProviderAccessKey is used as the PK for sharing an Lpa with a certificate provider.
func CertificateProviderAccessKey(code string) CertificateProviderAccessKeyType {
	return CertificateProviderAccessKeyType(certificateProviderAccessPrefix + "#" + code)
}

type AttorneyAccessKeyType string

func (t AttorneyAccessKeyType) PK() string { return string(t) }
func (t AttorneyAccessKeyType) access()    {} // mark as usable with AccessKey

// AttorneyAccessKey is used as the PK for sharing an Lpa with an attorney.
func AttorneyAccessKey(code string) AttorneyAccessKeyType {
	return AttorneyAccessKeyType(attorneyAccessPrefix + "#" + code)
}

type VoucherAccessKeyType string

func (t VoucherAccessKeyType) PK() string { return string(t) }
func (t VoucherAccessKeyType) access()    {} // mark as usable with AccessKey

// VoucherAccessKey is used as the PK for sharing an Lpa with a voucher.
func VoucherAccessKey(code string) VoucherAccessKeyType {
	return VoucherAccessKeyType(voucherAccessPrefix + "#" + code)
}

type ScheduledDayKeyType string

func (t ScheduledDayKeyType) PK() string { return string(t) }

// ScheduledDayKey is used as the PK for a scheduled.Event.
func ScheduledDayKey(at time.Time) ScheduledDayKeyType {
	return ScheduledDayKeyType(scheduledDayPrefix + "#" + at.Format(time.DateOnly))
}

func (t ScheduledDayKeyType) Handled() ScheduledDayKeyType {
	return ScheduledDayKeyType(string(t) + "#HANDLED")
}

type ScheduledKeyType string

func (t ScheduledKeyType) SK() string { return string(t) }

// ScheduledKey is used as the SK for a scheduled.Event.
func ScheduledKey(at time.Time, rnd string) ScheduledKeyType {
	return ScheduledKeyType(scheduledPrefix + "#" + at.Format(time.RFC3339) + "#" + rnd)
}

func PartialScheduledKey() ScheduledKeyType {
	return scheduledPrefix + "#"
}

type ReservedKeyType string

func (t ReservedKeyType) SK() string { return string(t) }

// ReservedKey is used to mark a key prefix as used. This allows creates for
// (A#abc, B#def) to check for the presence of any (A#abc, B#*) by instead using
// a transaction that writes [(A#abc, B#def), (A#abc, Reserved#B#)].
func ReservedKey[T SK](sk func(string) T) ReservedKeyType {
	return ReservedKeyType(reservedPrefix + "#" + sk("").SK())
}

type UIDKeyType string

func (t UIDKeyType) PK() string { return string(t) }

// UIDKey is used as the PK (with MetadataKey as SK) to ensure a UID can only be
// used once.
func UIDKey(uid string) UIDKeyType {
	return UIDKeyType(uidPrefix + "#" + uid)
}

type SessionKeyType string

func (t SessionKeyType) PK() string { return string(t) }

// SessionKey is used as the PK (with MetadataKey as SK) to store a session.
func SessionKey(uid string) SessionKeyType {
	return SessionKeyType(sessionPrefix + "#" + uid)
}

type ReuseKeyType string

func (t ReuseKeyType) PK() string { return string(t) }

// ReuseKey is used as the PK (with MetadataKey as SK) to store reusable data
// for a type of actor.
func ReuseKey(sessionID string, actorType string) ReuseKeyType {
	return ReuseKeyType(reusePrefix + "#" + sessionID + "#" + actorType)
}

type ActorAccessKeyType string

func (t ActorAccessKeyType) PK() string { return string(t) }

// ActorAccessKey is used as the PK (with MetadataKey as SK) to signal that an
// access code has been created for the actor.
func ActorAccessKey(actorUID string) ActorAccessKeyType {
	return ActorAccessKeyType(actorAccessPrefix + "#" + actorUID)
}

type AccessLimiterKeyType string

func (t AccessLimiterKeyType) PK() string { return string(t) }

// AccessLimiterKey is used as the PK (with MetadataKey as SK) to limit the rate
// at which a user is attempting to enter access codes.
func AccessLimiterKey(s string) AccessLimiterKeyType {
	return AccessLimiterKeyType(accessLimiterPrefix + "#" + s)
}
