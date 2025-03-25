package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	types "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

func marshalListOfMaps[T any](vs []T) (result []map[string]types.AttributeValue) {
	for _, v := range vs {
		marshalled, _ := attributevalue.MarshalMap(v)
		result = append(result, marshalled)
	}

	return result
}

func (c *mockDynamodbClient_One_Call) SetData(data any) {
	c.Run(func(_ context.Context, _ dynamo.PK, _ dynamo.SK, v any) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}

func (c *mockDynamodbClient_OneByPartialSK_Call) SetData(data any) {
	c.Run(func(_ context.Context, _ dynamo.PK, _ dynamo.SK, v any) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}
