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
	AttorneyInviteEmail Template = iota
	AttorneyNameChangeEmail
	CertificateProviderDigitalLpaDetailsChangedNotSeenLpaSMS
	CertificateProviderDigitalLpaDetailsChangedSeenLpaSMS
	CertificateProviderInviteEmail
	CertificateProviderNameChangeEmail
	CertificateProviderPaperLpaDetailsChangedSMS
	CertificateProviderPaperMeetingPromptSMS
	CertificateProviderReturnEmail
	ReplacementAttorneyInviteEmail
	ReplacementTrustCorporationInviteEmail
	SignatureCodeEmail
	SignatureCodeSMS
	TrustCorporationInviteEmail
)

var (
	productionTemplates = map[Template]string{
		AttorneyInviteEmail:     "9aaedb70-df4a-42a8-9c28-de435cb3d453",
		AttorneyNameChangeEmail: "1e0950c5-63fa-487e-8bf3-f40445412a12",
		CertificateProviderDigitalLpaDetailsChangedNotSeenLpaSMS: "19948d7d-a2df-4e85-930b-5d800978f41f",
		CertificateProviderDigitalLpaDetailsChangedSeenLpaSMS:    "71d21daa-11f9-4a2a-9ae2-bb5c2247bfb7",
		CertificateProviderInviteEmail:                           "a10341e3-3bbd-4452-b52f-ebb4f51a4d73",
		CertificateProviderNameChangeEmail:                       "9f8be86f-864a-4cda-a58a-5768522bd325",
		CertificateProviderPaperLpaDetailsChangedSMS:             "d363a56f-e802-4f88-bd09-80b8c9e9d650",
		CertificateProviderPaperMeetingPromptSMS:                 "6be11b4a-79f9-441e-8afe-adff96f7e7fc",
		CertificateProviderReturnEmail:                           "453917cd-d8bb-44af-90a1-d73ae0f3fd07",
		ReplacementAttorneyInviteEmail:                           "1c4d5b24-fc7d-45ee-be40-f1ccda96f101",
		ReplacementTrustCorporationInviteEmail:                   "1c4d5b24-fc7d-45ee-be40-f1ccda96f101",
		SignatureCodeEmail:                                       "95f7b0a2-1c3a-4ad9-818b-b358c549c88b",
		SignatureCodeSMS:                                         "e39849c0-ecab-4e16-87ec-6b22afb9d535",
		TrustCorporationInviteEmail:                              "9aaedb70-df4a-42a8-9c28-de435cb3d453",
	}
	testingTemplates = map[Template]string{
		AttorneyInviteEmail:     "9be88a99-21c0-4808-8d6a-52af366e44aa",
		AttorneyNameChangeEmail: "685bbdcc-71b8-48b9-b773-03941472d3b1",
		CertificateProviderDigitalLpaDetailsChangedNotSeenLpaSMS: "d7513751-49ba-4276-aef5-ad67361d29c4",
		CertificateProviderDigitalLpaDetailsChangedSeenLpaSMS:    "359fffa0-e1ec-444c-a886-6f046af374ab",
		CertificateProviderInviteEmail:                           "dd864a1a-64b4-4b4e-b810-86267ebd6476",
		CertificateProviderNameChangeEmail:                       "0f111ed1-5c58-47eb-a13f-931f2077523b",
		CertificateProviderPaperLpaDetailsChangedSMS:             "94477364-281a-4032-9a88-b215f969cd12",
		CertificateProviderPaperMeetingPromptSMS:                 "0eba4e55-c07e-4427-b4ad-b03e08dad8ca",
		CertificateProviderReturnEmail:                           "dd864a1a-64b4-4b4e-b810-86267ebd6476",
		ReplacementAttorneyInviteEmail:                           "bf79859b-72b7-4701-bfd3-22ac6f0908c8",
		ReplacementTrustCorporationInviteEmail:                   "bf79859b-72b7-4701-bfd3-22ac6f0908c8",
		SignatureCodeEmail:                                       "7e8564a0-2635-4f61-9155-0166ddbe5607",
		SignatureCodeSMS:                                         "dfa15e16-1f23-494a-bffb-a475513df6cc",
		TrustCorporationInviteEmail:                              "9be88a99-21c0-4808-8d6a-52af366e44aa",
	}
)

//go:generate mockery --testonly --inpackage --name Doer --structname mockDoer
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	baseURL   string
	doer      Doer
	issuer    string
	secretKey []byte
	now       func() time.Time
	templates map[Template]string
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
		baseURL:   baseURL,
		doer:      httpClient,
		issuer:    strings.Join(keyParts[1:6], "-"),
		secretKey: []byte(strings.Join(keyParts[6:11], "-")),
		now:       time.Now,
		templates: templates,
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
