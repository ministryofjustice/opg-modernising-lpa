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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(ctx, updated).
		Return(nil)

	err := handleFeeApproved(ctx, client, event, shareCodeSender, lpaStoreClient, page.AppData{}, func() time.Time { return now })
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

	err := handleFeeApproved(ctx, client, event, shareCodeSender, nil, page.AppData{}, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to send share code to certificate provider: %w", expectedError), err)
}

func TestHandleFeeApprovedWhenLpaStoreError(t *testing.T) {
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

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		SendLpa(ctx, updated).
		Return(expectedError)

	err := handleFeeApproved(ctx, client, event, shareCodeSender, lpaStoreClient, page.AppData{}, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to send to lpastore: %w", expectedError), err)
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

var donorSubmissionCompletedEvent = events.CloudWatchEvent{
	DetailType: "donor-submission-completed",
	Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
}

func TestHandleDonorSubmissionCompleted(t *testing.T) {
	appData := page.AppData{}

	donor := &lpastore.Lpa{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.ChannelOnline,
		},
	}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(dynamo.NotFoundError{})
	client.EXPECT().
		Put(ctx, &actor.DonorProvidedDetails{
			PK:        dynamo.LpaKey(testUuidString),
			SK:        dynamo.DonorKey("PAPER"),
			LpaID:     testUuidString,
			LpaUID:    "M-1111-2222-3333",
			CreatedAt: testNow,
			Version:   1,
		}).
		Return(nil)

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(donor, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendCertificateProviderInvite(ctx, appData, page.CertificateProviderInvite{
			CertificateProvider: donor.CertificateProvider,
		}).
		Return(nil)

	err := handleDonorSubmissionCompleted(ctx, client, donorSubmissionCompletedEvent, shareCodeSender, appData, lpaStoreClient, testUuidStringFn, testNowFn)
	assert.Nil(t, err)
}

func TestHandleDonorSubmissionCompletedWhenPaperCertificateProvider(t *testing.T) {
	appData := page.AppData{}

	donor := &lpastore.Lpa{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.ChannelPaper,
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
		Return(donor, nil)

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

	donor := &lpastore.Lpa{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.ChannelOnline,
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
		Return(donor, expectedError)

	err := handleDonorSubmissionCompleted(ctx, client, donorSubmissionCompletedEvent, nil, appData, lpaStoreClient, testUuidStringFn, testNowFn)
	assert.Equal(t, expectedError, err)
}

func TestHandleDonorSubmissionCompletedWhenShareCodeSenderError(t *testing.T) {
	appData := page.AppData{}

	donor := &lpastore.Lpa{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.ChannelOnline,
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
		Return(donor, nil)

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

	donor := &lpastore.Lpa{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.ChannelPaper,
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(donor, nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.EXPECT().
		SendAttorneys(ctx, appData, donor).
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

	err := handleCertificateProviderSubmissionCompleted(ctx, certificateProviderSubmissionCompletedEvent, factory)
	assert.Nil(t, err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenOnline(t *testing.T) {
	donor := &lpastore.Lpa{
		CertificateProvider: actor.CertificateProvider{
			CarryOutBy: actor.ChannelOnline,
		},
	}

	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(donor, nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)

	handler := &cloudWatchEventHandler{factory: factory}
	err := handler.Handle(ctx, certificateProviderSubmissionCompletedEvent)
	assert.Nil(t, err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenLpaStoreFactoryErrors(t *testing.T) {
	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(nil, expectedError)

	handler := &cloudWatchEventHandler{factory: factory}
	err := handler.Handle(ctx, certificateProviderSubmissionCompletedEvent)
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

	handler := &cloudWatchEventHandler{factory: factory}
	err := handler.Handle(ctx, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, fmt.Errorf("failed to retrieve lpa: %w", expectedError), err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenShareCodeSenderErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpastore.Lpa{
			CertificateProvider: actor.CertificateProvider{
				CarryOutBy: actor.ChannelPaper,
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

	handler := &cloudWatchEventHandler{factory: factory}
	err := handler.Handle(ctx, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, fmt.Errorf("failed to send share codes to attorneys: %w", expectedError), err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenShareCodeSenderFactoryErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpastore.Lpa{
			CertificateProvider: actor.CertificateProvider{
				CarryOutBy: actor.ChannelPaper,
			},
		}, nil)

	factory := newMockFactory(t)
	factory.EXPECT().
		LpaStoreClient().
		Return(lpaStoreClient, nil)
	factory.EXPECT().
		ShareCodeSender(ctx).
		Return(nil, expectedError)

	handler := &cloudWatchEventHandler{factory: factory}
	err := handler.Handle(ctx, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, expectedError, err)
}

func TestHandleCertificateProviderSubmissionCompletedWhenAppDataFactoryErrors(t *testing.T) {
	lpaStoreClient := newMockLpaStoreClient(t)
	lpaStoreClient.EXPECT().
		Lpa(ctx, "M-1111-2222-3333").
		Return(&lpastore.Lpa{
			CertificateProvider: actor.CertificateProvider{
				CarryOutBy: actor.ChannelPaper,
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

	handler := &cloudWatchEventHandler{factory: factory}
	err := handler.Handle(ctx, certificateProviderSubmissionCompletedEvent)
	assert.Equal(t, expectedError, err)
}
