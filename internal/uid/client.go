package uid

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	v4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

const apiGatewayServiceName = "execute-api"

//go:generate mockery --testonly --inpackage --name Doer --structname mockDoer
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

//go:generate mockery --testonly --inpackage --name v4Signer --structname mockV4Signer
type v4Signer interface {
	SignHTTP(context.Context, aws.Credentials, *http.Request, string, string, string, time.Time, ...func(options *v4.SignerOptions)) error
}

type Client struct {
	baseURL    string
	httpClient Doer
	cfg        aws.Config
	signer     v4Signer
	now        func() time.Time
}

func New(baseURL string, httpClient Doer, cfg aws.Config, signer v4Signer, now func() time.Time) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: httpClient,
		cfg:        cfg,
		signer:     signer,
		now:        now,
	}
}

type DonorDetails struct {
	Name     string    `json:"name"`
	Dob      date.Date `json:"dob"`
	Postcode string    `json:"postcode"`
}

type CreateCaseRequestBody struct {
	Type   string       `json:"type"`
	Source string       `json:"source"`
	Donor  DonorDetails `json:"donor"`
}

type CreateCaseResponse struct {
	UID              string                              `json:"uid"`
	BadRequestErrors []CreateCaseResponseBadRequestError `json:"errors"`
}

type CreateCaseResponseBadRequestError struct {
	Source string `json:"source"`
	Detail string `json:"detail"`
}

func (c *Client) CreateCase(ctx context.Context, body *CreateCaseRequestBody) (string, error) {
	if !body.Valid() {
		return "", errors.New("CreateCaseRequestBody missing details. Requires Type, Donor name, dob and postcode")
	}

	body.Source = "APPLICANT"
	data, _ := json.Marshal(body)

	r, err := http.NewRequest(http.MethodPost, c.baseURL+"/cases", bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	r.Header.Add("Content-Type", "application/json")

	err = c.sign(ctx, r, apiGatewayServiceName)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode > http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("error POSTing to UID service: (%d) %s", resp.StatusCode, string(body))
	}

	var createCaseResponse CreateCaseResponse

	if err := json.NewDecoder(resp.Body).Decode(&createCaseResponse); err != nil {
		return "", err
	}

	if len(createCaseResponse.BadRequestErrors) > 0 {
		return "", createCaseResponse.Error()
	}

	return createCaseResponse.UID, nil
}

func (c *Client) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequest(http.MethodGet, c.baseURL+"/health", nil)
	if err != nil {
		return err
	}

	if err = c.sign(ctx, req, apiGatewayServiceName); err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("expected 200 but got %d", resp.StatusCode)
	}

	return nil
}

func (b CreateCaseRequestBody) Valid() bool {
	return b.Type != "" &&
		b.Donor.Name != "" &&
		!b.Donor.Dob.IsZero() &&
		b.Donor.Postcode != ""
}

func (c *CreateCaseResponse) Error() error {
	if len(c.BadRequestErrors) == 0 {
		return nil
	}

	detail := fmt.Sprintf("error POSTing to UID service: (400) %s %s", c.BadRequestErrors[0].Source, c.BadRequestErrors[0].Detail)

	for _, err := range c.BadRequestErrors[1:] {
		detail = fmt.Sprintf("%s, %s %s", detail, err.Source, err.Detail)
	}

	return errors.New(detail)
}

func (c *Client) sign(ctx context.Context, req *http.Request, serviceName string) error {
	hash := sha256.New()

	if req.Body != nil {
		var reqBody bytes.Buffer

		if _, err := io.Copy(hash, io.TeeReader(req.Body, &reqBody)); err != nil {
			return err
		}

		req.Body = io.NopCloser(&reqBody)
	}

	encodedBody := hex.EncodeToString(hash.Sum(nil))

	credentials, err := c.cfg.Credentials.Retrieve(ctx)
	if err != nil {
		return err
	}

	return c.signer.SignHTTP(ctx, credentials, req, encodedBody, serviceName, c.cfg.Region, c.now().UTC())
}
