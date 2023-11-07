package s3

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func TestDeleteObject(t *testing.T) {
	s3Service := newMockS3Service(t)
	s3Service.
		On("DeleteObject", context.Background(), &s3.DeleteObjectInput{
			Bucket: aws.String("a-bucket"),
			Key:    aws.String("a-key"),
		}).
		Return(nil, nil)

	client := Client{bucket: "a-bucket", svc: s3Service}
	err := client.DeleteObject(context.Background(), "a-key")

	assert.Nil(t, err)
}

func TestDeleteObjectOnServiceError(t *testing.T) {
	s3Service := newMockS3Service(t)
	s3Service.
		On("DeleteObject", context.Background(), &s3.DeleteObjectInput{
			Bucket: aws.String("a-bucket"),
			Key:    aws.String("a-key"),
		}).
		Return(nil, expectedError)

	client := Client{bucket: "a-bucket", svc: s3Service}
	err := client.DeleteObject(context.Background(), "a-key")

	assert.Equal(t, expectedError, err)
}

func TestPutObject(t *testing.T) {
	s3Service := newMockS3Service(t)
	s3Service.
		On("PutObject", context.Background(), &s3.PutObjectInput{
			Bucket:               aws.String("a-bucket"),
			Key:                  aws.String("a-object-key"),
			Body:                 bytes.NewReader([]byte("a-body")),
			ServerSideEncryption: types.ServerSideEncryptionAwsKms,
			Tagging:              aws.String("replicate=true"),
		}).
		Return(nil, nil)

	client := Client{bucket: "a-bucket", svc: s3Service}
	err := client.PutObject(context.Background(), "a-object-key", []byte("a-body"))

	assert.Nil(t, err)
}

func TestPutObjectOnServiceError(t *testing.T) {
	s3Service := newMockS3Service(t)
	s3Service.
		On("PutObject", context.Background(), mock.Anything).
		Return(nil, expectedError)

	client := Client{bucket: "a-bucket", svc: s3Service}
	err := client.PutObject(context.Background(), "a-object-key", []byte("a-body"))

	assert.Equal(t, expectedError, err)
}

func TestGetObjectTags(t *testing.T) {
	expectedTagSet := []types.Tag{
		{Key: aws.String("a-tag-key"), Value: aws.String("a-value")},
		{Key: aws.String("another-tag-key"), Value: aws.String("another-value")},
	}

	s3Service := newMockS3Service(t)
	s3Service.
		On("GetObjectTagging", context.Background(), &s3.GetObjectTaggingInput{
			Bucket: aws.String("a-bucket"),
			Key:    aws.String("a-object-key"),
		}).
		Return(&s3.GetObjectTaggingOutput{
			TagSet: expectedTagSet,
		}, nil)

	client := Client{bucket: "a-bucket", svc: s3Service}
	tags, err := client.GetObjectTags(context.Background(), "a-object-key")

	assert.Nil(t, err)
	assert.Equal(t, expectedTagSet, tags)
}

func TestGetObjectTagsOnServiceError(t *testing.T) {
	s3Service := newMockS3Service(t)
	s3Service.
		On("GetObjectTagging", context.Background(), &s3.GetObjectTaggingInput{
			Bucket: aws.String("a-bucket"),
			Key:    aws.String("a-object-key"),
		}).
		Return(&s3.GetObjectTaggingOutput{}, expectedError)

	client := Client{bucket: "a-bucket", svc: s3Service}
	_, err := client.GetObjectTags(context.Background(), "a-object-key")

	assert.Equal(t, expectedError, err)
}

func TestPutObjectTagging(t *testing.T) {
	s3Service := newMockS3Service(t)
	s3Service.
		On("PutObjectTagging", context.Background(), &s3.PutObjectTaggingInput{
			Bucket: aws.String("a-bucket"),
			Key:    aws.String("a-object-key"),
			Tagging: &types.Tagging{
				TagSet: []types.Tag{
					{Key: aws.String("a-tag-key"), Value: aws.String("a-value")},
					{Key: aws.String("another-tag-key"), Value: aws.String("another-value")},
				},
			},
		}).
		Return(nil, expectedError)

	client := Client{bucket: "a-bucket", svc: s3Service}

	err := client.PutObjectTagging(context.Background(), "a-object-key", map[string]string{
		"a-tag-key":       "a-value",
		"another-tag-key": "another-value",
	})

	assert.Equal(t, expectedError, err)
}
