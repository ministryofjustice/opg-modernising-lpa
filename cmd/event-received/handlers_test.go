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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandleObjectTagsAdded(t *testing.T) {
	testCases := map[string]bool{
		"ok":       false,
		"infected": true,
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
					{Key: aws.String("virus-scan-status"), Value: aws.String(scanResult)},
				}, nil)

			dynamoClient := newMockDynamodbClient(t)
			dynamoClient.
				On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
				Return(func(ctx context.Context, uid string, v interface{}) error {
					b, _ := json.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
					json.Unmarshal(b, v)
					return nil
				})
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
			{Key: aws.String("not-virus-scan-status"), Value: aws.String("ok")},
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
			{Key: aws.String("virus-scan-status"), Value: aws.String("ok")},
		}, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			json.Unmarshal(b, v)
			return nil
		})
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
			{Key: aws.String("virus-scan-status"), Value: aws.String("ok")},
		}, nil)

	dynamoClient := newMockDynamodbClient(t)
	dynamoClient.
		On("OneByUID", ctx, "M-1111-2222-3333", mock.Anything).
		Return(func(ctx context.Context, uid string, v interface{}) error {
			b, _ := json.Marshal(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
			json.Unmarshal(b, v)
			return nil
		})
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

func TestGetDonorByLPAUID(t *testing.T) {
	expectedDonor := &donordata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("456"))}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil).
		SetData(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
	client.EXPECT().
		One(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(nil).
		SetData(expectedDonor)

	donor, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Equal(t, expectedDonor, donor)
	assert.Nil(t, err)
}

func TestGetDonorByLPAUIDWhenClientOneByUidError(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	lpa, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Nil(t, lpa)
	assert.Equal(t, fmt.Errorf("failed to resolve uid: %w", expectedError), err)
}

func TestGetDonorByLPAUIDWhenPKMissing(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil).
		SetData(dynamo.Keys{SK: dynamo.DonorKey("456")})

	donor, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Nil(t, donor)
	assert.Equal(t, errors.New("PK missing from LPA in response"), err)
}

func TestGetDonorByLPAUIDWhenClientOneError(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil).
		SetData(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("456")})
	client.EXPECT().
		One(ctx, dynamo.LpaKey("123"), dynamo.DonorKey("456"), mock.Anything).
		Return(expectedError)

	donor, err := getDonorByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Nil(t, donor)
	assert.Equal(t, fmt.Errorf("failed to get LPA: %w", expectedError), err)
}

func TestGetCertificateProviderByLpaUID(t *testing.T) {
	expectedCertificateProvider := certificateproviderdata.Provided{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("")}

	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil).
		SetData(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456")})
	client.EXPECT().
		OneByPartialSK(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey(""), mock.Anything).
		Return(nil).
		SetData(expectedCertificateProvider)

	certificateProvider, err := getCertificateProviderByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Nil(t, err)
	assert.Equal(t, &expectedCertificateProvider, certificateProvider)
}

func TestGetCertificateProviderByLpaUIDWhenClientOneByUidError(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(expectedError)

	certificateProvider, err := getCertificateProviderByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Nil(t, certificateProvider)
	assert.Equal(t, fmt.Errorf("failed to resolve uid: %w", expectedError), err)
}

func TestGetCertificateProviderByLpaUIDWhenPKMissing(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil).
		SetData(dynamo.Keys{SK: dynamo.CertificateProviderKey("456")})

	certificateProvider, err := getCertificateProviderByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Nil(t, certificateProvider)
	assert.Equal(t, errors.New("PK missing from LPA in response"), err)
}

func TestGetCertificateProviderByLpaUIDWhenClientOneByPartialSKError(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		OneByUID(ctx, "M-1111-2222-3333", mock.Anything).
		Return(nil).
		SetData(dynamo.Keys{PK: dynamo.LpaKey("123"), SK: dynamo.CertificateProviderKey("456")})
	client.EXPECT().
		OneByPartialSK(ctx, dynamo.LpaKey("123"), dynamo.CertificateProviderKey(""), mock.Anything).
		Return(expectedError)

	certificateProvider, err := getCertificateProviderByLpaUID(ctx, client, "M-1111-2222-3333")

	assert.Nil(t, certificateProvider)
	assert.Equal(t, fmt.Errorf("failed to get certificate provider: %w", expectedError), err)
}

func TestPutCertificateProvider(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		Put(ctx, &certificateproviderdata.Provided{PK: dynamo.LpaKey("123"), UpdatedAt: testNow}).
		Return(nil)

	err := putCertificateProvider(ctx, &certificateproviderdata.Provided{PK: dynamo.LpaKey("123"), UpdatedAt: testNow}, testNowFn, client)

	assert.Nil(t, err)
}

func TestPutCertificateProviderWhenClientError(t *testing.T) {
	client := newMockDynamodbClient(t)
	client.EXPECT().
		Put(ctx, &certificateproviderdata.Provided{PK: dynamo.LpaKey("123"), UpdatedAt: testNow}).
		Return(expectedError)

	err := putCertificateProvider(ctx, &certificateproviderdata.Provided{PK: dynamo.LpaKey("123"), UpdatedAt: testNow}, testNowFn, client)

	assert.Equal(t, expectedError, err)
}
