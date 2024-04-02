package dynamo

import (
	"encoding/base64"
	"fmt"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
)

func LpaKey(s string) string {
	return "LPA#" + s
}

func DonorKey(s string) string {
	return "#DONOR#" + s
}

func SubKey(s string) string {
	return "#SUB#" + s
}

func AttorneyKey(s string) string {
	return "#ATTORNEY#" + s
}

func CertificateProviderKey(s string) string {
	return "#CERTIFICATE_PROVIDER#" + s
}

func DocumentKey(s3Key string) string {
	return "#DOCUMENT#" + s3Key
}

func EvidenceReceivedKey() string {
	return "#EVIDENCE_RECEIVED"
}

func MemberKey(sessionID string) string {
	return "MEMBER#" + sessionID
}

func MemberInviteKey(email string) string {
	return fmt.Sprintf("MEMBERINVITE#%s", base64.StdEncoding.EncodeToString([]byte(email)))
}

func MemberIDKey(memberID string) string {
	return "MEMBERID#" + memberID
}

func OrganisationKey(organisationID string) string {
	return "ORGANISATION#" + organisationID
}

func MetadataKey(s string) string {
	return "#METADATA#" + s
}

func DonorShareKey(code string) string {
	return "DONORSHARE#" + code
}

func DonorInviteKey(organisationID, lpaID string) string {
	return "DONORINVITE#" + organisationID + "#" + lpaID
}

func ShareCodeKey(actorType actor.Type, shareCode string) (pk string, err error) {
	switch actorType {
	case actor.TypeDonor:
		return DonorShareKey(shareCode), nil
	// As attorneys and replacement attorneys share the same landing page we can't
	// differentiate between them
	case actor.TypeAttorney, actor.TypeReplacementAttorney, actor.TypeTrustCorporation, actor.TypeReplacementTrustCorporation:
		return "ATTORNEYSHARE#" + shareCode, nil
	case actor.TypeCertificateProvider:
		return "CERTIFICATEPROVIDERSHARE#" + shareCode, nil
	default:
		return "", fmt.Errorf("cannot have share code for actorType=%v", actorType)
	}
}
