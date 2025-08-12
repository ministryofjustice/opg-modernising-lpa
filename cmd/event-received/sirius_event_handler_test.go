package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled/scheduleddata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	unusedCertificateProviderStore = func(t *testing.T) *mockCertificateProviderStore {
		return newMockCertificateProviderStore(t)
	}
	unusedDynamoClient = func(t *testing.T) *mockDynamodbClient {
		return newMockDynamodbClient(t)
	}
	unusedLpaStoreClient = func(t *testing.T) *mockLpaStoreClient {
		return newMockLpaStoreClient(t)
	}
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
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123")}, nil)
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
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{}, expectedError)

	err := handleEvidenceReceived(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to resolve uid: %w", expectedError), err)
}

func TestHandleEvidenceReceivedWhenLpaMissingPK(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{}, nil)

	err := handleEvidenceReceived(ctx, client, event)
	assert.Equal(t, errors.New("PK missing from LPA in response"), err)
}

func TestHandleEvidenceReceivedWhenClientPutError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123")}, nil)
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

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCertificateProviderStarted(ctx, event.CertificateProviderStarted{
			UID: "M-1111-2222-3333",
		}).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendCertificateProviderPrompt(ctx, appcontext.Data{}, &completedDonorProvided).
		Return(nil)

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&donorProvided)
			attributevalue.Unmarshal(b, v)
			return nil
		})

	updatedDonorProvided := completedDonorProvided
	updatedDonorProvided.ReducedFeeDecisionAt = testNow
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
		AccessCodeSender(ctx).
		Return(accessCodeSender, nil)
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
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&donorProvided)
			attributevalue.Unmarshal(b, v)
			return nil
		})

	updatedDonorProvided := completedDonorProvided
	updatedDonorProvided.ReducedFeeDecisionAt = testNow
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
		AccessCodeSender(ctx).
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
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := attributevalue.Marshal(&donorProvided)
			attributevalue.Unmarshal(b, v)
			return nil
		})

	updatedDonorProvided := donorProvided
	updatedDonorProvided.ReducedFeeDecisionAt = testNow
	updatedDonorProvided.Tasks.PayForLpa = task.PaymentStateCompleted
	updatedDonorProvided.UpdateHash()
	updatedDonorProvided.UpdatedAt = testNow

	client.EXPECT().
		Put(ctx, &updatedDonorProvided).
		Return(nil)

	err := handleFeeApproved(ctx, client, event, nil, nil, appcontext.Data{}, testNowFn)
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
			client.EXPECT().
				OneByUID(ctx, "M-1111-2222-3333").
				Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
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

			err := handleFeeApproved(ctx, client, event, nil, nil, appcontext.Data{}, nil)
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
			client.EXPECT().
				OneByUID(ctx, "M-1111-2222-3333").
				Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
			client.
				On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
				Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
					b, _ := attributevalue.Marshal(donorProvided)
					attributevalue.Unmarshal(b, v)
					return nil
				})

			updatedDonorProvided := *donorProvided
			updatedDonorProvided.ReducedFeeDecisionAt = testNow
			updatedDonorProvided.Tasks.PayForLpa = tc.updatedTaskState
			updatedDonorProvided.FeeType = tc.approvedFeeType
			updatedDonorProvided.UpdateHash()
			updatedDonorProvided.UpdatedAt = testNow

			client.EXPECT().
				Put(ctx, &updatedDonorProvided).
				Return(nil)

			err := handleFeeApproved(ctx, client, event, nil, nil, appcontext.Data{}, testNowFn)
			assert.Nil(t, err)
		})
	}
}

