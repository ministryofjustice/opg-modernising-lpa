package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
)

var taskCount int
var entryPoint string

type TaskCountEvent struct {
	TaskCount int `json:"taskCount"`
}

func init() {
	flag.IntVar(&taskCount, "taskCount", 0, "Number of scheduled tasks to generate")
}

func handleAddScheduledTasks(ctx context.Context, taskCountEvent TaskCountEvent) error {
	flag.Parse()

	if taskCount == 0 {
		taskCount = taskCountEvent.TaskCount
	}

	cfg, err := config.LoadDefaultConfig(ctx)

	if err != nil {
		return fmt.Errorf("failed to load default config: %w", err)
	}

	awsBaseURL := cmp.Or(os.Getenv("AWS_BASE_URL"), "http://localstack:4566")

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
		return fmt.Errorf("failed to create dynamo client: %w", err)
	}

	const batchSize = 100
	donorChan := make(chan any, batchSize)
	eventChan := make(chan any, batchSize)

	start := time.Now()

	go func() {
		now := time.Now()
		for i := 0; i < taskCount; i++ {
			now = now.Add(time.Second * 1)

			donor := &donordata.Provided{
				LpaUID:           uuid.NewString(),
				PK:               dynamo.LpaKey(uuid.NewString()),
				SK:               dynamo.LpaOwnerKey(dynamo.DonorKey(uuid.NewString())),
				IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
				Donor:            donordata.Donor{Email: fmt.Sprintf("a%d@example.com", i)},
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
						return fmt.Errorf("failed to write remaining donors: %v", err)
					}
				}

				break
			}

			donorBatch = append(donorBatch, donor)
			if len(donorBatch) == batchSize {
				if err := client.BatchPut(ctx, donorBatch); err != nil {
					return fmt.Errorf("failed to write batched donors: %v", err)
				}

				donorBatch = donorBatch[:0]
			}

		case event, ok := <-eventChan:
			if !ok {
				if len(eventBatch) > 0 {
					if err := client.BatchPut(ctx, eventBatch); err != nil {
						return fmt.Errorf("failed to write batched events: %v", err)
					}

					addedEvents += len(eventBatch)
				}

				// as donor channel breaks when finished, this should be the last thing printed
				elapsed := time.Since(start)
				log.Printf("Execution time: %s", elapsed)
				log.Printf("Added %d tasks", addedEvents)

				return nil
			}

			eventBatch = append(eventBatch, event)
			if len(eventBatch) == batchSize {
				if err := client.BatchPut(ctx, eventBatch); err != nil {
					return fmt.Errorf("failed to write batched events: %v", err)
				}

				addedEvents += batchSize
				eventBatch = eventBatch[:0]
			}
		}
	}
}

func main() {
	if entryPoint == "local" {
		ctx := context.Background()

		if err := handleAddScheduledTasks(ctx, TaskCountEvent{TaskCount: taskCount}); err != nil {
			log.Fatal(err)
		}
	} else {
		lambda.Start(handleAddScheduledTasks)
	}
}
