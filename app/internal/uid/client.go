package uid

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

const uidServiceName = "execute-api"

//go:generate mockery --testonly --inpackage --name Doer --structname mockDoer
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

//go:generate mockery --testonly --inpackage --name RequestSigner --structname mockRequestSigner
type RequestSigner interface {
	Sign(context.Context, *http.Request, string) error
}

type Client struct {
	baseUrl    string
	httpClient Doer
	signer     RequestSigner
}

func New(baseUrl string, httpClient Doer, signer RequestSigner) *Client {
	return &Client{
		baseUrl:    baseUrl,
		httpClient: httpClient,
		signer:     signer,
	}
}

type DonorDetails struct {
	Name     string  `json:"name"`
	Dob      ISODate `json:"dob"`
	Postcode string  `json:"postcode"`
}

type CreateCaseRequestBody struct {
	Type   string       `json:"type"`
	Source string       `json:"source"`
	Donor  DonorDetails `json:"donor"`
}

type CreateCaseResponse struct {
	Uid              string                              `json:"uid"`
	BadRequestErrors []CreateCaseResponseBadRequestError `json:"errors"`
}

type CreateCaseResponseBadRequestError struct {
	Source string `json:"source"`
	Detail string `json:"detail"`
}

func (c *Client) CreateCase(ctx context.Context, body *CreateCaseRequestBody) (CreateCaseResponse, error) {
	if !body.Valid() {
		return CreateCaseResponse{}, errors.New("CreateCaseRequestBody missing details. Requires Type, Donor name, dob and postcode")
	}

	body.Source = "APPLICANT"
	data, _ := json.Marshal(body)

	r, err := http.NewRequest(http.MethodPost, c.baseUrl+"/cases", bytes.NewReader(data))
	if err != nil {
		return CreateCaseResponse{}, err
	}

	r.Header.Add("Content-Type", "application/json")

	err = c.signer.Sign(ctx, r, uidServiceName)
	if err != nil {
		return CreateCaseResponse{}, err
	}

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return CreateCaseResponse{}, err
	}

	defer resp.Body.Close()

	if resp.StatusCode > http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		return CreateCaseResponse{}, errors.New(fmt.Sprintf("error POSTing to UID service: (%d) %s", resp.StatusCode, string(body)))
	}

	var createCaseResponse CreateCaseResponse

	if err := json.NewDecoder(resp.Body).Decode(&createCaseResponse); err != nil {
		return CreateCaseResponse{}, err
	}

	if len(createCaseResponse.BadRequestErrors) > 0 {
		return CreateCaseResponse{}, createCaseResponse.Error()
	}

	return createCaseResponse, nil
}

type ISODate struct {
	time.Time
}

func (d ISODate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format("2000-01-02"))
}

func (b CreateCaseRequestBody) Valid() bool {
	return b.Type != "" &&
		b.Donor.Name != "" &&
		!b.Donor.Dob.IsZero() &&
		b.Donor.Postcode != ""
}

func (c *CreateCaseResponse) Error() error {
	if len(c.BadRequestErrors) > 0 {
		detail := c.BadRequestErrors[0].Detail

		c.BadRequestErrors = append(c.BadRequestErrors[:0], c.BadRequestErrors[0+1:]...)

		for _, err := range c.BadRequestErrors {
			detail = fmt.Sprintf("%s, %s", detail, err.Detail)
		}

		return errors.New(detail)
	} else {
		return nil
	}

}
