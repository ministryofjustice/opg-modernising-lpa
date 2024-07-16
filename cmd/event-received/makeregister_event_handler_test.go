package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMakeRegisterHandlerHandleUnknownEvent(t *testing.T) {
	handler := &makeregisterEventHandler{}

	err := handler.Handle(ctx, nil, events.CloudWatchEvent{DetailType: "some-event"})
	assert.Equal(t, fmt.Errorf("unknown makeregister event"), err)
}

func TestHandleUidRequested(t *testing.T) {
	e := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","organisationID":"org-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	dob := date.New("2000", "01", "02")

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, &uid.CreateCaseRequestBody{
			Type: "personal-welfare",
			Donor: uid.DonorDetails{
				Name:     "a donor",
				Dob:      dob,
				Postcode: "F1 1FF",
			},
		}).
		Return("M-1111-2222-3333", nil)

	uidStore := newMockUidStore(t)
	uidStore.EXPECT().
		Set(ctx, "an-id", "donor-id", "org-id", "M-1111-2222-3333").
		Return(nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	dynamoClient.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&actor.DonorProvidedDetails{
				Donor:     actor.Donor{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "a"}, DateOfBirth: dob},
				Type:      actor.LpaTypePersonalWelfare,
				CreatedAt: testNow,
				LpaUID:    "M-1111-2222-3333",
				PK:        dynamo.LpaKey("123"),
				SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
			})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendApplicationUpdated(ctx, event.ApplicationUpdated{
			UID:       "M-1111-2222-3333",
			Type:      actor.LpaTypePersonalWelfare.String(),
			CreatedAt: testNow,
			Donor: event.ApplicationUpdatedDonor{
				FirstNames:  "a",
				LastName:    "b",
				DateOfBirth: date.New("2000", "1", "2"),
				Address:     place.Address{Line1: "a"},
			},
		}).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		UidStore().
		Return(uidStore, nil)
	factory.EXPECT().
		UidClient().
		Return(uidClient)
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)
	factory.EXPECT().
		EventClient().
		Return(eventClient)

	handler := makeregisterEventHandler{}
	err := handler.Handle(ctx, factory, e)

	assert.Nil(t, err)
}

func TestHandleUidRequestedWhenUidClientErrors(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, mock.Anything).
		Return("", expectedError)

	err := handleUidRequested(ctx, nil, uidClient, event, nil, nil)
	assert.Equal(t, fmt.Errorf("failed to create case: %w", expectedError), err)
}

func TestHandleUidRequestedWhenUidStoreErrors(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, mock.Anything).
		Return("M-1111-2222-3333", nil)

	uidStore := newMockUidStore(t)
	uidStore.EXPECT().
		Set(ctx, "an-id", "donor-id", "", "M-1111-2222-3333").
		Return(expectedError)

	err := handleUidRequested(ctx, uidStore, uidClient, event, nil, nil)
	assert.Equal(t, fmt.Errorf("failed to set uid: %w", expectedError), err)
}

func TestHandleUidRequestedWhenEventClientErrors(t *testing.T) {
	e := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, mock.Anything).
		Return("M-1111-2222-3333", nil)

	uidStore := newMockUidStore(t)
	uidStore.EXPECT().
		Set(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.
		On("OneByUID", mock.Anything, mock.Anything, mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	dynamoClient.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&actor.DonorProvidedDetails{})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendApplicationUpdated(ctx, mock.Anything).
		Return(expectedError)

	err := handleUidRequested(ctx, uidStore, uidClient, e, dynamoClient, eventClient)
	assert.Equal(t, expectedError, err)
}
