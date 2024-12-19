package main

import (
	"cmp"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/google/uuid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
)

const batchSize = 100

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

	awsBaseURL := cmp.Or(os.Getenv("AWS_BASE_URL"), "http://localhost:4566")

	cfg.BaseEndpoint = aws.String(awsBaseURL)

	if !strings.Contains(awsBaseURL, "https") {
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

	var items []any
	now := time.Now()
	start := time.Now()

	for i := 0; i < taskCount; i++ {
		now = now.Add(time.Second * 1)

		lpaUID := uuid.NewString()
		lpaID := uuid.NewString()

		donor := &donordata.Provided{
			LpaUID:           lpaUID,
			PK:               dynamo.LpaKey(lpaID),
			SK:               dynamo.LpaOwnerKey(dynamo.DonorKey(uuid.NewString())),
			IdentityUserData: identity.UserData{Status: identity.StatusConfirmed},
			Donor:            donordata.Donor{Email: "a@b.com"},
			LpaID:            lpaID,
		}

		event := scheduled.Event{
			CreatedAt:         now,
			At:                now,
			Action:            scheduled.ActionExpireDonorIdentity,
			TargetLpaKey:      donor.PK,
			TargetLpaOwnerKey: donor.SK,
			LpaUID:            lpaUID,
			PK:                dynamo.ScheduledDayKey(now),
			SK:                dynamo.ScheduledKey(now, random.UuidString()),
		}

		items = append(items, donor, event)

		if len(items) == batchSize {
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

	log.Printf("Time taken: %s", time.Since(start))
	return nil
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
