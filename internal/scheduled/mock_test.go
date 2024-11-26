package scheduled

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
)

func (c *mockDynamoClient_AllScheduledEventsByUID_Call) SetData(data any) {
	c.Run(func(_ context.Context, _ string, v any) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}
