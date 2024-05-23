package dynamo

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	lpaUIDIndex      = "LpaUIDIndex"
	skUpdatedAtIndex = "SKUpdatedAtIndex"
)

type dynamoDB interface {
	Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	BatchGetItem(context.Context, *dynamodb.BatchGetItemInput, ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	TransactWriteItems(context.Context, *dynamodb.TransactWriteItemsInput, ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
	DeleteItem(context.Context, *dynamodb.DeleteItemInput, ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
	UpdateItem(context.Context, *dynamodb.UpdateItemInput, ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	BatchWriteItem(ctx context.Context, params *dynamodb.BatchWriteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.BatchWriteItemOutput, error)
}

type Client struct {
	table string
	svc   dynamoDB
}

type NotFoundError struct{}

func (n NotFoundError) Error() string {
	return "No results found"
}

type MultipleResultsError struct{}

func (n MultipleResultsError) Error() string {
	return "A single result was expected but multiple results found"
}

type ConditionalCheckFailedError struct{}

func (c ConditionalCheckFailedError) Error() string {
	return "Conditional checks failed"
}

func NewClient(cfg aws.Config, tableName string) (*Client, error) {
	return &Client{table: tableName, svc: dynamodb.NewFromConfig(cfg)}, nil
}

func (c *Client) One(ctx context.Context, pk PK, sk SK, v interface{}) error {
	result, err := c.svc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk.PK()},
			"SK": &types.AttributeValueMemberS{Value: sk.SK()},
		},
	})
	if err != nil {
		return err
	}
	if result.Item == nil {
		return NotFoundError{}
	}

	return attributevalue.UnmarshalMap(result.Item, v)
}

func (c *Client) OneByUID(ctx context.Context, uid string, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		IndexName:                aws.String(lpaUIDIndex),
		ExpressionAttributeNames: map[string]string{"#LpaUID": "LpaUID"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":LpaUID": &types.AttributeValueMemberS{Value: uid},
		},
		KeyConditionExpression: aws.String("#LpaUID = :LpaUID"),
	})

	if err != nil {
		return fmt.Errorf("failed to query UID: %w", err)
	}

	if len(response.Items) != 1 {
		return fmt.Errorf("expected to resolve LpaUID but got %d items", len(response.Items))
	}

	return attributevalue.UnmarshalMap(response.Items[0], v)
}

func (c *Client) AllBySK(ctx context.Context, sk SK, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		IndexName:                aws.String(skUpdatedAtIndex),
		ExpressionAttributeNames: map[string]string{"#SK": "SK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":SK": &types.AttributeValueMemberS{Value: sk.SK()},
		},
		KeyConditionExpression: aws.String("#SK = :SK"),
	})
	if err != nil {
		return err
	}

	return attributevalue.UnmarshalListOfMaps(response.Items, v)
}

func (c *Client) OneBySK(ctx context.Context, sk SK, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		IndexName:                aws.String(skUpdatedAtIndex),
		ExpressionAttributeNames: map[string]string{"#SK": "SK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":SK": &types.AttributeValueMemberS{Value: sk.SK()},
		},
		KeyConditionExpression: aws.String("#SK = :SK"),
	})
	if err != nil {
		return err
	}

	if len(response.Items) == 0 {
		return NotFoundError{}
	}

	if len(response.Items) > 1 {
		return fmt.Errorf("expected to resolve SK but got %d items", len(response.Items))
	}

	return attributevalue.UnmarshalMap(response.Items[0], v)
}

func (c *Client) LatestForActor(ctx context.Context, sk SK, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		IndexName:                aws.String(skUpdatedAtIndex),
		ExpressionAttributeNames: map[string]string{"#SK": "SK", "#UpdatedAt": "UpdatedAt"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":SK": &types.AttributeValueMemberS{Value: sk.SK()},
			// Specifying the condition UpdatedAt>2 filters out zero-value timestamps
			":UpdatedAt": &types.AttributeValueMemberS{Value: "2"},
		},
		KeyConditionExpression: aws.String("#SK = :SK and #UpdatedAt > :UpdatedAt"),
		ScanIndexForward:       aws.Bool(false),
		Limit:                  aws.Int32(1),
	})

	if err != nil {
		return err
	}

	if len(response.Items) == 0 {
		return nil
	}

	return attributevalue.UnmarshalMap(response.Items[0], v)
}

func (c *Client) AllKeysByPK(ctx context.Context, pk PK) ([]Keys, error) {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		ExpressionAttributeNames: map[string]string{"#PK": "PK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":PK": &types.AttributeValueMemberS{Value: pk.PK()},
		},
		KeyConditionExpression: aws.String("#PK = :PK"),
		ProjectionExpression:   aws.String("PK, SK"),
	})

	if err != nil {
		return nil, err
	}

	var keys []Keys
	err = attributevalue.UnmarshalListOfMaps(response.Items, &keys)

	return keys, err
}

