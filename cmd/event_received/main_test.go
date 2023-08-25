package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func TestHandleEvidenceReceived(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("GetOneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(page.Lpa{PK: "LPA#123"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("lpas"),
			Item: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "LPA#123"},
				"SK": &types.AttributeValueMemberS{Value: "#EVIDENCE_RECEIVED"},
			},
		}).
		Return(nil)

	err := handleEvidenceReceived(ctx, client, "lpas", event)
	assert.Nil(t, err)
}

func TestHandleEvidenceReceivedWhenClientGetError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("GetOneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	err := handleEvidenceReceived(ctx, client, "lpas", event)
	assert.Equal(t, fmt.Errorf("failed to resolve uid for 'evidence-received': %w", expectedError), err)
}

func TestHandleEvidenceReceivedWhenClientPutError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("GetOneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(page.Lpa{PK: "LPA#123"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("lpas"),
			Item: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "LPA#123"},
				"SK": &types.AttributeValueMemberS{Value: "#EVIDENCE_RECEIVED"},
			},
		}).
		Return(expectedError)

	err := handleEvidenceReceived(ctx, client, "lpas", event)
	assert.Equal(t, expectedError, err)
}
