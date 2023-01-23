package dynamo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamoDB interface {
	Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	GetItem(context.Context, *dynamodb.GetItemInput, ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error)
	PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
}

type Client struct {
	table string
	svc   dynamoDB
}

func NewClient(cfg aws.Config, tableName string) (*Client, error) {
	return &Client{table: tableName, svc: dynamodb.NewFromConfig(cfg)}, nil
}

func (c *Client) GetAll(ctx context.Context, pk string, v interface{}) error {
	pkey, err := attributevalue.Marshal(pk)
	if err != nil {
		return err
	}

	response, err := c.svc.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(c.table),
		ExpressionAttributeNames:  map[string]string{"#0": "PK"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":0": pkey},
		KeyConditionExpression:    aws.String("#0 = :0"),
	})
	if err != nil {
		return err
	}

	var items []types.AttributeValue
	for _, item := range response.Items {
		items = append(items, item["Data"])
	}

	return attributevalue.UnmarshalList(items, v)
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
		return nil
	}

	return attributevalue.Unmarshal(result.Item["Data"], v)
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
