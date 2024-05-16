package event

import (
	"context"
	"encoding/json"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

const source = "opg.poas.makeregister"

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
	return c.send(ctx, "uid-requested", event)
}

func (c *Client) SendApplicationUpdated(ctx context.Context, event ApplicationUpdated) error {
	return c.send(ctx, "application-updated", event)
}

func (c *Client) SendPreviousApplicationLinked(ctx context.Context, event PreviousApplicationLinked) error {
	return c.send(ctx, "previous-application-linked", event)
}

func (c *Client) SendReducedFeeRequested(ctx context.Context, event ReducedFeeRequested) error {
	return c.send(ctx, "reduced-fee-requested", event)
}

func (c *Client) SendNotificationSent(ctx context.Context, event NotificationSent) error {
	return c.send(ctx, "notification-sent", event)
}

func (c *Client) SendPaperFormRequested(ctx context.Context, event PaperFormRequested) error {
	return c.send(ctx, "paper-form-requested", event)
}

func (c *Client) SendPaymentCreated(ctx context.Context, event PaymentCreated) error {
	return c.send(ctx, "payment-created", event)
}

func (c *Client) send(ctx context.Context, detailType string, detail any) error {
	tracer := otel.GetTracerProvider().Tracer("mlpab")
	ctx, span := tracer.Start(ctx, detailType,
		trace.WithSpanKind(trace.SpanKindInternal),
	)
	defer span.End()

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