func TestHandleFeeApprovedWhenVoucherSelected(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

	donor := &donordata.Provided{
		PK:      dynamo.LpaKey("123"),
		SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		FeeType: pay.NoFee,
		Tasks:   donordata.Tasks{PayForLpa: task.PaymentStatePending},
		Voucher: donordata.Voucher{Allowed: true},
	}

	updatedDonor := *donor
	updatedDonor.Tasks.PayForLpa = task.PaymentStateCompleted
	updatedDonor.VoucherInvitedAt = testNow
	updatedDonor.ReducedFeeDecisionAt = testNow
	updatedDonor.UpdateHash()
	updatedDonor.UpdatedAt = testNow

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.EXPECT().
		One(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(nil).
		SetData(donor)
	client.EXPECT().
		Put(ctx, &updatedDonor).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendVoucherInvite(ctx, &donordata.Provided{
			PK:      dynamo.LpaKey("123"),
			SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
			FeeType: pay.NoFee,
			Tasks:   donordata.Tasks{PayForLpa: task.PaymentStateCompleted},
			Voucher: donordata.Voucher{Allowed: true},
		}, appcontext.Data{}).
		Return(nil)

	err := handleFeeApproved(ctx, client, event, accessCodeSender, nil, appcontext.Data{}, testNowFn)
	assert.Nil(t, err)
}

func TestHandleFeeApprovedWhenVoucherSelectedAndAccessCodeSenderError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

	donor := &donordata.Provided{
		PK:      dynamo.LpaKey("123"),
		SK:      dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		FeeType: pay.NoFee,
		Tasks:   donordata.Tasks{PayForLpa: task.PaymentStatePending},
		Voucher: donordata.Voucher{Allowed: true},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData(donor)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendVoucherInvite(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, accessCodeSender, nil, appcontext.Data{}, testNowFn)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandleFeeApprovedWhenDynamoClientPutError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
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

	err := handleFeeApproved(ctx, client, event, nil, nil, appcontext.Data{}, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to update donor provided details: %w", expectedError), err)
}

func TestHandleFeeApprovedWhenAccessCodeSenderError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
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

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCertificateProviderStarted(mock.Anything, mock.Anything).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendCertificateProviderPrompt(ctx, appcontext.Data{}, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, accessCodeSender, eventClient, appcontext.Data{}, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to send share code to certificate provider: %w", expectedError), err)
}

func TestHandleFeeApprovedWhenEventClientError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "reduced-fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","approvedType":"NoFee"}`),
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
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

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendCertificateProviderStarted(mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, nil, eventClient, appcontext.Data{}, testNowFn)
	assert.Equal(t, fmt.Errorf("failed to send certificate-provider-started event: %w", expectedError), err)
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
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
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
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
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
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
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

	updated := &donordata.Provided{
		PK:                   dynamo.LpaKey("123"),
		SK:                   dynamo.LpaOwnerKey(dynamo.DonorKey("456")),
		Tasks:                donordata.Tasks{PayForLpa: task.PaymentStateDenied},
		FeeType:              pay.FullFee,
		ReducedFeeDecisionAt: testNow,
		UpdatedAt:            testNow,
	}
	updated.UpdateHash()

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
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
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
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
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
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

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendLpaCertificateProviderPrompt(ctx, appData, dynamo.LpaKey(testUuidString), dynamo.LpaOwnerKey(dynamo.DonorKey("PAPER")), lpa).
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
					Action:            scheduleddata.ActionRemindCertificateProviderToComplete,
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
		AccessCodeSender(ctx).
		Return(accessCodeSender, nil)
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

func TestHandleDonorSubmissionCompletedWhenOnlineDonor(t *testing.T) {
	appData := appcontext.Data{}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{Donor: lpadata.Donor{Channel: lpadata.ChannelOnline}}, nil)

	err := handleDonorSubmissionCompleted(ctx, nil, donorSubmissionCompletedEvent, nil, appData, lpaStoreClient, testUuidStringFn, testNowFn)
	assert.Nil(t, err)
}

func TestHandleDonorSubmissionCompletedWhenWriteTransactionError(t *testing.T) {
	appData := appcontext.Data{}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendLpaCertificateProviderPrompt(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	client := newMockDynamodbClient(t)
	client.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleDonorSubmissionCompleted(ctx, client, donorSubmissionCompletedEvent, accessCodeSender, appData, lpaStoreClient, testUuidStringFn, testNowFn)
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

func TestHandleDonorSubmissionCompletedWhenAccessCodeSenderError(t *testing.T) {
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

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendLpaCertificateProviderPrompt(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleDonorSubmissionCompleted(ctx, nil, donorSubmissionCompletedEvent, accessCodeSender, appData, lpaStoreClient, testUuidStringFn, testNowFn)
	assert.ErrorIs(t, err, expectedError)
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
	lpaStoreClient.EXPECT().
		SendPaperCertificateProviderAccessOnline(ctx, lpa, "a@example.com").
		Return(nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("an-lpa"), SK: dynamo.DonorKey("a-donor")}, nil)
	dynamoClient.EXPECT().
		One(ctx, dynamo.LpaKey("an-lpa"), dynamo.DonorKey("a-donor"), mock.Anything).
		Return(nil).
		SetData(&donordata.Provided{PK: dynamo.LpaKey("an-lpa")})
	dynamoClient.EXPECT().
		Put(ctx, updatedDonor).
		Return(nil)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(ctx, appData, lpa).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllActionByUID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(&certificateproviderdata.Provided{
			LpaID: "lpa-id",
			SK:    dynamo.CertificateProviderKey("cp-sub"),
			Email: "a@example.com",
		}, nil)

	certificateProviderCtx := appcontext.ContextWithSession(ctx, &appcontext.Session{
		LpaID:     "lpa-id",
		SessionID: "cp-sub",
	})

	certificateProviderStore.EXPECT().
		Delete(certificateProviderCtx).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		AccessCodeSender(ctx).
		Return(accessCodeSender, nil)
	factory.EXPECT().
		AppData().
		Return(appData, nil)
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)
	factory.EXPECT().
		Now().
		Return(testNowFn)
	factory.EXPECT().
		ScheduledStore().
		Return(scheduledStore)
	factory.EXPECT().
		CertificateProviderStore().
		Return(certificateProviderStore)

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
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{
			CertificateProvider: lpadata.CertificateProvider{
				Channel: lpadata.ChannelPaper,
			},
		}, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(dynamo.Keys{}, expectedError)

	accessCodeSender := newMockAccessCodeSender(t)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		AccessCodeSender(ctx).
		Return(accessCodeSender, nil)
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
	lpaStoreClient.EXPECT().
		SendPaperCertificateProviderAccessOnline(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(dynamo.Keys{PK: dynamo.LpaKey("an-lpa"), SK: dynamo.DonorKey("a-donor")}, nil)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)
	dynamoClient.EXPECT().
		Put(mock.Anything, mock.Anything).
		Return(expectedError)

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(ctx, mock.Anything, mock.Anything).
		Return(nil)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllActionByUID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(&certificateproviderdata.Provided{
			LpaID: "lpa-id",
			SK:    dynamo.CertificateProviderKey("cp-sub"),
			Email: "a@example.com",
		}, nil)
	certificateProviderStore.EXPECT().
		Delete(mock.Anything).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		AccessCodeSender(ctx).
		Return(accessCodeSender, nil)
	factory.EXPECT().
		AppData().
		Return(appcontext.Data{}, nil)
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)
	factory.EXPECT().
		Now().
		Return(testNowFn)
	factory.EXPECT().
		ScheduledStore().
		Return(scheduledStore)
	factory.EXPECT().
		CertificateProviderStore().
		Return(certificateProviderStore)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandleCertificateProviderSubmissionCompletedWhenCertificateProviderStoreOneByUIDErrors(t *testing.T) {
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
		OneByUID(mock.Anything, mock.Anything).
		Return(dynamo.Keys{PK: dynamo.LpaKey("an-lpa"), SK: dynamo.DonorKey("a-donor")}, nil)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(&certificateproviderdata.Provided{
			LpaID: "lpa-id",
			SK:    dynamo.CertificateProviderKey("cp-sub"),
			Email: "a@example.com",
		}, expectedError)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		AccessCodeSender(ctx).
		Return(nil, nil)
	factory.EXPECT().
		AppData().
		Return(appcontext.Data{}, nil)
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)
	factory.EXPECT().
		Now().
		Return(testNowFn)
	factory.EXPECT().
		CertificateProviderStore().
		Return(certificateProviderStore)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandleCertificateProviderSubmissionCompletedWhenCertificateProviderStoreDeleteErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpadata.Lpa{
			CertificateProvider: lpadata.CertificateProvider{
				Channel: lpadata.ChannelPaper,
			},
		}, nil)
	lpaStoreClient.EXPECT().
		SendPaperCertificateProviderAccessOnline(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(dynamo.Keys{PK: dynamo.LpaKey("an-lpa"), SK: dynamo.DonorKey("a-donor")}, nil)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(&certificateproviderdata.Provided{
			LpaID: "lpa-id",
			SK:    dynamo.CertificateProviderKey("cp-sub"),
			Email: "a@example.com",
		}, nil)
	certificateProviderStore.EXPECT().
		Delete(mock.Anything).
		Return(expectedError)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		AccessCodeSender(ctx).
		Return(nil, nil)
	factory.EXPECT().
		AppData().
		Return(appcontext.Data{}, nil)
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)
	factory.EXPECT().
		Now().
		Return(testNowFn)
	factory.EXPECT().
		CertificateProviderStore().
		Return(certificateProviderStore)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandleCertificateProviderSubmissionCompletedWhenLpaStoreClientErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpadata.Lpa{
			CertificateProvider: lpadata.CertificateProvider{
				Channel: lpadata.ChannelPaper,
			},
		}, nil)
	lpaStoreClient.EXPECT().
		SendPaperCertificateProviderAccessOnline(mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("an-lpa"), SK: dynamo.DonorKey("a-donor")}, nil)
	dynamoClient.EXPECT().
		One(ctx, dynamo.LpaKey("an-lpa"), dynamo.DonorKey("a-donor"), mock.Anything).
		Return(nil).
		SetData(&donordata.Provided{PK: dynamo.LpaKey("an-lpa")})

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(&certificateproviderdata.Provided{
			LpaID: "lpa-id",
			SK:    dynamo.CertificateProviderKey("cp-sub"),
			Email: "a@example.com",
		}, nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		AccessCodeSender(ctx).
		Return(nil, nil)
	factory.EXPECT().
		AppData().
		Return(appcontext.Data{}, nil)
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)
	factory.EXPECT().
		Now().
		Return(testNowFn)
	factory.EXPECT().
		CertificateProviderStore().
		Return(certificateProviderStore)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandleCertificateProviderSubmissionCompletedWhenAccessCodeSenderErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{
			CertificateProvider: lpadata.CertificateProvider{
				Channel: lpadata.ChannelPaper,
			},
		}, nil)
	lpaStoreClient.EXPECT().
		SendPaperCertificateProviderAccessOnline(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(dynamo.Keys{PK: dynamo.LpaKey("an-lpa"), SK: dynamo.DonorKey("a-donor")}, nil)
	dynamoClient.EXPECT().
		One(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(nil).
		SetData(&donordata.Provided{PK: dynamo.LpaKey("an-lpa")})

	accessCodeSender := newMockAccessCodeSender(t)
	accessCodeSender.EXPECT().
		SendAttorneys(ctx, mock.Anything, mock.Anything).
		Return(expectedError)

	scheduledStore := newMockScheduledStore(t)
	scheduledStore.EXPECT().
		DeleteAllActionByUID(mock.Anything, mock.Anything, mock.Anything).
		Return(nil)

	certificateProviderStore := newMockCertificateProviderStore(t)
	certificateProviderStore.EXPECT().
		OneByUID(mock.Anything, mock.Anything).
		Return(&certificateproviderdata.Provided{
			LpaID: "lpa-id",
			SK:    dynamo.CertificateProviderKey("cp-sub"),
			Email: "a@example.com",
		}, nil)
	certificateProviderStore.EXPECT().
		Delete(mock.Anything).
		Return(nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		AccessCodeSender(ctx).
		Return(accessCodeSender, nil)
	factory.EXPECT().
		AppData().
		Return(appcontext.Data{}, nil)
	factory.EXPECT().
		DynamoClient().
		Return(dynamoClient)
	factory.EXPECT().
		Now().
		Return(testNowFn)
	factory.EXPECT().
		ScheduledStore().
		Return(scheduledStore)
	factory.EXPECT().
		CertificateProviderStore().
		Return(certificateProviderStore)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, fmt.Errorf("failed to send share codes to attorneys: %w", expectedError), err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenAccessCodeSenderFactoryErrors(t *testing.T) {
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
		AccessCodeSender(ctx).
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
		AccessCodeSender(ctx).
		Return(nil, nil)
	factory.EXPECT().
		AppData().
		Return(appcontext.Data{}, expectedError)

	handler := &siriusEventHandler{}
	err := handler.Handle(ctx, factory, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, expectedError, err)
}

func TestHandlePriorityCorrespondenceSent(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "priority-correspondence-sent",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","sentAt":"2024-01-18T00:00:00.000Z"}`),
	}

	updated := &donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456")), UpdatedAt: testNow}
	updated.UpdateHash()

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.EXPECT().
		One(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(nil).
		SetData(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
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

func TestHandlePriorityCorrespondenceSentWhenGetError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "priority-correspondence-sent",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","sentAt":"2024-01-18T00:00:00.000Z"}`),
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{}, expectedError)

	err := handlePriorityCorrespondenceSent(ctx, client, event, testNowFn)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandlePriorityCorrespondenceSentWhenPutError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "priority-correspondence-sent",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","sentAt":"2024-01-18T00:00:00.000Z"}`),
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.EXPECT().
		One(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(nil).
		SetData(donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(expectedError)

	err := handlePriorityCorrespondenceSent(ctx, client, event, testNowFn)
	assert.ErrorIs(t, err, expectedError)
}

func TestHandleImmaterialChangeConfirmed(t *testing.T) {
	testcases := map[string]struct {
		setupDynamoClient             func(*testing.T) *mockDynamodbClient
		setupLpaStoreClient           func(*testing.T) *mockLpaStoreClient
		setupCertificateProviderStore func(*testing.T) *mockCertificateProviderStore
	}{
		"certificateProvider": {
			setupDynamoClient: unusedDynamoClient,
			setupLpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				c := newMockLpaStoreClient(t)
				c.EXPECT().
					SendCertificateProviderConfirmIdentity(ctx, "M-1111-2222-3333", &certificateproviderdata.Provided{
						PK:                          dynamo.LpaKey("123"),
						SK:                          dynamo.CertificateProviderKey("789"),
						Tasks:                       certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
						IdentityDetailsMismatched:   true,
						ImmaterialChangeConfirmedAt: testNow,
					}).
					Return(nil)

				return c
			},
			setupCertificateProviderStore: func(t *testing.T) *mockCertificateProviderStore {
				s := newMockCertificateProviderStore(t)
				s.EXPECT().
					OneByUID(mock.Anything, mock.Anything).
					Return(&certificateproviderdata.Provided{
						PK:                        dynamo.LpaKey("123"),
						SK:                        dynamo.CertificateProviderKey("789"),
						Tasks:                     certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStatePending},
						IdentityDetailsMismatched: true,
					}, nil)
				s.EXPECT().
					Put(ctx, &certificateproviderdata.Provided{
						PK:                          dynamo.LpaKey("123"),
						SK:                          dynamo.CertificateProviderKey("789"),
						Tasks:                       certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
						IdentityDetailsMismatched:   true,
						ImmaterialChangeConfirmedAt: testNow,
					}).
					Return(nil)
				return s
			},
		},
	}

	for actorType, tc := range testcases {
		t.Run(actorType, func(t *testing.T) {
			event := &events.CloudWatchEvent{
				DetailType: "immaterial-change-confirmed",
				Detail:     json.RawMessage(fmt.Sprintf(`{"uid":"M-1111-2222-3333","actorUID":"740e5834-3a29-46b4-9a6f-16142fde533a","actorType":"%s"}`, strings.Split(actorType, " ")[0])),
			}

			factory := newMockFactory(t)
			factory.EXPECT().
				DynamoClient().
				Return(tc.setupDynamoClient(t))
			factory.EXPECT().
				LpaStoreClient().
				Return(tc.setupLpaStoreClient(t), nil)
			factory.EXPECT().
				Now().
				Return(testNowFn)
			factory.EXPECT().
				CertificateProviderStore().
				Return(tc.setupCertificateProviderStore(t))

			handler := &siriusEventHandler{}
			err := handler.Handle(ctx, factory, event)

			assert.Nil(t, err)
		})
	}
}

