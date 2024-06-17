package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSiriusEventHandlerHandleUnknownEvent(t *testing.T) {
	handler := &siriusEventHandler{}

	err := handler.Handle(ctx, nil, events.CloudWatchEvent{DetailType: "some-event"})
	assert.Equal(t, fmt.Errorf("unknown sirius event"), err)
}

func TestHandleEvidenceReceived(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "evidence-received",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123")})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, map[string]string{
			"PK": dynamo.LpaKey("123").PK(),
			"SK": dynamo.EvidenceReceivedKey().SK(),
		}).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		DynamoClient().
		Return(client)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, event)

	assert.Nil(t, err)
}

func TestHandleEvidenceReceivedWhenClientGetError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	err := handleEvidenceReceived(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to resolve uid: %w", expectedError), err)
}

func TestHandleEvidenceReceivedWhenLpaMissingPK(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	err := handleEvidenceReceived(ctx, client, event)
	assert.Equal(t, errors.New("PK missing from LPA in response"), err)
}

func TestHandleEvidenceReceivedWhenClientPutError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123")})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, map[string]string{
			"PK": dynamo.LpaKey("123").PK(),
			"SK": dynamo.EvidenceReceivedKey().SK(),
		}).
		Return(expectedError)

	err := handleEvidenceReceived(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to persist evidence received: %w", expectedError), err)
}

func TestHandleFeeApproved(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	donorProvided := actor.DonorProvidedDetails{
		PK:      dynamo.LpaKey("123"),
		SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		FeeType: pay.NoFee,
		Tasks:   actor.DonorTasks{PayForLpa: actor.PaymentTaskPending, ConfirmYourIdentityAndSign: actor.TaskCompleted},
	}

	completedDonorProvided := donorProvided
	completedDonorProvided.Tasks.PayForLpa = actor.PaymentTaskCompleted

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(ctx, &completedDonorProvided).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(ctx, page.AppData{}, &completedDonorProvided).
		Return(nil)

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&donorProvided)
			attributevalue.Unmarshal(b, v)
			return nil
		})

	now := time.Now()

	updatedDonorProvided := completedDonorProvided
	updatedDonorProvided.UpdateHash()
	updatedDonorProvided.UpdatedAt = now

	client.EXPECT().
		Put(ctx, &updatedDonorProvided).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		DynamoClient().
		Return(client)
	factory.EXPECT().
		AppData().
		Return(page.AppData{}, nil)
	factory.EXPECT().
		ShareCodeSender(ctx).
		Return(shareCodeSender, nil)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		DynamoClient().
		Return(client)
	factory.EXPECT().
		Now().
		Return(func() time.Time { return now })

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, event)

	assert.Nil(t, err)
}

func TestHandleFeeApprovedWhenNotPaid(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	donorProvided := actor.DonorProvidedDetails{
		PK:      dynamo.LpaKey("123"),
		SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		FeeType: pay.HalfFee,
		Tasks:   actor.DonorTasks{PayForLpa: actor.PaymentTaskPending, ConfirmYourIdentityAndSign: actor.TaskCompleted},
	}

	completedDonorProvided := donorProvided
	completedDonorProvided.Tasks.PayForLpa = actor.PaymentTaskApproved

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&donorProvided)
			attributevalue.Unmarshal(b, v)
			return nil
		})

	now := time.Now()

	updatedDonorProvided := completedDonorProvided
	updatedDonorProvided.UpdateHash()
	updatedDonorProvided.UpdatedAt = now

	client.EXPECT().
		Put(ctx, &updatedDonorProvided).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		DynamoClient().
		Return(client)
	factory.EXPECT().
		AppData().
		Return(page.AppData{}, nil)
	factory.EXPECT().
		ShareCodeSender(ctx).
		Return(nil, nil)
	factory.EXPECT().
		LpaStoreClient().
		Return(nil, nil)
	factory.EXPECT().
		DynamoClient().
		Return(client)
	factory.EXPECT().
		Now().
		Return(func() time.Time { return now })

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, event)

	assert.Nil(t, err)
}

