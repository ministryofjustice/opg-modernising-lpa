package app

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page/donor"
)

type evidenceStore struct {
	dynamoClient       DynamoClient
	s3Client           S3Client
	evidenceBucketName string
	randomUuid         func() string
}

func (e evidenceStore) Create(ctx context.Context, lpa *page.Lpa, files []donor.File) error {
	for _, file := range files {
		uuid := e.randomUuid()
		key := lpa.UID + "-evidence-" + uuid

		_, err := e.s3Client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:               aws.String(e.evidenceBucketName),
			Key:                  aws.String(key),
			Body:                 bytes.NewReader(file.Data),
			ServerSideEncryption: types.ServerSideEncryptionAwsKms,
		})
		if err != nil {
			return err
		}

		evidence := &page.Evidence{
			Key: key,
		}

		if err := e.dynamoClient.Create(ctx, evidence); err != nil {
			return err
		}
	}

	return nil
}
