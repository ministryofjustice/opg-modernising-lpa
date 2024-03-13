package actoruid

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
)

const makeRegisterPrefix = "urn:opg:poas:makeregister:users:"

type UID struct{ value string }

func New() UID {
	return UID{value: uuid.NewString()}
}

func FromRequest(r interface{ FormValue(string) string }) UID {
	return UID{value: r.FormValue("id")}
}

func FromPrefixedString(s string) (UID, error) {
	regex, err := regexp.Compile("urn:opg:poas:.*:users:")
	if err != nil {
		return UID{}, err
	}

	split := regex.Split(s, 2)
	if split[0] == s {
		return UID{}, fmt.Errorf("unexpected UID prefix: %s", s)
	}

	return UID{value: split[1]}, nil
}

func (u UID) IsZero() bool {
	return len(u.value) == 0
}

func (u UID) String() string {
	return u.value
}

func (u UID) PrefixedString() string {
	return makeRegisterPrefix + u.value
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

	uid, found := strings.CutPrefix(string(text), makeRegisterPrefix)
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
