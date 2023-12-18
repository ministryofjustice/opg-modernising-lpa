package lpastore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/secrets"
)

type responseError struct {
	name string
	body any
}

func (e responseError) Error() string { return e.name }
func (e responseError) Title() string { return e.name }
func (e responseError) Data() any     { return e.body }

//go:generate mockery --testonly --inpackage --name Doer --structname mockDoer
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

//go:generate mockery --testonly --inpackage --name SecretsClient --structname mockSecretsClient
type SecretsClient interface {
	Secret(ctx context.Context, name string) (string, error)
}

type Client struct {
	baseURL       string
	secretsClient SecretsClient
	doer          Doer
	now           func() time.Time
}

func New(baseURL string, secretsClient SecretsClient, lambdaClient Doer) *Client {
	return &Client{baseURL: baseURL, secretsClient: secretsClient, doer: lambdaClient, now: time.Now}
}

type lpaRequest struct {
	LpaType                                     string                        `json:"lpaType"`
	Donor                                       lpaRequestDonor               `json:"donor"`
	Attorneys                                   []lpaRequestAttorney          `json:"attorneys"`
	TrustCorporations                           []lpaRequestTrustCorporation  `json:"trustCorporations,omitempty"`
	CertificateProvider                         lpaRequestCertificateProvider `json:"certificateProvider"`
	PeopleToNotify                              []lpaRequestPersonToNotify    `json:"peopleToNotify"`
	HowAttorneysMakeDecisions                   string                        `json:"howAttorneysMakeDecisions"`
	HowAttorneysMakeDecisionsDetails            string                        `json:"howAttorneysMakeDecisionsDetails"`
	HowReplacementAttorneysMakeDecisions        string                        `json:"howReplacementAttorneysMakeDecisions"`
	HowReplacementAttorneysMakeDecisionsDetails string                        `json:"howReplacementAttorneysMakeDecisionsDetails"`
	HowReplacementAttorneysStepIn               string                        `json:"howReplacementAttorneysStepIn"`
	HowReplacementAttorneysStepInDetails        string                        `json:"howReplacementAttorneysStepInDetails"`
	Restrictions                                string                        `json:"restrictions"`
	WhenTheLpaCanBeUsed                         string                        `json:"whenTheLpaCanBeUsed,omitempty"`
	LifeSustainingTreatmentOption               string                        `json:"lifeSustainingTreatmentOption,omitempty"`
	SignedAt                                    time.Time                     `json:"signedAt"`
}

type lpaRequestDonor struct {
	FirstNames        string        `json:"firstNames"`
	Surname           string        `json:"surname"` // TODO: remove when API is changed to take lastName
	LastName          string        `json:"lastName"`
	DateOfBirth       date.Date     `json:"dateOfBirth"`
	Email             string        `json:"email"`
	Address           place.Address `json:"address"`
	OtherNamesKnownBy string        `json:"otherNamesKnownBy,omitempty"`
}

type lpaRequestAttorney struct {
	FirstNames  string        `json:"firstNames"`
	Surname     string        `json:"surname"`
	LastName    string        `json:"lastName"`
	DateOfBirth date.Date     `json:"dateOfBirth"`
	Email       string        `json:"email"`
	Address     place.Address `json:"address"`
	Status      string        `json:"status"`
}

type lpaRequestTrustCorporation struct {
	Name          string        `json:"name"`
	CompanyNumber string        `json:"companyNumber"`
	Email         string        `json:"email"`
	Address       place.Address `json:"address"`
	Status        string        `json:"status"`
}

type lpaRequestCertificateProvider struct {
	FirstNames string        `json:"firstNames"`
	LastName   string        `json:"lastName"`
	Email      string        `json:"email"`
	Address    place.Address `json:"address"`
	Channel    string        `json:"channel"`
}

type lpaRequestPersonToNotify struct {
	FirstNames string        `json:"firstNames"`
	LastName   string        `json:"lastName"`
	Address    place.Address `json:"address"`
}

