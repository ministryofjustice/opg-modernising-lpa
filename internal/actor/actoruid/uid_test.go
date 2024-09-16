package actoruid

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

func TestUID(t *testing.T) {
	uuidString := "2ea1a849-975e-481c-af19-1209d20ed362"
	uid, err := Parse(uuidString)

	assert.Nil(t, err)
	assert.Equal(t, uuidString, uid.String())
	assert.Equal(t, prefix+uuidString, uid.PrefixedString())

	_, err = Parse("what")
	assert.Error(t, err)
}

func TestUIDFromRequest(t *testing.T) {
	uuidString := "2ea1a849-975e-481c-af19-1209d20ed362"
	r, _ := http.NewRequest(http.MethodGet, "/?id="+uuidString, nil)

	assert.Equal(t, uuidString, FromRequest(r).String())
}

func TestUIDZero(t *testing.T) {
	assert.True(t, UID{}.IsZero())
	assert.False(t, New().IsZero())
}

func TestUIDJSON(t *testing.T) {
	uuidString := "2ea1a849-975e-481c-af19-1209d20ed362"
	uid, err := Parse(uuidString)
	assert.Nil(t, err)

	jsonData, _ := json.Marshal(uid)
	assert.Equal(t, `"`+uuidString+`"`, string(jsonData))

	var a UID
	err = json.Unmarshal([]byte(`"`+uuidString+`"`), &a)
	assert.Nil(t, err)
	assert.Equal(t, a, uid)

	emptyData, _ := json.Marshal(UID{})
	assert.Equal(t, `null`, string(emptyData))

	var b UID
	err = json.Unmarshal([]byte(`null`), &b)
	assert.Nil(t, err)
	assert.True(t, b.IsZero())

	var c UID
	err = json.Unmarshal([]byte(`""`), &c)
	assert.Nil(t, err)
	assert.True(t, c.IsZero())

	var d UID
	err = json.Unmarshal([]byte(`"abc"`), &d)
	assert.Error(t, err)
	assert.True(t, c.IsZero())
}

func TestUIDAttributeValue(t *testing.T) {
	uuidString := "2ea1a849-975e-481c-af19-1209d20ed362"
	uid, err := Parse(uuidString)
	assert.Nil(t, err)

	avData, _ := attributevalue.Marshal(uid)
	assert.Equal(t, &types.AttributeValueMemberS{Value: uuidString}, avData)

	var b UID
	err = attributevalue.Unmarshal(&types.AttributeValueMemberS{Value: uuidString}, &b)
	assert.Nil(t, err)
	assert.Equal(t, b, uid)

	var c UID
	err = attributevalue.Unmarshal(&types.AttributeValueMemberS{Value: "abc"}, &c)
	assert.Error(t, err)
	assert.True(t, c.IsZero())
}

func TestPrefixedJSON(t *testing.T) {
	uuidString := "2ea1a849-975e-481c-af19-1209d20ed362"
	uid, err := Parse(uuidString)
	assert.Nil(t, err)

	jsonData, _ := json.Marshal(Prefixed(uid))
	assert.Equal(t, `"`+prefix+uuidString+`"`, string(jsonData))

	emptyData, _ := json.Marshal(Prefixed(UID{}))
	assert.Equal(t, `null`, string(emptyData))
}
