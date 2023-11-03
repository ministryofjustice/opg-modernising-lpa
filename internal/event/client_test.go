package event

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/stretchr/testify/assert"
)

var expectedError = errors.New("err")

func TestClientSend(t *testing.T) {
	ctx := context.Background()

	svc := newMockEventbridgeClient(t)
	svc.
		On("PutEvents", ctx, &eventbridge.PutEventsInput{
			Entries: []types.PutEventsRequestEntry{{
				EventBusName: aws.String("my-bus"),
				Source:       aws.String("opg.poas.makeregister"),
				DetailType:   aws.String("my-detail"),
				Detail:       aws.String(`{"my":"event"}`),
			}},
		}).
		Return(nil, expectedError)

	client := Client{svc: svc, eventBusName: "my-bus"}
	err := client.send(ctx, "my-detail", map[string]string{"my": "event"})

	assert.Equal(t, expectedError, err)
}
