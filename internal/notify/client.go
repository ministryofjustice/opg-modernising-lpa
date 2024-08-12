// Package notify provides a client for GOV.UK's Notify service.
package notify

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var (
	allowResendAfter = 10 * time.Minute
	simulatedEmails  = []string{
		"simulate-delivered@notifications.service.gov.uk",
		"simulate-delivered-2@notifications.service.gov.uk",
		"simulate-delivered-3@notifications.service.gov.uk",
	}
	simulatedPhones = []string{
		"07700900000",
		"07700900111",
		"07700900222",
	}
)

type Logger interface {
	ErrorContext(ctx context.Context, msg string, args ...any)
}

type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type EventClient interface {
	SendNotificationSent(ctx context.Context, event event.NotificationSent) error
}

type Client struct {
	logger       Logger
	baseURL      string
	doer         Doer
	issuer       string
	secretKey    []byte
	now          func() time.Time
	isProduction bool
	eventClient  EventClient
	bundle       *localize.Bundle
}

func New(logger Logger, isProduction bool, baseURL, apiKey string, httpClient Doer, eventClient EventClient, bundle *localize.Bundle) (*Client, error) {
	keyParts := strings.Split(apiKey, "-")
	if len(keyParts) != 11 {
		return nil, errors.New("invalid apiKey format")
	}

	return &Client{
		logger:       logger,
		baseURL:      baseURL,
		doer:         httpClient,
		issuer:       strings.Join(keyParts[1:6], "-"),
		secretKey:    []byte(strings.Join(keyParts[6:11], "-")),
		now:          time.Now,
		isProduction: isProduction,
		eventClient:  eventClient,
		bundle:       bundle,
	}, nil
}

func (c *Client) EmailGreeting(lpa *lpadata.Lpa) string {
	localizer := c.bundle.For(lpa.Donor.ContactLanguagePreference)

	if lpa.Correspondent.FirstNames == "" {
		return localizer.Format("emailGreetingDonor", map[string]any{
			"DonorFullName": lpa.Donor.FullName(),
		})
	}

	return localizer.Format("emailGreetingCorrespondent", map[string]any{
		"LpaUID":                  lpa.LpaUID,
		"CorrespondentFullName":   lpa.Correspondent.FullName(),
		"DonorFullNamePossessive": localizer.Possessive(lpa.Donor.FullName()),
		"LpaType":                 localizer.T(lpa.Type.String()),
	})
}

type Sms struct {
	PhoneNumber     string            `json:"phone_number"`
	TemplateID      string            `json:"template_id"`
	Personalisation map[string]string `json:"personalisation,omitempty"`
	Reference       string            `json:"reference,omitempty"`
}

type response struct {
	ID            string                 `json:"id"`
	StatusCode    int                    `json:"status_code,omitempty"`
	Errors        errorsList             `json:"errors,omitempty"`
	Notifications []responseNotification `json:"notifications"`
}

type responseNotification struct {
	CreatedAt time.Time `json:"created_at"`
	Status    string    `json:"status"`
}

type errorsList []errorItem

func (es errorsList) Error() string {
	s := "error sending message"
	for _, e := range es {
		s += ": " + e.Message
	}
	return s
}

type errorItem struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type emailWrapper struct {
	EmailAddress    string `json:"email_address"`
	TemplateID      string `json:"template_id"`
	Personalisation any    `json:"personalisation,omitempty"`
	Reference       string `json:"reference"`
}

func (c *Client) SendEmail(ctx context.Context, to string, email Email) error {
	tracer := otel.GetTracerProvider().Tracer("mlpab")
	ctx, span := tracer.Start(ctx, "Email",
		trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.KeyValue{Key: "Template ID", Value: attribute.StringValue(email.emailID(c.isProduction))})
	defer span.End()

	req, err := c.newRequest(ctx, http.MethodPost, "/v2/notifications/email", emailWrapper{
		EmailAddress:    to,
		TemplateID:      email.emailID(c.isProduction),
		Personalisation: email,
	})
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "email send failed", slog.String("to", to))
		return err
	}
	span.SetAttributes(attribute.KeyValue{Key: "Notify ID", Value: attribute.StringValue(resp.ID)})

	return nil
}

