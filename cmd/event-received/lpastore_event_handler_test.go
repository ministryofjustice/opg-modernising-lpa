package main

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestLpaStoreEventHandlerHandleUnknownEvent(t *testing.T) {
	handler := &lpastoreEventHandler{}

	err := handler.Handle(ctx, nil, events.CloudWatchEvent{DetailType: "some-event"})
	assert.Equal(t, fmt.Errorf("unknown cloudwatch event"), err)
}

func TestHandleLpaUpdated(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"PERFECT"}`),
	}

	updated := &actor.DonorProvidedDetails{
		PK:        dynamo.LpaKey("123"),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		PerfectAt: testNow,
		UpdatedAt: testNow,
	}
	updated.Hash, _ = updated.GenerateHash()

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
			b, _ := json.Marshal(actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(nil)

	err := handleLpaUpdated(ctx, client, event, testNowFn)
	assert.Nil(t, err)
}

func TestHandleLpaUpdatedWhenChangeTypeNotPerfect(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"WHAT"}`),
	}

	err := handleLpaUpdated(ctx, nil, event, nil)
	assert.Nil(t, err)
}

func TestHandleLpaUpdatedWhenDynamoGetErrors(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"PERFECT"}`),
	}

	updated := &actor.DonorProvidedDetails{
		PK:        dynamo.LpaKey("123"),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		PerfectAt: testNow,
		UpdatedAt: testNow,
	}
	updated.Hash, _ = updated.GenerateHash()

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	err := handleLpaUpdated(ctx, client, event, testNowFn)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandleLpaUpdatedWhenDynamoPutErrors(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "lpa-updated",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","changeType":"PERFECT"}`),
	}

	updated := &actor.DonorProvidedDetails{
		PK:        dynamo.LpaKey("123"),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		PerfectAt: testNow,
		UpdatedAt: testNow,
	}
	updated.Hash, _ = updated.GenerateHash()

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
			b, _ := json.Marshal(actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(expectedError)

	err := handleLpaUpdated(ctx, client, event, testNowFn)
	assert.ErrorIs(t, err, expectedError)
}
