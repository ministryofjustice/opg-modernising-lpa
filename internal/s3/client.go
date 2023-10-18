package s3

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

//go:generate mockery --testonly --inpackage --name s3Service --structname mockS3Service
type s3Service interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	PutObjectTagging(context.Context, *s3.PutObjectTaggingInput, ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
}

type Client struct {
	bucket string
	svc    s3Service
}

func NewClient(cfg aws.Config, bucket string) *Client {
	return &Client{bucket: bucket, svc: s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})}
}

func (c *Client) DeleteObject(ctx context.Context, key string) error {
	_, err := c.svc.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	return err
}

func (c *Client) PutObjectTagging(ctx context.Context, key string, tagSet []types.Tag) error {
	_, err := c.svc.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket:  aws.String(c.bucket),
		Key:     aws.String(key),
		Tagging: &types.Tagging{TagSet: tagSet},
	})

	return err
}

func (c *Client) PutObject(ctx context.Context, key string, body []byte) error {
	_, err := c.svc.PutObject(ctx, &s3.PutObjectInput{
		Bucket:               aws.String(c.bucket),
		Key:                  aws.String(key),
		Body:                 bytes.NewReader(body),
		ServerSideEncryption: types.ServerSideEncryptionAwsKms,
	})

	return err
}
