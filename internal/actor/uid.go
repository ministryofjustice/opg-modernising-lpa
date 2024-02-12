package actor

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const uidPrefix = "urn:opg:poas:makeregister:users:"

type UID struct{ value string }

func NewUID() UID {
	return UID{value: uuid.NewString()}
}

func UIDFromRequest(r interface{ FormValue(string) string }) UID {
	return UID{value: r.FormValue("id")}
}

func (u UID) IsZero() bool {
	return len(u.value) == 0
}

func (u UID) String() string {
	return u.value
}

func (u UID) PrefixedString() string {
	return uidPrefix + u.value
}

func (u UID) MarshalJSON() ([]byte, error) {
	if u.value == "" {
		return []byte("null"), nil
	}

	return []byte(`"` + u.PrefixedString() + `"`), nil
}

func (u *UID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	uid, found := strings.CutPrefix(string(text), uidPrefix)
	if !found {
		return errors.New("invalid uid prefix")
	}

	u.value = uid
	return nil
}

func (u UID) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return attributevalue.Marshal(u.value)
}

func (u *UID) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	return attributevalue.Unmarshal(av, &u.value)
}
