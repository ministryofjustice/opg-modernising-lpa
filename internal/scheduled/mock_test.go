package scheduled

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

var (
	testUuidString   = "a-uuid"
	testUuidStringFn = func() string { return testUuidString }
)

func (c *mockDynamoClient_AllByLpaUIDAndPartialSK_Call) SetData(data any) {
	c.Run(func(_ context.Context, _ string, _ dynamo.SK, v any) {
		b, _ := attributevalue.Marshal(data)
		attributevalue.Unmarshal(b, v)
	})
}

func (c *mockDynamoClient_AnyByPK_Call) SetData(row *Event) {
	c.Run(func(_ context.Context, _ dynamo.PK, v any) {
		b, _ := attributevalue.Marshal(row)
		attributevalue.Unmarshal(b, v)
	})
}
