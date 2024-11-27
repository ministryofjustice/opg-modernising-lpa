package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLpaStoreEventHandlerHandleUnknownEvent(t *testing.T) {
	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, nil, &events.CloudWatchEvent{DetailType: "some-event"})
	assert.Equal(t, fmt.Errorf("unknown lpastore event"), err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedWhenChangeTypeNotExpected(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"WHAT"}`),
	}

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, nil, event)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedStatutoryWaitingPeriod(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"STATUTORY_WAITING_PERIOD"}`),
	}

	updated := &donordata.Provided{
		PK:                       dynamo.LpaKey("123"),
		SK:                       dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		StatutoryWaitingPeriodAt: testNow,
		UpdatedAt:                testNow,
	}
	updated.UpdateHash()

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil).
		SetData(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
	client.EXPECT().
		One(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(nil).
		SetData(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
	client.EXPECT().
		Put(ctx, updated).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(client)
	factory.EXPECT().Now().Return(testNowFn)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, event)
	assert.Nil(t, err)
}

func TestHandleStatutoryWaitingPeriodWhenDynamoErrors(t *testing.T) {
	updated := &donordata.Provided{
		PK:                       dynamo.LpaKey("123"),
		SK:                       dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		StatutoryWaitingPeriodAt: testNow,
		UpdatedAt:                testNow,
	}
	updated.UpdateHash()

	testcases := map[string]struct {
		dynamoClient  func() *mockDynamodbClient
		expectedError error
	}{
		"OneByUID": {
			dynamoClient: func() *mockDynamodbClient {
				client := newMockDynamodbClient(t)
				client.EXPECT().
					OneByUID(ctx, mock.Anything, mock.Anything).
					Return(expectedError)

				return client
			},
			expectedError: fmt.Errorf("failed to resolve uid: %w", expectedError),
		},
		"One": {
			dynamoClient: func() *mockDynamodbClient {
				client := newMockDynamodbClient(t)
				client.EXPECT().
					OneByUID(mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					SetData(dynamo.Keys{PK: dynamo.LpaKey("pk"), SK: dynamo.DonorKey("sk")})
				client.EXPECT().
					One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(expectedError)

				return client
			},
			expectedError: fmt.Errorf("failed to get LPA: %w", expectedError),
		},
		"Put": {
			dynamoClient: func() *mockDynamodbClient {
				client := newMockDynamodbClient(t)
				client.EXPECT().
					OneByUID(mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					SetData(dynamo.Keys{PK: dynamo.LpaKey("pk"), SK: dynamo.DonorKey("sk")})
				client.EXPECT().
					One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).
					SetData(updated)
				client.EXPECT().
					Put(mock.Anything, updated).
					Return(expectedError)

				return client
			},
			expectedError: fmt.Errorf("failed to update donor details: %w", expectedError),
		},
	}

	for testName, tc := range testcases {
		t.Run(testName, func(t *testing.T) {
			event := lpaUpdatedEvent{
				UID:        "M-1111-2222-3333",
				ChangeType: "STATUTORY_WAITING_PERIOD",
			}

			err := handleStatutoryWaitingPeriod(ctx, tc.dynamoClient(), testNowFn, event)
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestLpaStoreEventHandlerHandleLpaUpdatedCannotRegister(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"CANNOT_REGISTER"}`),
	}

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllByUID(ctx, "M-1111-2222-3333").
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().ScheduledStore().Return(scheduledStore)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, event)
	assert.Nil(t, err)
}

func TestHandleCannotRegisterWhenStoreErrors(t *testing.T) {
	event := lpaUpdatedEvent{
		UID:        "M-1111-2222-3333",
		ChangeType: "CANNOT_REGISTER",
	}

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllByUID(mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleCannotRegister(ctx, scheduledStore, event)
	assert.ErrorIs(t, err, expectedError)
}
