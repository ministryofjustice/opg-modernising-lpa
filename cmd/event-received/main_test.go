package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func TestHandleEvidenceReceived(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, map[string]string{
			"PK": "LPA#123",
			"SK": "#EVIDENCE_RECEIVED",
		}).
		Return(nil)

	err := handleEvidenceReceived(ctx, client, event)
	assert.Nil(t, err)
}

func TestHandleEvidenceReceivedWhenClientGetError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	err := handleEvidenceReceived(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to resolve uid for 'evidence-received': %w", expectedError), err)
}

func TestHandleEvidenceReceivedWhenLpaMissingPK(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{})
			json.Unmarshal(b, v)
			return nil
		})

	err := handleEvidenceReceived(ctx, client, event)
	assert.Equal(t, errors.New("PK missing from LPA in response to 'evidence-received'"), err)
}

func TestHandleEvidenceReceivedWhenClientPutError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, map[string]string{
			"PK": "LPA#123",
			"SK": "#EVIDENCE_RECEIVED",
		}).
		Return(expectedError)

	err := handleEvidenceReceived(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to persist evidence received for 'evidence-received': %w", expectedError), err)
}

func TestHandleFeeApproved(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskCompleted}}).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", ctx, notify.CertificateProviderInviteEmail, page.AppData{}, false, &page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskCompleted}}).
		Return(nil)

	err := handleFeeApproved(ctx, client, event, shareCodeSender, page.AppData{})
	assert.Nil(t, err)
}

func TestHandleFeeApprovedWhenDynamoClientOneByUIDError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, nil, page.AppData{})
	assert.Equal(t, fmt.Errorf("failed to resolve uid for 'fee-approved': %w", expectedError), err)
}

func TestHandleFeeApprovedWhenDynamoClientGetError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, nil, page.AppData{})
	assert.Equal(t, fmt.Errorf("failed to get LPA for 'fee-approved': %w", expectedError), err)
}

func TestHandleFeeApprovedWhenDynamoClientPutError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(page.Lpa{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskCompleted}}).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, nil, page.AppData{})
	assert.Equal(t, fmt.Errorf("failed to update LPA task status for 'fee-approved': %w", expectedError), err)
}

func TestHandleFeeApprovedWhenShareCodeSenderError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-approved",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskCompleted}}).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", ctx, notify.CertificateProviderInviteEmail, page.AppData{}, false, &page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskCompleted}}).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, shareCodeSender, page.AppData{})
	assert.Equal(t, fmt.Errorf("failed to send share code to certificate provider for 'fee-approved': %w", expectedError), err)
}

func TestHandleMoreEvidenceRequired(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "more-evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired}}).
		Return(nil)

	err := handleMoreEvidenceRequired(ctx, client, event)
	assert.Nil(t, err)
}

func TestHandleMoreEvidenceRequiredWhenOneByUIDError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "more-evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	err := handleMoreEvidenceRequired(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to resolve uid for 'more-evidence-required': %w", expectedError), err)
}

func TestHandleMoreEvidenceRequiredWhenPKMissing(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "more-evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{})
			json.Unmarshal(b, v)
			return nil
		})

	err := handleMoreEvidenceRequired(ctx, client, event)

	assert.Equal(t, errors.New("PK missing from LPA in response to 'more-evidence-required'"), err)
}

func TestHandleMoreEvidenceRequiredWhenGetError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "more-evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(expectedError)

	err := handleMoreEvidenceRequired(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to get LPA for 'more-evidence-required': %w", expectedError), err)
}

func TestHandleMoreEvidenceRequiredWhenPutError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "more-evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired}}).
		Return(expectedError)

	err := handleMoreEvidenceRequired(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to update LPA task status for 'more-evidence-required': %w", expectedError), err)
}

func TestHandleFeeDenied(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-denied",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskDenied}}).
		Return(nil)

	err := handleFeeDenied(ctx, client, event)
	assert.Nil(t, err)
}

func TestHandleFeeDeniedWhenOneByUIDError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-denied",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	err := handleFeeDenied(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to resolve uid for 'fee-denied': %w", expectedError), err)
}

func TestHandleFeeDeniedWhenPKMissing(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-denied",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{})
			json.Unmarshal(b, v)
			return nil
		})

	err := handleFeeDenied(ctx, client, event)

	assert.Equal(t, errors.New("PK missing from LPA in response to 'fee-denied'"), err)
}

func TestHandleFeeDeniedWhenGetError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-denied",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(expectedError)

	err := handleFeeDenied(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to get LPA for 'fee-denied': %w", expectedError), err)
}

func TestHandleFeeDeniedWhenPutError(t *testing.T) {
	ctx := context.Background()
	event := events.CloudWatchEvent{
		DetailType: "fee-denied",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("Put", ctx, page.Lpa{PK: "LPA#123", SK: "#DONOR#456", Tasks: page.Tasks{PayForLpa: actor.PaymentTaskDenied}}).
		Return(expectedError)

	err := handleFeeDenied(ctx, client, event)
	assert.Equal(t, fmt.Errorf("failed to update LPA task status for 'fee-denied': %w", expectedError), err)
}