func TestHandleFeeApprovedWhenNotSigned(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	donorProvided := actor.DonorProvidedDetails{
		PK:      dynamo.LpaKey("123"),
		SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		FeeType: pay.NoFee,
		Tasks:   actor.DonorTasks{PayForLpa: actor.PaymentTaskPending},
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&donorProvided)
			attributevalue.Unmarshal(b, v)
			return nil
		})

	now := time.Now()

	updatedDonorProvided := donorProvided
	updatedDonorProvided.Tasks.PayForLpa = actor.PaymentTaskCompleted
	updatedDonorProvided.UpdateHash()
	updatedDonorProvided.UpdatedAt = now

	client.EXPECT().
		Put(ctx, &updatedDonorProvided).
		Return(nil)

	err := handleFeeApproved(ctx, client, event, nil, nil, page.AppData{}, func() time.Time { return now })
	assert.Nil(t, err)
}

func TestHandleFeeApprovedWhenAlreadyPaidOrApproved(t *testing.T) {
	testcases := []actor.PaymentTask{
		actor.PaymentTaskCompleted,
		actor.PaymentTaskApproved,
	}

	for _, taskState := range testcases {
		t.Run(taskState.String(), func(t *testing.T) {
			event := events.CloudWatchEvent{
				DetailType: "reduced-fee-approved",
				Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
			}

			client := newMockDynamodbClient(t)
			client.
				On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
				Return(func(ctx context.Context, uid string, v interface{}) error {
					b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
					attributevalue.Unmarshal(b, v)
					return nil
				})
			client.
				On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
				Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
					b, _ := attributevalue.Marshal(&actor.DonorProvidedDetails{
						PK:      dynamo.LpaKey("123"),
						SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
						FeeType: pay.NoFee,
						Tasks:   actor.DonorTasks{PayForLpa: taskState},
					})
					attributevalue.Unmarshal(b, v)
					return nil
				})

			err := handleFeeApproved(ctx, client, event, nil, nil, page.AppData{}, nil)
			assert.Nil(t, err)
		})
	}
}

func TestHandleFeeApprovedWhenDynamoClientPutError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, nil, nil, page.AppData{}, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to update LPA task status: %w", expectedError), err)
}

func TestHandleFeeApprovedWhenShareCodeSenderError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(actor.DonorProvidedDetails{
				PK:      dynamo.LpaKey("123"),
				SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
				FeeType: pay.NoFee,
				Tasks:   actor.DonorTasks{PayForLpa: actor.PaymentTaskPending, ConfirmYourIdentityAndSign: actor.TaskCompleted},
			})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(mock.Anything, mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(ctx, page.AppData{}, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, shareCodeSender, lpaStoreClient, page.AppData{}, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to send share code to certificate provider: %w", expectedError), err)
}

func TestHandleFeeApprovedWhenLpaStoreError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(actor.DonorProvidedDetails{
				PK:      dynamo.LpaKey("123"),
				SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
				FeeType: pay.NoFee,
				Tasks:   actor.DonorTasks{PayForLpa: actor.PaymentTaskPending, ConfirmYourIdentityAndSign: actor.TaskCompleted},
			})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(ctx, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, nil, lpaStoreClient, page.AppData{}, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to send to lpastore: %w", expectedError), err)
}

func TestHandleFurtherInfoRequested(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "further-info-requested",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()
	updated := &actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired}, UpdatedAt: now}
	updated.UpdateHash()

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		DynamoClient().
		Return(client)
	factory.EXPECT().
		Now().
		Return(func() time.Time { return now })

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, event)

	assert.Nil(t, err)
}

func TestHandleFurtherInfoRequestedWhenPaymentTaskIsAlreadyMoreEvidenceRequired(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "further-info-requested",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&actor.DonorProvidedDetails{
				PK:    dynamo.LpaKey("123"),
				SK:    dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
				Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired},
			})

			attributevalue.Unmarshal(b, v)

			return nil
		})

	err := handleFurtherInfoRequested(ctx, client, event, func() time.Time { return now })
	assert.Nil(t, err)
}

