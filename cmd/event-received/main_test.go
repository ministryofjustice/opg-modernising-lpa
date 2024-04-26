package main

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/stretchr/testify/assert"
)

var (
	expectedError = errors.New("err")
	ctx           = context.Background()

	testNow   = time.Date(2023, time.April, 2, 3, 4, 5, 6, time.UTC)
	testNowFn = func() time.Time { return testNow }

	testUuidString   = "a-uuid"
	testUuidStringFn = func() string { return testUuidString }
)

func TestIsS3Event(t *testing.T) {
	s3Event := Event{S3Event: events.S3Event{Records: []events.S3EventRecord{{}, {}}}}

	assert.True(t, s3Event.isS3Event())

	s3Event.Records = []events.S3EventRecord{}

	assert.False(t, s3Event.isS3Event())
}
