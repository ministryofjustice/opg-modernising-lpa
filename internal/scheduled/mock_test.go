package scheduled

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	types "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

var (
	testUuidString   = "a-uuid"
	testUuidStringFn = func() string { return testUuidString }
)

func marshalListOfMaps[T any](vs []T) (result []map[string]types.AttributeValue) {
	for _, v := range vs {
		marshalled, _ := attributevalue.MarshalMap(v)
		result = append(result, marshalled)
	}

	return result
}

func (c *mockDynamoClient_OneByPK_Call) SetData(row *Event) {
	c.Run(func(_ context.Context, _ dynamo.PK, v any) {
		b, _ := attributevalue.Marshal(row)
		attributevalue.Unmarshal(b, v)
	})
}
