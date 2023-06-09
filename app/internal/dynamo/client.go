package dynamo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

//go:generate mockery --testonly --inpackage --name dynamoDB --structname mockDynamoDB
type dynamoDB interface {
	Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	BatchGetItem(context.Context, *dynamodb.BatchGetItemInput, ...func(*dynamodb.Options)) (*dynamodb.BatchGetItemOutput, error)
	PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	TransactWriteItems(ctx context.Context, params *dynamodb.TransactWriteItemsInput, optFns ...func(*dynamodb.Options)) (*dynamodb.TransactWriteItemsOutput, error)
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

func (c *Client) Get(ctx context.Context, pk, sk string, v interface{}) error {
	key, err := makeKey(pk, sk)
	if err != nil {
		return err
	}
	result, err := c.svc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(c.table),
		Key:       key,
	})
	if err != nil {
		return err
	}
	if result.Item == nil {
		return NotFoundError{}
	}
	return attributevalue.Unmarshal(result.Item["Data"], v)
}

func (c *Client) GetAllByGsi(ctx context.Context, gsi, sk string, v interface{}) error {
	skey, err := attributevalue.Marshal(sk)
	if err != nil {
		return err
	}

	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(c.table),
		IndexName:                 aws.String(gsi),
		ExpressionAttributeNames:  map[string]string{"#SK": "SK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":SK": skey},
		KeyConditionExpression:    aws.String("#SK = :SK"),
	})
	if err != nil {
		return err
	}

	return attributevalue.UnmarshalListOfMaps(response.Items, v)
}

type Key struct {
	PK string
	SK string
}

func (c *Client) GetAllByKeys(ctx context.Context, keys []Key, v interface{}) error {
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
		return err
	}

	return attributevalue.UnmarshalListOfMaps(result.Responses[c.table], &v)
}

func (c *Client) GetOneByPartialSk(ctx context.Context, pk, partialSk string, v interface{}) error {
	pkey, err := attributevalue.Marshal(pk)
	if err != nil {
		return err
	}

	partialSkey, err := attributevalue.Marshal(partialSk)
	if err != nil {
		return err
	}

	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(c.table),
		ExpressionAttributeNames:  map[string]string{"#PK": "PK", "#SK": "SK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":PK": pkey, ":SK": partialSkey},
		KeyConditionExpression:    aws.String("#PK = :PK and begins_with(#SK, :SK)"),
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

	return attributevalue.Unmarshal(response.Items[0]["Data"], v)
}

func (c *Client) Put(ctx context.Context, pk, sk string, v interface{}) error {
	item, err := makeKey(pk, sk)
	if err != nil {
		return err
	}

	data, err := attributevalue.Marshal(v)
	if err != nil {
		return err
	}
	item["Data"] = data

	_, err = c.svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.table),
		Item:      item,
	})

	return err
}

func (c *Client) Create(ctx context.Context, pk, sk string, v interface{}) error {
	item, err := makeKey(pk, sk)
	if err != nil {
		return err
	}

	data, err := attributevalue.Marshal(v)
	if err != nil {
		return err
	}
	item["Data"] = data

	_, err = c.svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName:           aws.String(c.table),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(PK) AND attribute_not_exists(SK)"),
	})

	return err
}

func makeKey(pk, sk string) (map[string]types.AttributeValue, error) {
	pkey, err := attributevalue.Marshal(pk)
	if err != nil {
		return nil, err
	}

	skey, err := attributevalue.Marshal(sk)
	if err != nil {
		return nil, err
	}

	return map[string]types.AttributeValue{"PK": pkey, "SK": skey}, nil
}
