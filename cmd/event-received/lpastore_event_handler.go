package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

type lpastoreEventHandler struct{}

type lpaUpdatedEvent struct {
	UID        string `json:"uid"`
	ChangeType string `json:"changeType"`
}

func (h *lpastoreEventHandler) Handle(ctx context.Context, factory factory, cloudWatchEvent *events.CloudWatchEvent) error {
	if cloudWatchEvent.DetailType == "lpa-updated" {
		var v lpaUpdatedEvent
		if err := json.Unmarshal(cloudWatchEvent.Detail, &v); err != nil {
			return fmt.Errorf("failed to unmarshal detail: %w", err)
		}

		switch v.ChangeType {
		case "STATUTORY_WAITING_PERIOD":
			return handleStatutoryWaitingPeriod(ctx, factory.DynamoClient(), factory.Now(), v)

		case "CANNOT_REGISTER":
			return handleCannotRegister(ctx, factory.ScheduledStore(), v)
		}
	}

	return fmt.Errorf("unknown lpastore event")
}

func handleStatutoryWaitingPeriod(ctx context.Context, client dynamodbClient, now func() time.Time, event lpaUpdatedEvent) error {
	donor, err := getDonorByLpaUID(ctx, client, event.UID)
	if err != nil {
		return err
	}

	donor.StatutoryWaitingPeriodAt = now()

	if err := putDonor(ctx, donor, now, client); err != nil {
		return fmt.Errorf("failed to update donor details: %w", err)
	}

	return nil
}

func handleCannotRegister(ctx context.Context, store ScheduledStore, event lpaUpdatedEvent) error {
	return store.DeleteAllByUID(ctx, event.UID)
}
