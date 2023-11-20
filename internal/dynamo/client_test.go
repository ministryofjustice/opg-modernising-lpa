package dynamo

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/smithy-go"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func TestOne(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("GetItem", ctx, &dynamodb.GetItemInput{
			TableName: aws.String("this"),
			Key:       map[string]types.AttributeValue{"PK": pkey, "SK": skey},
		}).
		Return(&dynamodb.GetItemOutput{Item: data}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var actual map[string]string
	err := c.One(ctx, "a-pk", "a-sk", &actual)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestOneWhenError(t *testing.T) {
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("GetItem", ctx, &dynamodb.GetItemInput{
			TableName: aws.String("this"),
			Key:       map[string]types.AttributeValue{"PK": pkey, "SK": skey},
		}).
		Return(&dynamodb.GetItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.One(ctx, "a-pk", "a-sk", &v)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", v)
}

func TestOneWhenNotFound(t *testing.T) {
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("GetItem", ctx, &dynamodb.GetItemInput{
			TableName: aws.String("this"),
			Key:       map[string]types.AttributeValue{"PK": pkey, "SK": skey},
		}).
		Return(&dynamodb.GetItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.One(ctx, "a-pk", "a-sk", &v)
	assert.Equal(t, NotFoundError{}, err)
	assert.Equal(t, "", v)
}

func TestOneByUID(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(uidIndex),
			ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#UID = :UID"),
		}).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{{
				"PK":     &types.AttributeValueMemberS{Value: "LPA#123"},
				"LpaUID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"},
			}},
		}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v actor.DonorProvidedDetails
	err := c.OneByUID(ctx, "M-1111-2222-3333", &v)

	assert.Nil(t, err)
	assert.Equal(t, actor.DonorProvidedDetails{PK: "LPA#123", LpaUID: "M-1111-2222-3333"}, v)
}

func TestOneByUIDWhenQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(uidIndex),
			ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#UID = :UID"),
		}).
		Return(&dynamodb.QueryOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.OneByUID(ctx, "M-1111-2222-3333", mock.Anything)

	assert.Equal(t, fmt.Errorf("failed to query UID: %w", expectedError), err)
}

func TestOneByUIDWhenNot1Item(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(uidIndex),
			ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#UID = :UID"),
		}).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"PK":  &types.AttributeValueMemberS{Value: "LPA#123"},
					"UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"},
				},
				{
					"PK":  &types.AttributeValueMemberS{Value: "LPA#123"},
					"UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"},
				},
			},
		}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.OneByUID(ctx, "M-1111-2222-3333", mock.Anything)

	assert.Equal(t, errors.New("expected to resolve UID but got 2 items"), err)
}

func TestOneByUIDWhenUnmarshalError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(uidIndex),
			ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#UID = :UID"),
		}).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{
				{
					"PK":  &types.AttributeValueMemberS{Value: "LPA#123"},
					"UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"},
				},
			},
		}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.OneByUID(ctx, "M-1111-2222-3333", "not an lpa")

	assert.IsType(t, &attributevalue.InvalidUnmarshalError{}, err)
}

func TestOneByPartialSk(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-partial-sk")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK", "#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey, ":SK": skey},
			KeyConditionExpression:    aws.String("#PK = :PK and begins_with(#SK, :SK)"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Nil(t, err)
	assert.Equal(t, expected, v)
}

func TestOneByPartialSkOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Equal(t, expectedError, err)
}

func TestOneByPartialSkWhenNotFound(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Equal(t, NotFoundError{}, err)
}

func TestOneByPartialSkWhenMultipleResults(t *testing.T) {
	ctx := context.Background()

	data, _ := attributevalue.MarshalMap(map[string]string{"Col": "Val"})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data, data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.OneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Equal(t, MultipleResultsError{}, err)
}

func TestAllByPartialSk(t *testing.T) {
	ctx := context.Background()

	expected := []map[string]string{{"Col": "Val"}, {"Other": "Thing"}}
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-partial-sk")
	data, _ := attributevalue.MarshalMap(expected[0])
	data2, _ := attributevalue.MarshalMap(expected[1])

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK", "#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey, ":SK": skey},
			KeyConditionExpression:    aws.String("#PK = :PK and begins_with(#SK, :SK)"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data, data2}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v []map[string]string
	err := c.AllByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Nil(t, err)
	assert.Equal(t, expected, v)
}

func TestAllByPartialSkOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.AllByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Equal(t, expectedError, err)
}

