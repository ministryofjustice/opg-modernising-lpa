package event

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func TestClientSendEvents(t *testing.T) {
	ctx := context.Background()

	testcases := map[string]func() (func(*Client) error, any){
		"uid-requested": func() (func(*Client) error, any) {
			event := UidRequested{LpaID: "5"}

			return func(client *Client) error { return client.SendUidRequested(ctx, event) }, event
		},
		"application-updated": func() (func(*Client) error, any) {
			event := ApplicationUpdated{UID: "a"}

			return func(client *Client) error { return client.SendApplicationUpdated(ctx, event) }, event
		},
		"previous-application-linked": func() (func(*Client) error, any) {
			event := PreviousApplicationLinked{UID: "a"}

			return func(client *Client) error { return client.SendPreviousApplicationLinked(ctx, event) }, event
		},
		"reduced-fee-requested": func() (func(*Client) error, any) {
			event := ReducedFeeRequested{UID: "a"}

			return func(client *Client) error { return client.SendReducedFeeRequested(ctx, event) }, event
		},
		"notification-sent": func() (func(*Client) error, any) {
			event := NotificationSent{UID: "a", NotificationID: random.UuidString()}

			return func(client *Client) error { return client.SendNotificationSent(ctx, event) }, event
		},
	}

	for eventName, setup := range testcases {
		t.Run(eventName, func(t *testing.T) {
			fn, event := setup()
			data, _ := json.Marshal(event)

			svc := newMockEventbridgeClient(t)
			svc.EXPECT().
				PutEvents(mock.Anything, &eventbridge.PutEventsInput{
					Entries: []types.PutEventsRequestEntry{{
						EventBusName: aws.String("my-bus"),
						Source:       aws.String("opg.poas.makeregister"),
						DetailType:   aws.String(eventName),
						Detail:       aws.String(string(data)),
					}},
				}).
				Return(nil, expectedError)

			client := &Client{svc: svc, eventBusName: "my-bus"}
			err := fn(client)

			assert.Equal(t, expectedError, err)
		})
	}
}
