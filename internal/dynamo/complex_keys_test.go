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

func TestKeysUnmarshalJSONWhenError(t *testing.T) {
	var v Keys
	err := json.Unmarshal([]byte(`hey`), &v)
	assert.Error(t, err)
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

func TestKeysUnmarshalAttributeValueWhenError(t *testing.T) {
	var v Keys
	err := attributevalue.Unmarshal(&types.AttributeValueMemberS{Value: "hey"}, &v)
	assert.Error(t, err)
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
		t.Run(str+"/SK", func(t *testing.T) {
			assert.Equal(t, str, key.SK())
		})

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

	t.Run("Equals", func(t *testing.T) {
		x := LpaOwnerKey(DonorKey("abc"))
		y := LpaOwnerKey(DonorKey("def"))
		z := LpaOwnerKey(OrganisationKey("abc"))

		assert.True(t, x.Equals(x))
		assert.True(t, x.Equals(DonorKey("abc")))
		assert.True(t, x.Equals(LpaOwnerKey(DonorKey("abc"))))
		assert.True(t, z.Equals(z))

		assert.False(t, x.Equals(y))
		assert.False(t, x.Equals(z))

		assert.False(t, LpaOwnerKey(nil).Equals(nil))
		assert.False(t, LpaOwnerKey(DonorKey("")).Equals(LpaOwnerKey(nil)))
		assert.False(t, LpaOwnerKey(nil).Equals(DonorKey("")))
	})

	t.Run("Organisation", func(t *testing.T) {
		key, ok := LpaOwnerKey(OrganisationKey("a")).Organisation()
		assert.Equal(t, OrganisationKey("a"), key)
		assert.True(t, ok)
		_, ok = LpaOwnerKey(DonorKey("")).Organisation()
		assert.False(t, ok)
		_, ok = LpaOwnerKey(nil).Organisation()
		assert.False(t, ok)
	})

	t.Run("malformed", func(t *testing.T) {
		var v LpaOwnerKeyType
		err := json.Unmarshal([]byte(`"WHAT"`), &v)
		assert.Error(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		var v LpaOwnerKeyType
		err := json.Unmarshal([]byte(`"ATTORNEY#123"`), &v)
		assert.Error(t, err)
	})

	t.Run("empty json", func(t *testing.T) {
		var v LpaOwnerKeyType
		err := json.Unmarshal([]byte(`""`), &v)
		assert.Nil(t, err)
		assert.Equal(t, LpaOwnerKeyType{}, v)

		data, err := json.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, `""`, string(data))
	})

	t.Run("empty attributevalue", func(t *testing.T) {
		var v LpaOwnerKeyType
		err := attributevalue.Unmarshal(&types.AttributeValueMemberS{Value: ""}, &v)
		assert.Nil(t, err)
		assert.Equal(t, LpaOwnerKeyType{}, v)

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
		t.Run(str+"/PK", func(t *testing.T) {
			assert.Equal(t, str, key.PK())
		})

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

	t.Run("malformed", func(t *testing.T) {
		var v ShareKeyType
		err := json.Unmarshal([]byte(`"WHAT"`), &v)
		assert.Error(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		var v ShareKeyType
		err := json.Unmarshal([]byte(`"ATTORNEY#123"`), &v)
		assert.Error(t, err)
	})

	t.Run("empty json", func(t *testing.T) {
		var v ShareKeyType
		err := json.Unmarshal([]byte(`""`), &v)
		assert.Nil(t, err)
		assert.Equal(t, ShareKeyType{}, v)

		data, err := json.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, `""`, string(data))
	})

	t.Run("empty attributevalue", func(t *testing.T) {
		var v ShareKeyType
		err := attributevalue.Unmarshal(&types.AttributeValueMemberS{Value: ""}, &v)
		assert.Nil(t, err)
		assert.Equal(t, ShareKeyType{}, v)

		av, err := attributevalue.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, &types.AttributeValueMemberS{}, av)
	})
}

func TestShareSortKey(t *testing.T) {
	for str, key := range map[string]ShareSortKeyType{
		"DONORINVITE#123#abc": ShareSortKey(DonorInviteKey(OrganisationKey("123"), LpaKey("abc"))),
		"METADATA#123":        ShareSortKey(MetadataKey("123")),
	} {
		t.Run(str+"/SK", func(t *testing.T) {
			assert.Equal(t, str, key.SK())
		})

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

	t.Run("malformed", func(t *testing.T) {
		var v ShareSortKeyType
		err := json.Unmarshal([]byte(`"WHAT"`), &v)
		assert.Error(t, err)
	})

	t.Run("invalid", func(t *testing.T) {
		var v ShareSortKeyType
		err := json.Unmarshal([]byte(`"ATTORNEY#123"`), &v)
		assert.Error(t, err)
	})

	t.Run("empty json", func(t *testing.T) {
		var v ShareSortKeyType
		err := json.Unmarshal([]byte(`""`), &v)
		assert.Nil(t, err)
		assert.Equal(t, ShareSortKeyType{}, v)

		data, err := json.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, `""`, string(data))
	})

	t.Run("empty attributevalue", func(t *testing.T) {
		var v ShareSortKeyType
		err := attributevalue.Unmarshal(&types.AttributeValueMemberS{Value: ""}, &v)
		assert.Nil(t, err)
		assert.Equal(t, ShareSortKeyType{}, v)

		av, err := attributevalue.Marshal(v)
		assert.Nil(t, err)
		assert.Equal(t, &types.AttributeValueMemberS{}, av)
	})
}
