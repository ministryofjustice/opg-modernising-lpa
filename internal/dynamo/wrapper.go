package dynamo

import (
	"errors"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// A PKType can be used to represent a PK that needs serialising generically.
type PKType struct{ pk PK }

// WrapPK returns a type suitable for representing any PK.
func WrapPK(pk PK) PKType {
	return PKType{pk: pk}
}

// Unwrap returns the underlying PK.
func (p PKType) Unwrap() PK {
	return p.pk
}

func (p PKType) MarshalText() ([]byte, error) {
	return []byte(p.pk.PK()), nil
}

func (p *PKType) UnmarshalText(text []byte) error {
	key, err := readKey(string(text))
	if err != nil {
		return err
	}

	pk, ok := key.(PK)
	if !ok {
		return errors.New("PKType unmarshalled to non-PK")
	}

	p.pk = pk
	return nil
}

func (p PKType) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return attributevalue.Marshal(p.pk.PK())
}

func (p *PKType) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var s string
	if err := attributevalue.Unmarshal(av, &s); err != nil {
		return err
	}

	return p.UnmarshalText([]byte(s))
}

// An SKType can be used to represent an SK that needs serialising generically.
type SKType struct{ sk SK }

// WrapSK returns a type suitable for representing any SK.
func WrapSK(sk SK) SKType {
	return SKType{sk: sk}
}

// SK returns the underlying SK.
func (s SKType) Unwrap() SK {
	return s.sk
}

func (s SKType) MarshalText() ([]byte, error) {
	return []byte(s.sk.SK()), nil
}

func (s *SKType) UnmarshalText(text []byte) error {
	key, err := readKey(string(text))
	if err != nil {
		return err
	}

	sk, ok := key.(SK)
	if !ok {
		return errors.New("SKType unmarshalled to non-SK")
	}

	s.sk = sk
	return nil
}

func (s SKType) MarshalDynamoDBAttributeValue() (types.AttributeValue, error) {
	return attributevalue.Marshal(s.sk.SK())
}

func (s *SKType) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	var str string
	if err := attributevalue.Unmarshal(av, &str); err != nil {
		return err
	}

	return s.UnmarshalText([]byte(str))
}