func (c *Client) SendActorEmail(ctx context.Context, to, lpaUID string, email Email) error {
	tracer := otel.GetTracerProvider().Tracer("mlpab")
	ctx, span := tracer.Start(ctx, "Email",
		trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.KeyValue{Key: "Template ID", Value: attribute.StringValue(email.emailID(c.isProduction))})
	defer span.End()

	if ok, err := c.recentlySent(ctx, c.makeReference(lpaUID, to, email)); err != nil || ok {
		return err
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/v2/notifications/email", emailWrapper{
		EmailAddress:    to,
		TemplateID:      email.emailID(c.isProduction),
		Personalisation: email,
		Reference:       c.makeReference(lpaUID, to, email),
	})
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "email send failed", slog.String("to", to))
		return err
	}
	span.SetAttributes(attribute.KeyValue{Key: "Notify ID", Value: attribute.StringValue(resp.ID)})

	if !slices.Contains(simulatedEmails, to) {
		if err := c.eventClient.SendNotificationSent(ctx, event.NotificationSent{
			UID:            lpaUID,
			NotificationID: resp.ID,
		}); err != nil {
			return err
		}
	}

	return nil
}

type smsWrapper struct {
	PhoneNumber     string `json:"phone_number"`
	TemplateID      string `json:"template_id"`
	Personalisation any    `json:"personalisation,omitempty"`
}

func (c *Client) SendActorSMS(ctx context.Context, to, lpaUID string, sms SMS) error {
	tracer := otel.GetTracerProvider().Tracer("mlpab")
	ctx, span := tracer.Start(ctx, "SMS",
		trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(attribute.KeyValue{Key: "Template ID", Value: attribute.StringValue(sms.smsID(c.isProduction))})
	defer span.End()

	req, err := c.newRequest(ctx, http.MethodPost, "/v2/notifications/sms", smsWrapper{
		PhoneNumber:     to,
		TemplateID:      sms.smsID(c.isProduction),
		Personalisation: sms,
	})
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	span.SetAttributes(attribute.KeyValue{Key: "Notify ID", Value: attribute.StringValue(resp.ID)})

	if !slices.Contains(simulatedPhones, to) {
		if err := c.eventClient.SendNotificationSent(ctx, event.NotificationSent{
			UID:            lpaUID,
			NotificationID: resp.ID,
		}); err != nil {
			return err
		}
	}

	return nil
}

func (c *Client) recentlySent(ctx context.Context, ref string) (bool, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/v2/notifications?reference="+ref, nil)
	if err != nil {
		return false, err
	}

	resp, err := c.do(req)
	if err != nil {
		return false, err
	}

	for _, notification := range resp.Notifications {
		if (notification.Status == "sending" || notification.Status == "delivered") &&
			notification.CreatedAt.After(c.now().Add(-allowResendAfter)) {
			return true, nil
		}
	}

	return false, nil
}

func (c *Client) newRequest(ctx context.Context, method, url string, wrapper any) (*http.Request, error) {
	var buf bytes.Buffer
	if wrapper != nil {
		if err := json.NewEncoder(&buf).Encode(wrapper); err != nil {
			return nil, err
		}
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:   c.issuer,
		IssuedAt: jwt.NewNumericDate(c.now()),
	}).SignedString(c.secretKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+url, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	return req, nil
}

func (c *Client) do(req *http.Request) (response, error) {
	var r response

	resp, err := c.doer.Do(req)
	if err != nil {
		return r, err
	}
	defer resp.Body.Close()

	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return r, err
	}

	if len(r.Errors) > 0 {
		return r, r.Errors
	}

	return r, nil
}

func (c *Client) makeReference(lpaUID, to string, email Email) string {
	hash := sha256.New()
	hash.Write([]byte(lpaUID))
	hash.Write([]byte{'|'})
	hash.Write([]byte(to))
	hash.Write([]byte{'|'})
	hash.Write([]byte(email.emailID(c.isProduction)))

	return base64.RawStdEncoding.EncodeToString(hash.Sum(nil))
}
