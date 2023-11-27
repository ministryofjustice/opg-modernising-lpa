package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

func handleUidRequested(ctx context.Context, uidStore UidStore, uidClient UidClient, e events.CloudWatchEvent) error {
	var v event.UidRequested
	if err := json.Unmarshal(e.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	uid, err := uidClient.CreateCase(ctx, &uid.CreateCaseRequestBody{Type: v.Type, Donor: v.Donor})
	if err != nil {
		return fmt.Errorf("failed to create case: %w", err)
	}

	if err := uidStore.Set(ctx, v.LpaID, v.DonorSessionID, uid); err != nil {
		return fmt.Errorf("failed to set uid: %w", err)
	}

	return nil
}

func handleEvidenceReceived(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	var key dynamo.Key
	if err := client.OneByUID(ctx, v.UID, &key); err != nil {
		return fmt.Errorf("failed to resolve uid: %w", err)
	}

	if key.PK == "" {
		return fmt.Errorf("PK missing from LPA in response")
	}

	if err := client.Put(ctx, map[string]string{"PK": key.PK, "SK": "#EVIDENCE_RECEIVED"}); err != nil {
		return fmt.Errorf("failed to persist evidence received: %w", err)
	}

	return nil
}

func handleFeeApproved(ctx context.Context, dynamoClient dynamodbClient, event events.CloudWatchEvent, shareCodeSender shareCodeSender, appData page.AppData, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	lpa, err := getDonorByLpaUID(ctx, dynamoClient, v.UID)
	if err != nil {
		return err
	}

	lpa.Tasks.PayForLpa = actor.PaymentTaskCompleted
	lpa.UpdatedAt = now()

	if err := dynamoClient.Put(ctx, lpa); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	if err := shareCodeSender.SendCertificateProvider(ctx, notify.CertificateProviderReturnEmail, appData, &lpa); err != nil {
		return fmt.Errorf("failed to send share code to certificate provider: %w", err)
	}

	return nil
}

func handleMoreEvidenceRequired(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	donor.Tasks.PayForLpa = actor.PaymentTaskMoreEvidenceRequired
	donor.UpdatedAt = now()

	if err := client.Put(ctx, donor); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	return nil
}

func handleFeeDenied(ctx context.Context, client dynamodbClient, event events.CloudWatchEvent, now func() time.Time) error {
	var v uidEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	donor.Tasks.PayForLpa = actor.PaymentTaskDenied
	donor.UpdatedAt = now()

	if err := client.Put(ctx, donor); err != nil {
		return fmt.Errorf("failed to update LPA task status: %w", err)
	}

	return nil
}

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

func getDonorByLpaUID(ctx context.Context, client dynamodbClient, uid string) (actor.DonorProvidedDetails, error) {
	var key dynamo.Key
	if err := client.OneByUID(ctx, uid, &key); err != nil {
		return actor.DonorProvidedDetails{}, fmt.Errorf("failed to resolve uid: %w", err)
	}

	if key.PK == "" {
		return actor.DonorProvidedDetails{}, fmt.Errorf("PK missing from LPA in response")
	}

	var donor actor.DonorProvidedDetails
	if err := client.One(ctx, key.PK, key.SK, &donor); err != nil {
		return actor.DonorProvidedDetails{}, fmt.Errorf("failed to get LPA: %w", err)
	}

	return donor, nil
}
