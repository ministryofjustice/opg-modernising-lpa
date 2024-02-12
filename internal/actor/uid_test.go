package actor

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestUID(t *testing.T) {
	uid := UID{value: "abc"}

	assert.Equal(t, "abc", uid.String())
}

func TestUIDJSON(t *testing.T) {
	uid := UID{value: "abc"}

	jsonData, _ := json.Marshal(uid)
	assert.Equal(t, `"urn:opg:poas:makeregister:users:abc"`, string(jsonData))

	var a UID
	_ = json.Unmarshal([]byte(`"urn:opg:poas:makeregister:users:abc"`), &a)
	assert.Equal(t, a, uid)
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
