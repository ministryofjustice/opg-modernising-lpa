package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

//go:generate mockery --testonly --inpackage --name Doer --structname mockDoer
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	baseURL      string
	doer         Doer
	issuer       string
	secretKey    []byte
	now          func() time.Time
	isProduction bool
}

func New(isProduction bool, baseURL, apiKey string, httpClient Doer) (*Client, error) {
	keyParts := strings.Split(apiKey, "-")
	if len(keyParts) != 11 {
		return nil, errors.New("invalid apiKey format")
	}

	return &Client{
		baseURL:      baseURL,
		doer:         httpClient,
		issuer:       strings.Join(keyParts[1:6], "-"),
		secretKey:    []byte(strings.Join(keyParts[6:11], "-")),
		now:          time.Now,
		isProduction: isProduction,
	}, nil
}

type Sms struct {
	PhoneNumber     string            `json:"phone_number"`
	TemplateID      string            `json:"template_id"`
	Personalisation map[string]string `json:"personalisation,omitempty"`
	Reference       string            `json:"reference,omitempty"`
}

type response struct {
	ID         string     `json:"id"`
	StatusCode int        `json:"status_code,omitempty"`
	Errors     errorsList `json:"errors,omitempty"`
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
}

func (c *Client) SendEmail(ctx context.Context, to string, email Email) (string, error) {
	req, err := c.newRequest(ctx, "/v2/notifications/email", emailWrapper{
		EmailAddress:    to,
		TemplateID:      email.emailID(c.isProduction),
		Personalisation: email,
	})
	if err != nil {
		return "", err
	}

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

type smsWrapper struct {
	PhoneNumber     string `json:"phone_number"`
	TemplateID      string `json:"template_id"`
	Personalisation any    `json:"personalisation,omitempty"`
}

func (c *Client) SendSMS(ctx context.Context, to string, sms SMS) (string, error) {
	req, err := c.newRequest(ctx, "/v2/notifications/sms", smsWrapper{
		PhoneNumber:     to,
		TemplateID:      sms.smsID(c.isProduction),
		Personalisation: sms,
	})
	if err != nil {
		return "", err
	}

	resp, err := c.do(req)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (c *Client) newRequest(ctx context.Context, url string, wrapper any) (*http.Request, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(wrapper); err != nil {
		return nil, err
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:   c.issuer,
		IssuedAt: jwt.NewNumericDate(c.now()),
	}).SignedString(c.secretKey)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+url, &buf)
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
