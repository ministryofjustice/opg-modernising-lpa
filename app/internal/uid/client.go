package uid

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
)

//go:generate mockery --testonly --inpackage --name Doer --structname mockDoer
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

type Client struct {
	baseUrl    string
	httpClient Doer
}

func New(baseUrl string, httpClient Doer) *Client {
	return &Client{
		baseUrl:    baseUrl,
		httpClient: httpClient,
	}
}

type DonorDetails struct {
	Name     string    `json:"name"`
	Dob      date.Date `json:"dob"`
	Postcode string    `json:"postcode"`
}

type CreateCaseBody struct {
	Type   string       `json:"type"`
	Source string       `json:"source"`
	Donor  DonorDetails `json:"donor"`
}

type CreateCaseResponse struct {
	Uid string
}

func (c *Client) CreateCase(body CreateCaseBody) (CreateCaseResponse, error) {
	body.Source = "APPLICANT"
	data, _ := json.Marshal(body)

	r, err := http.NewRequest(http.MethodGet, c.baseUrl+"/cases", bytes.NewReader(data))
	if err != nil {
		return CreateCaseResponse{}, err
	}

	r.Header.Add("Content-Type", "application/json")

	resp, err := c.httpClient.Do(r)
	if err != nil {
		return CreateCaseResponse{}, err
	}

	defer resp.Body.Close()

	var createCaseResponse CreateCaseResponse

	if err := json.NewDecoder(resp.Body).Decode(&createCaseResponse); err != nil {
		return CreateCaseResponse{}, err
	}

	return createCaseResponse, nil
}
