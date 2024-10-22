package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMakeRegisterHandlerHandleUnknownEvent(t *testing.T) {
	handler := &makeregisterEventHandler{}

	err := handler.Handle(ctx, nil, &events.CloudWatchEvent{DetailType: "some-event"})
	assert.Equal(t, fmt.Errorf("unknown makeregister event"), err)
}

func TestHandleUidRequestedDonor(t *testing.T) {
	e := &events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail: json.RawMessage(
			`{"lpaID":"lpa-id","donorSessionID":"donor-session-id","organisationID":"","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`,
		),
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
		Set(ctx, "lpa-id", "donor-session-id", "", "M-1111-2222-3333").
		Return(nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.
		On("One", ctx, dynamo.LpaKey("lpa-id"), dynamo.DonorKey("donor-session-id"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&donordata.Provided{
				Donor:     donordata.Donor{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "a"}, DateOfBirth: dob},
				Type:      lpadata.LpaTypePersonalWelfare,
				CreatedAt: testNow,
				PK:        dynamo.LpaKey("lpa-id"),
				SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("donor-session-id")),
			})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendApplicationUpdated(ctx, event.ApplicationUpdated{
			UID:       "M-1111-2222-3333",
			Type:      lpadata.LpaTypePersonalWelfare.String(),
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

func TestHandleUidRequestedOrganisation(t *testing.T) {
	e := &events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail: json.RawMessage(
			`{"lpaID":"lpa-id","donorSessionID":"","organisationID":"organisation-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`,
		),
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
		Set(ctx, "lpa-id", "", "organisation-id", "M-1111-2222-3333").
		Return(nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.
		On("One", ctx, dynamo.LpaKey("lpa-id"), dynamo.OrganisationKey("organisation-id"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&donordata.Provided{
				Donor:     donordata.Donor{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "a"}, DateOfBirth: dob},
				Type:      lpadata.LpaTypePersonalWelfare,
				CreatedAt: testNow,
				PK:        dynamo.LpaKey("lpa-id"),
				SK:        dynamo.LpaOwnerKey(dynamo.OrganisationKey("organisation-id")),
			})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendApplicationUpdated(ctx, event.ApplicationUpdated{
			UID:       "M-1111-2222-3333",
			Type:      lpadata.LpaTypePersonalWelfare.String(),
			CreatedAt: testNow,
			Donor: event.ApplicationUpdatedDonor{
				FirstNames:  "a",
				LastName:    "b",
				DateOfBirth: date.New("2000", "1", "2"),
				Address:     place.Address{Line1: "a"},
			},
		}).
		Return(nil)

	err := handleUidRequested(ctx, uidStore, uidClient, e, dynamoClient, eventClient)

	assert.Nil(t, err)
}

func TestHandleUidRequestedWhenLpaUIDExists(t *testing.T) {
	e := &events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail: json.RawMessage(
			`{"lpaID":"lpa-id","donorSessionID":"","organisationID":"organisation-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`,
		),
	}

	dob := date.New("2000", "01", "02")

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.
		On("One", ctx, dynamo.LpaKey("lpa-id"), dynamo.OrganisationKey("organisation-id"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&donordata.Provided{
				LpaUID:    "M-1111-2222-3333",
				Donor:     donordata.Donor{FirstNames: "a", LastName: "b", Address: place.Address{Line1: "a"}, DateOfBirth: dob},
				Type:      lpadata.LpaTypePersonalWelfare,
				CreatedAt: testNow,
				PK:        dynamo.LpaKey("lpa-id"),
				SK:        dynamo.LpaOwnerKey(dynamo.OrganisationKey("organisation-id")),
			})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	err := handleUidRequested(ctx, nil, nil, e, dynamoClient, nil)

	assert.Nil(t, err)
}

func TestHandleUidRequestedWhenDynamoClientError(t *testing.T) {
	e := &events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleUidRequested(ctx, nil, nil, e, dynamoClient, nil)
	assert.Equal(t, fmt.Errorf("failed to get donor: %w", expectedError), err)
}

func TestHandleUidRequestedWhenUidClientErrors(t *testing.T) {
	e := &events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, mock.Anything).
		Return("", expectedError)

	err := handleUidRequested(ctx, nil, uidClient, e, dynamoClient, nil)
	assert.Equal(t, fmt.Errorf("failed to create case: %w", expectedError), err)
}

func TestHandleUidRequestedWhenUidStoreErrors(t *testing.T) {
	e := &events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, mock.Anything).
		Return("", nil)

	uidStore := newMockUidStore(t)
	uidStore.EXPECT().
		Set(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleUidRequested(ctx, uidStore, uidClient, e, dynamoClient, nil)
	assert.Equal(t, fmt.Errorf("failed to set uid: %w", expectedError), err)
}

func TestHandleUidRequestedWhenEventClientErrors(t *testing.T) {
	e := &events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, mock.Anything).
		Return("", nil)

	uidStore := newMockUidStore(t)
	uidStore.EXPECT().
		Set(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendApplicationUpdated(ctx, mock.Anything).
		Return(expectedError)

	err := handleUidRequested(ctx, uidStore, uidClient, e, dynamoClient, eventClient)
	assert.Equal(t, expectedError, err)
}
