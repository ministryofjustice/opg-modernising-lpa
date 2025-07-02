package dynamo

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

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
}

type Client struct {
	table string
	svc   dynamoDB
}

type NotFoundError struct{}

func (n NotFoundError) Error() string {
	return "No results found"
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

func (c *Client) OneByUID(ctx context.Context, uid string) (Keys, error) {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(c.table),
		IndexName: aws.String(lpaUIDIndex),
		ExpressionAttributeNames: map[string]string{
			"#LpaUID": "LpaUID",
			"#PK":     "PK",
			"#SK":     "SK",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":LpaUID": &types.AttributeValueMemberS{Value: uid},
			":PK":     &types.AttributeValueMemberS{Value: LpaKey("").PK()},
			":SK":     &types.AttributeValueMemberS{Value: DonorKey("").SK()},
		},
		KeyConditionExpression: aws.String("#LpaUID = :LpaUID"),
		FilterExpression:       aws.String("begins_with(#PK, :PK) and begins_with(#SK, :SK)"),
	})
	if err != nil {
		return Keys{}, fmt.Errorf("failed to query UID: %w", err)
	}
	if len(response.Items) == 0 {
		return Keys{}, NotFoundError{}
	}

	// limits are applied before filters so we need to manually handle > 1
	// see https://docs.aws.amazon.com/amazondynamodb/latest/developerguide/Query.Other.html#Query.Limit
	if len(response.Items) > 1 {
		return Keys{}, fmt.Errorf("expected to resolve partial PK and SK with LpaUID %s but got %d items", uid, len(response.Items))
	}

	var keys Keys
	err = attributevalue.UnmarshalMap(response.Items[0], &keys)
	return keys, err
}

func (c *Client) AllByLpaUIDAndPartialSK(ctx context.Context, uid string, partialSK SK) ([]Keys, error) {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName: aws.String(c.table),
		IndexName: aws.String(lpaUIDIndex),
		ExpressionAttributeNames: map[string]string{
			"#LpaUID": "LpaUID",
			"#SK":     "SK",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":LpaUID": &types.AttributeValueMemberS{Value: uid},
			":SK":     &types.AttributeValueMemberS{Value: partialSK.SK()},
		},
		KeyConditionExpression: aws.String("#LpaUID = :LpaUID"),
		FilterExpression:       aws.String("begins_with(#SK, :SK)"),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to query UID: %w", err)
	}
	if len(response.Items) == 0 {
		return nil, NotFoundError{}
	}

	var v []Keys
	err = attributevalue.UnmarshalListOfMaps(response.Items, &v)
	return v, err
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
		Limit:                  aws.Int32(1),
	})

	if err != nil {
		return err
	}

	if len(response.Items) == 0 {
		return NotFoundError{}
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
		Limit:                  aws.Int32(1),
	})

	if err != nil {
		return err
	}

	if len(response.Items) == 0 {
		return NotFoundError{}
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

// Create writes data ensuring that the (PK, SK) combination is unique.
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

// CreateOnly writes data ensuring that the PK is unique.
func (c *Client) CreateOnly(ctx context.Context, v interface{}) error {
	item, err := attributevalue.MarshalMap(v)
	if err != nil {
		return err
	}

	_, err = c.svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(c.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK)"),
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

func (c *Client) Update(ctx context.Context, pk PK, sk SK, names map[string]string, values map[string]types.AttributeValue, expression string) error {
	_, err := c.svc.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: pk.PK()},
			"SK": &types.AttributeValueMemberS{Value: sk.SK()},
		},
		ExpressionAttributeNames:  names,
		ExpressionAttributeValues: values,
		UpdateExpression:          aws.String(expression),
	})

	return err
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

func (c *Client) Move(ctx context.Context, oldKeys Keys, value any) error {
	v, err := attributevalue.MarshalMap(value)
	if err != nil {
		return err
	}

	_, err = c.svc.TransactWriteItems(ctx, &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				Delete: &types.Delete{
					TableName: aws.String(c.table),
					Key: map[string]types.AttributeValue{
						"PK": &types.AttributeValueMemberS{Value: oldKeys.PK.PK()},
						"SK": &types.AttributeValueMemberS{Value: oldKeys.SK.SK()},
					},
					ConditionExpression: aws.String("attribute_exists(PK) and attribute_exists(SK)"),
				},
			},
			{
				Put: &types.Put{
					TableName: aws.String(c.table),
					Item:      v,
				},
			},
		},
	})

	var conflictException *types.TransactionConflictException
	if errors.As(err, &conflictException) {
		return ConditionalCheckFailedError{}
	}

	var canceledException *types.TransactionCanceledException
	if errors.As(err, &canceledException) {
		for _, reason := range canceledException.CancellationReasons {
			if *reason.Code == "ConditionalCheckFailed" {
				return ConditionalCheckFailedError{}
			}
		}
	}

	return err
}

func (c *Client) AnyByPK(ctx context.Context, pk PK, v interface{}) error {
	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		ExpressionAttributeNames: map[string]string{"#PK": "PK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":PK": &types.AttributeValueMemberS{Value: pk.PK()},
		},
		KeyConditionExpression: aws.String("#PK = :PK"),
		Limit:                  aws.Int32(1),
	})

	if err != nil {
		return err
	}

	if len(response.Items) == 0 {
		return NotFoundError{}
	}

	return attributevalue.UnmarshalMap(response.Items[0], v)
}

func (c *Client) OneActive(ctx context.Context, pk PK, sk ExpiresAtKeyType, now time.Time, v interface{}) error {
	nowEpochSeconds := strconv.FormatInt(now.UTC().Unix(), 10)

	output, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                aws.String(c.table),
		ExpressionAttributeNames: map[string]string{"#PK": "PK", "#SK": "SK", "#ExpiresAt": "ExpiresAt"},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":PK":  &types.AttributeValueMemberS{Value: pk.PK()},
			":SK":  &types.AttributeValueMemberS{Value: sk.Partial()},
			":now": &types.AttributeValueMemberN{Value: nowEpochSeconds},
		},
		KeyConditionExpression: aws.String("#PK = :PK and begins_with(#SK, :SK)"),
		Limit:                  aws.Int32(1),
		FilterExpression:       aws.String("#ExpiresAt > :now"),
	})

	if err != nil {
		return err
	}

	if len(output.Items) == 0 {
		return NotFoundError{}
	}

	return attributevalue.UnmarshalMap(output.Items[0], v)
}
