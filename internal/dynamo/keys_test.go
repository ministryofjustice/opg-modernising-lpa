package dynamo

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
)

func TestStringKeys(t *testing.T) {
	testcases := map[string]struct {
		fn     func(string) string
		prefix string
	}{
		"LpaKey":                 {LpaKey, "LPA#"},
		"DonorKey":               {DonorKey, "#DONOR#"},
		"SubKey":                 {SubKey, "#SUB#"},
		"AttorneyKey":            {AttorneyKey, "#ATTORNEY#"},
		"CertificateProviderKey": {CertificateProviderKey, "#CERTIFICATE_PROVIDER#"},
		"DocumentKey":            {DocumentKey, "#DOCUMENT#"},
		"MemberKey":              {MemberKey, "MEMBER#"},
		"MemberIDKey":            {MemberIDKey, "MEMBERID#"},
		"OrganisationKey":        {OrganisationKey, "ORGANISATION#"},
		"MetadataKey":            {MetadataKey, "#METADATA#"},
		"DonorShareKey":          {DonorShareKey, "DONORSHARE#"},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.prefix+"S", tc.fn("S"))
		})
	}
}

func TestEvidenceReceivedKey(t *testing.T) {
	assert.Equal(t, "#EVIDENCE_RECEIVED", EvidenceReceivedKey())
}

func TestMemberInviteKey(t *testing.T) {
	assert.Equal(t, "MEMBERINVITE#ZW1haWxAZXhhbXBsZS5jb20=", MemberInviteKey("email@example.com"))
}

func TestDonorInviteKey(t *testing.T) {
	assert.Equal(t, "DONORINVITE#org-id#lpa-id", DonorInviteKey("org-id", "lpa-id"))
}

func TestShareCodeKey(t *testing.T) {
	testcases := map[actor.Type]string{
		actor.TypeDonor:                       "DONORSHARE#",
		actor.TypeAttorney:                    "ATTORNEYSHARE#",
		actor.TypeReplacementAttorney:         "ATTORNEYSHARE#",
		actor.TypeTrustCorporation:            "ATTORNEYSHARE#",
		actor.TypeReplacementTrustCorporation: "ATTORNEYSHARE#",
		actor.TypeCertificateProvider:         "CERTIFICATEPROVIDERSHARE#",
	}

	for actorType, prefix := range testcases {
		t.Run(actorType.String(), func(t *testing.T) {
			pk, err := ShareCodeKey(actorType, "S")
			assert.Nil(t, err)
			assert.Equal(t, prefix+"S", pk)
		})
	}
}

func TestShareCodeKeyWhenUnknownType(t *testing.T) {
	_, err := ShareCodeKey(actor.TypeAuthorisedSignatory, "S")
	assert.NotNil(t, err)
}
