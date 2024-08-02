// Package actoruid provides an identifier for the various actors on an LPA.
package actoruid

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const prefix = "urn:opg:poas:makeregister:users:"

// Service is used when the application does an action.
var Service = UID{value: [16]byte{0, 0, 0, 0, 0, 0, 64}}

type UID struct{ value uuid.UUID }

func New() UID {
	return UID{value: uuid.New()}
}

func Parse(uid string) (UID, error) {
	u, err := uuid.Parse(uid)
	if err != nil {
		return UID{}, err
	}

	return UID{value: u}, nil
}

func FromRequest(r interface{ FormValue(string) string }) UID {
	uid, _ := Parse(r.FormValue("id"))
	return uid
}

func (u UID) IsZero() bool {
	return u == UID{}
}

func (u UID) String() string {
	return u.value.String()
}

func (u UID) PrefixedString() string {
	return prefix + u.String()
}

func (u UID) MarshalJSON() ([]byte, error) {
	if u.IsZero() {
		return []byte("null"), nil
	}

	return []byte(`"` + u.String() + `"`), nil
}

func (u *UID) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	uid, err := Parse(string(text))
	if err != nil {
		return err
	}

	*u = uid
	return nil
}

func (u UID) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return attributevalue.Marshal(u.value.String())
}

func (u *UID) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	if err := attributevalue.Unmarshal(av, &s); err != nil {
		return err
	}

	return u.UnmarshalText([]byte(s))
}
