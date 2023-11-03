package event

import (
	"context"
	"encoding/json"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge"
	"github.com/aws/aws-sdk-go-v2/service/eventbridge/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/uid"
)

const source = "opg.poas.makeregister"

//go:generate mockery --testonly --inpackage --name eventbridgeClient --structname mockEventbridgeClient
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

type UidRequested struct {
	ID             string
	DonorSessionID string
	Type           string
	Donor          uid.DonorDetails
}

func (c *Client) SendUidRequested(ctx context.Context, event UidRequested) error {
	return c.send(ctx, "uid-requested", event)
}

type ApplicationUpdated struct {
	UID       string                  `json:"uid"`
	Type      string                  `json:"type"`
	CreatedAt time.Time               `json:"createdAt"`
	Donor     ApplicationUpdatedDonor `json:"donor"`
}

type ApplicationUpdatedDonor struct {
	FirstNames  string        `json:"firstNames"`
	LastName    string        `json:"lastName"`
	DateOfBirth date.Date     `json:"dob"`
	Address     place.Address `json:"address"`
}

func (c *Client) SendApplicationUpdated(ctx context.Context, event ApplicationUpdated) error {
	return c.send(ctx, "application-updated", event)
}

type PreviousApplicationLinked struct {
	UID                       string `json:"uid"`
	PreviousApplicationNumber string `json:"previousApplicationNumber"`
}

func (c *Client) SendPreviousApplicationLinked(ctx context.Context, event PreviousApplicationLinked) error {
	return c.send(ctx, "previous-application-linked", event)
}

type ReducedFeeRequested struct {
	UID         string   `json:"uid"`
	RequestType string   `json:"requestType"`
	Evidence    []string `json:"evidence"`
}

func (c *Client) SendReducedFeeRequested(ctx context.Context, event ReducedFeeRequested) error {
	return c.send(ctx, "reduced-fee-requested", event)
}

func (c *Client) send(ctx context.Context, detailType string, detail any) error {
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
