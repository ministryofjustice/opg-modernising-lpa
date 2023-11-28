package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")
var ctx = context.Background()

func TestHandleUidRequested(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"hw","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	uidClient := newMockUidClient(t)
	uidClient.
		On("CreateCase", ctx, &uid.CreateCaseRequestBody{
			Type: "hw",
			Donor: uid.DonorDetails{
				Name:     "a donor",
				Dob:      date.New("2000", "01", "02"),
				Postcode: "F1 1FF",
			},
		}).
		Return("M-1111-2222-3333", nil)

	uidStore := newMockUidStore(t)
	uidStore.
		On("Set", ctx, "an-id", "donor-id", "M-1111-2222-3333").
		Return(nil)

	err := handleUidRequested(ctx, uidStore, uidClient, event)
	assert.Nil(t, err)
}

func TestHandleUidRequestedWhenUidClientErrors(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"hw","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	uidClient := newMockUidClient(t)
	uidClient.
		On("CreateCase", ctx, mock.Anything).
		Return("", expectedError)

	err := handleUidRequested(ctx, nil, uidClient, event)
	assert.Equal(t, fmt.Errorf("failed to create case: %w", expectedError), err)
}

func TestHandleUidRequestedWhenUidStoreErrors(t *testing.T) {
	event := events.CloudWatchEvent{
		DetailType: "uid-requested",
		Detail:     json.RawMessage(`{"lpaID":"an-id","donorSessionID":"donor-id","type":"hw","donor":{"name":"a donor","dob":"2000-01-02","postcode":"F1 1FF"}}`),
	}

	uidClient := newMockUidClient(t)
	uidClient.
		On("CreateCase", ctx, mock.Anything).
		Return("M-1111-2222-3333", nil)

	uidStore := newMockUidStore(t)
	uidStore.
		On("Set", ctx, "an-id", "donor-id", "M-1111-2222-3333").
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
	event := events.CloudWatchEvent{
		DetailType: "evidence-required",
		Detail:     json.RawMessage(`{"uid":"M-1111-2222-3333"}`),
	}

	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
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
	client.
		On("Put", ctx, map[string]string{
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
	client.
		On("Put", ctx, actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted}, UpdatedAt: now}).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", ctx, notify.CertificateProviderProvideCertificatePromptEmail, page.AppData{}, &actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted}, UpdatedAt: now}).
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
	client.
		On("Put", ctx, actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted}, UpdatedAt: now}).
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
	client.
		On("Put", ctx, actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted}, UpdatedAt: now}).
		Return(nil)

	shareCodeSender := newMockShareCodeSender(t)
	shareCodeSender.
		On("SendCertificateProvider", ctx, notify.CertificateProviderProvideCertificatePromptEmail, page.AppData{}, &actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskCompleted}, UpdatedAt: now}).
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
	client.
		On("Put", ctx, actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired}, UpdatedAt: now}).
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
	client.
		On("Put", ctx, actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskMoreEvidenceRequired}, UpdatedAt: now}).
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
	client.
		On("Put", ctx, actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}, UpdatedAt: now}).
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
	client.
		On("Put", ctx, actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskDenied}, UpdatedAt: now}).
		Return(expectedError)

	err := handleFeeDenied(ctx, client, event, func() time.Time { return now })
	assert.Equal(t, fmt.Errorf("failed to update LPA task status: %w", expectedError), err)
}

