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

const (
	issuer            = "opg.poas.makeregister"
	statusActive      = "active"
	statusReplacement = "replacement"
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
	LpaType                                     actor.LpaType                    `json:"lpaType"`
	Donor                                       lpaRequestDonor                  `json:"donor"`
	Attorneys                                   []lpaRequestAttorney             `json:"attorneys"`
	TrustCorporations                           []lpaRequestTrustCorporation     `json:"trustCorporations,omitempty"`
	CertificateProvider                         lpaRequestCertificateProvider    `json:"certificateProvider"`
	PeopleToNotify                              []lpaRequestPersonToNotify       `json:"peopleToNotify,omitempty"`
	HowAttorneysMakeDecisions                   actor.AttorneysAct               `json:"howAttorneysMakeDecisions,omitempty"`
	HowAttorneysMakeDecisionsDetails            string                           `json:"howAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysMakeDecisions        actor.AttorneysAct               `json:"howReplacementAttorneysMakeDecisions,omitempty"`
	HowReplacementAttorneysMakeDecisionsDetails string                           `json:"howReplacementAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysStepIn               actor.ReplacementAttorneysStepIn `json:"howReplacementAttorneysStepIn,omitempty"`
	HowReplacementAttorneysStepInDetails        string                           `json:"howReplacementAttorneysStepInDetails,omitempty"`
	Restrictions                                string                           `json:"restrictions"`
	WhenTheLpaCanBeUsed                         actor.CanBeUsedWhen              `json:"whenTheLpaCanBeUsed,omitempty"`
	LifeSustainingTreatmentOption               actor.LifeSustainingTreatment    `json:"lifeSustainingTreatmentOption,omitempty"`
	SignedAt                                    time.Time                        `json:"signedAt"`
}

type lpaRequestDonor struct {
	FirstNames        string        `json:"firstNames"`
	LastName          string        `json:"lastName"`
	DateOfBirth       date.Date     `json:"dateOfBirth"`
	Email             string        `json:"email"`
	Address           place.Address `json:"address"`
	OtherNamesKnownBy string        `json:"otherNamesKnownBy,omitempty"`
}

type lpaRequestAttorney struct {
	FirstNames  string        `json:"firstNames"`
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
	FirstNames string                              `json:"firstNames"`
	LastName   string                              `json:"lastName"`
	Email      string                              `json:"email,omitempty"`
	Address    place.Address                       `json:"address"`
	Channel    actor.CertificateProviderCarryOutBy `json:"channel"`
}

type lpaRequestPersonToNotify struct {
	FirstNames string        `json:"firstNames"`
	LastName   string        `json:"lastName"`
	Address    place.Address `json:"address"`
}

func (c *Client) SendLpa(ctx context.Context, donor *actor.DonorProvidedDetails) error {
	body := lpaRequest{
		LpaType: donor.Type,
		Donor: lpaRequestDonor{
			FirstNames:        donor.Donor.FirstNames,
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
			Channel:    donor.CertificateProvider.CarryOutBy,
		},
		Restrictions: donor.Restrictions,
		SignedAt:     donor.SignedAt,
	}

	switch donor.Type {
	case actor.LpaTypePropertyAndAffairs:
		body.WhenTheLpaCanBeUsed = donor.WhenCanTheLpaBeUsed
	case actor.LpaTypePersonalWelfare:
		body.LifeSustainingTreatmentOption = donor.LifeSustainingTreatmentOption
	}

	if donor.Attorneys.Len() > 1 {
		body.HowAttorneysMakeDecisions = donor.AttorneyDecisions.How
		body.HowAttorneysMakeDecisionsDetails = donor.AttorneyDecisions.Details
	}

	if donor.ReplacementAttorneys.Len() > 0 && donor.AttorneyDecisions.How.IsJointlyAndSeverally() {
		body.HowReplacementAttorneysStepIn = donor.HowShouldReplacementAttorneysStepIn
		body.HowReplacementAttorneysStepInDetails = donor.HowShouldReplacementAttorneysStepInDetails
	}

	if donor.ReplacementAttorneys.Len() > 1 && (donor.HowShouldReplacementAttorneysStepIn.IsWhenAllCanNoLongerAct() || !donor.AttorneyDecisions.How.IsJointlyAndSeverally()) {
		body.HowReplacementAttorneysMakeDecisions = donor.ReplacementAttorneyDecisions.How
		body.HowReplacementAttorneysMakeDecisionsDetails = donor.ReplacementAttorneyDecisions.Details
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		body.Attorneys = append(body.Attorneys, lpaRequestAttorney{
			FirstNames:  attorney.FirstNames,
			LastName:    attorney.LastName,
			DateOfBirth: attorney.DateOfBirth,
			Email:       attorney.Email,
			Address:     attorney.Address,
			Status:      statusActive,
		})
	}

	if trustCorporation := donor.Attorneys.TrustCorporation; trustCorporation.Name != "" {
		body.TrustCorporations = append(body.TrustCorporations, lpaRequestTrustCorporation{
			Name:          trustCorporation.Name,
			CompanyNumber: trustCorporation.CompanyNumber,
			Email:         trustCorporation.Email,
			Address:       trustCorporation.Address,
			Status:        statusActive,
		})
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		body.Attorneys = append(body.Attorneys, lpaRequestAttorney{
			FirstNames:  attorney.FirstNames,
			LastName:    attorney.LastName,
			DateOfBirth: attorney.DateOfBirth,
			Email:       attorney.Email,
			Address:     attorney.Address,
			Status:      statusReplacement,
		})
	}

	if trustCorporation := donor.ReplacementAttorneys.TrustCorporation; trustCorporation.Name != "" {
		body.TrustCorporations = append(body.TrustCorporations, lpaRequestTrustCorporation{
			Name:          trustCorporation.Name,
			CompanyNumber: trustCorporation.CompanyNumber,
			Email:         trustCorporation.Email,
			Address:       trustCorporation.Address,
			Status:        statusReplacement,
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

	return c.do(ctx, req)
}

type updateRequest struct {
	Type    string                `json:"type"`
	Changes []updateRequestChange `json:"changes"`
}

type updateRequestChange struct {
	Key string `json:"key"`
	Old any    `json:"old"`
	New any    `json:"new"`
}

func (c *Client) SendCertificateProvider(ctx context.Context, lpaUID string, certificateProvider *actor.CertificateProviderProvidedDetails) error {
	body := updateRequest{
		Type: "CERTIFICATE_PROVIDER_SIGN",
		Changes: []updateRequestChange{
			{Key: "/certificateProvider/signedAt", New: certificateProvider.Certificate.Agreed},
			{Key: "/certificateProvider/contactLanguagePreference", New: certificateProvider.ContactLanguagePreference.String()},
		},
	}

	if certificateProvider.HomeAddress.Line1 != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/line1", New: certificateProvider.HomeAddress.Line1})
	}

	if certificateProvider.HomeAddress.Line2 != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/line2", New: certificateProvider.HomeAddress.Line2})
	}

	if certificateProvider.HomeAddress.Line3 != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/line3", New: certificateProvider.HomeAddress.Line3})
	}

	if certificateProvider.HomeAddress.TownOrCity != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/town", New: certificateProvider.HomeAddress.TownOrCity})
	}

	if certificateProvider.HomeAddress.Postcode != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/postcode", New: certificateProvider.HomeAddress.Postcode})
	}

	if certificateProvider.HomeAddress.Country != "" {
		body.Changes = append(body.Changes, updateRequestChange{Key: "/certificateProvider/address/country", New: certificateProvider.HomeAddress.Country})
	}

	return c.sendUpdate(ctx, lpaUID, body)
}