func TestHandleChangeConfirmedWhenIdentityTaskNotPending(t *testing.T) {
	testcases := map[string]struct {
		setupDynamoClient             func(*testing.T) *mockDynamodbClient
		setupCertificateProviderStore func(*testing.T) *mockCertificateProviderStore
	}{
		"certificateProvider": {
			setupDynamoClient: unusedDynamoClient,
			setupCertificateProviderStore: func(t *testing.T) *mockCertificateProviderStore {
				s := newMockCertificateProviderStore(t)
				s.EXPECT().
					OneByUID(ctx, "M-1111-2222-3333").
					Return(&certificateproviderdata.Provided{
						PK:                        dynamo.LpaKey("123"),
						SK:                        dynamo.CertificateProviderKey("789"),
						Tasks:                     certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateProblem},
						IdentityDetailsMismatched: true,
					}, nil)
				return s
			},
		},
	}

	for actorType, tc := range testcases {
		t.Run(actorType, func(t *testing.T) {
			event := &events.CloudWatchEvent{
				DetailType: "immaterial-change-confirmed",
				Detail:     json.RawMessage(fmt.Sprintf(`{"uid":"M-1111-2222-3333","actorUID":"740e5834-3a29-46b4-9a6f-16142fde533a","actorType":"%s"}`, actorType)),
			}

			err := handleChangeConfirmed(ctx, tc.setupDynamoClient(t), tc.setupCertificateProviderStore(t), event, testNowFn, nil, false)

			assert.Nil(t, err)
		})
	}
}

