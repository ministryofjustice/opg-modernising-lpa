package uid

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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
	Name     string  `json:"name"`
	Dob      ISODate `json:"dob"`
	Postcode string  `json:"postcode"`
}

type CreateCaseBody struct {
	Type   string       `json:"type"`
	Source string       `json:"source"`
	Donor  DonorDetails `json:"donor"`
}

type CreateCaseResponse struct {
	Uid string
}

func (c *Client) CreateCase(lpa *page.Lpa) (CreateCaseResponse, error) {
	if !Valid(lpa) {
		return CreateCaseResponse{}, errors.New("LPA missing details. Requires Type, Donor name, dob and postcode")
	}

	data, _ := json.Marshal(CreateCaseBody{
		Source: "APPLICANT",
		Type:   lpa.Type,
		Donor: DonorDetails{
			Name:     lpa.Donor.FullName(),
			Dob:      ISODate{lpa.Donor.DateOfBirth.Time()},
			Postcode: lpa.Donor.Address.Postcode,
		},
	})

	r, err := http.NewRequest(http.MethodPost, c.baseUrl+"/cases", bytes.NewReader(data))
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

type ISODate struct {
	time.Time
}

func (d ISODate) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Time.Format("2000-01-02"))
}

func Valid(lpa *page.Lpa) bool {
	return lpa.Type != "" &&
		lpa.Donor.FullName() != " " &&
		!lpa.Donor.DateOfBirth.IsZero() &&
		lpa.Donor.Address.Postcode != ""
}
