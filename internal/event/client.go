// Package event provides a client for AWS EventBridge.
package event

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
)

const source = "opg.poas.makeregister"

var events = map[any]string{
	(*UidRequested)(nil):                  "uid-requested",
	(*ApplicationDeleted)(nil):            "application-deleted",
	(*ApplicationUpdated)(nil):            "application-updated",
	(*ReducedFeeRequested)(nil):           "reduced-fee-requested",
	(*NotificationSent)(nil):              "notification-sent",
	(*PaperFormRequested)(nil):            "paper-form-requested",
	(*PaymentReceived)(nil):               "payment-received",
	(*CertificateProviderStarted)(nil):    "certificate-provider-started",
	(*AttorneyStarted)(nil):               "attorney-started",
	(*IdentityCheckMismatched)(nil):       "identity-check-mismatched",
	(*CorrespondentUpdated)(nil):          "correspondent-updated",
	(*LpaAccessGranted)(nil):              "lpa-access-granted",
	(*LetterRequested)(nil):               "letter-requested",
	(*ConfirmAtPostOfficeSelected)(nil):   "confirm-at-post-office-selected",
	(*RegisterWithCourtOfProtection)(nil): "register-with-court-of-protection",
	(*Metrics)(nil):                       "metric",
}

type eventbridgeClient interface {
	PutEvents(ctx context.Context, params *eventbridge.PutEventsInput, optFns ...func(*eventbridge.Options)) (*eventbridge.PutEventsOutput, error)
}

type Client struct {
	svc          eventbridgeClient
	eventBusName string
	environment  string
	now          func() time.Time
}

func NewClient(cfg aws.Config, eventBusName, environment string) *Client {
	return &Client{
		svc:          eventbridge.NewFromConfig(cfg),
		eventBusName: eventBusName,
		environment:  environment,
		now:          time.Now,
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

func (c *Client) SendConfirmAtPostOfficeSelected(ctx context.Context, event ConfirmAtPostOfficeSelected) error {
	return send[ConfirmAtPostOfficeSelected](ctx, c, event)
}

func (c *Client) SendRegisterWithCourtOfProtection(ctx context.Context, event RegisterWithCourtOfProtection) error {
	return send[RegisterWithCourtOfProtection](ctx, c, event)
}

func (c *Client) SendMetric(ctx context.Context, key string, category Category, measure Measure) error {
	hashedKey := sha256.Sum256([]byte(key))

	return send[Metrics](ctx, c, Metrics{
		Metrics: []MetricWrapper{{
			Metric: Metric{
				Project:          "MRLPA",
				Category:         category,
				Subcategory:      hex.EncodeToString(hashedKey[:]),
				Environment:      c.environment,
				MeasureName:      measure,
				MeasureValue:     "1",
				MeasureValueType: "BIGINT",
				Time:             strconv.FormatInt(c.now().UnixMilli(), 10),
			},
		}},
	})
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
