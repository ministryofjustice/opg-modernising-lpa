package dynamo

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Keys struct {
	PK PK
	SK SK
}

func (k *Keys) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var skeys struct{ PK, SK string }
	err := attributevalue.Unmarshal(av, &skeys)
	if err != nil {
		return err
	}

	k.PK, err = readPK(skeys.PK)
	if err != nil {
		return err
	}

	k.SK, err = readSK(skeys.SK)
	if err != nil {
		return err
	}

	return nil
}

func (k *Keys) UnmarshalJSON(text []byte) error {
	var skeys struct{ PK, SK string }
	err := json.Unmarshal(text, &skeys)
	if err != nil {
		return err
	}

	k.PK, err = readPK(skeys.PK)
	if err != nil {
		return err
	}

	k.SK, err = readSK(skeys.SK)
	if err != nil {
		return err
	}

	return nil
}

func readPK(s string) (PK, error) {
	prefix, _, ok := strings.Cut(s, "#")
	if !ok {
		return nil, errors.New("malformed pk")
	}

	switch prefix {
	case "LPA":
		return LpaKeyType(s), nil
	case "ORGANISATION":
		return OrganisationKeyType(s), nil
	case "DONORSHARE":
		return DonorShareKeyType(s), nil
	case "CERTIFICATEPROVIDERSHARE":
		return CertificateProviderShareKeyType(s), nil
	case "ATTORNEYSHARE":
		return AttorneyShareKeyType(s), nil
	default:
		return nil, errors.New("unknown pk prefix")
	}
}

func readSK(s string) (SK, error) {
	prefix, _, ok := strings.Cut(s, "#")
	if !ok {
		return nil, errors.New("malformed sk")
	}

	switch prefix {
	case "DONOR":
		return DonorKeyType(s), nil
	case "SUB":
		return SubKeyType(s), nil
	case "ATTORNEY":
		return AttorneyKeyType(s), nil
	case "CERTIFICATE_PROVIDER":
		return CertificateProviderKeyType(s), nil
	case "DOCUMENT":
		return DocumentKeyType(s), nil
	case "EVIDENCE_RECEIVED":
		return EvidenceReceivedKey(), nil
	case "ORGANISATION":
		return OrganisationKeyType(s), nil
	case "MEMBER":
		return MemberKeyType(s), nil
	case "MEMBERINVITE":
		return MemberInviteKeyType(s), nil
	case "MEMBERID":
		return MemberIDKeyType(s), nil
	case "METADATA":
		return MetadataKeyType(s), nil
	case "DONORINVITE":
		return DonorInviteKeyType(s), nil
	default:
		return nil, errors.New("unknown sk prefix")
	}
}

type PK interface{ PK() string }
type SK interface{ SK() string }

type LpaOwnerKeyType struct{ sk SK }

func LpaOwnerKey(sk SK) LpaOwnerKeyType {
	return LpaOwnerKeyType{sk: sk}
}

func (k LpaOwnerKeyType) MarshalJSON() ([]byte, error) {
	if k.sk == nil {
		return []byte("null"), nil
	}

	return []byte(`"` + k.sk.SK() + `"`), nil
}

func (k LpaOwnerKeyType) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	if k.sk == nil {
		return &types.AttributeValueMemberNULL{Value: true}, nil
	}

	return attributevalue.Marshal(k.sk.SK())
}

func (k *LpaOwnerKeyType) UnmarshalJSON(text []byte) error {
	var s string
	err := json.Unmarshal(text, &s)
	if err != nil {
		return err
	}

	// TODO: lock down allowed types

	if s != "" {
		k.sk, err = readSK(s)
	}
	return err
}

func (k *LpaOwnerKeyType) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	err := attributevalue.Unmarshal(av, &s)
	if err != nil {
		return err
	}

	// TODO: lock down allowed types

	if s != "" {
		k.sk, err = readSK(s)
	}
	return err
}

func (k LpaOwnerKeyType) Equals(sk SK) bool {
	return k.sk == sk
}

func (k LpaOwnerKeyType) SK() string {
	return k.sk.SK()
}

