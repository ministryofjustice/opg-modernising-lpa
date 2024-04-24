package dynamo

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestKeysJSON(t *testing.T) {
	keys := Keys{PK: LpaKey("abc"), SK: DonorKey("123")}

	data, err := json.Marshal(keys)
	assert.Nil(t, err)
	assert.Equal(t, `{"PK":"LPA#abc","SK":"DONOR#123"}`, string(data))

	var v Keys
	err = json.Unmarshal(data, &v)
	assert.Nil(t, err)
	assert.Equal(t, keys, v)
}

func TestKeysAttributeValue(t *testing.T) {
	keys := Keys{PK: LpaKey("abc"), SK: DonorKey("123")}

	data, err := attributevalue.Marshal(keys)
	assert.Nil(t, err)
	assert.Equal(t, &types.AttributeValueMemberM{
		Value: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: "LPA#abc"},
			"SK": &types.AttributeValueMemberS{Value: "DONOR#123"},
		},
	}, data)

	var v Keys
	err = attributevalue.Unmarshal(data, &v)
	assert.Nil(t, err)
	assert.Equal(t, keys, v)
}

func TestKeysWhenMalformed(t *testing.T) {
	testcases := map[string]string{
		"empty":        `{}`,
		"malformed pk": `{"PK":"WHAT","SK":"DONOR#123"}`,
		"bad pk":       `{"PK":"ATTORNEY#123","SK":"DONOR#123"}`,
		"malformed sk": `{"PK":"LPA#123","SK":"WHAT"}`,
		"bad sk":       `{"PK":"LPA#123","SK":"LPA#123"}`,
	}

	for name, str := range testcases {
		t.Run(name, func(t *testing.T) {
			var v Keys
			err := json.Unmarshal([]byte(str), &v)
			assert.Error(t, err)
		})
	}
}

func TestLpaOwnerKey(t *testing.T) {
	for str, key := range map[string]LpaOwnerKeyType{
		"DONOR#123":        LpaOwnerKey(DonorKey("123")),
		"ORGANISATION#123": LpaOwnerKey(OrganisationKey("123")),
	} {
		t.Run(str+"/json", func(t *testing.T) {
			data, err := json.Marshal(key)
			assert.Nil(t, err)
			assert.Equal(t, `"`+str+`"`, string(data))

			var v LpaOwnerKeyType
			err = json.Unmarshal(data, &v)
			assert.Nil(t, err)
			assert.Equal(t, key, v)
		})

		t.Run(str+"/attributevalue", func(t *testing.T) {
			data, err := attributevalue.Marshal(key)
			assert.Nil(t, err)
			assert.Equal(t, &types.AttributeValueMemberS{Value: str}, data)

			var v LpaOwnerKeyType
			err = attributevalue.Unmarshal(data, &v)
			assert.Nil(t, err)
			assert.Equal(t, key, v)
		})
	}

	t.Run("invalid", func(t *testing.T) {
		var v LpaOwnerKeyType
		err := json.Unmarshal([]byte(`"ATTORNEY#123"`), &v)
		assert.Error(t, err)
	})

	t.Run("empty", func(t *testing.T) {
		var v LpaOwnerKeyType
		data, err := json.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, `""`, string(data))

		av, err := attributevalue.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, &types.AttributeValueMemberS{}, av)
	})
}

func TestShareKey(t *testing.T) {
	for str, key := range map[string]ShareKeyType{
		"DONORSHARE#123":               ShareKey(DonorShareKey("123")),
		"CERTIFICATEPROVIDERSHARE#123": ShareKey(CertificateProviderShareKey("123")),
		"ATTORNEYSHARE#123":            ShareKey(AttorneyShareKey("123")),
	} {
		t.Run(str+"/json", func(t *testing.T) {
			data, err := json.Marshal(key)
			assert.Nil(t, err)
			assert.Equal(t, `"`+str+`"`, string(data))

			var v ShareKeyType
			err = json.Unmarshal(data, &v)
			assert.Nil(t, err)
			assert.Equal(t, key, v)
		})

		t.Run(str+"/attributevalue", func(t *testing.T) {
			data, err := attributevalue.Marshal(key)
			assert.Nil(t, err)
			assert.Equal(t, &types.AttributeValueMemberS{Value: str}, data)

			var v ShareKeyType
			err = attributevalue.Unmarshal(data, &v)
			assert.Nil(t, err)
			assert.Equal(t, key, v)
		})
	}

	t.Run("invalid", func(t *testing.T) {
		var v ShareKeyType
		err := json.Unmarshal([]byte(`"ATTORNEY#123"`), &v)
		assert.Error(t, err)
	})

	t.Run("empty", func(t *testing.T) {
		var v ShareKeyType
		data, err := json.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, `""`, string(data))

		av, err := attributevalue.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, &types.AttributeValueMemberS{}, av)
	})
}

func TestShareSortKey(t *testing.T) {
	for str, key := range map[string]ShareSortKeyType{
		"DONORINVITE#123#abc": ShareSortKey(DonorInviteKey("123", "abc")),
		"METADATA#123":        ShareSortKey(MetadataKey("123")),
	} {
		t.Run(str+"/json", func(t *testing.T) {
			data, err := json.Marshal(key)
			assert.Nil(t, err)
			assert.Equal(t, `"`+str+`"`, string(data))

			var v ShareSortKeyType
			err = json.Unmarshal(data, &v)
			assert.Nil(t, err)
			assert.Equal(t, key, v)
		})

		t.Run(str+"/attributevalue", func(t *testing.T) {
			data, err := attributevalue.Marshal(key)
			assert.Nil(t, err)
			assert.Equal(t, &types.AttributeValueMemberS{Value: str}, data)

			var v ShareSortKeyType
			err = attributevalue.Unmarshal(data, &v)
			assert.Nil(t, err)
			assert.Equal(t, key, v)
		})
	}

	t.Run("invalid", func(t *testing.T) {
		var v ShareSortKeyType
		err := json.Unmarshal([]byte(`"ATTORNEY#123"`), &v)
		assert.Error(t, err)
	})

	t.Run("empty", func(t *testing.T) {
		var v ShareSortKeyType
		data, err := json.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, `""`, string(data))

		av, err := attributevalue.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, &types.AttributeValueMemberS{}, av)
	})
}
