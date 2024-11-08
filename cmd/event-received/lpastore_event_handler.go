package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

type lpastoreEventHandler struct{}

func (h *lpastoreEventHandler) Handle(ctx context.Context, factory factory, cloudWatchEvent *events.CloudWatchEvent) error {
	switch cloudWatchEvent.DetailType {
	case "lpa-updated":
		return handleLpaUpdated(ctx, factory.DynamoClient(), cloudWatchEvent, factory.Now())

	default:
		return fmt.Errorf("unknown lpastore event")
	}
}

type lpaUpdatedEvent struct {
	UID        string `json:"uid"`
	ChangeType string `json:"changeType"`
}

func handleLpaUpdated(ctx context.Context, client dynamodbClient, event *events.CloudWatchEvent, now func() time.Time) error {
	var v lpaUpdatedEvent
	if err := json.Unmarshal(event.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	if v.ChangeType != "STATUTORY_WAITING_PERIOD" {
		return nil
	}

	donor, err := getDonorByLpaUID(ctx, client, v.UID)
	if err != nil {
		return err
	}

	donor.StatutoryWaitingPeriodAt = now()

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update donor details: %w", err)
	}

	return nil
}