func (k LpaOwnerKeyType) IsOrganisation() bool {
	_, ok := k.sk.(OrganisationKeyType)
	return ok
}

type ShareKeyType struct{ pk PK }

func ShareKey(pk PK) ShareKeyType {
	return ShareKeyType{pk: pk}
}

func (k ShareKeyType) MarshalJSON() ([]byte, error) {
	if k.pk == nil {
		return []byte("null"), nil
	}

	return []byte(`"` + k.pk.PK() + `"`), nil
}

func (k *ShareKeyType) UnmarshalJSON(text []byte) error {
	var s string
	err := json.Unmarshal(text, &s)
	if err != nil {
		return err
	}

	// TODO: lock down allowed types

	if s != "" {
		k.pk, err = readPK(s)
	}
	return err
}

func (k ShareKeyType) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	if k.pk == nil {
		return &types.AttributeValueMemberNULL{Value: true}, nil
	}

	return attributevalue.Marshal(k.pk.PK())
}

func (k *ShareKeyType) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	err := attributevalue.Unmarshal(av, &s)
	if err != nil {
		return err
	}

	// TODO: lock down allowed types

	if s != "" {
		k.pk, err = readPK(s)
	}
	return err
}

func (k ShareKeyType) PK() string {
	return k.pk.PK()
}

type ShareKeySKType struct{ sk SK }

func ShareKeySK(sk SK) ShareKeySKType {
	return ShareKeySKType{sk: sk}
}

func (k ShareKeySKType) MarshalJSON() ([]byte, error) {
	if k.sk == nil {
		return []byte("null"), nil
	}

	return []byte(`"` + k.sk.SK() + `"`), nil
}

func (k *ShareKeySKType) UnmarshalJSON(text []byte) error {
	var s string
	err := json.Unmarshal(text, &s)
	if err != nil {
		return err
	}

	// TODO: lock down allowed types

	if s != "" {
		k.sk, err = readSK(s)
	}
	return err
}

func (k ShareKeySKType) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	if k.sk == nil {
		return &types.AttributeValueMemberNULL{Value: true}, nil
	}

	return attributevalue.Marshal(k.sk.SK())
}

func (k *ShareKeySKType) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	err := attributevalue.Unmarshal(av, &s)
	if err != nil {
		return err
	}

	// TODO: lock down allowed types

	if s != "" {
		k.sk, err = readSK(s)
	}
	return err
}

func (k ShareKeySKType) SK() string {
	return k.sk.SK()
}

type LpaKeyType string

func (t LpaKeyType) PK() string { return string(t) }

// LpaKey is used as the PK for all Lpa related information.
func LpaKey(s string) LpaKeyType {
	return LpaKeyType("LPA#" + s)
}

type DonorKeyType string

func (t DonorKeyType) SK() string { return string(t) }

// DonorKey is used as the SK (with LpaKey as PK) for donor entered
// information. It is set to PAPER when the donor information has been provided
// from paper forms.
func DonorKey(s string) DonorKeyType {
	return DonorKeyType("DONOR#" + s)
}

type SubKeyType string

func (t SubKeyType) SK() string { return string(t) }

// SubKey is used as the SK (with LpaKey as PK) to allow queries on a OneLogin
// sub against all Lpas an actor may have provided information on.
func SubKey(s string) SubKeyType {
	return SubKeyType("SUB#" + s)
}

type AttorneyKeyType string

func (t AttorneyKeyType) SK() string { return string(t) }

// AttorneyKey is used as the SK (with LpaKey as PK) for attorney entered
// information.
func AttorneyKey(s string) AttorneyKeyType {
	return AttorneyKeyType("ATTORNEY#" + s)
}

type CertificateProviderKeyType string

func (t CertificateProviderKeyType) SK() string { return string(t) }

// CertificateProviderKey is used as the SK (with LpaKey as PK) for certificate
// provider entered information.
func CertificateProviderKey(s string) CertificateProviderKeyType {
	return CertificateProviderKeyType("CERTIFICATE_PROVIDER#" + s)
}

