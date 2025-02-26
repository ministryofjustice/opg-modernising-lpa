package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
)

type sqsClient struct {
	svc      *sqs.Client
	queueURL string
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

func (c sqsClient) SendMessage(ctx context.Context, source, detailType string, detail json.RawMessage) error {
	v, err := json.Marshal(CloudWatchEvent{
		Version:    "0",
		ID:         "63eb7e5f-1f10-4744-bba9-e16d327c3b98",
		DetailType: detailType,
		Source:     source,
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
