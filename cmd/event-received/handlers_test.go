package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleObjectTagsAdded(t *testing.T) {
	testCases := map[string]bool{
		"NO_THREATS_FOUND": false,
		"THREATS_FOUND":    true,
	}

	for scanResult, hasVirus := range testCases {
		t.Run(scanResult, func(t *testing.T) {
			event := Event{
				S3Event: &events.S3Event{Records: []events.S3EventRecord{
					{S3: events.S3Entity{Object: events.S3Object{Key: "M-1111-2222-3333/evidence/a-uid"}}},
				}},
			}

			s3Client := newMockS3Client(t)
			s3Client.EXPECT().
				GetObjectTags(ctx, "M-1111-2222-3333/evidence/a-uid").
				Return([]types.Tag{
					{Key: aws.String("GuardDutyMalwareScanStatus"), Value: aws.String(scanResult)},
				}, nil)

			dynamoClient := newMockDynamodbClient(t)
			dynamoClient.EXPECT().
				OneByUID(ctx, "M-1111-2222-3333").
				Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
			dynamoClient.
				On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
				Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
					b, _ := json.Marshal(donordata.Provided{LpaID: "123", Tasks: donordata.Tasks{PayForLpa: task.PaymentStatePending}})
					json.Unmarshal(b, v)
					return nil
				})

			documentStore := newMockDocumentStore(t)
			documentStore.EXPECT().
				UpdateScanResults(ctx, "123", "M-1111-2222-3333/evidence/a-uid", hasVirus).
				Return(nil)

			err := handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, documentStore)
			assert.Nil(t, err)
		})
	}
}

func TestHandleObjectTagsAddedWhenScannedTagMissing(t *testing.T) {
	event := Event{
		S3Event: &events.S3Event{Records: []events.S3EventRecord{
			{S3: events.S3Entity{Object: events.S3Object{Key: "M-1111-2222-3333/evidence/a-uid"}}},
		}},
	}

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		GetObjectTags(ctx, "M-1111-2222-3333/evidence/a-uid").
		Return([]types.Tag{
			{Key: aws.String("NotGuardDutyMalwareScanStatus"), Value: aws.String("NO_THREATS_FOUND")},
		}, nil)

	err := handleObjectTagsAdded(ctx, nil, event.S3Event, s3Client, nil)
	assert.Nil(t, err)
}

func TestHandleObjectTagsAddedWhenObjectKeyMissing(t *testing.T) {
	event := Event{
		S3Event: &events.S3Event{Records: []events.S3EventRecord{
			{S3: events.S3Entity{Object: events.S3Object{}}},
		}},
	}

	err := handleObjectTagsAdded(ctx, nil, event.S3Event, nil, nil)
	assert.Equal(t, fmt.Errorf("object key missing"), err)
}

func TestHandleObjectTagsAddedWhenS3ClientGetObjectTagsError(t *testing.T) {
	event := Event{
		S3Event: &events.S3Event{Records: []events.S3EventRecord{
			{S3: events.S3Entity{Object: events.S3Object{Key: "M-1111-2222-3333/evidence/a-uid"}}},
		}},
	}

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		GetObjectTags(ctx, "M-1111-2222-3333/evidence/a-uid").
		Return([]types.Tag{}, expectedError)

	err := handleObjectTagsAdded(ctx, nil, event.S3Event, s3Client, nil)
	assert.Equal(t, fmt.Errorf("failed to get tags for object: %w", expectedError), err)
}

func TestHandleObjectTagsAddedWhenDynamoClientOneByUIDError(t *testing.T) {
	event := Event{
		S3Event: &events.S3Event{Records: []events.S3EventRecord{
			{S3: events.S3Entity{Object: events.S3Object{Key: "M-1111-2222-3333/evidence/a-uid"}}},
		}},
	}

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		GetObjectTags(ctx, "M-1111-2222-3333/evidence/a-uid").
		Return([]types.Tag{
			{Key: aws.String("GuardDutyMalwareScanStatus"), Value: aws.String("NO_THREATS_FOUND")},
		}, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	dynamoClient.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(donordata.Provided{LpaID: "123", Tasks: donordata.Tasks{PayForLpa: task.PaymentStatePending}})
			json.Unmarshal(b, v)
			return expectedError
		})

	err := handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, nil)
	assert.Equal(t, fmt.Errorf("failed to get LPA: %w", expectedError), err)
}

func TestHandleObjectTagsAddedWhenDocumentStoreUpdateScanResultsError(t *testing.T) {
	event := Event{
		S3Event: &events.S3Event{Records: []events.S3EventRecord{
			{S3: events.S3Entity{Object: events.S3Object{Key: "M-1111-2222-3333/evidence/a-uid"}}},
		}},
	}

	s3Client := newMockS3Client(t)
	s3Client.EXPECT().
		GetObjectTags(ctx, "M-1111-2222-3333/evidence/a-uid").
		Return([]types.Tag{
			{Key: aws.String("GuardDutyMalwareScanStatus"), Value: aws.String("NO_THREATS_FOUND")},
		}, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	dynamoClient.
		On("One", ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(func(ctx context.Context, pk dynamo.PK, sk dynamo.SK, v interface{}) error {
			b, _ := json.Marshal(donordata.Provided{LpaID: "123", Tasks: donordata.Tasks{PayForLpa: task.PaymentStatePending}})
			json.Unmarshal(b, v)
			return nil
		})

	documentStore := newMockDocumentStore(t)
	documentStore.EXPECT().
		UpdateScanResults(ctx, "123", "M-1111-2222-3333/evidence/a-uid", false).
		Return(expectedError)

	err := handleObjectTagsAdded(ctx, dynamoClient, event.S3Event, s3Client, documentStore)
	assert.Equal(t, fmt.Errorf("failed to update scan results: %w", expectedError), err)
}

func TestGetDonorByLpaUID(t *testing.T) {
	expectedDonor := &donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.EXPECT().
		One(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(nil).
		SetData(expectedDonor)

	donor, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Equal(t, expectedDonor, donor)
	assert.Nil(t, err)
}

func TestGetDonorByLpaUIDWhenClientOneByUidError(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{}, expectedError)

	lpa, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Nil(t, lpa)
	assert.Equal(t, fmt.Errorf("failed to resolve uid: %w", expectedError), err)
}

func TestGetDonorByLpaUIDWhenPKMissing(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{SK: dynamo.DonorKey("456")}, nil)

	donor, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Nil(t, donor)
	assert.Equal(t, errors.New("PK missing from LPA in response"), err)
}

func TestGetDonorByLpaUIDWhenClientOneError(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333").
		Return(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")}, nil)
	client.EXPECT().
		One(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(expectedError)

	donor, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Nil(t, donor)
	assert.Equal(t, fmt.Errorf("failed to get LPA: %w", expectedError), err)
}
