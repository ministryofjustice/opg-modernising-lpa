// Event logger is a tool to capture sqs events and display the most recent as a
// HTML page. It keeps the last 10 messages received, and those results can be
// filtered using ?detail-type and ?detail query parameters. It will wait 10
// seconds for a result when filtering.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/ministryofjustice/opg-go-common/env"
)

type message struct {
	Time       time.Time
	Detail     json.RawMessage
	DetailType string `json:"detail-type"`
	Source     string
}

const messageCount = 10

func main() {
	ctx := context.Background()

	var (
		awsBaseURL = env.Get("AWS_BASE_URL", "")
		port       = env.Get("PORT", "8080")
		queueName  = env.Get("QUEUE_NAME", "event-queue")
		queueURL   string
	)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to load SDK config: %w", err))
	}

	cfg.Region = "eu-west-1"
	cfg.EndpointResolverWithOptions = aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			PartitionID:   "aws",
			URL:           awsBaseURL,
			SigningRegion: "eu-west-1",
		}, nil
	})

	client := sqs.NewFromConfig(cfg)

	var messages []message

	go func() {
		for range time.Tick(time.Second) {
			if queueURL == "" {
				urlResponse, err := client.GetQueueUrl(ctx, &sqs.GetQueueUrlInput{
					QueueName: aws.String(queueName),
				})
				if err != nil {
					log.Println("failed to get queue url:", err)
					continue
				}

				queueURL = *urlResponse.QueueUrl
			}

			messageResponse, err := client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
				MessageAttributeNames: []string{string(types.QueueAttributeNameAll)},
				QueueUrl:              aws.String(queueURL),
				MaxNumberOfMessages:   5,
				VisibilityTimeout:     5,
			})
			if err != nil {
				log.Println("failed to retrieve message:", err)
				continue
			}

			log.Println("received", len(messageResponse.Messages))
			if len(messageResponse.Messages) == 0 {
				continue
			}

			var toDelete []types.DeleteMessageBatchRequestEntry

			for _, m := range messageResponse.Messages {
				toDelete = append(toDelete, types.DeleteMessageBatchRequestEntry{Id: m.MessageId, ReceiptHandle: m.ReceiptHandle})

				var v message
				if err := json.Unmarshal([]byte(*m.Body), &v); err != nil {
					log.Println("could not unmarshal message: ", err)
					continue
				}

				messages = append(messages, v)
			}

			deleteResponse, err := client.DeleteMessageBatch(ctx, &sqs.DeleteMessageBatchInput{
				QueueUrl: aws.String(queueURL),
				Entries:  toDelete,
			})
			if err != nil {
				log.Println("problem deleting messages:", err)
				continue
			}
			log.Println("deleting messages:", len(deleteResponse.Successful), "success,", len(deleteResponse.Failed), "failed")

			// trim to last N messages
			sort.Slice(messages, func(i, j int) bool {
				return messages[i].Time.After(messages[j].Time)
			})

			if len(messages) > messageCount {
				messages = messages[:messageCount]
			}
		}
	}()

	waitForMessages := func(detailType, detail string) []message {
		var matching []message
		done := make(chan struct{})
		count := 0

		go func() {
			for range time.Tick(time.Second) {
				count++
				matching = []message{}

				for _, m := range messages {
					if m.DetailType == detailType && strings.Contains(string(m.Detail), detail) {
						matching = append(matching, m)
					}
				}

				if len(matching) > 0 || count > 10 {
					done <- struct{}{}
				}
			}
		}()

		<-done
		return matching
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<!DOCTYPE html><body><table><thead><tr><th>Time</th><th>Source</th><th>DetailType</th><th>Detail</th></thead><tbody>")

		messages := messages
		if detailType := r.FormValue("detail-type"); detailType != "" {
			messages = waitForMessages(detailType, r.FormValue("detail"))
		}

		for _, m := range messages {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>", m.Time, m.Source, m.DetailType, m.Detail)
		}
		fmt.Fprint(w, "</tbody></table></body>")
	})

	http.ListenAndServe(":"+port, nil)
}
