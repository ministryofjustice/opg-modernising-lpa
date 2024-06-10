package dynamo

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestTransactWriteItems(t *testing.T) {
	items := []types.TransactWriteItem{
		{
			Put: &types.Put{
				Item:                map[string]types.AttributeValue{"1": &types.AttributeValueMemberS{Value: "1"}},
				TableName:           aws.String("this"),
				ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
			},
		},
		{
			Put: &types.Put{
				Item:                map[string]types.AttributeValue{"2": &types.AttributeValueMemberS{Value: "2"}},
				TableName:           aws.String("this"),
				ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
			},
		},
		{
			Put: &types.Put{
				Item:      map[string]types.AttributeValue{"3": &types.AttributeValueMemberS{Value: "3"}},
				TableName: aws.String("this"),
			},
		},
		{
			Put: &types.Put{
				Item:      map[string]types.AttributeValue{"4": &types.AttributeValueMemberS{Value: "4"}},
				TableName: aws.String("this"),
			},
		},
		{
			Delete: &types.Delete{Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "PK-1"},
				"SK": &types.AttributeValueMemberS{Value: "SK-1"},
			}, TableName: aws.String("this")},
		},
		{
			Delete: &types.Delete{Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "PK-2"},
				"SK": &types.AttributeValueMemberS{Value: "SK-2"},
			}, TableName: aws.String("this")},
		},
	}

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		TransactWriteItems(context.Background(), &dynamodb.TransactWriteItemsInput{
			TransactItems: items,
		}).
		Return(nil, nil)

	c := &Client{table: "this", svc: dynamoDB}
	err := c.WriteTransaction(context.Background(), &Transaction{
		Creates: []any{
			map[string]string{"1": "1"},
			map[string]string{"2": "2"},
		},
		Puts: []any{
			map[string]string{"3": "3"},
			map[string]string{"4": "4"},
		},
		Deletes: []Keys{
			{PK: testPK("PK-1"), SK: testSK("SK-1")},
			{PK: testPK("PK-2"), SK: testSK("SK-2")},
		},
	})

	assert.Nil(t, err)
}

func TestTransactWriteItemsWhenNoTransactions(t *testing.T) {
	c := &Client{table: "this", svc: nil}
	err := c.WriteTransaction(context.Background(), &Transaction{})

	assert.Equal(t, errors.New("WriteTransaction requires at least one transaction"), err)
}

func TestTransactWriteItemsWhenErrorsBuildingTransaction(t *testing.T) {
	testcases := map[string]struct {
		creates []any
		puts    []any
	}{
		"create": {
			creates: []any{map[string]string{"": "1"}},
		},
		"put": {
			puts: []any{map[string]string{"": "1"}},
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			c := &Client{table: "this", svc: nil}
			err := c.WriteTransaction(context.Background(), &Transaction{
				Creates: tc.creates,
				Puts:    tc.puts,
			})

			assert.Error(t, err)
		})
	}
}

func TestTransactWriteItemsWhenTransactWriteItemsError(t *testing.T) {
	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		TransactWriteItems(context.Background(), mock.Anything).
		Return(nil, expectedError)

	c := &Client{table: "this", svc: dynamoDB}
	err := c.WriteTransaction(context.Background(), &Transaction{
		Puts: []any{map[string]string{"1": "1"}},
	})

	assert.Equal(t, expectedError, err)
}

func TestTransactWriteItemsWhenTransactionCancelled(t *testing.T) {
	exception := &types.TransactionCanceledException{
		CancellationReasons: []types.CancellationReason{
			{Code: aws.String("None")},
		},
	}

	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		TransactWriteItems(mock.Anything, mock.Anything).
		Return(nil, exception)

	c := &Client{table: "this", svc: dynamoDB}
	err := c.WriteTransaction(context.Background(), &Transaction{Puts: []any{"1"}})

	assert.Equal(t, exception, err)
}

func TestTransactWriteItemsWhenConditionalCheckFailed(t *testing.T) {
	dynamoDB := newMockDynamoDB(t)
	dynamoDB.EXPECT().
		TransactWriteItems(mock.Anything, mock.Anything).
		Return(nil, &types.TransactionCanceledException{
			CancellationReasons: []types.CancellationReason{
				{Code: aws.String("None")},
				{Code: aws.String("ConditionalCheckFailed")},
			},
		})

	c := &Client{table: "this", svc: dynamoDB}
	err := c.WriteTransaction(context.Background(), &Transaction{Puts: []any{"1"}})

	assert.Equal(t, ConditionalCheckFailedError{}, err)
}

func TestNewTransaction(t *testing.T) {
	assert.Equal(t, &Transaction{}, NewTransaction())
}

func TestTransactionPut(t *testing.T) {
	putA := map[string]string{"a": "a"}
	putB := map[string]string{"b": "b"}
	transaction := NewTransaction().
		Put(putA).
		Put(putB)

	assert.Equal(t, []any{putA, putB}, transaction.Puts)
}

func TestTransactionDelete(t *testing.T) {
	deleteA := Keys{PK: testPK("PK-A"), SK: testSK("SK-A")}
	deleteB := Keys{PK: testPK("PK-B"), SK: testSK("SK-B")}

	transaction := NewTransaction().
		Delete(deleteA).
		Delete(deleteB)

	assert.Equal(t, []Keys{deleteA, deleteB}, transaction.Deletes)
}

func TestTransactionCreate(t *testing.T) {
	putA := map[string]string{"a": "a"}
	putB := map[string]string{"b": "b"}
	transaction := NewTransaction().
		Create(putA).
		Create(putB)

	assert.Equal(t, []any{putA, putB}, transaction.Creates)
}