func (c *Client) AllByKeys(ctx context.Context, keys []Keys) ([]map[string]types.AttributeValue, error) {
	var keyAttrs []map[string]types.AttributeValue
	for _, key := range keys {
		keyAttrs = append(keyAttrs, map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: key.PK.PK()},
			"SK": &types.AttributeValueMemberS{Value: key.SK.SK()},
		})
	}

	result, err := c.svc.BatchGetItem(ctx, &dynamodb.BatchGetItemInput{
		RequestItems: map[string]types.KeysAndAttributes{
			c.table: {
				Keys: keyAttrs,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	return result.Responses[c.table], nil
}

func (c *Client) OneByPK(ctx context.Context, pk PK, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		ExpressionAttributeNames: map[string]string{"#PK": "PK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":PK": &types.AttributeValueMemberS{Value: pk.PK()},
		},
		KeyConditionExpression: aws.String("#PK = :PK"),
	})

	if err != nil {
		return err
	}

	if len(response.Items) == 0 {
		return NotFoundError{}
	}

	if len(response.Items) > 1 {
		return MultipleResultsError{}
	}

	return attributevalue.UnmarshalMap(response.Items[0], v)
}

func (c *Client) OneByPartialSK(ctx context.Context, pk PK, partialSK SK, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		ExpressionAttributeNames: map[string]string{"#PK": "PK", "#SK": "SK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":PK": &types.AttributeValueMemberS{Value: pk.PK()},
			":SK": &types.AttributeValueMemberS{Value: partialSK.SK()},
		},
		KeyConditionExpression: aws.String("#PK = :PK and begins_with(#SK, :SK)"),
	})

	if err != nil {
		return err
	}

	if len(response.Items) == 0 {
		return NotFoundError{}
	}

	if len(response.Items) > 1 {
		return MultipleResultsError{}
	}

	return attributevalue.UnmarshalMap(response.Items[0], v)
}

func (c *Client) AllByPartialSK(ctx context.Context, pk PK, partialSk SK, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		ExpressionAttributeNames: map[string]string{"#PK": "PK", "#SK": "SK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":PK": &types.AttributeValueMemberS{Value: pk.PK()},
			":SK": &types.AttributeValueMemberS{Value: partialSk.SK()},
		},
		KeyConditionExpression: aws.String("#PK = :PK and begins_with(#SK, :SK)"),
	})
	if err != nil {
		return err
	}

	return attributevalue.UnmarshalListOfMaps(response.Items, v)
}

func (c *Client) Put(ctx context.Context, v interface{}) error {
	item, err := attributevalue.MarshalMap(v)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(c.table),
		Item:      item,
	}

	// Tracking Value equality against data on write allows for optimistic locking
	if currentVersion, exists := item["Version"]; exists {
		var v int
		err = attributevalue.Unmarshal(currentVersion, &v)
		if err != nil {
			return err
		}

		v++
		newVersion, err := attributevalue.Marshal(v)
		if err != nil {
			return err
		}

		item["Version"] = newVersion

		input.Item = item
		input.ConditionExpression = aws.String("Version = :version")
		input.ExpressionAttributeValues = map[string]types.AttributeValue{
			":version": currentVersion,
		}
	}

	_, err = c.svc.PutItem(ctx, input)

	if err != nil {
		var ccf *types.ConditionalCheckFailedException
		if errors.As(err, &ccf) {
			return ConditionalCheckFailedError{}
		}

		return err
	}

	return nil
}

func (c *Client) Create(ctx context.Context, v interface{}) error {
	item, err := attributevalue.MarshalMap(v)
	if err != nil {
		return err
	}

	_, err = c.svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(c.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	})

	return err
}

func (c *Client) DeleteKeys(ctx context.Context, keys []Keys) error {
	items := make([]types.TransactWriteItem, len(keys))

	for i, key := range keys {
		items[i] = types.TransactWriteItem{
			Delete: &types.Delete{
				TableName: aws.String(c.table),
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: key.PK.PK()},
					"SK": &types.AttributeValueMemberS{Value: key.SK.SK()},
				},
			},
		}
	}

	_, err := c.svc.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: items,
	})

	return err
}

func (c *Client) DeleteOne(ctx context.Context, pk PK, sk SK) error {
	_, err := c.svc.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk.PK()},
			"SK": &types.AttributeValueMemberS{Value: sk.SK()},
		},
	})

	return err
}

func (c *Client) Update(ctx context.Context, pk PK, sk SK, values map[string]types.AttributeValue, expression string) error {
	_, err := c.svc.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk.PK()},
			"SK": &types.AttributeValueMemberS{Value: sk.SK()},
		},
		ExpressionAttributeValues: values,
		UpdateExpression:          aws.String(expression),
	})

	return err
}

func (c *Client) UpdateReturn(ctx context.Context, pk PK, sk SK, values map[string]types.AttributeValue, expression string) (map[string]types.AttributeValue, error) {
	resp, err := c.svc.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk.PK()},
			"SK": &types.AttributeValueMemberS{Value: sk.SK()},
		},
		ExpressionAttributeValues: values,
		UpdateExpression:          aws.String(expression),
		ReturnValues:              types.ReturnValueAllNew,
	})
	if err != nil {
		return nil, err
	}

	return resp.Attributes, nil
}

func (c *Client) BatchPut(ctx context.Context, values []interface{}) error {
	items := make([]types.TransactWriteItem, len(values))

	for i, value := range values {
		v, err := attributevalue.MarshalMap(value)
		if err != nil {
			return err
		}

		items[i] = types.TransactWriteItem{
			Put: &types.Put{
				TableName: aws.String(c.table),
				Item:      v,
			},
		}
	}

	_, err := c.svc.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: items,
	})

	return err
}

func (c *Client) WriteTransaction(ctx context.Context, transaction *Transaction) error {
	if len(transaction.Puts) == 0 && len(transaction.Deletes) == 0 {
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
