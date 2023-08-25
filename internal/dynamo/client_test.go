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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func TestGet(t *testing.T) {
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
	err := c.Get(ctx, "a-pk", "a-sk", &actual)
	assert.Nil(t, err)
	assert.Equal(t, expected, actual)
}

func TestGetWhenError(t *testing.T) {
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
	err := c.Get(ctx, "a-pk", "a-sk", &v)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", v)
}

func TestGetWhenNotFound(t *testing.T) {
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
	err := c.Get(ctx, "a-pk", "a-sk", &v)
	assert.Equal(t, NotFoundError{}, err)
	assert.Equal(t, "", v)
}

func TestPut(t *testing.T) {
	ctx := context.Background()
	data, _ := attributevalue.MarshalMap(map[string]string{"Col": "Val"})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("PutItem", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("this"),
			Item:      data,
		}).
		Return(&dynamodb.PutItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, map[string]string{"Col": "Val"})
	assert.Nil(t, err)
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

func TestGetOneByPartialSk(t *testing.T) {
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
	err := c.GetOneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Nil(t, err)
	assert.Equal(t, expected, v)
}

func TestGetOneByPartialSkOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.GetOneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Equal(t, expectedError, err)
}

func TestGetOneByPartialSkWhenNotFound(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.GetOneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Equal(t, NotFoundError{}, err)
}

func TestGetOneByPartialSkWhenMultipleResults(t *testing.T) {
	ctx := context.Background()

	data, _ := attributevalue.MarshalMap(map[string]string{"Col": "Val"})

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data, data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v map[string]string
	err := c.GetOneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Equal(t, MultipleResultsError{}, err)
}

func TestGetAllByGsi(t *testing.T) {
	ctx := context.Background()

	expected := map[string]string{"Col": "Val"}
	skey, _ := attributevalue.Marshal("a-partial-sk")
	data, _ := attributevalue.MarshalMap(expected)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String("index-name"),
			ExpressionAttributeNames:  map[string]string{"#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":SK": skey},
			KeyConditionExpression:    aws.String("#SK = :SK"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{data, data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v []map[string]string
	err := c.GetAllByGsi(ctx, "index-name", "a-partial-sk", &v)
	assert.Nil(t, err)
	assert.Equal(t, []map[string]string{expected, expected}, v)
}

func TestGetAllByGsiWhenNotFound(t *testing.T) {
	ctx := context.Background()
	skey, _ := attributevalue.Marshal("a-partial-sk")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String("index-name"),
			ExpressionAttributeNames:  map[string]string{"#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":SK": skey},
			KeyConditionExpression:    aws.String("#SK = :SK"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.GetAllByGsi(ctx, "index-name", "a-partial-sk", &v)
	assert.Nil(t, err)
}

func TestGetAllByGsiOnQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.GetAllByGsi(ctx, "index-name", "a-partial-sk", &v)
	assert.Equal(t, expectedError, err)
}

func TestGetAllByKeys(t *testing.T) {
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

	v, err := c.GetAllByKeys(ctx, []Key{{PK: "pk", SK: "sk"}})
	assert.Nil(t, err)
	assert.Equal(t, []map[string]types.AttributeValue{data}, v)
}

func TestGetAllByKeysWhenQueryErrors(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("BatchGetItem", ctx, mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	_, err := c.GetAllByKeys(ctx, []Key{{PK: "pk", SK: "sk"}})
	assert.Equal(t, expectedError, err)
}

func TestGetOneByUID(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String("UidIndex"),
			ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#UID = :UID"),
		}).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{{
				"PK":  &types.AttributeValueMemberS{Value: "LPA#123"},
				"UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"},
			}},
		}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v page.Lpa
	err := c.GetOneByUID(ctx, "M-1111-2222-3333", &v)

	assert.Nil(t, err)
	assert.Equal(t, page.Lpa{PK: "LPA#123", UID: "M-1111-2222-3333"}, v)
}

func TestGetOneByUIDWhenQueryError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String("UidIndex"),
			ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#UID = :UID"),
		}).
		Return(&dynamodb.QueryOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.GetOneByUID(ctx, "M-1111-2222-3333", mock.Anything)

	assert.Equal(t, fmt.Errorf("failed to query UID: %w", expectedError), err)
}

func TestGetOneByUIDWhenNot1Item(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			IndexName:                 aws.String("UidIndex"),
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

	err := c.GetOneByUID(ctx, "M-1111-2222-3333", mock.Anything)

	assert.Equal(t, errors.New("expected to resolve UID but got 2 items"), err)
}
