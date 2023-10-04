package dynamo

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	uidIndex            = "UidIndex"
	actorUpdatedAtIndex = "ActorUpdatedAtIndex"
)

//go:generate mockery --testonly --inpackage --name dynamoDB --structname mockDynamoDB
type dynamoDB interface {
	Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	BatchGetItem(context.Context, *dynamodb.BatchGetItemInput, ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	TransactWriteItems(context.Context, *dynamodb.TransactWriteItemsInput, ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
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

func NewClient(cfg aws.Config, tableName string) (*Client, error) {
	return &Client{table: tableName, svc: dynamodb.NewFromConfig(cfg)}, nil
}

func (c *Client) One(ctx context.Context, pk, sk string, v interface{}) error {
	result, err := c.svc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk},
			"SK": &types.AttributeValueMemberS{Value: sk},
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
		IndexName:                aws.String(uidIndex),
		ExpressionAttributeNames: map[string]string{"#UID": "UID"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":UID": &types.AttributeValueMemberS{Value: uid},
		},
		KeyConditionExpression: aws.String("#UID = :UID"),
	})

	if err != nil {
		return fmt.Errorf("failed to query UID: %w", err)
	}

	if len(response.Items) != 1 {
		return fmt.Errorf("expected to resolve UID but got %d items", len(response.Items))
	}

	return attributevalue.UnmarshalMap(response.Items[0], v)
}

func (c *Client) AllForActor(ctx context.Context, sk string, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		IndexName:                aws.String(actorUpdatedAtIndex),
		ExpressionAttributeNames: map[string]string{"#SK": "SK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":SK": &types.AttributeValueMemberS{Value: sk},
		},
		KeyConditionExpression: aws.String("#SK = :SK"),
	})
	if err != nil {
		return err
	}

	return attributevalue.UnmarshalListOfMaps(response.Items, v)
}

func (c *Client) LatestForActor(ctx context.Context, sk string, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		IndexName:                aws.String(actorUpdatedAtIndex),
		ExpressionAttributeNames: map[string]string{"#SK": "SK", "#UpdatedAt": "UpdatedAt"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":SK": &types.AttributeValueMemberS{Value: sk},
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

type Key struct {
	PK string
	SK string
}

func (c *Client) AllByKeys(ctx context.Context, keys []Key) ([]map[string]types.AttributeValue, error) {
	var keyAttrs []map[string]types.AttributeValue
	for _, key := range keys {
		keyAttrs = append(keyAttrs, map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: key.PK},
			"SK": &types.AttributeValueMemberS{Value: key.SK},
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

func (c *Client) OneByPartialSk(ctx context.Context, pk, partialSk string, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		ExpressionAttributeNames: map[string]string{"#PK": "PK", "#SK": "SK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":PK": &types.AttributeValueMemberS{Value: pk},
			":SK": &types.AttributeValueMemberS{Value: partialSk},
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

func (c *Client) Put(ctx context.Context, v interface{}) error {
	item, err := attributevalue.MarshalMap(v)
	if err != nil {
		return err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(c.table),
		Item:      item,
	}

	if updatedAt, exists := item["UpdatedAt"]; exists {
		var updatedAtTime time.Time
		err = attributevalue.Unmarshal(updatedAt, &updatedAtTime)
		if err != nil {
			return err
		}

		if !updatedAtTime.IsZero() {
			input.ConditionExpression = aws.String("UpdatedAt < :updatedAt")
			input.ExpressionAttributeValues = map[string]types.AttributeValue{
				":updatedAt": updatedAt,
			}
		}
	}

	_, err = c.svc.PutItem(ctx, input)

	return err
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
