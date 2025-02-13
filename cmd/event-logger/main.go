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
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
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

type sqsClient struct {
	svc      *sqs.Client
	queueURL string
}

type message struct {
	Time       time.Time
	Detail     json.RawMessage
	DetailType string `json:"detail-type"`
	Source     string
}

var cfg aws.Config

func main() {
	var (
		awsBaseURL = env.Get("AWS_BASE_URL", "")
		port       = env.Get("PORT", "8080")
		queueName  = env.Get("QUEUE_NAME", "event-queue")

		ctx      = context.Background()
		queueURL string
		messages []message
	)

	var err error
	cfg, err = config.LoadDefaultConfig(ctx)
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
				MessageAttributeNames: []string{string(sqstypes.QueueAttributeNameAll)},
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

			var toDelete []sqstypes.DeleteMessageBatchRequestEntry

			for _, m := range messageResponse.Messages {
				toDelete = append(toDelete, sqstypes.DeleteMessageBatchRequestEntry{Id: m.MessageId, ReceiptHandle: m.ReceiptHandle})

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

	http.HandleFunc("/emit-sirius-event", func(w http.ResponseWriter, r *http.Request) {
		sqsClient := &sqsClient{
			svc:      sqs.NewFromConfig(cfg),
			queueURL: "http://localhost:4566/000000000000/event-bus-queue",
		}

		switch r.FormValue("detailType") {
		case "immaterial-change-confirmed":
			if err = sqsClient.SendImmaterialChangeConfirmed(
				r.Context(),
				r.FormValue("uid"),
				r.FormValue("actorType"),
				r.FormValue("actorUID"),
			); err != nil {
				log.Printf("failed to send immaterial-change-confirmed: %v", err)
			}
		case "material-change-confirmed":
			if err = sqsClient.SendMaterialChangeConfirmed(
				r.Context(),
				r.FormValue("uid"),
				r.FormValue("actorType"),
				r.FormValue("actorUID"),
			); err != nil {
				log.Printf("failed to send material-change-confirmed: %v", err)
			}
		default:
			log.Println("unsupported event type:", r.FormValue("detailType"))
		}

		log.Println("successfully handled", r.FormValue("detailType"))

	})

	http.ListenAndServe(":"+port, nil)
}

type CloudWatchEvent struct {
	Version    string          `json:"version"`
	ID         string          `json:"id"`
	DetailType string          `json:"detail-type"`
	Source     string          `json:"source"`
	Account    string          `json:"account"`
	Time       time.Time       `json:"time"`
	Region     string          `json:"region"`
	Resources  []string        `json:"resources"`
	Detail     json.RawMessage `json:"detail"`
}

type ChangeConfirmedEvent struct {
	UID       string `json:"uid"`
	ActorType string `json:"actorType"`
	ActorUID  string `json:"actorUID"`
}

func (c sqsClient) SendImmaterialChangeConfirmed(ctx context.Context, lpaUID, actorType, actorUID string) error {
	detail, _ := json.Marshal(ChangeConfirmedEvent{
		UID:       lpaUID,
		ActorType: actorType,
		ActorUID:  actorUID,
	})

	return c.SendMessage(ctx, "immaterial-change-confirmed", detail)
}

func (c sqsClient) SendMaterialChangeConfirmed(ctx context.Context, lpaUID, actorType, actorUID string) error {
	detail, _ := json.Marshal(ChangeConfirmedEvent{
		UID:       lpaUID,
		ActorType: actorType,
		ActorUID:  actorUID,
	})

	return c.SendMessage(ctx, "material-change-confirmed", detail)
}

func (c sqsClient) SendMessage(ctx context.Context, detailType string, detail json.RawMessage) error {
	v, err := json.Marshal(CloudWatchEvent{
		Version:    "0",
		ID:         "63eb7e5f-1f10-4744-bba9-e16d327c3b98",
		DetailType: detailType,
		Source:     "opg.poas.sirius",
		Account:    "653761790766",
		Time:       time.Now().UTC(),
		Region:     "eu-west-1",
		Resources:  []string{},
		Detail:     detail,
	})
	if err != nil {
		return err
	}

	if _, err = c.svc.SendMessage(ctx, &sqs.SendMessageInput{
		QueueUrl:    aws.String(c.queueURL),
		MessageBody: aws.String(string(v)),
	}); err != nil {
		return fmt.Errorf("failed to send %s message: %w", detailType, err)
	}

	return nil
}
