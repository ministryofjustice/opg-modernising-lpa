package main

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/stretchr/testify/assert"
)

var expectedError = errors.New("err")

func TestHandleEvidenceReceived(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	store := newMockStore(t)
	store.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("lpas"),
			IndexName:                 aws.String("UidIndex"),
			ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#UID = :UID"),
		}).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{{
				"PK": &types.AttributeValueMemberS{Value: "LPA#123"},
			}},
		}, nil)
	store.
		On("PutItem", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("lpas"),
			Item: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "LPA#123"},
				"SK": &types.AttributeValueMemberS{Value: "#EVIDENCE_RECEIVED"},
			},
		}).
		Return(nil, nil)

	err := handleEvidenceReceived(ctx, store, "lpas", event)
	assert.Nil(t, err)
}

func TestHandleEvidenceReceivedWhenUIDNotFound(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	store := newMockStore(t)
	store.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("lpas"),
			IndexName:                 aws.String("UidIndex"),
			ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#UID = :UID"),
		}).
		Return(&dynamodb.QueryOutput{}, nil)

	err := handleEvidenceReceived(ctx, store, "lpas", event)
	assert.EqualError(t, err, "failed to resolve uid for 'evidence-received': expected to resolve UID but got 0 items")
}

func TestHandleEvidenceReceivedWhenUIDErrors(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	store := newMockStore(t)
	store.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("lpas"),
			IndexName:                 aws.String("UidIndex"),
			ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#UID = :UID"),
		}).
		Return(&dynamodb.QueryOutput{}, expectedError)

	err := handleEvidenceReceived(ctx, store, "lpas", event)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandleEvidenceReceivedWhenPutItemErrors(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	store := newMockStore(t)
	store.
		On("Query", ctx, &dynamodb.QueryInput{
			TableName:                 aws.String("lpas"),
			IndexName:                 aws.String("UidIndex"),
			ExpressionAttributeNames:  map[string]string{"#UID": "UID"},
			ExpressionAttributeValues: map[string]types.AttributeValue{":UID": &types.AttributeValueMemberS{Value: "M-1111-2222-3333"}},
			KeyConditionExpression:    aws.String("#UID = :UID"),
		}).
		Return(&dynamodb.QueryOutput{
			Items: []map[string]types.AttributeValue{{
				"PK": &types.AttributeValueMemberS{Value: "LPA#123"},
			}},
		}, nil)
	store.
		On("PutItem", ctx, &dynamodb.PutItemInput{
			TableName: aws.String("lpas"),
			Item: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "LPA#123"},
				"SK": &types.AttributeValueMemberS{Value: "#EVIDENCE_RECEIVED"},
			},
		}).
		Return(nil, expectedError)

	err := handleEvidenceReceived(ctx, store, "lpas", event)
	assert.ErrorIs(t, err, expectedError)
}