func (c *Client) SendLpa(ctx context.Context, donor *actor.DonorProvidedDetails) error {
	var lpaType string
	switch donor.Type {
	case actor.LpaTypeHealthWelfare:
		lpaType = "personal-welfare"
	case actor.LpaTypePropertyFinance:
		lpaType = "property-and-affairs"
	}

	howAttorneysMakeDecisions := donor.AttorneyDecisions.How.String()
	if howAttorneysMakeDecisions == "mixed" {
		howAttorneysMakeDecisions = "jointly-for-some-severally-for-others"
	}

	howReplacementAttorneysMakeDecisions := donor.ReplacementAttorneyDecisions.How.String()
	if howReplacementAttorneysMakeDecisions == "mixed" {
		howReplacementAttorneysMakeDecisions = "jointly-for-some-severally-for-others"
	}

	var howReplacementAttorneysStepIn string
	switch donor.HowShouldReplacementAttorneysStepIn {
	case actor.ReplacementAttorneysStepInWhenAllCanNoLongerAct:
		howReplacementAttorneysStepIn = "all-can-no-longer-act"
	case actor.ReplacementAttorneysStepInWhenOneCanNoLongerAct:
		howReplacementAttorneysStepIn = "one-can-no-longer-act"
	case actor.ReplacementAttorneysStepInAnotherWay:
		howReplacementAttorneysStepIn = "another-way"
	}

	body := lpaRequest{
		LpaType: lpaType,
		Donor: lpaRequestDonor{
			FirstNames:        donor.Donor.FirstNames,
			Surname:           donor.Donor.LastName,
			LastName:          donor.Donor.LastName,
			DateOfBirth:       donor.Donor.DateOfBirth,
			Email:             donor.Donor.Email,
			Address:           donor.Donor.Address,
			OtherNamesKnownBy: donor.Donor.OtherNames,
		},
		CertificateProvider: lpaRequestCertificateProvider{
			FirstNames: donor.CertificateProvider.FirstNames,
			LastName:   donor.CertificateProvider.LastName,
			Email:      donor.CertificateProvider.Email,
			Address:    donor.CertificateProvider.Address,
			Channel:    donor.CertificateProvider.CarryOutBy.String(),
		},
		HowAttorneysMakeDecisions:                   howAttorneysMakeDecisions,
		HowAttorneysMakeDecisionsDetails:            donor.AttorneyDecisions.Details,
		HowReplacementAttorneysMakeDecisions:        howReplacementAttorneysMakeDecisions,
		HowReplacementAttorneysMakeDecisionsDetails: donor.ReplacementAttorneyDecisions.Details,
		HowReplacementAttorneysStepIn:               howReplacementAttorneysStepIn,
		HowReplacementAttorneysStepInDetails:        donor.HowShouldReplacementAttorneysStepInDetails,
		Restrictions:                                donor.Restrictions,
		SignedAt:                                    donor.SignedAt,
	}

	switch donor.Type {
	case actor.LpaTypePropertyFinance:
		body.WhenTheLpaCanBeUsed = donor.WhenCanTheLpaBeUsed.String()
	case actor.LpaTypeHealthWelfare:
		body.LifeSustainingTreatmentOption = donor.LifeSustainingTreatmentOption.String()
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		body.Attorneys = append(body.Attorneys, lpaRequestAttorney{
			FirstNames:  attorney.FirstNames,
			Surname:     attorney.LastName,
			LastName:    attorney.LastName,
			DateOfBirth: attorney.DateOfBirth,
			Email:       attorney.Email,
			Address:     attorney.Address,
			Status:      "active",
		})
	}

	if trustCorporation := donor.Attorneys.TrustCorporation; trustCorporation.Name != "" {
		body.TrustCorporations = append(body.TrustCorporations, lpaRequestTrustCorporation{
			Name:          trustCorporation.Name,
			CompanyNumber: trustCorporation.CompanyNumber,
			Email:         trustCorporation.Email,
			Address:       trustCorporation.Address,
			Status:        "active",
		})
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		body.Attorneys = append(body.Attorneys, lpaRequestAttorney{
			FirstNames:  attorney.FirstNames,
			Surname:     attorney.LastName,
			LastName:    attorney.LastName,
			DateOfBirth: attorney.DateOfBirth,
			Email:       attorney.Email,
			Address:     attorney.Address,
			Status:      "replacement",
		})
	}

	if trustCorporation := donor.ReplacementAttorneys.TrustCorporation; trustCorporation.Name != "" {
		body.TrustCorporations = append(body.TrustCorporations, lpaRequestTrustCorporation{
			Name:          trustCorporation.Name,
			CompanyNumber: trustCorporation.CompanyNumber,
			Email:         trustCorporation.Email,
			Address:       trustCorporation.Address,
			Status:        "replacement",
		})
	}

	for _, person := range donor.PeopleToNotify {
		body.PeopleToNotify = append(body.PeopleToNotify, lpaRequestPersonToNotify{
			FirstNames: person.FirstNames,
			LastName:   person.LastName,
			Address:    person.Address,
		})
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/lpas/"+donor.LpaUID, &buf)
	if err != nil {
		return err
	}
	req.Header.Add("Content-Type", "application/json")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:   "opg.poas.makeregister",
		IssuedAt: jwt.NewNumericDate(c.now()),
		Subject:  "todo",
	})

	secretKey, err := c.secretsClient.Secret(ctx, secrets.LpaStoreJwtSecretKey)
	if err != nil {
		return err
	}

	auth, err := token.SignedString([]byte(secretKey))
	if err != nil {
		return err
	}
	req.Header.Add("X-Jwt-Authorization", "Bearer "+auth)

	resp, err := c.doer.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)

		return responseError{
			name: fmt.Sprintf("expected 201 response but got %d", resp.StatusCode),
			body: string(body),
		}
	}

	return nil
}

func (c *Client) CheckHealth(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/health-check", nil)
	if err != nil {
		return err
	}

	resp, err := c.doer.Do(req)
	if err != nil {
		return err
	}
	if err := resp.Body.Close(); err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return responseError{name: fmt.Sprintf("expected 200 response but got %d", resp.StatusCode)}
	}

	return nil
}