func TestHandleChangeConfirmedWhenLpaStoreClientError(t *testing.T) {
	testcases := map[string]struct {
		setupDynamoClient             func(*testing.T) *mockDynamodbClient
		setupLpaStoreClient           func(*testing.T) *mockLpaStoreClient
		setupCertificateProviderStore func(*testing.T) *mockCertificateProviderStore
		expectedError                 error
		actorType                     string
	}{
		"certificateProvider": {
			setupDynamoClient: unusedDynamoClient,
			setupLpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				c := newMockLpaStoreClient(t)
				c.EXPECT().
					SendCertificateProviderConfirmIdentity(ctx, mock.Anything, mock.Anything).
					Return(expectedError)
				return c
			},
			setupCertificateProviderStore: func(t *testing.T) *mockCertificateProviderStore {
				s := newMockCertificateProviderStore(t)
				s.EXPECT().
					OneByUID(mock.Anything, mock.Anything).
					Return(&certificateproviderdata.Provided{
						PK:                        dynamo.LpaKey("123"),
						SK:                        dynamo.CertificateProviderKey("789"),
						Tasks:                     certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStatePending},
						IdentityDetailsMismatched: true,
					}, nil)
				return s
			},
			expectedError: fmt.Errorf("failed to send certificate provider confirmed identity to lpa store: %w", expectedError),
			actorType:     "certificateProvider",
		},
		"certificateProvider LPA not found": {
			setupDynamoClient: unusedDynamoClient,
			setupLpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				c := newMockLpaStoreClient(t)
				c.EXPECT().
					SendCertificateProviderConfirmIdentity(ctx, mock.Anything, mock.Anything).
					Return(lpastore.ErrNotFound)
				return c
			},
			setupCertificateProviderStore: func(t *testing.T) *mockCertificateProviderStore {
				s := newMockCertificateProviderStore(t)
				s.EXPECT().
					OneByUID(mock.Anything, mock.Anything).
					Return(&certificateproviderdata.Provided{
						PK:                        dynamo.LpaKey("123"),
						SK:                        dynamo.CertificateProviderKey("789"),
						Tasks:                     certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStatePending},
						IdentityDetailsMismatched: true,
					}, nil)
				s.EXPECT().
					Put(mock.Anything, mock.Anything).
					Return(nil)
				return s
			},
			actorType: "certificateProvider",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			event := &events.CloudWatchEvent{
				DetailType: "immaterial-change-confirmed",
				Detail:     json.RawMessage(fmt.Sprintf(`{"uid":"M-1111-2222-3333","actorUID":"740e5834-3a29-46b4-9a6f-16142fde533a","actorType":"%s"}`, tc.actorType)),
			}

			err := handleChangeConfirmed(ctx, tc.setupDynamoClient(t), tc.setupCertificateProviderStore(t), event, testNowFn, tc.setupLpaStoreClient(t), false)

			assert.Equal(t, tc.expectedError, err)
		})
	}
}

