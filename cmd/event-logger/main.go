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

const (
	// duration of ticks for receiving messages
	receiveTick = time.Second
	// maximum number of messages to remember
	maxMessages = 10
	// duration of ticks when filtering messages
	waitTick = time.Second
	// when filtering messages how many matches to wait for
	waitMinimum = 1
	// when filtering messages how many "ticks" to wait before returning a response
	waitMaxTicks = 10
)

type message struct {
	Time       time.Time
	Detail     json.RawMessage
	DetailType string `json:"detail-type"`
	Source     string
}

func main() {
	var (
		awsBaseURL = env.Get("AWS_BASE_URL", "")
		port       = env.Get("PORT", "8080")
		queueName  = env.Get("QUEUE_NAME", "event-queue")

		ctx      = context.Background()
		queueURL string
		messages []message
	)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Fatal(fmt.Errorf("unable to load SDK config: %w", err))
	}

	cfg.Region = "eu-west-1"
	cfg.BaseEndpoint = aws.String(awsBaseURL)

	client := sqs.NewFromConfig(cfg)

	go func() {
		for range time.Tick(receiveTick) {
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
				MaxNumberOfMessages:   10, // may as well ask for as many as possible
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

			if len(messages) > maxMessages {
				messages = messages[:maxMessages]
			}
		}
	}()

	filterMessages := func(detailType, detail string) []message {
		if detailType == "" {
			return messages
		}

		var matching []message
		done := make(chan struct{})
		count := 0

		go func() {
			for range time.Tick(waitTick) {
				count++

				for _, m := range messages {
					if m.DetailType == detailType && strings.Contains(string(m.Detail), detail) {
						matching = append(matching, m)
					}
				}

				if len(matching) >= waitMinimum || count > waitMaxTicks {
					done <- struct{}{}
					break
				}
			}
		}()

		<-done
		return matching
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<!DOCTYPE html><body><table><thead><tr><th>Time</th><th>Source</th><th>DetailType</th><th>Detail</th></thead><tbody>")

		for _, m := range filterMessages(r.FormValue("detail-type"), r.FormValue("detail")) {
			fmt.Fprintf(w, "<tr><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>", m.Time, m.Source, m.DetailType, m.Detail)
		}

		fmt.Fprint(w, "</tbody></table></body>")
	})

	http.ListenAndServe(":"+port, nil)
}