type DocumentKeyType string

func (t DocumentKeyType) SK() string { return string(t) }

// DocumentKey is used as the SK (with LpaKey as PK) for any documents uploaded
// as evidence for reduced fees.
func DocumentKey(s3Key string) DocumentKeyType {
	return DocumentKeyType("DOCUMENT#" + s3Key)
}

type EvidenceReceivedKeyType string

func (t EvidenceReceivedKeyType) SK() string { return string(t) }

// EvidenceReceivedKey is used as the SK (with LpaKey as PK) to show that paper
// evidence has been submitted for an Lpa.
func EvidenceReceivedKey() EvidenceReceivedKeyType {
	return EvidenceReceivedKeyType("EVIDENCE_RECEIVED")
}

type OrganisationKeyType string

func (t OrganisationKeyType) PK() string { return string(t) }
func (t OrganisationKeyType) SK() string { return string(t) }

// OrganisationKey is used as the PK to group organisation data; or as the SK
// (with OrganisationKey as PK) for the organisation itself; or as the SK (with
// LpaKey as PK) for the donor information entered by a member of an
// organisation.
func OrganisationKey(organisationID string) OrganisationKeyType {
	return OrganisationKeyType("ORGANISATION#" + organisationID)
}

type MemberKeyType string

func (t MemberKeyType) SK() string { return string(t) }

// MemberKey is used as the SK (with OrganisationKey as PK) for a member of an
// organisation.
func MemberKey(sessionID string) MemberKeyType {
	return MemberKeyType("MEMBER#" + sessionID)
}

type MemberInviteKeyType string

func (t MemberInviteKeyType) SK() string { return string(t) }

// MemberInviteKey is used as the SK (with OrganisationKey as PK) for a member
// invite.
func MemberInviteKey(email string) MemberInviteKeyType {
	return MemberInviteKeyType("MEMBERINVITE#" + base64.StdEncoding.EncodeToString([]byte(email)))
}

type MemberIDKeyType string

func (t MemberIDKeyType) SK() string { return string(t) }

// MemberIDKey is used as the SK (with OrganisationKey as PK) to allow
// retrieving a member using their ID instead of their OneLogin sub.
func MemberIDKey(memberID string) MemberIDKeyType {
	return MemberIDKeyType("MEMBERID#" + memberID)
}

type MetadataKeyType string

func (t MetadataKeyType) SK() string { return string(t) }

// MetadataKey is used as the SK when the value of the SK is not used for any purpose.
func MetadataKey(s string) MetadataKeyType {
	return MetadataKeyType("METADATA#" + s)
}

type DonorShareKeyType string

func (t DonorShareKeyType) PK() string { return string(t) }

// DonorShareKey is used as the PK for sharing an Lpa with a donor.
func DonorShareKey(code string) DonorShareKeyType {
	return DonorShareKeyType("DONORSHARE#" + code)
}

type DonorInviteKeyType string

func (t DonorInviteKeyType) SK() string { return string(t) }

// DonorInviteKey is used as the SK (with DonorShareKey as PK) for an invitation
// to a donor to link an Lpa being created by a member of an organisation.
func DonorInviteKey(organisationID, lpaID string) DonorInviteKeyType {
	return DonorInviteKeyType("DONORINVITE#" + organisationID + "#" + lpaID)
}

type CertificateProviderShareKeyType string

func (t CertificateProviderShareKeyType) PK() string { return string(t) }

// CertificateProviderShareKey is used as the PK for sharing an Lpa with a donor.
func CertificateProviderShareKey(code string) CertificateProviderShareKeyType {
	return CertificateProviderShareKeyType("CERTIFICATEPROVIDERSHARE#" + code)
}

type AttorneyShareKeyType string

func (t AttorneyShareKeyType) PK() string { return string(t) }

// AttorneyShareKey is used as the PK for sharing an Lpa with a donor.
func AttorneyShareKey(code string) AttorneyShareKeyType {
	return AttorneyShareKeyType("ATTORNEYSHARE#" + code)
}
