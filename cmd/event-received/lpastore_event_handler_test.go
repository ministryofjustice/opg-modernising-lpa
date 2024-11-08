package main

import (
	"context"
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

func TestLpaStoreEventHandlerHandleLpaUpdated(t *testing.T) {
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
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			json.Unmarshal(b, v)
			return nil
		})
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

func TestLpaStoreEventHandlerHandleLpaUpdatedWhenChangeTypeNotStatutoryWaitingPeriod(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"WHAT"}`),
	}

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(nil)
	factory.EXPECT().Now().Return(testNowFn)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, event)
	assert.Nil(t, err)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedWhenDynamoGetErrors(t *testing.T) {
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
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(client)
	factory.EXPECT().Now().Return(testNowFn)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, event)
	assert.ErrorIs(t, err, expectedError)
}

func TestLpaStoreEventHandlerHandleLpaUpdatedWhenDynamoPutErrors(t *testing.T) {
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
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(expectedError)

	factory := newMockFactory(t)
	factory.EXPECT().DynamoClient().Return(client)
	factory.EXPECT().Now().Return(testNowFn)

	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, factory, event)
	assert.ErrorIs(t, err, expectedError)
}
