package reuse

import (
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type orString[T any] struct {
	v T
	s string
}

func (c *orString[T]) UnmarshalDynamoDBAttributeValue(av types.AttributeValue) error {
	if err := attributevalue.Unmarshal(av, &c.v); err != nil {
		if err := attributevalue.Unmarshal(av, &c.s); err != nil {
			return err
		}
	}

	return nil
}
