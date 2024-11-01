package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

type makeregisterEventHandler struct{}

func (h *makeregisterEventHandler) Handle(ctx context.Context, factory factory, cloudWatchEvent *events.CloudWatchEvent) error {
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

func handleUidRequested(ctx context.Context, uidStore UidStore, uidClient UidClient, e *events.CloudWatchEvent, dynamoClient dynamodbClient, eventClient EventClient) error {
	var v event.UidRequested
	if err := json.Unmarshal(e.Detail, &v); err != nil {
		return fmt.Errorf("failed to unmarshal detail: %w", err)
	}

	var sk dynamo.SK = dynamo.DonorKey(v.DonorSessionID)
	if v.OrganisationID != "" {
		sk = dynamo.OrganisationKey(v.OrganisationID)
	}

	var donor donordata.Provided
	if err := dynamoClient.One(ctx, dynamo.LpaKey(v.LpaID), sk, &donor); err != nil {
		return fmt.Errorf("failed to get donor: %w", err)
	}

	if donor.LpaUID != "" {
		return nil
	}

	uid, err := uidClient.CreateCase(ctx, &uid.CreateCaseRequestBody{Type: v.Type, Donor: v.Donor})
	if err != nil {
		return fmt.Errorf("failed to create case: %w", err)
	}

	if err := uidStore.Set(ctx, &donor, uid); err != nil {
		return fmt.Errorf("failed to set uid: %w", err)
	}

	if err := eventClient.SendApplicationUpdated(ctx, event.ApplicationUpdated{
		UID:       uid,
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
