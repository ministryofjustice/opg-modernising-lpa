package main

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
)

func main() {
	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("eu-west-1"),
		config.WithEndpointResolver(aws.EndpointResolverFunc(
			func(service, region string) (aws.Endpoint, error) {
				return aws.Endpoint{
					URL: "http://localhost:4566",
				}, nil
			},
		)),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test",
			"test",
			"test",
		)),
	)

	if err != nil {
		log.Fatal("failed to load default config: %w", err)
	}

	client, err := dynamo.NewClient(cfg, "lpas")
	if err != nil {
		log.Fatal("failed to create dynamo client: %w", err)
	}

	var items []any
	now := time.Now()

	for i := 0; i < 1; i++ {
		now = now.Add(time.Second * 1)

		donor := &donordata.Provided{
			LpaUID:           uuid.NewString(),
			PK:               dynamo.LpaKey(uuid.NewString()),
			SK:               dynamo.LpaOwnerKey(dynamo.DonorKey(uuid.NewString())),
			IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			Donor:            donordata.Donor{Email: "a@b.com"},
		}

		event := scheduled.Event{
			At:                now,
			Action:            scheduled.ActionExpireDonorIdentity,
			TargetLpaKey:      donor.PK,
			TargetLpaOwnerKey: donor.SK,
			PK:                dynamo.ScheduledDayKey(now),
			SK:                dynamo.ScheduledKey(now, int(scheduled.ActionExpireDonorIdentity)),
		}

		items = append(items, donor, event)

		if len(items) == 100 {
			if err := client.BatchPut(ctx, items); err != nil {
				log.Fatal(err)
			}

			items = []any{}
		}
	}

	if len(items) > 0 {
		if err := client.BatchPut(ctx, items); err != nil {
			log.Fatal(err)
		}
	}
}