func TestHandleFurtherInfoRequestedWhenPutError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "further-info-requested",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()
	updated := &actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired}, UpdatedAt: now}
	updated.UpdateHash()

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(expectedError)

	err := handleFurtherInfoRequested(ctx, client, event, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to update LPA task status: %w", expectedError), err)
}

func TestHandleFeeDenied(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "reduced-fee-declined",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()
	updated := &actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}, FeeType: pay.FullFee, UpdatedAt: now}
	updated.UpdateHash()

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		DynamoClient().
		Return(client)
	factory.EXPECT().
		Now().
		Return(func() time.Time { return now })

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, event)

	assert.Nil(t, err)
}

func TestHandleFeeDeniedWhenTaskAlreadyDenied(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "reduced-fee-declined",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	err := handleFeeDenied(ctx, client, event, func() time.Time { return now })
	assert.Nil(t, err)
}

func TestHandleFeeDeniedWhenPutError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "reduced-fee-declined",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(actor.DonorProvidedDetails{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(expectedError)

	err := handleFeeDenied(ctx, client, event, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to update LPA task status: %w", expectedError), err)
}

var donorSubmissionCompletedEvent = events.CloudWatchEvent{
	DetailType: "donor-submission-completed",
	Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
}

func TestHandleDonorSubmissionCompleted(t *testing.T) {
	appData := page.AppData{}
	uid := actoruid.New()

	lpa := &lpastore.Lpa{
		Donor: lpastore.Donor{FirstNames: "Dave", LastName: "Smith"},
		CertificateProvider: lpastore.CertificateProvider{
			Channel:    actor.ChannelOnline,
			UID:        uid,
			FirstNames: "John",
			LastName:   "Smith",
			Email:      "john@example.com",
		},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(dynamo.NotFoundError{})
	client.EXPECT().
		Put(ctx, &actor.DonorProvidedDetails{
			PK:        dynamo.LpaKey(testUuidString),
			SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
			LpaID:     testUuidString,
			LpaUID:    "M-1111-2222-3333",
			CreatedAt: testNow,
			Version:   1,
		}).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderInvite(ctx, appData, page.CertificateProviderInvite{
			DonorFirstNames:             "Dave",
			DonorFullName:               "Dave Smith",
			CertificateProviderUID:      uid,
			CertificateProviderFullName: "John Smith",
			CertificateProviderEmail:    "john@example.com",
		}).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		AppData().
		Return(appData, nil)
	factory.EXPECT().
		ShareCodeSender(ctx).
		Return(shareCodeSender, nil)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		DynamoClient().
		Return(client)
	factory.EXPECT().
		UuidString().
		Return(testUuidStringFn)
	factory.EXPECT().
		Now().
		Return(testNowFn)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, donorSubmissionCompletedEvent)

	assert.Nil(t, err)
}

func TestHandleDonorSubmissionCompletedWhenPaperCertificateProvider(t *testing.T) {
	appData := page.AppData{}

	lpa := &lpastore.Lpa{
		CertificateProvider: lpastore.CertificateProvider{
			Channel: actor.ChannelPaper,
		},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(dynamo.NotFoundError{})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	err := handleDonorSubmissionCompleted(ctx, client, donorSubmissionCompletedEvent, nil, appData, lpaStoreClient, testUuidStringFn, testNowFn)
	assert.Nil(t, err)
}

func TestHandleDonorSubmissionCompletedWhenDynamoExists(t *testing.T) {
	appData := page.AppData{}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil)

	err := handleDonorSubmissionCompleted(ctx, client, donorSubmissionCompletedEvent, nil, appData, nil, nil, nil)
	assert.Nil(t, err)
}

func TestHandleDonorSubmissionCompletedWhenDynamoOneByUIDError(t *testing.T) {
	appData := page.AppData{}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	err := handleDonorSubmissionCompleted(ctx, client, donorSubmissionCompletedEvent, nil, appData, nil, nil, nil)
	assert.Equal(t, expectedError, err)
}

func TestHandleDonorSubmissionCompletedWhenDynamoPutError(t *testing.T) {
	appData := page.AppData{}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(dynamo.NotFoundError{})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(expectedError)

	err := handleDonorSubmissionCompleted(ctx, client, donorSubmissionCompletedEvent, nil, appData, nil, testUuidStringFn, testNowFn)
	assert.Equal(t, expectedError, err)
}

func TestHandleDonorSubmissionCompletedWhenLpaStoreError(t *testing.T) {
	appData := page.AppData{}

	lpa := &lpastore.Lpa{
		CertificateProvider: lpastore.CertificateProvider{
			Channel: actor.ChannelOnline,
		},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(dynamo.NotFoundError{})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, expectedError)

	err := handleDonorSubmissionCompleted(ctx, client, donorSubmissionCompletedEvent, nil, appData, lpaStoreClient, testUuidStringFn, testNowFn)
	assert.Equal(t, expectedError, err)
}

func TestHandleDonorSubmissionCompletedWhenShareCodeSenderError(t *testing.T) {
	appData := page.AppData{}

	lpa := &lpastore.Lpa{
		CertificateProvider: lpastore.CertificateProvider{
			Channel: actor.ChannelOnline,
		},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(dynamo.NotFoundError{})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderInvite(ctx, mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleDonorSubmissionCompleted(ctx, client, donorSubmissionCompletedEvent, shareCodeSender, appData, lpaStoreClient, testUuidStringFn, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to send share code to certificate provider: %w", expectedError), err)
}

var certificateProviderSubmissionCompletedEvent = events.CloudWatchEvent{
	DetailType: "certificate-provider-submission-completed",
	Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
}

func TestHandleCertificateProviderSubmissionCompleted(t *testing.T) {
	appData := page.AppData{}

	lpa := &lpastore.Lpa{
		CertificateProvider: lpastore.CertificateProvider{
			Channel: actor.ChannelPaper,
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendAttorneys(ctx, appData, lpa).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		ShareCodeSender(ctx).
		Return(shareCodeSender, nil)
	factory.EXPECT().
		AppData().
		Return(appData, nil)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)

	assert.Nil(t, err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenOnline(t *testing.T) {
	lpa := &lpastore.Lpa{
		CertificateProvider: lpastore.CertificateProvider{
			Channel: actor.ChannelOnline,
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.Nil(t, err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenLpaStoreFactoryErrors(t *testing.T) {
	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(nil, expectedError)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, expectedError, err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenLpaStoreErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(nil, expectedError)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, fmt.Errorf("failed to retrieve lpa: %w", expectedError), err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenShareCodeSenderErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpastore.Lpa{
			CertificateProvider: lpastore.CertificateProvider{
				Channel: actor.ChannelPaper,
			},
		}, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendAttorneys(ctx, mock.Anything, mock.Anything).
		Return(expectedError)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		ShareCodeSender(ctx).
		Return(shareCodeSender, nil)
	factory.EXPECT().
		AppData().
		Return(page.AppData{}, nil)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, fmt.Errorf("failed to send share codes to attorneys: %w", expectedError), err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenShareCodeSenderFactoryErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpastore.Lpa{
			CertificateProvider: lpastore.CertificateProvider{
				Channel: actor.ChannelPaper,
			},
		}, nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		ShareCodeSender(ctx).
		Return(nil, expectedError)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, expectedError, err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenAppDataFactoryErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpastore.Lpa{
			CertificateProvider: lpastore.CertificateProvider{
				Channel: actor.ChannelPaper,
			},
		}, nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		ShareCodeSender(ctx).
		Return(nil, nil)
	factory.EXPECT().
		AppData().
		Return(page.AppData{}, expectedError)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, expectedError, err)
}
