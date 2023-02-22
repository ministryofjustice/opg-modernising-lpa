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
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func TestGetAll(t *testing.T) {
	ctx := context.Background()

	pkey, _ := attributevalue.Marshal("a-pk")
	data, _ := attributevalue.Marshal("hello")

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("this"),
			ExpressionAttributeNames:  map[string]string{"#PK": "PK"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey},
			KeyConditionExpression:    aws.String("#PK = :PK"),
		}).
		Return(&dynamodb.QueryOutput{Items: []map[string]types.AttributeValue{{"Data": data}}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.GetAll(ctx, "a-pk", &v)
	assert.Nil(t, err)
	assert.Equal(t, []string{"hello"}, v)
}

func TestGetAllWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.
		On("Query", ctx, mock.Anything).
		Return(&dynamodb.QueryOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v []string
	err := c.GetAll(ctx, "a-pk", &v)
	assert.Equal(t, expectedError, err)
	assert.Empty(t, v)
}

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
	assert.Nil(t, err)
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