func TestHandleObjectTagsAdded(t *testing.T) {
	testCases := map[string]bool{
		"ok":       false,
		"infected": true,
	}

	for scanResult, hasVirus := range testCases {
		t.Run(scanResult, func(t *testing.T) {
			event := Event{
				S3Event: events.S3Event{Records: []events.S3EventRecord{
					{S3: events.S3Entity{Object: events.S3Object{Key: "M-1111-2222-3333/evidence/a-uid"}}},
				}},
			}

			s3Client := newMockS3Client(t)
			s3Client.
				On("GetObjectTags", ctx, "M-1111-2222-3333/evidence/a-uid").
				Return([]types.Tag{
					{Key: aws.String("virus-scan-status"), Value: aws.String(scanResult)},
				}, nil)

			dynamoClient := newMockDynamodbClient(t)
			dynamoClient.
				On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
				Return(func(ctx context.Context, uid string, v interface{}) error {
					b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
					json.Unmarshal(b, v)
					return nil
				})
			dynamoClient.
				On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
				Return(func(ctx context.Context, pk, sk string, v interface{}) error {
					b, _ := json.Marshal(actor.DonorProvidedDetails{LpaID: "123", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
					json.Unmarshal(b, v)
					return nil
				})

			documentStore := newMockDocumentStore(t)
			documentStore.
				On("UpdateScanResults", ctx, "123", "M-1111-2222-3333/evidence/a-uid", hasVirus).
				Return(nil)

			err := handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, documentStore)
			assert.Nil(t, err)
		})
	}
}

func TestHandleObjectTagsAddedWhenScannedTagMissing(t *testing.T) {
	event := Event{
		S3Event: events.S3Event{Records: []events.S3EventRecord{
			{S3: events.S3Entity{Object: events.S3Object{Key: "M-1111-2222-3333/evidence/a-uid"}}},
		}},
	}

	s3Client := newMockS3Client(t)
	s3Client.
		On("GetObjectTags", ctx, "M-1111-2222-3333/evidence/a-uid").
		Return([]types.Tag{
			{Key: aws.String("not-virus-scan-status"), Value: aws.String("ok")},
		}, nil)

	err := handleObjectTagsAdded(ctx, nil, event.S3Event, s3Client, nil)
	assert.Nil(t, err)
}

func TestHandleObjectTagsAddedWhenObjectKeyMissing(t *testing.T) {
	event := Event{
		S3Event: events.S3Event{Records: []events.S3EventRecord{
			{S3: events.S3Entity{Object: events.S3Object{}}},
		}},
	}

	err := handleObjectTagsAdded(ctx, nil, event.S3Event, nil, nil)
	assert.Equal(t, fmt.Errorf("object key missing"), err)
}

func TestHandleObjectTagsAddedWhenS3ClientGetObjectTagsError(t *testing.T) {
	event := Event{
		S3Event: events.S3Event{Records: []events.S3EventRecord{
			{S3: events.S3Entity{Object: events.S3Object{Key: "M-1111-2222-3333/evidence/a-uid"}}},
		}},
	}

	s3Client := newMockS3Client(t)
	s3Client.
		On("GetObjectTags", ctx, "M-1111-2222-3333/evidence/a-uid").
		Return([]types.Tag{}, expectedError)

	err := handleObjectTagsAdded(ctx, nil, event.S3Event, s3Client, nil)
	assert.Equal(t, fmt.Errorf("failed to get tags for object: %w", expectedError), err)
}

func TestHandleObjectTagsAddedWhenDynamoClientOneByUIDError(t *testing.T) {
	event := Event{
		S3Event: events.S3Event{Records: []events.S3EventRecord{
			{S3: events.S3Entity{Object: events.S3Object{Key: "M-1111-2222-3333/evidence/a-uid"}}},
		}},
	}

	s3Client := newMockS3Client(t)
	s3Client.
		On("GetObjectTags", ctx, "M-1111-2222-3333/evidence/a-uid").
		Return([]types.Tag{
			{Key: aws.String("virus-scan-status"), Value: aws.String("ok")},
		}, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	dynamoClient.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(actor.DonorProvidedDetails{LpaID: "123", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return expectedError
		})

	err := handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, nil)
	assert.Equal(t, fmt.Errorf("failed to get LPA: %w", expectedError), err)
}

func TestHandleObjectTagsAddedWhenDocumentStoreUpdateScanResultsError(t *testing.T) {
	event := Event{
		S3Event: events.S3Event{Records: []events.S3EventRecord{
			{S3: events.S3Entity{Object: events.S3Object{Key: "M-1111-2222-3333/evidence/a-uid"}}},
		}},
	}

	s3Client := newMockS3Client(t)
	s3Client.
		On("GetObjectTags", ctx, "M-1111-2222-3333/evidence/a-uid").
		Return([]types.Tag{
			{Key: aws.String("virus-scan-status"), Value: aws.String("ok")},
		}, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{PK: "LPA#123", SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})
	dynamoClient.
		On("One", ctx, "LPA#123", "#DONOR#456", mock.Anything).
		Return(func(ctx context.Context, pk, sk string, v interface{}) error {
			b, _ := json.Marshal(actor.DonorProvidedDetails{LpaID: "123", Tasks: actor.DonorTasks{PayForLpa: actor.PaymentTaskPending}})
			json.Unmarshal(b, v)
			return nil
		})

	documentStore := newMockDocumentStore(t)
	documentStore.
		On("UpdateScanResults", ctx, "123", "M-1111-2222-3333/evidence/a-uid", false).
		Return(expectedError)

	err := handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, documentStore)
	assert.Equal(t, fmt.Errorf("failed to update scan results: %w", expectedError), err)
}

func TestGetLpaByUID(t *testing.T) {
	expectedDonor := actor.DonorProvidedDetails{PK: "LPA#123", SK: "#DONOR#456"}

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
			b, _ := json.Marshal(expectedDonor)
			json.Unmarshal(b, v)
			return nil
		})

	lpa, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Equal(t, expectedDonor, lpa)
	assert.Nil(t, err)
}

func TestGetLpaByUIDWhenClientOneByUidError(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	lpa, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Equal(t, actor.DonorProvidedDetails{}, lpa)
	assert.Equal(t, fmt.Errorf("failed to resolve uid: %w", expectedError), err)
}

func TestGetLpaByUIDWhenPKMissing(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Key{SK: "#DONOR#456"})
			json.Unmarshal(b, v)
			return nil
		})

	lpa, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Equal(t, actor.DonorProvidedDetails{}, lpa)
	assert.Equal(t, errors.New("PK missing from LPA in response"), err)
}

func TestGetLpaByUIDWhenClientOneError(t *testing.T) {
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

	lpa, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Equal(t, actor.DonorProvidedDetails{}, lpa)
	assert.Equal(t, fmt.Errorf("failed to get LPA: %w", expectedError), err)
}
