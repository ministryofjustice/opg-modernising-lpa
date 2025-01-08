package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestSiriusEventHandlerHandleUnknownEvent(t *testing.T) {
	handler := &siriusEventHandler{}

	err := handler.Handle(ctx, nil, &events.CloudWatchEvent{DetailType: "some-event"})
	assert.Equal(t, fmt.Errorf("unknown sirius event"), err)
}

func TestHandleEvidenceReceived(t *testing.T) {
	event := &events.CloudWatchEvent{
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
	event := &events.CloudWatchEvent{
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
	event := &events.CloudWatchEvent{
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
	event := &events.CloudWatchEvent{
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
	e := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

	donorProvided := donordata.Provided{
		PK:      dynamo.LpaKey("123"),
		SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		FeeType: pay.NoFee,
		Tasks:   donordata.Tasks{PayForLpa: task.PaymentStatePending, SignTheLpa: task.StateCompleted},
	}

	completedDonorProvided := donorProvided
	completedDonorProvided.Tasks.PayForLpa = task.PaymentStateCompleted

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(ctx, completedDonorProvided.LpaUID, lpastore.CreateLpaFromDonorProvided(&completedDonorProvided)).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCertificateProviderStarted(ctx, event.CertificateProviderStarted{
			UID: "M-1111-2222-3333",
		}).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(ctx, appcontext.Data{}, &completedDonorProvided).
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

	updatedDonorProvided := completedDonorProvided
	updatedDonorProvided.UpdateHash()
	updatedDonorProvided.UpdatedAt = testNow

	client.EXPECT().
		Put(ctx, &updatedDonorProvided).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		DynamoClient().
		Return(client)
	factory.EXPECT().
		AppData().
		Return(appcontext.Data{}, nil)
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
		Return(testNowFn)
	factory.EXPECT().
		EventClient().
		Return(eventClient)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, e)

	assert.Nil(t, err)
}

func TestHandleFeeApprovedWhenNotPaid(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"HalfFee"}`),
	}

	donorProvided := donordata.Provided{
		PK:      dynamo.LpaKey("123"),
		SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		FeeType: pay.HalfFee,
		Tasks:   donordata.Tasks{PayForLpa: task.PaymentStatePending, SignTheLpa: task.StateCompleted},
	}

	completedDonorProvided := donorProvided
	completedDonorProvided.Tasks.PayForLpa = task.PaymentStateApproved

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

	updatedDonorProvided := completedDonorProvided
	updatedDonorProvided.UpdateHash()
	updatedDonorProvided.UpdatedAt = testNow

	client.EXPECT().
		Put(ctx, &updatedDonorProvided).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		DynamoClient().
		Return(client)
	factory.EXPECT().
		AppData().
		Return(appcontext.Data{}, nil)
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
		Return(testNowFn)
	factory.EXPECT().
		EventClient().
		Return(nil)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, event)

	assert.Nil(t, err)
}

func TestHandleFeeApprovedWhenNotSigned(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

	donorProvided := donordata.Provided{
		PK:      dynamo.LpaKey("123"),
		SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		FeeType: pay.NoFee,
		Tasks:   donordata.Tasks{PayForLpa: task.PaymentStatePending},
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

	updatedDonorProvided := donorProvided
	updatedDonorProvided.Tasks.PayForLpa = task.PaymentStateCompleted
	updatedDonorProvided.UpdateHash()
	updatedDonorProvided.UpdatedAt = testNow

	client.EXPECT().
		Put(ctx, &updatedDonorProvided).
		Return(nil)

	err := handleFeeApproved(ctx, client, event, nil, nil, nil, appcontext.Data{}, testNowFn)
	assert.Nil(t, err)
}

func TestHandleFeeApprovedWhenAlreadyPaidOrApproved(t *testing.T) {
	testcases := []task.PaymentState{
		task.PaymentStateCompleted,
		task.PaymentStateApproved,
	}

	for _, taskState := range testcases {
		t.Run(taskState.String(), func(t *testing.T) {
			event := &events.CloudWatchEvent{
				DetailType: "reduced-fee-approved",
				Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
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
					b, _ := attributevalue.Marshal(donordata.Provided{
						PK:      dynamo.LpaKey("123"),
						SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
						FeeType: pay.NoFee,
						Tasks:   donordata.Tasks{PayForLpa: taskState},
					})
					attributevalue.Unmarshal(b, v)
					return nil
				})

			err := handleFeeApproved(ctx, client, event, nil, nil, nil, appcontext.Data{}, nil)
			assert.Nil(t, err)
		})
	}
}

func TestHandleFeeApprovedWhenApprovedTypeDiffers(t *testing.T) {
	testcases := map[string]struct {
		requestedFeeType              pay.FeeType
		requestedRepeatApplicationFee pay.CostOfRepeatApplication
		previousFeeType               pay.PreviousFee
		approvedFeeType               pay.FeeType
		updatedTaskState              task.PaymentState
		payment                       donordata.Payment
	}{
		"Requested HalfFee, got QuarterFee": {
			requestedFeeType: pay.HalfFee,
			approvedFeeType:  pay.QuarterFee,
			updatedTaskState: task.PaymentStateCompleted,
			payment:          donordata.Payment{Amount: 4100},
		},
		"Requested HalfFee, got NoFee": {
			requestedFeeType: pay.HalfFee,
			approvedFeeType:  pay.NoFee,
			updatedTaskState: task.PaymentStateCompleted,
			payment:          donordata.Payment{Amount: 4100},
		},
		"Requested NoFee, got HalfFee": {
			requestedFeeType: pay.NoFee,
			approvedFeeType:  pay.HalfFee,
			updatedTaskState: task.PaymentStateApproved,
		},
		"Requested NoFee, got QuarterFee": {
			requestedFeeType: pay.NoFee,
			approvedFeeType:  pay.QuarterFee,
			updatedTaskState: task.PaymentStateApproved,
		},
		"Requested HalfFee RepeatApplicationFee (previously paid FullFee), got QuarterFee": {
			requestedFeeType:              pay.RepeatApplicationFee,
			requestedRepeatApplicationFee: pay.CostOfRepeatApplicationHalfFee,
			previousFeeType:               pay.PreviousFeeFull,
			approvedFeeType:               pay.QuarterFee,
			updatedTaskState:              task.PaymentStateCompleted,
			payment:                       donordata.Payment{Amount: 4100},
		},
		"Requested HalfFee RepeatApplicationFee (previously paid FullFee), got NoFee": {
			requestedFeeType:              pay.RepeatApplicationFee,
			requestedRepeatApplicationFee: pay.CostOfRepeatApplicationHalfFee,
			previousFeeType:               pay.PreviousFeeFull,
			approvedFeeType:               pay.NoFee,
			updatedTaskState:              task.PaymentStateCompleted,
			payment:                       donordata.Payment{Amount: 4100},
		},
		"Requested HalfFee RepeatApplicationFee (previously paid HalfFee), got QuarterFee": {
			requestedFeeType:              pay.RepeatApplicationFee,
			requestedRepeatApplicationFee: pay.CostOfRepeatApplicationHalfFee,
			previousFeeType:               pay.PreviousFeeHalf,
			approvedFeeType:               pay.QuarterFee,
			updatedTaskState:              task.PaymentStateCompleted,
			payment:                       donordata.Payment{Amount: 4100},
		},
		"Requested HalfFee RepeatApplicationFee (previously paid HalfFee), got NoFee": {
			requestedFeeType:              pay.RepeatApplicationFee,
			requestedRepeatApplicationFee: pay.CostOfRepeatApplicationHalfFee,
			previousFeeType:               pay.PreviousFeeHalf,
			approvedFeeType:               pay.NoFee,
			updatedTaskState:              task.PaymentStateCompleted,
			payment:                       donordata.Payment{Amount: 4100},
		},
		"Requested HalfFee RepeatApplicationFee (previously paid NoFee), got HalfFee": {
			requestedFeeType:              pay.RepeatApplicationFee,
			requestedRepeatApplicationFee: pay.CostOfRepeatApplicationHalfFee,
			previousFeeType:               pay.PreviousFeeExemption,
			approvedFeeType:               pay.HalfFee,
			updatedTaskState:              task.PaymentStateApproved,
		},
		"Requested HalfFee RepeatApplicationFee (previously paid NoFee), got NoFee": {
			requestedFeeType:              pay.RepeatApplicationFee,
			requestedRepeatApplicationFee: pay.CostOfRepeatApplicationHalfFee,
			previousFeeType:               pay.PreviousFeeExemption,
			approvedFeeType:               pay.NoFee,
			updatedTaskState:              task.PaymentStateCompleted,
		},
		"Requested NoFee RepeatApplicationFee, got HalfFee": {
			requestedFeeType:              pay.RepeatApplicationFee,
			requestedRepeatApplicationFee: pay.CostOfRepeatApplicationNoFee,
			approvedFeeType:               pay.HalfFee,
			updatedTaskState:              task.PaymentStateApproved,
		},
		"Requested NoFee RepeatApplicationFee, got QuarterFee": {
			requestedFeeType:              pay.RepeatApplicationFee,
			requestedRepeatApplicationFee: pay.CostOfRepeatApplicationNoFee,
			approvedFeeType:               pay.QuarterFee,
			updatedTaskState:              task.PaymentStateApproved,
		},
		"Requested HardshipFee, got HalfFee": {
			requestedFeeType: pay.HardshipFee,
			approvedFeeType:  pay.HalfFee,
			updatedTaskState: task.PaymentStateApproved,
		},
		"Requested HardshipFee, got QuarterFee": {
			requestedFeeType: pay.HardshipFee,
			approvedFeeType:  pay.QuarterFee,
			updatedTaskState: task.PaymentStateApproved,
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			event := &events.CloudWatchEvent{
				DetailType: "reduced-fee-approved",
				Detail:     json.RawMessage(fmt.Sprintf(`{"uid":"M-1111-2222-3333","approvedType":"%s"}`, tc.approvedFeeType.String())),
			}

			donorProvided := &donordata.Provided{
				PK:                      dynamo.LpaKey("123"),
				SK:                      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
				FeeType:                 tc.requestedFeeType,
				Tasks:                   donordata.Tasks{PayForLpa: task.PaymentStatePending},
				PaymentDetails:          []donordata.Payment{tc.payment},
				PreviousFee:             tc.previousFeeType,
				CostOfRepeatApplication: tc.requestedRepeatApplicationFee,
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
					b, _ := attributevalue.Marshal(donorProvided)
					attributevalue.Unmarshal(b, v)
					return nil
				})

			updatedDonorProvided := *donorProvided
			updatedDonorProvided.Tasks.PayForLpa = tc.updatedTaskState
			updatedDonorProvided.FeeType = tc.approvedFeeType
			updatedDonorProvided.UpdateHash()
			updatedDonorProvided.UpdatedAt = testNow

			client.EXPECT().
				Put(ctx, &updatedDonorProvided).
				Return(nil)

			err := handleFeeApproved(ctx, client, event, nil, nil, nil, appcontext.Data{}, testNowFn)
			assert.Nil(t, err)
		})
	}
}

func TestHandleFeeApprovedWhenDynamoClientPutError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := attributevalue.Marshal(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: donordata.Tasks{PayForLpa: task.PaymentStatePending}})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, nil, nil, nil, appcontext.Data{}, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to update LPA task status: %w", expectedError), err)
}

func TestHandleFeeApprovedWhenShareCodeSenderError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

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
			b, _ := attributevalue.Marshal(donordata.Provided{
				PK:      dynamo.LpaKey("123"),
				SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
				FeeType: pay.NoFee,
				Tasks:   donordata.Tasks{PayForLpa: task.PaymentStatePending, SignTheLpa: task.StateCompleted},
			})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCertificateProviderStarted(mock.Anything, mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(ctx, appcontext.Data{}, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, shareCodeSender, lpaStoreClient, eventClient, appcontext.Data{}, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to send share code to certificate provider: %w", expectedError), err)
}

func TestHandleFeeApprovedWhenEventClientError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

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
			b, _ := attributevalue.Marshal(donordata.Provided{
				PK:      dynamo.LpaKey("123"),
				SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
				FeeType: pay.NoFee,
				Tasks:   donordata.Tasks{PayForLpa: task.PaymentStatePending, SignTheLpa: task.StateCompleted},
			})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCertificateProviderStarted(mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, nil, lpaStoreClient, eventClient, appcontext.Data{}, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to send certificate-provider-started event: %w", expectedError), err)
}

func TestHandleFeeApprovedWhenLpaStoreError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

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
			b, _ := attributevalue.Marshal(donordata.Provided{
				PK:      dynamo.LpaKey("123"),
				SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
				FeeType: pay.NoFee,
				Tasks:   donordata.Tasks{PayForLpa: task.PaymentStatePending, SignTheLpa: task.StateCompleted},
			})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(ctx, mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, nil, lpaStoreClient, nil, appcontext.Data{}, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to send to lpastore: %w", expectedError), err)
}

func TestHandleFurtherInfoRequested(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "further-info-requested",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	updated := &donordata.Provided{
		PK:                     dynamo.LpaKey("123"),
		SK:                     dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		Tasks:                  donordata.Tasks{PayForLpa: task.PaymentStateMoreEvidenceRequired},
		UpdatedAt:              testNow,
		MoreEvidenceRequiredAt: testNow,
	}
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
			b, _ := attributevalue.Marshal(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: donordata.Tasks{PayForLpa: task.PaymentStatePending}})
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
		Return(testNowFn)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, event)

	assert.Nil(t, err)
}

func TestHandleFurtherInfoRequestedWhenPaymentTaskIsAlreadyMoreEvidenceRequired(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "further-info-requested",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

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
			b, _ := attributevalue.Marshal(&donordata.Provided{
				PK:    dynamo.LpaKey("123"),
				SK:    dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
				Tasks: donordata.Tasks{PayForLpa: task.PaymentStateMoreEvidenceRequired},
			})

			attributevalue.Unmarshal(b, v)

			return nil
		})

	err := handleFurtherInfoRequested(ctx, client, event, testNowFn)
	assert.Nil(t, err)
}

func TestHandleFurtherInfoRequestedWhenPutError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "further-info-requested",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	updated := &donordata.Provided{
		PK:                     dynamo.LpaKey("123"),
		SK:                     dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		Tasks:                  donordata.Tasks{PayForLpa: task.PaymentStateMoreEvidenceRequired},
		UpdatedAt:              testNow,
		MoreEvidenceRequiredAt: testNow,
	}
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
			b, _ := attributevalue.Marshal(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: donordata.Tasks{PayForLpa: task.PaymentStatePending}})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(expectedError)

	err := handleFurtherInfoRequested(ctx, client, event, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to update LPA task status: %w", expectedError), err)
}

func TestHandleFeeDenied(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-declined",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	updated := &donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: donordata.Tasks{PayForLpa: task.PaymentStateDenied}, FeeType: pay.FullFee, UpdatedAt: testNow}
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
			b, _ := attributevalue.Marshal(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: donordata.Tasks{PayForLpa: task.PaymentStatePending}})
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
		Return(testNowFn)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, event)

	assert.Nil(t, err)
}

func TestHandleFeeDeniedWhenTaskAlreadyDenied(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-declined",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

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
			b, _ := attributevalue.Marshal(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: donordata.Tasks{PayForLpa: task.PaymentStateDenied}})
			attributevalue.Unmarshal(b, v)
			return nil
		})

	err := handleFeeDenied(ctx, client, event, testNowFn)
	assert.Nil(t, err)
}

func TestHandleFeeDeniedWhenPutError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-declined",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

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
			b, _ := attributevalue.Marshal(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), Tasks: donordata.Tasks{PayForLpa: task.PaymentStatePending}})
			attributevalue.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(expectedError)

	err := handleFeeDenied(ctx, client, event, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to update LPA task status: %w", expectedError), err)
}

var donorSubmissionCompletedEvent = &events.CloudWatchEvent{
	DetailType: "donor-submission-completed",
	Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
}

func TestHandleDonorSubmissionCompleted(t *testing.T) {
	appData := appcontext.Data{}
	uid := actoruid.New()

	lpa := &lpadata.Lpa{
		Donor: lpadata.Donor{FirstNames: "Dave", LastName: "Smith"},
		CertificateProvider: lpadata.CertificateProvider{
			Channel:    lpadata.ChannelOnline,
			UID:        uid,
			FirstNames: "John",
			LastName:   "Smith",
			Email:      "john@example.com",
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderInvite(ctx, appData, sharecode.CertificateProviderInvite{
			DonorFirstNames:             "Dave",
			DonorFullName:               "Dave Smith",
			CertificateProviderUID:      uid,
			CertificateProviderFullName: "John Smith",
		}, notify.ToLpaCertificateProvider(&certificateproviderdata.Provided{ContactLanguagePreference: localize.En}, lpa)).
		Return(nil)

	client := newMockDynamodbClient(t)
	client.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Creates: []any{
				&donordata.Provided{
					PK:                           dynamo.LpaKey(testUuidString),
					SK:                           dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
					LpaID:                        testUuidString,
					LpaUID:                       "M-1111-2222-3333",
					CreatedAt:                    testNow,
					Version:                      1,
					CertificateProviderInvitedAt: testNow,
				},
				scheduled.Event{
					PK:                dynamo.ScheduledDayKey(testNow.AddDate(0, 3, 1)),
					SK:                dynamo.ScheduledKey(testNow.AddDate(0, 3, 1), testUuidString),
					CreatedAt:         testNow,
					At:                testNow.AddDate(0, 3, 1),
					Action:            scheduled.ActionRemindCertificateProviderToComplete,
					TargetLpaKey:      dynamo.LpaKey(testUuidString),
					TargetLpaOwnerKey: dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")),
					LpaUID:            "M-1111-2222-3333",
				},
				dynamo.Keys{PK: dynamo.UIDKey("M-1111-2222-3333"), SK: dynamo.MetadataKey("")},
				dynamo.Keys{PK: dynamo.LpaKey(testUuidString), SK: dynamo.ReservedKey(dynamo.DonorKey)},
			},
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

func TestHandleDonorSubmissionCompletedWhenWriteTransactionError(t *testing.T) {
	appData := appcontext.Data{}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderInvite(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	client := newMockDynamodbClient(t)
	client.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleDonorSubmissionCompleted(ctx, client, donorSubmissionCompletedEvent, shareCodeSender, appData, lpaStoreClient, testUuidStringFn, testNowFn)
	assert.Equal(t, expectedError, err)
}

func TestHandleDonorSubmissionCompletedWhenLpaStoreError(t *testing.T) {
	appData := appcontext.Data{}

	lpa := &lpadata.Lpa{
		CertificateProvider: lpadata.CertificateProvider{
			Channel: lpadata.ChannelOnline,
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, expectedError)

	err := handleDonorSubmissionCompleted(ctx, nil, donorSubmissionCompletedEvent, nil, appData, lpaStoreClient, testUuidStringFn, testNowFn)
	assert.Equal(t, expectedError, err)
}

func TestHandleDonorSubmissionCompletedWhenShareCodeSenderError(t *testing.T) {
	appData := appcontext.Data{}

	lpa := &lpadata.Lpa{
		CertificateProvider: lpadata.CertificateProvider{
			Channel: lpadata.ChannelOnline,
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderInvite(ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleDonorSubmissionCompleted(ctx, nil, donorSubmissionCompletedEvent, shareCodeSender, appData, lpaStoreClient, testUuidStringFn, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to send share code to certificate provider: %w", expectedError), err)
}

var certificateProviderSubmissionCompletedEvent = &events.CloudWatchEvent{
	DetailType: "certificate-provider-submission-completed",
	Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
}

func TestHandleCertificateProviderSubmissionCompleted(t *testing.T) {
	appData := appcontext.Data{}

	lpa := &lpadata.Lpa{
		CertificateProvider: lpadata.CertificateProvider{
			Channel: lpadata.ChannelPaper,
		},
	}

	updatedDonor := &donordata.Provided{
		PK:                 dynamo.LpaKey("an-lpa"),
		UpdatedAt:          testNow,
		AttorneysInvitedAt: testNow,
	}
	updatedDonor.UpdateHash()

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(lpa, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil).
		SetData(&donordata.Provided{PK: dynamo.LpaKey("an-lpa"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor"))})
	dynamoClient.EXPECT().
		One(ctx, dynamo.LpaKey("an-lpa"), dynamo.DonorKey("a-donor"), mock.Anything).
		Return(nil).
		SetData(&donordata.Provided{PK: dynamo.LpaKey("an-lpa")})
	dynamoClient.EXPECT().
		Put(ctx, updatedDonor).
		Return(nil)

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
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)
	factory.EXPECT().
		Now().
		Return(testNowFn)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)

	assert.Nil(t, err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenOnline(t *testing.T) {
	lpa := &lpadata.Lpa{
		CertificateProvider: lpadata.CertificateProvider{
			Channel: lpadata.ChannelOnline,
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

func TestHandleCertificateProviderSubmissionCompletedWhenDonorGetErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpadata.Lpa{
			CertificateProvider: lpadata.CertificateProvider{
				Channel: lpadata.ChannelPaper,
			},
		}, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	shareCodeSender := newMockShareCodeSender(t)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		ShareCodeSender(ctx).
		Return(shareCodeSender, nil)
	factory.EXPECT().
		AppData().
		Return(appcontext.Data{}, nil)
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandleCertificateProviderSubmissionCompletedWhenDonorPutErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpadata.Lpa{
			CertificateProvider: lpadata.CertificateProvider{
				Channel: lpadata.ChannelPaper,
			},
		}, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData(&donordata.Provided{PK: dynamo.LpaKey("an-lpa"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor"))})
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	dynamoClient.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendAttorneys(ctx, mock.Anything, mock.Anything).
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
		Return(appcontext.Data{}, nil)
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)
	factory.EXPECT().
		Now().
		Return(testNowFn)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandleCertificateProviderSubmissionCompletedWhenShareCodeSenderErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpadata.Lpa{
			CertificateProvider: lpadata.CertificateProvider{
				Channel: lpadata.ChannelPaper,
			},
		}, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil).
		SetData(&donordata.Provided{PK: dynamo.LpaKey("an-lpa"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor"))})
	dynamoClient.EXPECT().
		One(ctx, dynamo.LpaKey("an-lpa"), dynamo.DonorKey("a-donor"), mock.Anything).
		Return(nil).
		SetData(&donordata.Provided{PK: dynamo.LpaKey("an-lpa")})

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
		Return(appcontext.Data{}, nil)
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)
	factory.EXPECT().
		Now().
		Return(testNowFn)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, fmt.Errorf("failed to send share codes to attorneys: %w", expectedError), err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenShareCodeSenderFactoryErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpadata.Lpa{
			CertificateProvider: lpadata.CertificateProvider{
				Channel: lpadata.ChannelPaper,
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
		Return(&lpadata.Lpa{
			CertificateProvider: lpadata.CertificateProvider{
				Channel: lpadata.ChannelPaper,
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
		Return(appcontext.Data{}, expectedError)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, expectedError, err)
}
