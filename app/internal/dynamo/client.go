package dynamo

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type dynamoDB interface {
	GetItemWithContext(aws.Context, *dynamodb.GetItemInput, ...request.Option) (*dynamodb.GetItemOutput, error)
	PutItemWithContext(aws.Context, *dynamodb.PutItemInput, ...request.Option) (*dynamodb.PutItemOutput, error)
}

type Client struct {
	table string
	svc   dynamoDB
}

func NewClient(sess *session.Session, tableName string) (*Client, error) {
	return &Client{table: tableName, svc: dynamodb.New(sess)}, nil
}

func (c *Client) Get(ctx context.Context, id string, v interface{}) error {
	result, err := c.svc.GetItemWithContext(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(c.table),
		Key: map[string]*dynamodb.AttributeValue{
			"Id": {
				S: aws.String(id),
			},
		},
	})
	if err != nil {
		return err
	}
	if result.Item == nil {
		return nil
	}

	return json.Unmarshal(result.Item["Data"].B, v)
}

func (c *Client) Put(ctx context.Context, id string, v interface{}) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}

	_, err = c.svc.PutItemWithContext(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(c.table),
		Item: map[string]*dynamodb.AttributeValue{
			"Id":   {S: aws.String(id)},
			"Data": {B: data},
		},
	})

	return err
}
