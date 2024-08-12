package dynamo

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestReadKeyMalformed(t *testing.T) {
	testcases := map[string]string{
		"empty":          "",
		"no hash":        "DONOR",
		"unknown prefix": "WHAT#123",
	}

	for name, key := range testcases {
		t.Run(name, func(t *testing.T) {
			_, err := readKey(key)
			assert.Error(t, err)
		})
	}
}

func TestPK(t *testing.T) {
	testcases := map[string]struct {
		key PK
		str string
	}{
		"LpaKey":                      {LpaKey("S"), "LPA#S"},
		"OrganisationKey":             {OrganisationKey("S"), "ORGANISATION#S"},
		"DonorShareKey":               {DonorShareKey("S"), "DONORSHARE#S"},
		"CertificateProviderShareKey": {CertificateProviderShareKey("S"), "CERTIFICATEPROVIDERSHARE#S"},
		"AttorneyShareKey":            {AttorneyShareKey("S"), "ATTORNEYSHARE#S"},
		"VoucherShareKey":             {VoucherShareKey("S"), "VOUCHERSHARE#S"},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.str, tc.key.PK())
		})

		t.Run(name+"/read", func(t *testing.T) {
			pk, err := readKey(tc.str)
			assert.Nil(t, err)
			assert.Equal(t, tc.key, pk)
		})

		t.Run(name+"/json", func(t *testing.T) {
			data, err := json.Marshal(tc.key)
			assert.Nil(t, err)
			assert.Equal(t, `"`+tc.str+`"`, string(data))
		})

		t.Run(name+"/attributevalue", func(t *testing.T) {
			data, err := attributevalue.Marshal(tc.key)
			assert.Nil(t, err)
			assert.Equal(t, &types.AttributeValueMemberS{Value: tc.str}, data)
		})
	}
}

func TestSK(t *testing.T) {
	testcases := map[string]struct {
		key SK
		str string
	}{
		"DonorKey":               {DonorKey("S"), "DONOR#S"},
		"SubKey":                 {SubKey("S"), "SUB#S"},
		"AttorneyKey":            {AttorneyKey("S"), "ATTORNEY#S"},
		"CertificateProviderKey": {CertificateProviderKey("S"), "CERTIFICATE_PROVIDER#S"},
		"DocumentKey":            {DocumentKey("S"), "DOCUMENT#S"},
		"EvidenceReceivedKey":    {EvidenceReceivedKey(), "EVIDENCE_RECEIVED#"},
		"MemberKey":              {MemberKey("S"), "MEMBER#S"},
		"MemberInviteKey":        {MemberInviteKey("email@example.com"), "MEMBERINVITE#ZW1haWxAZXhhbXBsZS5jb20="},
		"MemberIDKey":            {MemberIDKey("S"), "MEMBERID#S"},
		"OrganisationKey":        {OrganisationKey("S"), "ORGANISATION#S"},
		"MetadataKey":            {MetadataKey("S"), "METADATA#S"},
		"DonorInviteKey":         {DonorInviteKey(OrganisationKey("org-id"), LpaKey("lpa-id")), "DONORINVITE#org-id#lpa-id"},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.str, tc.key.SK())
		})

		t.Run(name+"/read", func(t *testing.T) {
			sk, err := readKey(tc.str)
			assert.Nil(t, err)
			assert.Equal(t, tc.key, sk)
		})

		t.Run(name+"/json", func(t *testing.T) {
			data, err := json.Marshal(tc.key)
			assert.Nil(t, err)
			assert.Equal(t, `"`+tc.str+`"`, string(data))
		})

		t.Run(name+"/attributevalue", func(t *testing.T) {
			data, err := attributevalue.Marshal(tc.key)
			assert.Nil(t, err)
			assert.Equal(t, &types.AttributeValueMemberS{Value: tc.str}, data)
		})
	}
}

func TestLpaOwnerKeyTypes(t *testing.T) {
	for _, key := range []interface{ lpaOwner() }{DonorKey("hey"), OrganisationKey("what")} {
		key.lpaOwner()
	}
}

func TestShareKeyTypes(t *testing.T) {
	for _, key := range []interface{ share() }{DonorShareKey("hey"), CertificateProviderShareKey("what"), AttorneyShareKey("hello")} {
		key.share()
	}
}

func TestShareSortKeyTypes(t *testing.T) {
	for _, key := range []interface{ shareSort() }{MetadataKey("hey"), DonorInviteKey(OrganisationKey("what"), LpaKey("hello"))} {
		key.shareSort()
	}
}
