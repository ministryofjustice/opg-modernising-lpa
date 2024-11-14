package main

import (
	"cmp"
	"context"
	"flag"
	"log"
	"os"
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
	taskCount := flag.Int("taskCount", 10, "the number of scheduled tasks to generate")
	flag.Parse()

	ctx := context.Background()

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		log.Fatal("failed to load default config: %w", err)
	}

	awsBaseURL := cmp.Or(os.Getenv("AWS_BASE_URL"), "http://localhost:4566")

	cfg.BaseEndpoint = aws.String(awsBaseURL)

	if awsBaseURL == "http://localhost:4566" {
		cfg.Credentials = credentials.NewStaticCredentialsProvider(
			"test",
			"test",
			"test",
		)

		cfg.Region = "eu-west-1"
	}

	client, err := dynamo.NewClient(cfg, "lpas")
	if err != nil {
		log.Fatal("failed to create dynamo client: %w", err)
	}

	const batchSize = 100
	donorChan := make(chan any, batchSize)
	eventChan := make(chan any, batchSize)

	start := time.Now()

	go func() {
		now := time.Now()
		for i := 0; i < *taskCount; i++ {
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

			eventChan <- event
			donorChan <- donor
		}
		close(eventChan)
		close(donorChan)
	}()

	addedEvents := 0

	var donorBatch, eventBatch []any

	for {
		select {
		case donor, ok := <-donorChan:
			if !ok {
				if len(donorBatch) > 0 {
					if err := client.BatchPut(ctx, donorBatch); err != nil {
						log.Printf("failed to write remaining donors: %v", err)
					}
				}

				break
			}

			donorBatch = append(donorBatch, donor)
			if len(donorBatch) == batchSize {
				if err := client.BatchPut(ctx, donorBatch); err != nil {
					log.Printf("failed to write batched donors: %v", err)
				}

				donorBatch = donorBatch[:0]
			}

		case event, ok := <-eventChan:
			if !ok {
				if len(eventBatch) > 0 {
					if err := client.BatchPut(ctx, eventBatch); err != nil {
						log.Printf("failed to write batched events: %v", err)
					}

					addedEvents += len(eventBatch)
				}

				elapsed := time.Since(start)
				log.Printf("Execution time: %s", elapsed)
				log.Printf("Added %d tasks", addedEvents)

				return
			}

			eventBatch = append(eventBatch, event)
			if len(eventBatch) == batchSize {
				if err := client.BatchPut(ctx, eventBatch); err != nil {
					log.Printf("failed to write batched events: %v", err)
				}

				addedEvents += batchSize
				eventBatch = eventBatch[:0]
			}
		}
	}

}
