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
	simulatedEmails = []string{
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

type Bundle interface {
	For(lang localize.Lang) localize.Localizer
}

type Client struct {
	logger      Logger
	baseURL     string
	doer        Doer
	issuer      string
	secretKey   []byte
	now         func() time.Time
	eventClient EventClient
	bundle      Bundle
}

func New(logger Logger, baseURL, apiKey string, httpClient Doer, eventClient EventClient, bundle Bundle) (*Client, error) {
	keyParts := strings.Split(apiKey, "-")
	if len(keyParts) != 11 {
		return nil, errors.New("invalid apiKey format")
	}

	return &Client{
		logger:      logger,
		baseURL:     baseURL,
		doer:        httpClient,
		issuer:      strings.Join(keyParts[1:6], "-"),
		secretKey:   []byte(strings.Join(keyParts[6:11], "-")),
		now:         time.Now,
		eventClient: eventClient,
		bundle:      bundle,
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
		"LpaUID":                lpa.LpaUID,
		"CorrespondentFullName": lpa.Correspondent.FullName(),
		"DonorFullName":         lpa.Donor.FullName(),
		"LpaType":               localizer.T(lpa.Type.String()),
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

func (c *Client) SendEmail(ctx context.Context, to ToEmail, email Email) error {
	if to.ignore() {
		return nil
	}

	address, lang := to.toEmail()
	templateID := email.emailID(lang)

	ctx, span := newSpan(ctx, "Email", templateID, address)
	defer span.End()

	req, err := c.newRequest(ctx, http.MethodPost, "/v2/notifications/email", emailWrapper{
		EmailAddress:    address,
		TemplateID:      templateID,
		Personalisation: email,
	})
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "email send failed", slog.String("to", address))
		return err
	}
	span.SetAttributes(attribute.KeyValue{Key: "notify_id", Value: attribute.StringValue(resp.ID)})

	return nil
}

func (c *Client) SendActorEmail(ctx context.Context, to ToEmail, lpaUID string, email Email) error {
	if to.ignore() {
		return nil
	}

	address, lang := to.toEmail()

	templateID := email.emailID(lang)

	ctx, span := newSpan(ctx, "Email", templateID, address)
	defer span.End()

	if ok, err := c.recentlySent(ctx, c.makeReference(lpaUID, address, templateID)); err != nil || ok {
		return err
	}

	req, err := c.newRequest(ctx, http.MethodPost, "/v2/notifications/email", emailWrapper{
		EmailAddress:    address,
		TemplateID:      templateID,
		Personalisation: email,
		Reference:       c.makeReference(lpaUID, address, templateID),
	})
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		c.logger.ErrorContext(ctx, "email send failed", slog.String("to", address))
		return err
	}
	span.SetAttributes(attribute.KeyValue{Key: "notify_id", Value: attribute.StringValue(resp.ID)})

	if !slices.Contains(simulatedEmails, address) {
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

func (c *Client) SendActorSMS(ctx context.Context, to ToMobile, lpaUID string, sms SMS) error {
	if to.ignore() {
		return nil
	}

	number, lang := to.toMobile()

	templateID := sms.smsID(lang)

	ctx, span := newSpan(ctx, "SMS", templateID, number)
	defer span.End()

	req, err := c.newRequest(ctx, http.MethodPost, "/v2/notifications/sms", smsWrapper{
		PhoneNumber:     number,
		TemplateID:      templateID,
		Personalisation: sms,
	})
	if err != nil {
		return err
	}

	resp, err := c.do(req)
	if err != nil {
		return err
	}
	span.SetAttributes(attribute.KeyValue{Key: "notification_id", Value: attribute.StringValue(resp.ID)})

	if !slices.Contains(simulatedPhones, number) {
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
		if notification.Status == "sending" || notification.Status == "delivered" {
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

func (c *Client) makeReference(lpaUID, to, templateID string) string {
	hash := sha256.New()
	hash.Write([]byte(lpaUID))
	hash.Write([]byte{'|'})
	hash.Write([]byte(to))
	hash.Write([]byte{'|'})
	hash.Write([]byte(templateID))

	return base64.RawStdEncoding.EncodeToString(hash.Sum(nil))
}

func newSpan(ctx context.Context, label, templateID, to string) (context.Context, trace.Span) {
	tracer := otel.GetTracerProvider().Tracer("mlpab")
	ctx, span := tracer.Start(ctx, label,
		trace.WithSpanKind(trace.SpanKindInternal))
	span.SetAttributes(
		attribute.KeyValue{Key: "template_id", Value: attribute.StringValue(templateID)},
		attribute.KeyValue{Key: "to", Value: attribute.StringValue(to)})

	return ctx, span
}
