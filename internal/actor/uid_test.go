package actor

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestUID(t *testing.T) {
	uid := UID{value: "abc"}

	assert.Equal(t, "abc", uid.String())
	assert.Equal(t, uidPrefix+"abc", uid.PrefixedString())
}

func TestUIDFromRequest(t *testing.T) {
	r, _ := http.NewRequest(http.MethodGet, "/?id=abc", nil)

	assert.Equal(t, UID{value: "abc"}, UIDFromRequest(r))
}

func TestUIDZero(t *testing.T) {
	assert.True(t, UID{}.IsZero())
	assert.False(t, NewUID().IsZero())
}

func TestUIDJSON(t *testing.T) {
	uid := UID{value: "abc"}

	jsonData, _ := json.Marshal(uid)
	assert.Equal(t, `"urn:opg:poas:makeregister:users:abc"`, string(jsonData))

	var a UID
	err := json.Unmarshal([]byte(`"urn:opg:poas:makeregister:users:abc"`), &a)
	assert.Nil(t, err)
	assert.Equal(t, a, uid)

	emptyData, _ := json.Marshal(UID{})
	assert.Equal(t, `null`, string(emptyData))

	var b UID
	err = json.Unmarshal([]byte(`null`), &b)
	assert.Nil(t, err)
	assert.True(t, b.IsZero())
}

func TestUIDJSONInvalidPrefix(t *testing.T) {
	var v UID
	err := json.Unmarshal([]byte(`"urn:opg:poas:makeregister:users2:abc"`), &v)
	assert.ErrorContains(t, err, "invalid uid prefix")
}

func TestUIDAttributeValue(t *testing.T) {
	uid := UID{value: "abc"}

	avData, _ := attributevalue.Marshal(uid)
	assert.Equal(t, &types.AttributeValueMemberS{Value: "abc"}, avData)

	var b UID
	_ = attributevalue.Unmarshal(&types.AttributeValueMemberS{Value: "abc"}, &b)
	assert.Equal(t, b, uid)
}
