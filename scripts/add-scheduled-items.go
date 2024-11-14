package main

import (
	"context"
	"log"
	"sync"
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
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			"test",
			"test",
			"test",
		)),
	)

	cfg.BaseEndpoint = aws.String("http://localhost:4566")

	if err != nil {
		log.Fatal("failed to load default config: %w", err)
	}

	client, err := dynamo.NewClient(cfg, "lpas")
	if err != nil {
		log.Fatal("failed to create dynamo client: %w", err)
	}

	const batchSize = 100
	itemChan := make(chan any, batchSize)
	var wg sync.WaitGroup

	start := time.Now()

	go func() {
		now := time.Now()
		for i := 0; i < 10000; i++ {
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

			itemChan <- donor
			itemChan <- event
		}
		close(itemChan)
	}()

	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			batchItems := make([]any, 0, batchSize)
			for item := range itemChan {
				batchItems = append(batchItems, item)
				if len(batchItems) == batchSize {
					if err := client.BatchPut(ctx, batchItems); err != nil {
						log.Printf("failed to write batch: %v", err)
					}
					batchItems = batchItems[:0]
				}
			}
			if len(batchItems) > 0 {
				if err := client.BatchPut(ctx, batchItems); err != nil {
					log.Printf("failed to write batch: %v", err)
				}
			}
		}()
	}

	wg.Wait()

	elapsed := time.Since(start)
	log.Printf("Execution time: %s", elapsed)
}
