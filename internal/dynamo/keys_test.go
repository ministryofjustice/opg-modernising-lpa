package dynamo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringKeys(t *testing.T) {
	testcases := map[string]struct {
		fn     func(string) string
		prefix string
	}{
		"LpaKey":                      {LpaKey, "LPA#"},
		"DonorKey":                    {DonorKey, "DONOR#"},
		"SubKey":                      {SubKey, "SUB#"},
		"AttorneyKey":                 {AttorneyKey, "ATTORNEY#"},
		"CertificateProviderKey":      {CertificateProviderKey, "CERTIFICATE_PROVIDER#"},
		"DocumentKey":                 {DocumentKey, "DOCUMENT#"},
		"MemberKey":                   {MemberKey, "MEMBER#"},
		"MemberIDKey":                 {MemberIDKey, "MEMBERID#"},
		"OrganisationKey":             {OrganisationKey, "ORGANISATION#"},
		"MetadataKey":                 {MetadataKey, "METADATA#"},
		"DonorShareKey":               {DonorShareKey, "DONORSHARE#"},
		"CertificateProviderShareKey": {CertificateProviderShareKey, "CERTIFICATEPROVIDERSHARE#"},
		"AttorneyShareKey":            {AttorneyShareKey, "ATTORNEYSHARE#"},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.prefix+"S", tc.fn("S"))
		})
	}
}

func TestEvidenceReceivedKey(t *testing.T) {
	assert.Equal(t, "EVIDENCE_RECEIVED", EvidenceReceivedKey())
}

func TestMemberInviteKey(t *testing.T) {
	assert.Equal(t, "MEMBERINVITE#ZW1haWxAZXhhbXBsZS5jb20=", MemberInviteKey("email@example.com"))
}

func TestDonorInviteKey(t *testing.T) {
	assert.Equal(t, "DONORINVITE#org-id#lpa-id", DonorInviteKey("org-id", "lpa-id"))
}
