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
)

func Handler(ctx context.Context, event events.CloudWatchEvent) error {
	tableName := os.Getenv("LPAS_TABLE")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
	}

	switch event.DetailType {
	case "evidence-received":
		var v struct {
			UID string `json:"uid"`
		}
		if err := json.Unmarshal(event.Detail, &v); err != nil {
			return fmt.Errorf("failed to unmarshal 'evidence-received' detail: %w", err)
		}

		item, err := attributevalue.MarshalMap(map[string]any{
			"PK": "LPA#" + v.UID,
			"SK": "#EVIDENCE_RECEIVED",
		})
		if err != nil {
			return fmt.Errorf("failed to marshal item in response to 'evidence-received': %w", err)
		}

		_, err = dynamodb.NewFromConfig(cfg).PutItem(ctx, &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      item,
		})

		return err
	default:
		return fmt.Errorf("unknown event received: %s", event.DetailType)
	}
}

func main() {
	lambda.Start(Handler)
}
