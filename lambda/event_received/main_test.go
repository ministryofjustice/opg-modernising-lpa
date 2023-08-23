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
	"github.com/ministryofjustice/opg-modernising-lpa/shared/notify"
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

func TestHandleFeeApproved(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-approved",
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
				"SK": &types.AttributeValueMemberS{Value: "#DONOR#an-id"},
				"Tasks": &types.AttributeValueMemberM{
					Value: map[string]types.AttributeValue{"PayForLpa": &types.AttributeValueMemberN{Value: "3"}},
				},
			}},
		}, nil)
	store.
		On("UpdateItem", ctx, &dynamodb.UpdateItemInput{
			TableName: aws.String("lpas"),
			Key: map[string]types.AttributeValue{
				"PK": &types.AttributeValueMemberS{Value: "LPA#123"},
				"SK": &types.AttributeValueMemberS{Value: "#DONOR#an-id"},
			},
			UpdateExpression: aws.String("SET #tasks.#payForLpa = :status"),
			ExpressionAttributeNames: map[string]string{
				"#tasks": "Tasks", "#payForLpa": "PayForLpa",
			},
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":status": &types.AttributeValueMemberN{Value: "5"},
			},
		}).
		Return(nil, nil)

	notifyClient := newMockNotifyClient(t)
	notifyClient.On("Email", ctx, notify.Email{
		TemplateID:   "template-id",
		EmailAddress: lpa.CertificateProvider.Email,
		Personalisation: map[string]string{
			"cpFullName":                  lpa.CertificateProvider.FullName(),
			"donorFullName":               lpa.Donor.FullName(),
			"lpaLegalTerm":                appData.Localizer.T(lpa.Type.LegalTermTransKey()),
			"donorFirstNames":             lpa.Donor.FirstNames,
			"certificateProviderStartURL": fmt.Sprintf("%s%s", s.appPublicURL, Paths.CertificateProviderStart),
			"donorFirstNamesPossessive":   appData.Localizer.Possessive(lpa.Donor.FirstNames),
			"shareCode":                   shareCode,
		},
	})

	err := handleFeeApproved(ctx, store, "lpas", event, nil)
	assert.Nil(t, err)
}