func TestHandleChangeConfirmedWhenPutError(t *testing.T) {
	testcases := map[string]struct {
		setupDynamoClient             func(*testing.T) *mockDynamodbClient
		setupLpaStoreClient           func(*testing.T) *mockLpaStoreClient
		setupCertificateProviderStore func(*testing.T) *mockCertificateProviderStore
	}{
		"certificateProvider": {
			setupDynamoClient: unusedDynamoClient,
			setupLpaStoreClient: func(t *testing.T) *mockLpaStoreClient {
				c := newMockLpaStoreClient(t)
				c.EXPECT().
					SendCertificateProviderConfirmIdentity(mock.Anything, mock.Anything, mock.Anything).
					Return(nil)

				return c
			},
			setupCertificateProviderStore: func(t *testing.T) *mockCertificateProviderStore {
				s := newMockCertificateProviderStore(t)
				s.EXPECT().
					OneByUID(mock.Anything, mock.Anything).
					Return(&certificateproviderdata.Provided{
						PK:                        dynamo.LpaKey("123"),
						SK:                        dynamo.CertificateProviderKey("789"),
						Tasks:                     certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStatePending},
						IdentityDetailsMismatched: true,
					}, nil)
				s.EXPECT().
					Put(ctx, &certificateproviderdata.Provided{
						PK:                          dynamo.LpaKey("123"),
						SK:                          dynamo.CertificateProviderKey("789"),
						Tasks:                       certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateCompleted},
						IdentityDetailsMismatched:   true,
						ImmaterialChangeConfirmedAt: testNow,
					}).
					Return(expectedError)
				return s
			},
		},
	}

	for actorType, tc := range testcases {
		t.Run(actorType, func(t *testing.T) {
			event := &events.CloudWatchEvent{
				DetailType: "immaterial-change-confirmed",
				Detail:     json.RawMessage(fmt.Sprintf(`{"uid":"M-1111-2222-3333","actorUID":"740e5834-3a29-46b4-9a6f-16142fde533a","actorType":"%s"}`, actorType)),
			}

			err := handleChangeConfirmed(ctx, tc.setupDynamoClient(t), tc.setupCertificateProviderStore(t), event, testNowFn, tc.setupLpaStoreClient(t), false)

			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestHandleImmaterialChangeConfirmedWhenUnexpectedActorType(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "immaterial-change-confirmed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333","actorUID":"740e5834-3a29-46b4-9a6f-16142fde533a","actorType":"attorney"}`),
	}

	err := handleChangeConfirmed(ctx, nil, nil, event, testNowFn, nil, false)

	assert.ErrorContains(t, err, "invalid actorType, got attorney, want donor or certificateProvider")
}

func TestHandleMaterialChangeConfirmed(t *testing.T) {
	testcases := map[string]struct {
		setupDynamoClient             func(*testing.T) *mockDynamodbClient
		setupCertificateProviderStore func(*testing.T) *mockCertificateProviderStore
	}{
		"certificateProvider": {
			setupDynamoClient: unusedDynamoClient,
			setupCertificateProviderStore: func(t *testing.T) *mockCertificateProviderStore {
				s := newMockCertificateProviderStore(t)
				s.EXPECT().
					OneByUID(mock.Anything, mock.Anything).
					Return(&certificateproviderdata.Provided{
						PK:                        dynamo.LpaKey("123"),
						SK:                        dynamo.CertificateProviderKey("789"),
						Tasks:                     certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStatePending},
						IdentityDetailsMismatched: true,
					}, nil)
				s.EXPECT().
					Put(ctx, &certificateproviderdata.Provided{
						PK:                        dynamo.LpaKey("123"),
						SK:                        dynamo.CertificateProviderKey("789"),
						Tasks:                     certificateproviderdata.Tasks{ConfirmYourIdentity: task.IdentityStateProblem},
						IdentityDetailsMismatched: true,
						MaterialChangeConfirmedAt: testNow,
					}).
					Return(nil)
				return s
			},
		},
	}

	for actorType, tc := range testcases {
		t.Run(actorType, func(t *testing.T) {
			event := &events.CloudWatchEvent{
				DetailType: "material-change-confirmed",
				Detail:     json.RawMessage(fmt.Sprintf(`{"uid":"M-1111-2222-3333","actorUID":"740e5834-3a29-46b4-9a6f-16142fde533a","actorType":"%s"}`, strings.Split(actorType, " ")[0])),
			}

			factory := newMockFactory(t)
			factory.EXPECT().
				DynamoClient().
				Return(tc.setupDynamoClient(t))
			factory.EXPECT().
				CertificateProviderStore().
				Return(tc.setupCertificateProviderStore(t))
			factory.EXPECT().
				LpaStoreClient().
				Return(unusedLpaStoreClient(t), nil)
			factory.EXPECT().
				Now().
				Return(testNowFn)

			handler := &siriusEventHandler{}
			err := handler.Handle(ctx, factory, event)

			assert.Nil(t, err)
		})
	}
}

func TestHandleCertificateProviderIdentityCheckFailed(t *testing.T) {
	testcases := map[lpadata.Channel]struct {
		notifyClient func(*testing.T, *lpadata.Lpa) *mockNotifyClient
		eventClient  func(*testing.T, *lpadata.Lpa) *mockEventClient
	}{
		lpadata.ChannelOnline: {
			notifyClient: func(t *testing.T, lpa *lpadata.Lpa) *mockNotifyClient {
				notifyClient := newMockNotifyClient(t)
				notifyClient.EXPECT().
					EmailGreeting(lpa).
					Return("greeting")
				notifyClient.EXPECT().
					SendActorEmail(ctx, notify.ToLpaDonor(lpa), "M-1111-2222-3333", notify.InformDonorPaperCertificateProviderIdentityCheckFailed{
						Greeting:                    "greeting",
						CertificateProviderFullName: "a b",
						LpaType:                     "property and affairs",
						DonorStartPageURL:           "app:///start",
					}).
					Return(nil)

				return notifyClient
			},
			eventClient: func(_ *testing.T, _ *lpadata.Lpa) *mockEventClient { return nil },
		},
		lpadata.ChannelPaper: {
			notifyClient: func(_ *testing.T, _ *lpadata.Lpa) *mockNotifyClient { return nil },
			eventClient: func(t *testing.T, lpa *lpadata.Lpa) *mockEventClient {
				eventClient := newMockEventClient(t)
				eventClient.EXPECT().
					SendLetterRequested(ctx, event.LetterRequested{
						UID:        lpa.LpaUID,
						LetterType: "INFORM_DONOR_CERTIFICATE_PROVIDER_HAS_NOT_CONFIRMED_IDENTITY",
						ActorType:  actor.TypeDonor,
						ActorUID:   lpa.Donor.UID,
					}).
					Return(nil)

				return eventClient
			},
		},
	}

	event := &events.CloudWatchEvent{
		DetailType: "certificate-provider-identity-check-failed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	lpa := &lpadata.Lpa{
		LpaUID:              "lpa-uid",
		Type:                lpadata.LpaTypePropertyAndAffairs,
		Donor:               lpadata.Donor{ContactLanguagePreference: localize.En},
		CertificateProvider: lpadata.CertificateProvider{FirstNames: "a", LastName: "b"},
	}

	for channel, tc := range testcases {
		t.Run(channel.String(), func(t *testing.T) {
			lpa := lpa
			lpa.Donor.Channel = channel

			lpaStoreClient := newMockLpaStoreClient(t)
			lpaStoreClient.EXPECT().
				Lpa(ctx, "M-1111-2222-3333").
				Return(lpa, nil)

			localizer := newMockLocalizer(t)
			localizer.EXPECT().
				T("property-and-affairs").
				Return("Property and affairs").
				Maybe()

			bundle := newMockBundle(t)
			bundle.EXPECT().
				For(localize.En).
				Return(localizer).
				Maybe()

			factory := newMockFactory(t)
			factory.EXPECT().
				LpaStoreClient().
				Return(lpaStoreClient, nil)
			factory.EXPECT().
				NotifyClient(ctx).
				Return(tc.notifyClient(t, lpa), nil)
			factory.EXPECT().
				Bundle().
				Return(bundle, nil)
			factory.EXPECT().
				DonorStartURL().
				Return("app:///start")
			factory.EXPECT().
				EventClient().
				Return(tc.eventClient(t, lpa))

			handler := &siriusEventHandler{}
			err := handler.Handle(ctx, factory, event)

			assert.Nil(t, err)
		})
	}
}

func TestHandleCertificateProviderIdentityCheckFailedWhenLpaStoreError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "certificate-provider-identity-check-failed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{}, expectedError)

	err := handleCertificateProviderIdentityCheckedFailed(ctx, lpaStoreClient, nil, nil, nil, "", event)

	assert.ErrorIs(t, err, expectedError)
}

func TestHandleCertificateProviderIdentityCheckFailedWhenNotifyError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "certificate-provider-identity-check-failed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{}, nil)

	localizer := newMockLocalizer(t)
	localizer.EXPECT().
		T(mock.Anything).
		Return("")

	bundle := newMockBundle(t)
	bundle.EXPECT().
		For(mock.Anything).
		Return(localizer)

	notifyClient := newMockNotifyClient(t)
	notifyClient.EXPECT().
		EmailGreeting(mock.Anything).
		Return("")
	notifyClient.EXPECT().
		SendActorEmail(ctx, mock.Anything, mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleCertificateProviderIdentityCheckedFailed(ctx, lpaStoreClient, notifyClient, nil, bundle, "", event)

	assert.ErrorIs(t, err, expectedError)
}

func TestHandleCertificateProviderIdentityCheckFailedWhenEventError(t *testing.T) {
	event := &events.CloudWatchEvent{
		DetailType: "certificate-provider-identity-check-failed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(mock.Anything, mock.Anything).
		Return(&lpadata.Lpa{Donor: lpadata.Donor{Channel: lpadata.ChannelPaper}}, nil)

	eventClient := newMockEventClient(t)
	eventClient.EXPECT().
		SendLetterRequested(mock.Anything, mock.Anything).
		Return(expectedError)

	err := handleCertificateProviderIdentityCheckedFailed(ctx, lpaStoreClient, nil, eventClient, nil, "", event)

	assert.ErrorIs(t, err, expectedError)
}

func TestHandleCertificateProviderIdentityCheckFailedWhenFactoryErrors(t *testing.T) {
	testcases := map[string]func(*testing.T) *mockFactory{
		"NotifyClient": func(t *testing.T) *mockFactory {
			f := newMockFactory(t)
			f.EXPECT().
				NotifyClient(ctx).
				Return(newMockNotifyClient(t), expectedError)
			return f
		},
		"Bundle": func(t *testing.T) *mockFactory {
			f := newMockFactory(t)
			f.EXPECT().
				NotifyClient(ctx).
				Return(newMockNotifyClient(t), nil)
			f.EXPECT().
				Bundle().
				Return(newMockBundle(t), expectedError)
			return f
		},
	}

	for name, setupFactory := range testcases {
		t.Run(name, func(t *testing.T) {
			event := &events.CloudWatchEvent{
				DetailType: "certificate-provider-identity-check-failed",
				Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
			}

			handler := &siriusEventHandler{}
			err := handler.Handle(ctx, setupFactory(t), event)

			assert.ErrorIs(t, err, expectedError)
		})
	}
}
