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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type evidenceReceivedEvent struct {
	UID string `json:"uid"`
}

//go:generate mockery --testonly --inpackage --name dynamodbClient --structname mockDynamodbClient
type dynamodbClient interface {
	Put(ctx context.Context, v interface{}) error
	GetOneByUID(context.Context, string, interface{}) error
}

func Handler(ctx context.Context, event events.CloudWatchEvent) error {
	tableName := os.Getenv("LPAS_TABLE")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
	}

	dynamoClient, err := dynamo.NewClient(cfg, tableName)
	if err != nil {
		return fmt.Errorf("failed to create dynamodb client: %w", err)
	}

	switch event.DetailType {
	case "evidence-received":
		return handleEvidenceReceived(ctx, dynamoClient, tableName, event)
	default:
		return fmt.Errorf("unknown event received: %s", event.DetailType)
	}
}

func handleEvidenceReceived(ctx context.Context, client dynamodbClient, tableName string, event events.CloudWatchEvent) error {
	var v evidenceReceivedEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal 'evidence-received' detail: %w", err)
	}

	var lpa page.Lpa
	err := client.GetOneByUID(ctx, v.UID, &lpa)
	if err != nil {
		return fmt.Errorf("failed to resolve uid for 'evidence-received': %w", err)
	}

	item, err := attributevalue.MarshalMap(map[string]any{"PK": lpa.PK, "SK": "#EVIDENCE_RECEIVED"})
	if err != nil {
		return fmt.Errorf("failed to marshal item in response to 'evidence-received': %w", err)
	}

	return client.Put(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
}

func main() {
	lambda.Start(Handler)
}
