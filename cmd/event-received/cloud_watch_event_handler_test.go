package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMakeAppData(t *testing.T) {
	handler := &cloudWatchEventHandler{}

	appData, err := handler.makeAppData()
	assert.Error(t, err)
	assert.Equal(t, page.AppData{}, appData)
}

func TestMakeLambdaClient(t *testing.T) {
	handler := &cloudWatchEventHandler{}
	client := handler.makeLambdaClient()

	assert.NotNil(t, client)
}

func TestMakeShareCodeSender(t *testing.T) {
	ctx := context.Background()
	handler := &cloudWatchEventHandler{}

	secretsClient := newMockSecretsClient(t)
	secretsClient.EXPECT().
		Secret(ctx, secrets.GovUkNotify).
		Return("a-b-c-d-e-f-g-h-i-j-k", nil)

	sender, err := handler.makeShareCodeSender(ctx, secretsClient)
	assert.Nil(t, err)
	assert.NotNil(t, sender)
}

func TestMakeLpaStoreClient(t *testing.T) {
	handler := &cloudWatchEventHandler{}

	secretsClient := newMockSecretsClient(t)

	client := handler.makeLpaStoreClient(secretsClient)
	assert.NotNil(t, client)
}

func TestHandleUnknownEvent(t *testing.T) {
	handler := &cloudWatchEventHandler{}

	err := handler.Handle(ctx, events.CloudWatchEvent{DetailType: "some-event"})
	assert.Equal(t, fmt.Errorf("unknown cloudwatch event"), err)
}

func TestHandleUidRequested(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","organisationID":"org-id","type":"personal-welfare","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	uidClient := newMockUidClient(t)
	uidClient.EXPECT().
		CreateCase(ctx, &uid.CreateCaseRequestBody{
			Type: "personal-welfare",
			Donor: uid.DonorDetails{
				Name:     "a donor",
				Dob:      date.New("2000", "01", "02"),
				Postcode: "F1 1FF",
			},
		}).
		Return("M-1111-2222-3333", nil)

	uidStore := newMockUidStore(t)
	uidStore.EXPECT().
		Set(ctx, "an-id", "donor-id", "org-id", "M-1111-2222-3333").
		Return(nil)

	err := handleUidRequested(ctx, uidStore, uidClient, event)
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

	err := handleUidRequested(ctx, nil, uidClient, event)
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

	err := handleUidRequested(ctx, uidStore, uidClient, event)
	assert.Equal(t, fmt.Errorf("failed to set uid: %w", expectedError), err)
}

func TestHandleEvidenceReceived(t *testing.T) {
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
	client.EXPECT().
		Put(ctx, map[string]string{
			"PK": "LPA#123",
			"SK": "#EVIDENCE_RECEIVED",
		}).
		Return(nil)

	err := handleEvidenceReceived(ctx, client, event)
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
			b, _ := json.Marshal(dynamo.Key{})
			json.Unmarshal(b, v)
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
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123"})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, map[string]string{
			"PK": "LPA#123",
			"SK": "#EVIDENCE_RECEIVED",
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

	now := time.Now()
	updated := &actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted}, UpdatedAt: now}
	updated.Hash, _ = updated.GenerateHash()

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
			b, _ := json.Marshal(actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(ctx, page.AppData{}, updated).
		Return(nil)

	err := handleFeeApproved(ctx, client, event, shareCodeSender, page.AppData{}, func() time.Time { return now })
	assert.Nil(t, err)
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
			b, _ := json.Marshal(actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, nil, page.AppData{}, func() time.Time { return now })
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
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderPrompt(ctx, page.AppData{}, mock.Anything).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, shareCodeSender, page.AppData{}, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to send share code to certificate provider: %w", expectedError), err)
}

func TestHandleMoreEvidenceRequired(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "more-evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()
	updated := &actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired}, UpdatedAt: now}
	updated.Hash, _ = updated.GenerateHash()

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
			b, _ := json.Marshal(actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(nil)

	err := handleMoreEvidenceRequired(ctx, client, event, func() time.Time { return now })
	assert.Nil(t, err)
}

func TestHandleMoreEvidenceRequiredWhenPutError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "more-evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()
	updated := &actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired}, UpdatedAt: now}
	updated.Hash, _ = updated.GenerateHash()

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
			b, _ := json.Marshal(actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(expectedError)

	err := handleMoreEvidenceRequired(ctx, client, event, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to update LPA task status: %w", expectedError), err)
}

func TestHandleFeeDenied(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "reduced-fee-declined",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	now := time.Now()
	updated := &actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}, UpdatedAt: now}
	updated.Hash, _ = updated.GenerateHash()

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
			b, _ := json.Marshal(actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, updated).
		Return(nil)

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
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	client.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})
	client.EXPECT().
		Put(ctx, mock.Anything).
		Return(expectedError)

	err := handleFeeDenied(ctx, client, event, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to update LPA task status: %w", expectedError), err)
}

func TestHandleDonorSubmissionCompleted(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "donor-submission-completed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	appData := page.AppData{}

	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.Online,
		},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(dynamo.NotFoundError{})

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(donor, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderInvite(ctx, appData, donor).
		Return(nil)

	err := handleDonorSubmissionCompleted(ctx, client, event, shareCodeSender, appData, lpaStoreClient)
	assert.Nil(t, err)
}

func TestHandleDonorSubmissionCompletedWhenPaperCertificateProvider(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "donor-submission-completed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	appData := page.AppData{}

	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.Paper,
		},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(dynamo.NotFoundError{})

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(donor, nil)

	err := handleDonorSubmissionCompleted(ctx, client, event, nil, appData, lpaStoreClient)
	assert.Nil(t, err)
}

func TestHandleDonorSubmissionCompletedWhenDynamoExists(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "donor-submission-completed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	appData := page.AppData{}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil)

	err := handleDonorSubmissionCompleted(ctx, client, event, nil, appData, nil)
	assert.Nil(t, err)
}

func TestHandleDonorSubmissionCompletedWhenDynamoError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "donor-submission-completed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	appData := page.AppData{}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	err := handleDonorSubmissionCompleted(ctx, client, event, nil, appData, nil)
	assert.Equal(t, expectedError, err)
}

func TestHandleDonorSubmissionCompletedWhenLpaStoreError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "donor-submission-completed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	appData := page.AppData{}

	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.Online,
		},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(dynamo.NotFoundError{})

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(donor, expectedError)

	err := handleDonorSubmissionCompleted(ctx, client, event, nil, appData, lpaStoreClient)
	assert.Equal(t, expectedError, err)
}

func TestHandleDonorSubmissionCompletedWhenShareCodeSenderError(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "donor-submission-completed",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	appData := page.AppData{}

	donor := &actor.DonorProvidedDetails{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.Online,
		},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(dynamo.NotFoundError{})

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(donor, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderInvite(ctx, appData, donor).
		Return(expectedError)

	err := handleDonorSubmissionCompleted(ctx, client, event, shareCodeSender, appData, lpaStoreClient)
	assert.Equal(t, expectedError, err)
}
