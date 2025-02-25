package event

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var expectedError = errors.New("err")

func TestClientSendEvents(t *testing.T) {
	ctx := context.Background()
	uid := actoruid.New()

	testcases := map[string]func() (func(*Client) error, any){
		"uid-requested": func() (func(*Client) error, any) {
			event := UidRequested{LpaID: "5"}

			return func(client *Client) error { return client.SendUidRequested(ctx, event) }, event
		},
		"application-deleted": func() (func(*Client) error, any) {
			event := ApplicationDeleted{UID: "a"}

			return func(client *Client) error { return client.SendApplicationDeleted(ctx, event) }, event
		},
		"application-updated": func() (func(*Client) error, any) {
			event := ApplicationUpdated{UID: "a"}

			return func(client *Client) error { return client.SendApplicationUpdated(ctx, event) }, event
		},
		"reduced-fee-requested": func() (func(*Client) error, any) {
			event := ReducedFeeRequested{UID: "a"}

			return func(client *Client) error { return client.SendReducedFeeRequested(ctx, event) }, event
		},
		"notification-sent": func() (func(*Client) error, any) {
			event := NotificationSent{UID: "a", NotificationID: random.UuidString()}

			return func(client *Client) error { return client.SendNotificationSent(ctx, event) }, event
		},
		"paper-form-requested": func() (func(*Client) error, any) {
			event := PaperFormRequested{UID: "a", ActorType: "attorney", ActorUID: actoruid.New()}

			return func(client *Client) error { return client.SendPaperFormRequested(ctx, event) }, event
		},
		"payment-received": func() (func(*Client) error, any) {
			event := PaymentReceived{UID: "a", PaymentID: "xyz", Amount: 8200}

			return func(client *Client) error { return client.SendPaymentReceived(ctx, event) }, event
		},
		"certificate-provider-started": func() (func(*Client) error, any) {
			event := CertificateProviderStarted{UID: "a"}

			return func(client *Client) error { return client.SendCertificateProviderStarted(ctx, event) }, event
		},
		"attorney-started": func() (func(*Client) error, any) {
			event := AttorneyStarted{LpaUID: "a", ActorUID: uid}

			return func(client *Client) error { return client.SendAttorneyStarted(ctx, event) }, event
		},
		"identity-check-mismatched": func() (func(*Client) error, any) {
			event := IdentityCheckMismatched{LpaUID: "a", ActorUID: uid}

			return func(client *Client) error { return client.SendIdentityCheckMismatched(ctx, event) }, event
		},
		"correspondent-updated": func() (func(*Client) error, any) {
			event := CorrespondentUpdated{UID: "a"}

			return func(client *Client) error { return client.SendCorrespondentUpdated(ctx, event) }, event
		},
		"lpa-access-granted": func() (func(*Client) error, any) {
			event := LpaAccessGranted{UID: "a"}

			return func(client *Client) error { return client.SendLpaAccessGranted(ctx, event) }, event
		},
		"letter-requested": func() (func(*Client) error, any) {
			event := LetterRequested{UID: "a"}

			return func(client *Client) error { return client.SendLetterRequested(ctx, event) }, event
		},
		"confirm-at-post-office-selected": func() (func(*Client) error, any) {
			event := ConfirmAtPostOfficeSelected{UID: "a"}

			return func(client *Client) error { return client.SendConfirmAtPostOfficeSelected(ctx, event) }, event
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

func TestSendUnknownType(t *testing.T) {
	assert.Error(t, send[int](nil, nil, nil))
}

func TestEventsAllHaveSchemas(t *testing.T) {
	for _, name := range events {
		if info, _ := os.Stat("testdata/" + name + ".json"); info == nil {
			t.Fail()
			t.Log("missing schema for", name)
		}
	}
}
