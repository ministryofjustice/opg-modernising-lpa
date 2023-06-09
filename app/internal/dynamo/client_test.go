package dynamo

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

var expectedError = errors.New("err")

func TestGet(t *testing.T) {
	ctx := context.Background()

	result := "hello"
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")
	data, _ := attributevalue.Marshal(result)

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("GetItem", ctx, &dynamodb.GetItemInput{
			TableName: aws.String("this"),
			Key:       map[string]types.AttributeValue{"PK": pkey, "SK": skey},
		}).
		Return(&dynamodb.GetItemOutput{Item: map[string]types.AttributeValue{"Data": data}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.Get(ctx, "a-pk", "a-sk", &v)
	assert.Nil(t, err)
	assert.Equal(t, result, v)
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
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")
	data, _ := attributevalue.Marshal("hello")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("PutItem", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("this"),
			Item: map[string]types.AttributeValue{
				"PK":   pkey,
				"SK":   skey,
				"Data": data,
			},
		}).
		Return(&dynamodb.PutItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, "a-pk", "a-sk", "hello")
	assert.Nil(t, err)
}

func TestPutWhenError(t *testing.T) {
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")
	data, _ := attributevalue.Marshal("hello")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("PutItem", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("this"),
			Item: map[string]types.AttributeValue{
				"PK":   pkey,
				"SK":   skey,
				"Data": data,
			},
		}).
		Return(&dynamodb.PutItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, "a-pk", "a-sk", "hello")
	assert.Equal(t, expectedError, err)
}

func TestCreate(t *testing.T) {
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")
	data, _ := attributevalue.Marshal("hello")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("PutItem", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("this"),
			Item: map[string]types.AttributeValue{
				"PK":   pkey,
				"SK":   skey,
				"Data": data,
			},
			ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
		}).
		Return(&dynamodb.PutItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Create(ctx, "a-pk", "a-sk", "hello")
	assert.Nil(t, err)
}

func TestCreateWhenError(t *testing.T) {
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-sk")
	data, _ := attributevalue.Marshal("hello")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("PutItem", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("this"),
			Item: map[string]types.AttributeValue{
				"PK":   pkey,
				"SK":   skey,
				"Data": data,
			},
			ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
		}).
		Return(&dynamodb.PutItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Create(ctx, "a-pk", "a-sk", "hello")
	assert.Equal(t, expectedError, err)
}

func TestGetOneByPartialSk(t *testing.T) {
	result, _ := attributevalue.Marshal("some data")
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-partial-sk")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK", "#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey, ":SK": skey},
			KeyConditionExpression:    aws.String("#PK = :PK and begins_with(#SK, :SK)"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{{"Data": result}}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.GetOneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Nil(t, err)
	assert.Equal(t, "some data", v)
}

func TestGetOneByPartialSkOnQueryError(t *testing.T) {
	result, _ := attributevalue.Marshal("some data")
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-partial-sk")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK", "#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey, ":SK": skey},
			KeyConditionExpression:    aws.String("#PK = :PK and begins_with(#SK, :SK)"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{{"Data": result}}}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.GetOneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Equal(t, expectedError, err)
}

func TestGetOneByPartialSkWhenNotFound(t *testing.T) {
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-partial-sk")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK", "#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey, ":SK": skey},
			KeyConditionExpression:    aws.String("#PK = :PK and begins_with(#SK, :SK)"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.GetOneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Equal(t, NotFoundError{}, err)
}

func TestGetOneByPartialSkWhenMultipleResults(t *testing.T) {
	result, _ := attributevalue.Marshal("some data")
	ctx := context.Background()
	pkey, _ := attributevalue.Marshal("a-pk")
	skey, _ := attributevalue.Marshal("a-partial-sk")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK", "#SK": "SK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey, ":SK": skey},
			KeyConditionExpression:    aws.String("#PK = :PK and begins_with(#SK, :SK)"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{{"Data": result}, {"Data": result}}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.GetOneByPartialSk(ctx, "a-pk", "a-partial-sk", &v)
	assert.Equal(t, MultipleResultsError{}, err)
}

func TestGetAllByGsi(t *testing.T) {
	result, _ := attributevalue.Marshal("some data")
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
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{{"Data": result}, {"Data": result}}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v []struct {
		Data string
	}
	err := c.GetAllByGsi(ctx, "index-name", "a-partial-sk", &v)
	assert.Nil(t, err)
	assert.Len(t, v, 2)
	assert.Equal(t, "some data", v[0].Data)
	assert.Equal(t, "some data", v[1].Data)
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
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{}}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.GetAllByGsi(ctx, "index-name", "a-partial-sk", &v)
	assert.Equal(t, expectedError, err)
}
