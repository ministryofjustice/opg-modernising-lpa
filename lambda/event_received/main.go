package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/ministryofjustice/opg-modernising-lpa/shared/notify"
)

type evidenceReceivedEvent struct {
	UID string `json:"uid"`
}

type feeApprovedEvent struct {
	UID string `json:"uid"`
}

type primaryKey struct {
	PK, SK string
}

//go:generate mockery --testonly --inpackage --name store --structname mockStore
type store interface {
	PutItem(context.Context, *dynamodb.PutItemInput, ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error)
	Query(context.Context, *dynamodb.QueryInput, ...func(*dynamodb.Options)) (*dynamodb.QueryOutput, error)
	UpdateItem(context.Context, *dynamodb.UpdateItemInput, ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

//go:generate mockery --testonly --inpackage --name notifyClient --structname mockNotifyClient
type notifyClient interface {
	Email(context.Context, notify.Email) (string, error)
}

func Handler(ctx context.Context, event events.CloudWatchEvent) error {
	tableName := os.Getenv("LPAS_TABLE")
	notifyIsProduction := os.Getenv("GOVUK_NOTIFY_IS_PRODUCTION") == "1"
	notifyBaseURL := os.Getenv("GOVUK_NOTIFY_BASE_URL")

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
	}

	db := dynamodb.NewFromConfig(cfg)

	secretsClient := secretsmanager.NewFromConfig(cfg)

	notifyApiKey, err := secretsClient.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String("gov-uk-notify-api-key"),
	})

	notifyClient, err := notify.New(notifyIsProduction, notifyBaseURL, *notifyApiKey.SecretString, http.DefaultClient)

	switch event.DetailType {
	case "evidence-received":
		return handleEvidenceReceived(ctx, db, tableName, event)
	case "fee-approved":
		return handleFeeApproved(ctx, db, tableName, event, notifyClient)
	default:
		return fmt.Errorf("unknown event received: %s", event.DetailType)
	}
}

func handleEvidenceReceived(ctx context.Context, db store, tableName string, event events.CloudWatchEvent) error {
	var v evidenceReceivedEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal 'evidence-received' detail: %w", err)
	}

	lpa, err := resolveUID(ctx, db, tableName, v.UID)
	if err != nil {
		return fmt.Errorf("failed to resolve uid for 'evidence-received': %w", err)
	}

	item, err := attributevalue.MarshalMap(map[string]any{"PK": lpa.PK, "SK": "#EVIDENCE_RECEIVED"})
	if err != nil {
		return fmt.Errorf("failed to marshal item in response to 'evidence-received': %w", err)
	}

	_, err = db.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})

	return err
}

func handleFeeApproved(ctx context.Context, db store, tableName string, event events.CloudWatchEvent, notifyClient notifyClient) error {
	var v feeApprovedEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal 'fee-approved' detail: %w", err)
	}

	lpa, err := resolveUID(ctx, db, tableName, v.UID)
	if err != nil {
		return fmt.Errorf("failed to resolve uid for 'fee-approved': %w", err)
	}

	if err != nil {
		return fmt.Errorf("failed to marshal item in response to 'fee-approved': %w", err)
	}

	_, err = db.UpdateItem(ctx, &dynamodb.UpdateItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"PK": &types.AttributeValueMemberS{Value: lpa.PK},
			"SK": &types.AttributeValueMemberS{Value: lpa.SK},
		},
		UpdateExpression: aws.String("SET #tasks.#payForLpa = :status"),
		ExpressionAttributeNames: map[string]string{
			"#tasks": "Tasks", "#payForLpa": "PayForLpa",
		},
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":status": &types.AttributeValueMemberN{Value: "5"},
		},
	})

	//awslocal dynamodb update-item
	//--table-name lpas
	//--key '{"PK": { "S": "LPA#6dd7e006-74e3-455c-a8af-cec490388da5" }, "SK": {"S": "#DONOR#czRoc3BtYlNtUXFpUnZqSw=="} }'
	//--update-expression 'SET #tasks.#payForLpa = :status'
	//--expression-attribute-names '{"#tasks": "Tasks", "#payForLpa": "PayForLpa"}'
	//--expression-attribute-values '{":status": {"N": "5"}}'

	return err
}

func resolveUID(ctx context.Context, db store, tableName, uid string) (primaryKey, error) {
	skey, err := attributevalue.Marshal(uid)
	if err != nil {
		return primaryKey{}, fmt.Errorf("failed to marshal UID: %w", err)
	}

	response, err := db.Query(ctx, &dynamodb.QueryInput{
		TableName:                 aws.String(tableName),
		IndexName:                 aws.String("UidIndex"),
		ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
		ExpressionAttributeValues: map[string]types.AttributeValue{":UID": skey},
		KeyConditionExpression:    aws.String("#UID = :UID"),
	})
	if err != nil {
		return primaryKey{}, fmt.Errorf("failed to query UID: %w", err)
	}
	if len(response.Items) != 1 {
		return primaryKey{}, fmt.Errorf("expected to resolve UID but got %d items", len(response.Items))
	}

	var pk primaryKey
	if err := attributevalue.UnmarshalMap(response.Items[0], &pk); err != nil {
		return primaryKey{}, fmt.Errorf("failed to unmarshal UID response: %w", err)
	}

	return pk, nil
}

func main() {
	lambda.Start(Handler)
}