func TestAllForActor(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	skey, _ := attributevalue.Marshal("a-partial-sk")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(actorUpdatedAtIndex),
			ExpressionAttributeNames:  map[string]string{"#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":SK": skey},
			KeyConditionExpression:    aws.String("#SK = :SK"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data, data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v []map[string]string
	err := c.AllForActor(ctx, "a-partial-sk", &v)
	assert.Nil(t, err)
	assert.Equal(t, []map[string]string{expected, expected}, v)
}

func TestAllForActorWhenNotFound(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.AllForActor(ctx, "a-partial-sk", &v)
	assert.Nil(t, err)
	assert.Empty(t, v)
}

func TestAllForActorOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.AllForActor(ctx, "a-partial-sk", &v)
	assert.Equal(t, expectedError, err)
}

func TestLatestForActor(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	skey, _ := attributevalue.Marshal("a-partial-sk")
	updated, _ := attributevalue.Marshal("2")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String(actorUpdatedAtIndex),
			ExpressionAttributeNames:  map[string]string{"#SK": "SK", "#UpdatedAt": "UpdatedAt"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":SK": skey, ":UpdatedAt": updated},
			KeyConditionExpression:    aws.String("#SK = :SK and #UpdatedAt > :UpdatedAt"),
			ScanIndexForward:          aws.Bool(false),
			Limit:                     aws.Int32(1),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.LatestForActor(ctx, "a-partial-sk", &v)
	assert.Nil(t, err)
	assert.Equal(t, expected, v)
}

func TestLatestForActorWhenNotFound(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v interface{}
	err := c.LatestForActor(ctx, "a-partial-sk", &v)
	assert.Nil(t, err)
	assert.Nil(t, v)
}

func TestLatestForActorOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.LatestForActor(ctx, "a-partial-sk", &v)
	assert.Equal(t, expectedError, err)
}

func TestAllKeysByPk(t *testing.T) {
	ctx := context.Background()

	keys := []Key{
		{PK: "pk", SK: "sk1"},
		{PK: "pk", SK: "sk2"},
	}

	item1, _ := attributevalue.MarshalMap(keys[0])
	item2, _ := attributevalue.MarshalMap(keys[1])

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                aws.String("this"),
			ExpressionAttributeNames: map[string]string{"#PK": "PK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":PK": &types.AttributeValueMemberS{Value: "pk"},
			},
			KeyConditionExpression: aws.String("#PK = :PK"),
			ProjectionExpression:   aws.String("PK, SK"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{item1, item2}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	result, err := c.AllKeysByPk(ctx, "pk")
	assert.Nil(t, err)
	assert.Equal(t, keys, result)
}

func TestAllKeysByPkWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	_, err := c.AllKeysByPk(ctx, "pk")
	assert.Equal(t, expectedError, err)
}

func TestAllByKeys(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("BatchGetItem", ctx, &dynamodb.BatchGetItemInput{
			RequestItems: map[string]types.KeysAndAttributes{
				"this": {
					Keys: []map[string]types.AttributeValue{{
						"PK": &types.AttributeValueMemberS{Value: "pk"},
						"SK": &types.AttributeValueMemberS{Value: "sk"},
					}},
				},
			},
		}).
		Return(&dynamodb.BatchGetItemOutput{
			Responses: map[string][]map[string]types.AttributeValue{
				"this": {data},
			},
		}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	v, err := c.AllByKeys(ctx, []Key{{PK: "pk", SK: "sk"}})
	assert.Nil(t, err)
	assert.Equal(t, []map[string]types.AttributeValue{data}, v)
}

func TestAllByKeysWhenQueryErrors(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("BatchGetItem", ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	_, err := c.AllByKeys(ctx, []Key{{PK: "pk", SK: "sk"}})
	assert.Equal(t, expectedError, err)
}

func TestPut(t *testing.T) {
	testCases := map[string]map[string]string{
		"Without UpdatedAt": {"Col": "Val"},
		"Zero UpdatedAt":    {"Col": "Val", "UpdatedAt": "0001-01-01T00:00:00Z"},
	}

	for name, dataMap := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			data, _ := attributevalue.MarshalMap(dataMap)

			dynamoDB := newMockDynamoDB(t)
			dynamoDB.
				On("PutItem", ctx, &dynamodb.PutItemInput{
					TableName: aws.String("this"),
					Item:      data,
				}).
				Return(&dynamodb.PutItemOutput{}, nil)

			c := &Client{table: "this", svc: dynamoDB}

			err := c.Put(ctx, dataMap)
			assert.Nil(t, err)
		})
	}
}

func TestPutWhenStructHasVersion(t *testing.T) {
	ctx := context.Background()
	data, _ := attributevalue.MarshalMap(map[string]any{"Col": "Val", "Version": 2})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("PutItem", ctx, &dynamodb.PutItemInput{
			TableName:                 aws.String("this"),
			Item:                      data,
			ConditionExpression:       aws.String("Version = :version"),
			ExpressionAttributeValues: map[string]types.AttributeValue{":version": &types.AttributeValueMemberN{Value: "1"}},
		}).
		Return(&dynamodb.PutItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, map[string]any{"Col": "Val", "Version": 1})
	assert.Nil(t, err)
}

