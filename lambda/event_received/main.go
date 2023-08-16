package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type evidenceReceivedEvent struct {
	UID string `json:"uid"`
}

func Handler(ctx context.Context, event events.CloudWatchEvent) error {
	tableName := os.Getenv("LPAS_TABLE")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
	}

	db := dynamodb.NewFromConfig(cfg)

	switch event.DetailType {
	case "evidence-received":
		var v evidenceReceivedEvent
		if err := json.Unmarshal(event.Detail, &v); err != nil {
			return fmt.Errorf("failed to unmarshal 'evidence-received' detail: %w", err)
		}

		id, err := resolveUID(ctx, db, tableName, v.UID)
		if err != nil {
			return fmt.Errorf("failed to resolve uid for 'evidence-received': %w", err)
		}

		item, err := attributevalue.MarshalMap(map[string]any{
			"PK": id,
			"SK": "#EVIDENCE_RECEIVED",
		})
		if err != nil {
			return fmt.Errorf("failed to marshal item in response to 'evidence-received': %w", err)
		}

		_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      item,
		})

		return err
	default:
		return fmt.Errorf("unknown event received: %s", event.DetailType)
	}
}

func resolveUID(ctx context.Context, db *dynamodb.Client, tableName, uid string) (string, error) {
	skey, err := attributevalue.Marshal(uid)
	if err != nil {
		return "", fmt.Errorf("failed to marshal UID: %w", err)
	}

	response, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("UidIndex"),
		ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":UID": skey},
		KeyConditionExpression:    aws.String("#UID = :UID"),
	})
	if err != nil {
		return "", fmt.Errorf("failed to query UID: %w", err)
	}
	if len(response.Items) != 1 {
		return "", fmt.Errorf("expected to resolve UID but got %d items", len(response.Items))
	}

	var v struct{ PK string }
	if err := attributevalue.UnmarshalMap(response.Items[0], &v); err != nil {
		return "", fmt.Errorf("failed to unmarshal UID response: %w", err)
	}

	return v.PK, nil
}

func main() {
	lambda.Start(Handler)
}
