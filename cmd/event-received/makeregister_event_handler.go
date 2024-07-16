package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type makeregisterEventHandler struct{}

func (h *makeregisterEventHandler) Handle(ctx context.Context, factory factory, cloudWatchEvent events.CloudWatchEvent) error {
	switch cloudWatchEvent.DetailType {
	case "uid-requested":
		uidStore, err := factory.UidStore()
		if err != nil {
			return err
		}

		uidClient := factory.UidClient()
		dynamoClient := factory.DynamoClient()
		eventClient := factory.EventClient()

		return handleUidRequested(ctx, uidStore, uidClient, cloudWatchEvent, dynamoClient, eventClient)

	default:
		return fmt.Errorf("unknown makeregister event")
	}
}

func handleUidRequested(ctx context.Context, uidStore UidStore, uidClient UidClient, e events.CloudWatchEvent, dynamoClient dynamodbClient, eventClient EventClient) error {
	var v event.UidRequested
	if err := json.Unmarshal(e.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	uid, err := uidClient.CreateCase(ctx, &uid.CreateCaseRequestBody{Type: v.Type, Donor: v.Donor})
	if err != nil {
		return fmt.Errorf("failed to create case: %w", err)
	}

	if err := uidStore.Set(ctx, v.LpaID, v.DonorSessionID, v.OrganisationID, uid); err != nil {
		return fmt.Errorf("failed to set uid: %w", err)
	}

	donor, err := getDonorByLpaUID(ctx, dynamoClient, uid)
	if err != nil {
		return err
	}

	if err := eventClient.SendApplicationUpdated(ctx, event.ApplicationUpdated{
		UID:       donor.LpaUID,
		Type:      donor.Type.String(),
		CreatedAt: donor.CreatedAt,
		Donor: event.ApplicationUpdatedDonor{
			FirstNames:  donor.Donor.FirstNames,
			LastName:    donor.Donor.LastName,
			DateOfBirth: donor.Donor.DateOfBirth,
			Address:     donor.Donor.Address,
		},
	}); err != nil {
		return err
	}

	return nil
}
