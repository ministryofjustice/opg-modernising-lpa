package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type Template uint8

const (
	CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS Template = iota
	CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS
	CertificateProviderCertificateProvidedEmail
	CertificateProviderInviteEmail
	CertificateProviderActingOnPaperDetailsChangedSMS
	CertificateProviderActingOnPaperMeetingPromptSMS
	CertificateProviderProvideCertificatePromptEmail
	WitnessCodeSMS
)

var (
	productionTemplates = map[Template]string{
		CertificateProviderCertificateProvidedEmail:                                                "2915acb9-2a52-4b62-812a-b31b19c6a94b",
		CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS: "19948d7d-a2df-4e85-930b-5d800978f41f",
		CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS:    "71d21daa-11f9-4a2a-9ae2-bb5c2247bfb7",
		CertificateProviderInviteEmail:                                                             "13df4493-20b0-4c20-b742-cab3844e69b2",
		CertificateProviderActingOnPaperDetailsChangedSMS:                                          "ab90c6be-806e-411a-a354-de10f7a70c47",
		CertificateProviderActingOnPaperMeetingPromptSMS:                                           "b5cd2c1b-e9b4-4f3e-8cf1-504aff93b16d",
		CertificateProviderProvideCertificatePromptEmail:                                           "eada8a4f-5e7c-4f6b-b3fb-d4e92eeeb0ed",
		WitnessCodeSMS: "e39849c0-ecab-4e16-87ec-6b22afb9d535",
	}
	testingTemplates = map[Template]string{
		CertificateProviderCertificateProvidedEmail:                                                "c916f964-bf30-4dee-a9f3-b9bf0043e64d",
		CertificateProviderActingDigitallyHasNotConfirmedPersonalDetailsLPADetailsChangedPromptSMS: "d7513751-49ba-4276-aef5-ad67361d29c4",
		CertificateProviderActingDigitallyHasConfirmedPersonalDetailsLPADetailsChangedPromptSMS:    "359fffa0-e1ec-444c-a886-6f046af374ab",
		CertificateProviderInviteEmail:                                                             "4ab51290-5ac6-44ea-88f4-a27c37f285f8",
		CertificateProviderActingOnPaperDetailsChangedSMS:                                          "94477364-281a-4032-9a88-b215f969cd12",
		CertificateProviderActingOnPaperMeetingPromptSMS:                                           "ee39cd81-5802-44bb-b967-27da7e25e897",
		CertificateProviderProvideCertificatePromptEmail:                                           "a445edda-dea1-4554-be9b-ad11adad9e89",
		WitnessCodeSMS: "dfa15e16-1f23-494a-bffb-a475513df6cc",
	}
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
	templates    map[Template]string
	isProduction bool
}

func New(isProduction bool, baseURL, apiKey string, httpClient Doer) (*Client, error) {
	keyParts := strings.Split(apiKey, "-")
	if len(keyParts) != 11 {
		return nil, errors.New("invalid apiKey format")
	}

	templates := testingTemplates
	if isProduction {
		templates = productionTemplates
	}

	return &Client{
		baseURL:      baseURL,
		doer:         httpClient,
		issuer:       strings.Join(keyParts[1:6], "-"),
		secretKey:    []byte(strings.Join(keyParts[6:11], "-")),
		now:          time.Now,
		templates:    templates,
		isProduction: isProduction,
	}, nil
}

type Email struct {
	EmailAddress    string            `json:"email_address"`
	TemplateID      string            `json:"template_id"`
	Personalisation map[string]string `json:"personalisation,omitempty"`
	Reference       string            `json:"reference,omitempty"`
	EmailReplyToID  string            `json:"email_reply_to_id,omitempty"`
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

func (c *Client) TemplateID(id Template) string {
	return c.templates[id]
}

func (c *Client) Email(ctx context.Context, email Email) (string, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(email); err != nil {
		return "", err
	}

	req, err := c.request(ctx, "/v2/notifications/email", &buf)
	if err != nil {
		return "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

type SendableEmail interface {
	emailID(bool) string
}

type emailWrapper struct {
	EmailAddress    string `json:"email_address"`
	TemplateID      string `json:"template_id"`
	Personalisation any    `json:"personalisation,omitempty"`
}

func (c *Client) SendEmail(ctx context.Context, to string, email SendableEmail) (string, error) {
	wrapper := emailWrapper{
		EmailAddress:    to,
		TemplateID:      email.emailID(c.isProduction),
		Personalisation: email,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(wrapper); err != nil {
		return "", err
	}

	req, err := c.request(ctx, "/v2/notifications/email", &buf)
	if err != nil {
		return "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (c *Client) Sms(ctx context.Context, sms Sms) (string, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(sms); err != nil {
		return "", err
	}

	req, err := c.request(ctx, "/v2/notifications/sms", &buf)
	if err != nil {
		return "", err
	}

	resp, err := c.doRequest(req)
	if err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (c *Client) request(ctx context.Context, url string, body io.Reader) (*http.Request, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, &jwt.RegisteredClaims{
		Issuer:   c.issuer,
		IssuedAt: jwt.NewNumericDate(c.now()),
	}).SignedString(c.secretKey)
	if err != nil {
		return &http.Request{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+url, body)
	if err != nil {
		return &http.Request{}, err
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", "Bearer "+token)

	return req, nil
}

func (c *Client) doRequest(req *http.Request) (response, error) {
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