func TestPutWhenConditionalCheckFailedException(t *testing.T) {
	ctx := context.Background()
	data, _ := attributevalue.MarshalMap(map[string]any{"Col": "Val", "Version": 2})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("PutItem", ctx, &dynamodb.PutItemInput{
			TableName:                 aws.String("this"),
			Item:                      data,
			ConditionExpression:       aws.String("Version = :version"),
			ExpressionAttributeValues: map[string]types.AttributeValue{":version": &types.AttributeValueMemberN{Value: "1"}},
		}).
		Return(&dynamodb.PutItemOutput{}, &smithy.OperationError{Err: &types.ConditionalCheckFailedException{}})

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, map[string]any{"Col": "Val", "Version": 1})
	assert.Equal(t, ConditionalCheckFailedError{}, err)
}

func TestPutWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("PutItem", ctx, mock.Anything).
		Return(&dynamodb.PutItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, "hello")
	assert.Equal(t, expectedError, err)
}

func TestPutWhenUnmarshalError(t *testing.T) {
	ctx := context.Background()

	c := &Client{table: "this", svc: newMockDynamoDB(t)}

	err := c.Put(ctx, map[string]string{"Col": "Val", "Version": "not an int"})
	assert.NotNil(t, err)
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	data, _ := attributevalue.MarshalMap(map[string]string{"Col": "Val"})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("PutItem", ctx, &dynamodb.PutItemInput{
			TableName:           aws.String("this"),
			Item:                data,
			ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
		}).
		Return(&dynamodb.PutItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Create(ctx, map[string]string{"Col": "Val"})
	assert.Nil(t, err)
}

func TestCreateWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("PutItem", ctx, mock.Anything).
		Return(&dynamodb.PutItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Create(ctx, map[string]string{"Col": "Val"})
	assert.Equal(t, expectedError, err)
}

func TestDeleteKeys(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("TransactWriteItems", ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{
				{
					Delete: &types.Delete{
						TableName: aws.String("this"),
						Key: map[string]types.AttributeValue{
							"PK": &types.AttributeValueMemberS{Value: "pk"},
							"SK": &types.AttributeValueMemberS{Value: "sk1"},
						},
					},
				},
				{
					Delete: &types.Delete{
						TableName: aws.String("this"),
						Key: map[string]types.AttributeValue{
							"PK": &types.AttributeValueMemberS{Value: "pk"},
							"SK": &types.AttributeValueMemberS{Value: "sk2"},
						},
					},
				},
			},
		}).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.DeleteKeys(ctx, []Key{{PK: "pk", SK: "sk1"}, {PK: "pk", SK: "sk2"}})
	assert.Equal(t, expectedError, err)
}

func TestUpdate(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("UpdateItem", ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String("table-name"),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "a-pk"},
				"SK": &types.AttributeValueMemberS{Value: "a-sk"},
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{"prop": &types.AttributeValueMemberS{Value: "val"}},
			UpdateExpression:          aws.String("some = expression"),
		}).
		Return(nil, nil)

	c := &Client{table: "table-name", svc: dynamoDB}

	err := c.Update(ctx, "a-pk", "a-sk", map[string]types.AttributeValue{"prop": &types.AttributeValueMemberS{Value: "val"}}, "some = expression")

	assert.Nil(t, err)
}

func TestUpdateOnServiceError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("UpdateItem", ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String("table-name"),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "a-pk"},
				"SK": &types.AttributeValueMemberS{Value: "a-sk"},
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{"Col": &types.AttributeValueMemberS{Value: "Val"}},
			UpdateExpression:          aws.String("some = expression"),
		}).
		Return(nil, expectedError)

	c := &Client{table: "table-name", svc: dynamoDB}

	err := c.Update(ctx, "a-pk", "a-sk", map[string]types.AttributeValue{"Col": &types.AttributeValueMemberS{Value: "Val"}}, "some = expression")

	assert.Equal(t, expectedError, err)
}

func TestBatchPutOneBatch(t *testing.T) {
	ctx := context.Background()

	values := []any{map[string]string{"a": "b"}, map[string]string{"x": "y"}}
	itemA, _ := attributevalue.MarshalMap(values[0])
	itemB, _ := attributevalue.MarshalMap(values[1])

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("TransactWriteItems", ctx, &dynamodb.TransactWriteItemsInput{
			TransactItems: []types.TransactWriteItem{
				{
					Put: &types.Put{
						TableName: aws.String("table-name"),
						Item:      itemA,
					},
				},
				{
					Put: &types.Put{
						TableName: aws.String("table-name"),
						Item:      itemB,
					},
				},
			},
		}).
		Return(nil, expectedError)

	c := &Client{table: "table-name", svc: dynamoDB}
	err := c.BatchPut(ctx, values)

	assert.Equal(t, expectedError, err)
}
