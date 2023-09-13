package app

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Create", ctx, &page.Evidence{
			PK:  "LPA#lpa-id",
			SK:  "#FEE_EVIDENCE#a-uuid",
			Key: "lpa-uid-evidence-a-uuid",
		}).
		Return(nil)

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", ctx, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return assert.Equal(t, aws.String("bucket-name"), input.Bucket) &&
				assert.Equal(t, aws.String("lpa-uid-evidence-a-uuid"), input.Key) &&
				assert.Equal(t, types.ServerSideEncryptionAwsKms, input.ServerSideEncryption)
		})).
		Return(nil, nil)

	store := evidenceStore{dynamoClient: dynamoClient, s3Client: s3Client, evidenceBucketName: "bucket-name", randomUuid: func() string { return "a-uuid" }}

	evidence, err := store.Create(ctx, &page.Lpa{UID: "lpa-uid", ID: "lpa-id", PK: "LPA#lpa-id"}, []byte("file contents"))

	assert.Nil(t, err)
	assert.Equal(t, &page.Evidence{
		PK:  "LPA#lpa-id",
		SK:  "#FEE_EVIDENCE#a-uuid",
		Key: "lpa-uid-evidence-a-uuid",
	}, evidence)
}

func TestCreateWhenS3ClientError(t *testing.T) {
	ctx := context.Background()

	evidence := &page.Evidence{
		PK:  "LPA#lpa-id",
		SK:  "#FEE_EVIDENCE#a-uuid",
		Key: "lpa-uid-evidence-a-uuid",
	}

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", ctx, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return assert.Equal(t, aws.String("bucket-name"), input.Bucket) &&
				assert.Equal(t, aws.String("lpa-uid-evidence-a-uuid"), input.Key) &&
				assert.Equal(t, types.ServerSideEncryptionAwsKms, input.ServerSideEncryption)
		})).
		Return(nil, expectedError)

	store := evidenceStore{dynamoClient: nil, s3Client: s3Client, evidenceBucketName: "bucket-name", randomUuid: func() string { return "a-uuid" }}

	evidence, err := store.Create(ctx, &page.Lpa{UID: "lpa-uid", ID: "lpa-id", PK: "LPA#lpa-id"}, []byte(""))

	assert.Equal(t, expectedError, err)
	assert.Equal(t, &page.Evidence{}, evidence)
}

func TestCreateWhenDynamoError(t *testing.T) {
	ctx := context.Background()

	evidence := &page.Evidence{
		PK:  "LPA#lpa-id",
		SK:  "#FEE_EVIDENCE#a-uuid",
		Key: "lpa-uid-evidence-a-uuid",
	}

	s3Client := newMockS3Client(t)
	s3Client.
		On("PutObject", ctx, mock.MatchedBy(func(input *s3.PutObjectInput) bool {
			return assert.Equal(t, aws.String("bucket-name"), input.Bucket) &&
				assert.Equal(t, aws.String("lpa-uid-evidence-a-uuid"), input.Key) &&
				assert.Equal(t, types.ServerSideEncryptionAwsKms, input.ServerSideEncryption)
		})).
		Return(nil, nil)

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Create", ctx, evidence).
		Return(expectedError)

	store := evidenceStore{dynamoClient: dynamoClient, s3Client: s3Client, evidenceBucketName: "bucket-name", randomUuid: func() string { return "a-uuid" }}

	evidence, err := store.Create(ctx, &page.Lpa{UID: "lpa-uid", ID: "lpa-id", PK: "LPA#lpa-id"}, []byte(""))

	assert.Equal(t, expectedError, err)
	assert.Equal(t, &page.Evidence{}, evidence)
}
