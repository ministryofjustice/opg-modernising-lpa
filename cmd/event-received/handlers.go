package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
)

func handleObjectTagsAdded(ctx context.Context, dynamodbClient dynamodbClient, event events.S3Event, s3Client s3Client, documentStore DocumentStore) error {
	objectKey := event.Records[0].S3.Object.Key
	if objectKey == "" {
		return fmt.Errorf("object key missing")
	}

	tags, err := s3Client.GetObjectTags(ctx, objectKey)
	if err != nil {
		return fmt.Errorf("failed to get tags for object: %w", err)
	}

	hasScannedTag := false
	hasVirus := false

	for _, tag := range tags {
		if *tag.Key == "virus-scan-status" {
			hasScannedTag = true
			hasVirus = *tag.Value == virusFound
			break
		}
	}

	if !hasScannedTag {
		return nil
	}

	parts := strings.Split(objectKey, "/")

	donor, err := getDonorByLpaUID(ctx, dynamodbClient, parts[0])
	if err != nil {
		return err
	}

	err = documentStore.UpdateScanResults(ctx, donor.LpaID, objectKey, hasVirus)
	if err != nil {
		return fmt.Errorf("failed to update scan results: %w", err)
	}

	return nil
}

func putDonor(ctx context.Context, donor *donordata.Provided, now func() time.Time, client dynamodbClient) error {
	donor.UpdatedAt = now()
	if err := donor.UpdateHash(); err != nil {
		return err
	}

	return client.Put(ctx, donor)
}

func getDonorByLpaUID(ctx context.Context, client dynamodbClient, uid string) (*donordata.Provided, error) {
	var key dynamo.Keys
	if err := client.OneByUID(ctx, uid, &key); err != nil {
		return nil, fmt.Errorf("failed to resolve uid: %w", err)
	}

	if key.PK == nil {
		return nil, fmt.Errorf("PK missing from LPA in response")
	}

	var donor donordata.Provided
	if err := client.One(ctx, key.PK, key.SK, &donor); err != nil {
		return nil, fmt.Errorf("failed to get LPA: %w", err)
	}

	return &donor, nil
}
