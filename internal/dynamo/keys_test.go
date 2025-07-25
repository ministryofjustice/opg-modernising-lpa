package dynamo

import (
	"encoding/json"
	"testing"
	"time"

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
		"DonorShareKey":               {DonorAccessKey("S"), "DONORSHARE#S"},
		"CertificateProviderShareKey": {CertificateProviderAccessKey("S"), "CERTIFICATEPROVIDERSHARE#S"},
		"AttorneyShareKey":            {AttorneyAccessKey("S"), "ATTORNEYSHARE#S"},
		"VoucherShareKey":             {VoucherAccessKey("S"), "VOUCHERSHARE#S"},
		"ScheduledDayKey":             {ScheduledDayKey(time.Date(2024, time.January, 2, 12, 13, 14, 15, time.UTC)), "SCHEDULEDDAY#2024-01-02"},
		"UIDKey":                      {UIDKey("S"), "UID#S"},
		"SessionKey":                  {SessionKey("S"), "SESSION#S"},
		"ReuseKey":                    {ReuseKey("S", "T"), "REUSE#S#T"},
		"ActorAccessKey":              {ActorAccessKey("S"), "ACTORACCESS#S"},
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
		"VoucherShareSortKey":    {VoucherShareSortKey(LpaKey("S")), "VOUCHERSHARESORT#S"},
		"DonorInviteKey":         {DonorInviteKey(OrganisationKey("org-id"), LpaKey("lpa-id")), "DONORINVITE#org-id#lpa-id"},
		"VoucherKey":             {VoucherKey("S"), "VOUCHER#S"},
		"ScheduledKey":           {ScheduledKey(time.Date(2024, time.January, 2, 12, 13, 14, 15, time.UTC), "some-string"), "SCHEDULED#2024-01-02T12:13:14Z#some-string"},
		"ReservedKey":            {ReservedKey(VoucherKey), "RESERVED#VOUCHER#"},
		"PartialScheduledKey":    {PartialScheduledKey(), "SCHEDULED#"},
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
	for _, key := range []interface{ share() }{
		DonorAccessKey("hey"),
		CertificateProviderAccessKey("what"),
		AttorneyAccessKey("hello"),
	} {
		key.share()
	}
}

func TestShareSortKeyTypes(t *testing.T) {
	for _, key := range []interface{ shareSort() }{
		MetadataKey("hey"),
		DonorInviteKey(OrganisationKey("what"), LpaKey("hello")),
		VoucherShareSortKey(LpaKey("hi")),
	} {
		key.shareSort()
	}
}

func TestDonorKeyTypeToSub(t *testing.T) {
	assert.Equal(t, SubKey("xyz"), DonorKey("xyz").ToSub())
}

func TestScheduledDayKeyTypeHandled(t *testing.T) {
	key := ScheduledDayKey(time.Now())

	assert.Equal(t, key.PK()+"#HANDLED", key.Handled().PK())
}

func TestCertificateProviderKeyTypeSub(t *testing.T) {
	assert.Equal(t, "xyz", CertificateProviderKey("xyz").Sub())
}
