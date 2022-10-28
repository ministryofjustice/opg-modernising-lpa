package dynamo

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type dynamoDB interface {
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

func (c *Client) Get(ctx context.Context, id string, v interface{}) error {
	keyID, err := attributevalue.Marshal(id)
	if err != nil {
		return err
	}

	result, err := c.svc.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(c.table),
		Key: map[string]types.AttributeValue{
			"Id": keyID,
		},
	})
	if err != nil {
		return err
	}
	if result.Item == nil {
		return nil
	}

	return attributevalue.Unmarshal(result.Item["Data"], v)
}

type combined struct {
	ID   string
	Data interface{}
}

func (c *Client) Put(ctx context.Context, id string, v interface{}) error {
	keyID, err := attributevalue.Marshal(id)
	if err != nil {
		return err
	}

	data, err := attributevalue.Marshal(v)
	if err != nil {
		return err
	}

	_, err = c.svc.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.table),
		Item:      map[string]types.AttributeValue{"Id": keyID, "Data": data},
	})

	return err
}
