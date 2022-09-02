package dynamo

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

type mockDynamoDB struct {
	mock.Mock
}

func (m *mockDynamoDB) GetItemWithContext(ctx aws.Context, input *dynamodb.GetItemInput, opts ...request.Option) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *mockDynamoDB) PutItemWithContext(ctx aws.Context, input *dynamodb.PutItemInput, opts ...request.Option) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, input)
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func TestGet(t *testing.T) {
	ctx := context.Background()

	dynamoDB := &mockDynamoDB{}
	dynamoDB.
		On("GetItemWithContext", ctx, &dynamodb.GetItemInput{
			TableName: aws.String("this"),
			Key: map[string]*dynamodb.AttributeValue{
				"Id": {
					S: aws.String("some-id"),
				},
			},
		}).
		Return(&dynamodb.GetItemOutput{Item: map[string]*dynamodb.AttributeValue{
			"Data": {B: []byte(`"hello"`)},
		}}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.Get(ctx, "some-id", &v)
	assert.Nil(t, err)
	assert.Equal(t, "hello", v)
}

func TestGetWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := &mockDynamoDB{}
	dynamoDB.
		On("GetItemWithContext", ctx, &dynamodb.GetItemInput{
			TableName: aws.String("this"),
			Key: map[string]*dynamodb.AttributeValue{
				"Id": {
					S: aws.String("some-id"),
				},
			},
		}).
		Return(&dynamodb.GetItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.Get(ctx, "some-id", &v)
	assert.Equal(t, expectedError, err)
	assert.Equal(t, "", v)
}

func TestGetWhenNotFound(t *testing.T) {
	ctx := context.Background()

	dynamoDB := &mockDynamoDB{}
	dynamoDB.
		On("GetItemWithContext", ctx, &dynamodb.GetItemInput{
			TableName: aws.String("this"),
			Key: map[string]*dynamodb.AttributeValue{
				"Id": {
					S: aws.String("some-id"),
				},
			},
		}).
		Return(&dynamodb.GetItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	var v string
	err := c.Get(ctx, "some-id", &v)
	assert.Nil(t, err)
	assert.Equal(t, "", v)
}

func TestPut(t *testing.T) {
	ctx := context.Background()

	dynamoDB := &mockDynamoDB{}
	dynamoDB.
		On("PutItemWithContext", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("this"),
			Item: map[string]*dynamodb.AttributeValue{
				"Id":   {S: aws.String("some-id")},
				"Data": {B: []byte(`"hello"`)},
			},
		}).
		Return(&dynamodb.PutItemOutput{}, nil)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, "some-id", "hello")
	assert.Nil(t, err)
}

func TestPutWhenError(t *testing.T) {
	ctx := context.Background()

	dynamoDB := &mockDynamoDB{}
	dynamoDB.
		On("PutItemWithContext", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("this"),
			Item: map[string]*dynamodb.AttributeValue{
				"Id":   {S: aws.String("some-id")},
				"Data": {B: []byte(`"hello"`)},
			},
		}).
		Return(&dynamodb.PutItemOutput{}, expectedError)

	c := &Client{table: "this", svc: dynamoDB}

	err := c.Put(ctx, "some-id", "hello")
	assert.Equal(t, expectedError, err)
}
