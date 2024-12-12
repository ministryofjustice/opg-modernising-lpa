// Package event provides a client for AWS EventBridge.
package event

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

const source = "opg.poas.makeregister"

var events = map[any]string{
	(*UidRequested)(nil):               "uid-requested",
	(*ApplicationDeleted)(nil):         "application-deleted",
	(*ApplicationUpdated)(nil):         "application-updated",
	(*ReducedFeeRequested)(nil):        "reduced-fee-requested",
	(*NotificationSent)(nil):           "notification-sent",
	(*PaperFormRequested)(nil):         "paper-form-requested",
	(*PaymentReceived)(nil):            "payment-received",
	(*CertificateProviderStarted)(nil): "certificate-provider-started",
	(*AttorneyStarted)(nil):            "attorney-started",
	(*IdentityCheckMismatched)(nil):    "identity-check-mismatched",
	(*CorrespondentUpdated)(nil):       "correspondent-updated",
	(*LpaAccessGranted)(nil):           "lpa-access-granted",
	(*LetterRequested)(nil):            "letter-requested",
}

type eventbridgeClient interface {
	PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type Client struct {
	svc          eventbridgeClient
	eventBusName string
}

func NewClient(cfg aws.Config, eventBusName string) *Client {
	return &Client{
		svc:          eventbridge.NewFromConfig(cfg),
		eventBusName: eventBusName,
	}
}

func (c *Client) SendUidRequested(ctx context.Context, event UidRequested) error {
	return send[UidRequested](ctx, c, event)
}

func (c *Client) SendApplicationDeleted(ctx context.Context, event ApplicationDeleted) error {
	return send[ApplicationDeleted](ctx, c, event)
}

func (c *Client) SendApplicationUpdated(ctx context.Context, event ApplicationUpdated) error {
	return send[ApplicationUpdated](ctx, c, event)
}

func (c *Client) SendReducedFeeRequested(ctx context.Context, event ReducedFeeRequested) error {
	return send[ReducedFeeRequested](ctx, c, event)
}

func (c *Client) SendNotificationSent(ctx context.Context, event NotificationSent) error {
	return send[NotificationSent](ctx, c, event)
}

func (c *Client) SendPaperFormRequested(ctx context.Context, event PaperFormRequested) error {
	return send[PaperFormRequested](ctx, c, event)
}

func (c *Client) SendPaymentReceived(ctx context.Context, event PaymentReceived) error {
	return send[PaymentReceived](ctx, c, event)
}

func (c *Client) SendCertificateProviderStarted(ctx context.Context, event CertificateProviderStarted) error {
	return send[CertificateProviderStarted](ctx, c, event)
}

func (c *Client) SendAttorneyStarted(ctx context.Context, event AttorneyStarted) error {
	return send[AttorneyStarted](ctx, c, event)
}

func (c *Client) SendIdentityCheckMismatched(ctx context.Context, event IdentityCheckMismatched) error {
	return send[IdentityCheckMismatched](ctx, c, event)
}

func (c *Client) SendCorrespondentUpdated(ctx context.Context, event CorrespondentUpdated) error {
	return send[CorrespondentUpdated](ctx, c, event)
}

func (c *Client) SendLpaAccessGranted(ctx context.Context, event LpaAccessGranted) error {
	return send[LpaAccessGranted](ctx, c, event)
}

func (c *Client) SendLetterRequested(ctx context.Context, event LetterRequested) error {
	return send[LetterRequested](ctx, c, event)
}

func send[T any](ctx context.Context, c *Client, detail any) error {
	detailType, ok := events[(*T)(nil)]
	if !ok {
		return errors.New("event send of unknown type")
	}

	v, err := json.Marshal(detail)
	if err != nil {
		return err
	}

	_, err = c.svc.PutEvents(ctx, &eventbridge.PutEventsInput{
		Entries: []types.PutEventsRequestEntry{{
			EventBusName: aws.String(c.eventBusName),
			Source:       aws.String(source),
			DetailType:   aws.String(detailType),
			Detail:       aws.String(string(v)),
		}},
	})

	return err
}
