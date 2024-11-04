package dynamo

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func (c *Client) WriteTransaction(ctx context.Context, transaction *Transaction) error {
	if len(transaction.Creates) == 0 && len(transaction.Puts) == 0 && len(transaction.Deletes) == 0 {
		return errors.New("WriteTransaction requires at least one transaction")
	}

	var items []types.TransactWriteItem

	for _, cr := range transaction.Creates {
		values, err := attributevalue.MarshalMap(cr)
		if err != nil {
			return err
		}

		items = append(items, types.TransactWriteItem{Put: &types.Put{
			TableName:           aws.String(c.table),
			Item:                values,
			ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
		}})
	}

	for _, p := range transaction.Puts {
		values, err := attributevalue.MarshalMap(p)
		if err != nil {
			return err
		}

		items = append(items, types.TransactWriteItem{Put: &types.Put{
			TableName: aws.String(c.table),
			Item:      values,
		}})
	}

	for _, d := range transaction.Deletes {
		items = append(items, types.TransactWriteItem{Delete: &types.Delete{
			TableName: aws.String(c.table),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: d.PK.PK()},
				"SK": &types.AttributeValueMemberS{Value: d.SK.SK()},
			},
		}})
	}

	_, err := c.svc.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: items,
	})

	var tce *types.TransactionCanceledException
	if errors.As(err, &tce) {
		for _, reason := range tce.CancellationReasons {
			if *reason.Code == "ConditionalCheckFailed" {
				return ConditionalCheckFailedError{}
			}
		}
	}

	return err
}

type Transaction struct {
	Creates []any
	Puts    []any
	Deletes []Keys
}

func NewTransaction() *Transaction {
	return &Transaction{}
}

func (t *Transaction) Create(v interface{}) *Transaction {
	t.Creates = append(t.Creates, v)
	return t
}

func (t *Transaction) Put(v interface{}) *Transaction {
	t.Puts = append(t.Puts, v)
	return t
}

func (t *Transaction) Delete(keys Keys) *Transaction {
	t.Deletes = append(t.Deletes, keys)
	return t
}
