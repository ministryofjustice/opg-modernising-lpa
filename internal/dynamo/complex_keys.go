package dynamo

import (
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type Keys struct {
	PK PK
	SK SK
}

func (k *Keys) read(pkStr, skStr string) error {
	v, err := readKey(pkStr)
	if err != nil {
		return err
	}

	var ok bool
	k.PK, ok = v.(PK)
	if !ok {
		return errors.New("newKeys pk not pk")
	}

	v, err = readKey(skStr)
	if err != nil {
		return err
	}

	k.SK, ok = v.(SK)
	if !ok {
		return errors.New("newKeys sk not sk")
	}

	return nil
}

func (k *Keys) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var skeys struct{ PK, SK string }
	if err := attributevalue.Unmarshal(av, &skeys); err != nil {
		return err
	}

	return k.read(skeys.PK, skeys.SK)
}

func (k *Keys) UnmarshalJSON(text []byte) error {
	var skeys struct{ PK, SK string }
	if err := json.Unmarshal(text, &skeys); err != nil {
		return err
	}

	return k.read(skeys.PK, skeys.SK)
}

type LpaOwnerKeyType struct{ sk SK }

// LpaOwnerKey is used as the SK (with LpaKey as PK) to allow both donors and
// organisations to "own" LPAs.
func LpaOwnerKey(sk interface {
	SK
	lpaOwner()
}) LpaOwnerKeyType {
	return LpaOwnerKeyType{sk: sk}
}

func (k LpaOwnerKeyType) MarshalText() ([]byte, error) {
	if k.sk == nil {
		return []byte(nil), nil
	}

	return []byte(k.sk.SK()), nil
}

func (k *LpaOwnerKeyType) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	v, err := readKey(string(text))
	if err != nil {
		return err
	}

	sk, ok := v.(interface {
		SK
		lpaOwner()
	})
	if !ok {
		return errors.New("invalid key")
	}

	k.sk = sk
	return err
}

func (k LpaOwnerKeyType) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	text, _ := k.MarshalText()

	return attributevalue.Marshal(string(text))
}

func (k *LpaOwnerKeyType) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	if err := attributevalue.Unmarshal(av, &s); err != nil {
		return err
	}

	return k.UnmarshalText([]byte(s))
}

func (k LpaOwnerKeyType) Equals(sk SK) bool {
	if k.sk == nil || sk == nil {
		return false
	}

	return k.SK() == sk.SK()
}

func (k LpaOwnerKeyType) SK() string {
	if k.sk == nil {
		return ""
	}

	return k.sk.SK()
}

func (k LpaOwnerKeyType) Organisation() (OrganisationKeyType, bool) {
	v, ok := k.sk.(OrganisationKeyType)
	return v, ok
}

type ShareKeyType struct{ pk PK }

// ShareKey is used as the PK (with ShareSortKey as SK) for sharing an LPA with
// another actor.
func ShareKey(pk interface {
	PK
	share()
}) ShareKeyType {
	return ShareKeyType{pk: pk}
}

func (k ShareKeyType) MarshalText() ([]byte, error) {
	if k.pk == nil {
		return []byte(nil), nil
	}

	return []byte(k.pk.PK()), nil
}

func (k *ShareKeyType) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	v, err := readKey(string(text))
	if err != nil {
		return err
	}

	pk, ok := v.(interface {
		PK
		share()
	})
	if !ok {
		return errors.New("invalid key")
	}

	k.pk = pk
	return err
}

func (k ShareKeyType) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	text, _ := k.MarshalText()

	return attributevalue.Marshal(string(text))
}

func (k *ShareKeyType) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	if err := attributevalue.Unmarshal(av, &s); err != nil {
		return err
	}

	return k.UnmarshalText([]byte(s))
}

func (k ShareKeyType) PK() string {
	return k.pk.PK()
}

type ShareSortKeyType struct{ sk SK }

// ShareSortKey is used as the SK (with ShareKey as the PK) for sharing an LPA
// with another actor.
func ShareSortKey(sk interface {
	SK
	shareSort()
}) ShareSortKeyType {
	return ShareSortKeyType{sk: sk}
}

func (k ShareSortKeyType) MarshalText() ([]byte, error) {
	if k.sk == nil {
		return []byte(nil), nil
	}

	return []byte(k.sk.SK()), nil
}

func (k *ShareSortKeyType) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return nil
	}

	v, err := readKey(string(text))
	if err != nil {
		return err
	}

	sk, ok := v.(interface {
		SK
		shareSort()
	})
	if !ok {
		return errors.New("invalid key")
	}

	k.sk = sk
	return nil
}

func (k ShareSortKeyType) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	text, _ := k.MarshalText()

	return attributevalue.Marshal(string(text))
}

func (k *ShareSortKeyType) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	if err := attributevalue.Unmarshal(av, &s); err != nil {
		return err
	}

	return k.UnmarshalText([]byte(s))
}

func (k ShareSortKeyType) SK() string {
	return k.sk.SK()
}
