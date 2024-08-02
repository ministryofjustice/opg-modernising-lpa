// Package s3 provides a client for AWS S3.
package s3

import (
	"bytes"
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type s3Service interface {
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	DeleteObject(context.Context, *s3.DeleteObjectInput, ...func(*s3.Options)) (*s3.DeleteObjectOutput, error)
	GetObjectTagging(context.Context, *s3.GetObjectTaggingInput, ...func(*s3.Options)) (*s3.GetObjectTaggingOutput, error)
	DeleteObjects(context.Context, *s3.DeleteObjectsInput, ...func(*s3.Options)) (*s3.DeleteObjectsOutput, error)
	PutObjectTagging(context.Context, *s3.PutObjectTaggingInput, ...func(*s3.Options)) (*s3.PutObjectTaggingOutput, error)
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

func (c *Client) DeleteObjects(ctx context.Context, keys []string) error {
	var objectIdentifier []types.ObjectIdentifier
	for _, k := range keys {
		objectIdentifier = append(objectIdentifier, types.ObjectIdentifier{Key: aws.String(k)})
	}

	_, err := c.svc.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(c.bucket),
		Delete: &types.Delete{
			Objects: objectIdentifier,
		},
	})

	return err
}

func (c *Client) GetObjectTags(ctx context.Context, key string) ([]types.Tag, error) {
	output, err := c.svc.GetObjectTagging(ctx, &s3.GetObjectTaggingInput{
		Bucket: aws.String(c.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return []types.Tag{}, err
	}

	return output.TagSet, nil
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

func (c *Client) PutObjectTagging(ctx context.Context, key string, tags map[string]string) error {
	tagSet := make([]types.Tag, 0, len(tags))
	for k, v := range tags {
		tagSet = append(tagSet, types.Tag{Key: aws.String(k), Value: aws.String(v)})
	}

	_, err := c.svc.PutObjectTagging(ctx, &s3.PutObjectTaggingInput{
		Bucket:  aws.String(c.bucket),
		Key:     aws.String(key),
		Tagging: &types.Tagging{TagSet: tagSet},
	})

	return err
}
