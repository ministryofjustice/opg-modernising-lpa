package main

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

var expectedError = errors.New("err")
var ctx = context.Background()

func TestIsS3Event(t *testing.T) {
	s3Event := Event{S3Event: events.S3Event{Records: []events.S3EventRecord{{}, {}}}}

	assert.True(t, s3Event.isS3Event())

	s3Event.Records = []events.S3EventRecord{}

	assert.False(t, s3Event.isS3Event())
}

func TestIsCloudWatchEvent(t *testing.T) {
	cloudwatchEvents := []Event{
		{CloudWatchEvent: events.CloudWatchEvent{Source: "aws.cloudwatch"}},
		{CloudWatchEvent: events.CloudWatchEvent{Source: "opg.poas.makeregister"}},
		{CloudWatchEvent: events.CloudWatchEvent{Source: "opg.poas.sirius"}},
	}

	for _, e := range cloudwatchEvents {
		assert.True(t, e.isCloudWatchEvent())
	}

	assert.False(t, Event{CloudWatchEvent: events.CloudWatchEvent{Source: "somewhere else"}}.isCloudWatchEvent())
}
