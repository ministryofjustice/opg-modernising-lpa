package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

func (c *mockDynamodbClient_OneByUID_Call) SetData(data any) {
	c.Run(func(_ context.Context, _ string, v any) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}

func (c *mockDynamodbClient_One_Call) SetData(data any) {
	c.Run(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}