func (c *Client) SendAttorney(ctx context.Context, donor *actor.DonorProvidedDetails, attorney *actor.AttorneyProvidedDetails) error {
	var attorneyKey string
	if attorney.IsTrustCorporation && attorney.IsReplacement {
		attorneyKey = "/trustCorporations/1"
	} else if attorney.IsTrustCorporation {
		attorneyKey = "/trustCorporations/0"
	} else if attorney.IsReplacement {
		attorneyKey = fmt.Sprintf("/attorneys/%d", len(donor.Attorneys.Attorneys)+donor.ReplacementAttorneys.Index(attorney.ID))
	} else {
		attorneyKey = fmt.Sprintf("/attorneys/%d", donor.Attorneys.Index(attorney.ID))
	}

	body := updateRequest{
		Type: "ATTORNEY_SIGN",
		Changes: []updateRequestChange{
			{Key: attorneyKey + "/mobile", New: attorney.Mobile},
			{Key: attorneyKey + "/contactLanguagePreference", New: attorney.ContactLanguagePreference.String()},
		},
	}

	if attorney.IsTrustCorporation {
		body.Changes = append(body.Changes,
			updateRequestChange{Key: attorneyKey + "/signatories/0/firstNames", New: attorney.AuthorisedSignatories[0].FirstNames},
			updateRequestChange{Key: attorneyKey + "/signatories/0/lastName", New: attorney.AuthorisedSignatories[0].LastName},
			updateRequestChange{Key: attorneyKey + "/signatories/0/professionalTitle", New: attorney.AuthorisedSignatories[0].ProfessionalTitle},
			updateRequestChange{Key: attorneyKey + "/signatories/0/signedAt", New: attorney.AuthorisedSignatories[0].Confirmed},
		)

		if !attorney.AuthorisedSignatories[1].Confirmed.IsZero() {
			body.Changes = append(body.Changes,
				updateRequestChange{Key: attorneyKey + "/signatories/1/firstNames", New: attorney.AuthorisedSignatories[1].FirstNames},
				updateRequestChange{Key: attorneyKey + "/signatories/1/lastName", New: attorney.AuthorisedSignatories[1].LastName},
				updateRequestChange{Key: attorneyKey + "/signatories/1/professionalTitle", New: attorney.AuthorisedSignatories[1].ProfessionalTitle},
				updateRequestChange{Key: attorneyKey + "/signatories/1/signedAt", New: attorney.AuthorisedSignatories[1].Confirmed},
			)
		}
	} else {
		body.Changes = append(body.Changes, updateRequestChange{Key: attorneyKey + "/signedAt", New: attorney.Confirmed})
	}

	return c.sendUpdate(ctx, donor.LpaUID, body)
}

func (c *Client) sendUpdate(ctx context.Context, uid string, body updateRequest) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/lpas/"+uid+"/updates", &buf)
	if err != nil {
		return err
	}

	return c.do(ctx, req)
}

func (c *Client) do(ctx context.Context, req *http.Request) error {
	req.Header.Add("Content-Type", "application/json")

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:   issuer,
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
