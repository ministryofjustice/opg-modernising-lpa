package lpastore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type abstractError struct {
	Detail string `json:"detail"`
}

type lpaRequest struct {
	LpaType                                     lpadata.LpaType                    `json:"lpaType"`
	Channel                                     lpadata.Channel                    `json:"channel"`
	Donor                                       lpaRequestDonor                    `json:"donor"`
	Attorneys                                   []lpaRequestAttorney               `json:"attorneys"`
	TrustCorporations                           []lpaRequestTrustCorporation       `json:"trustCorporations,omitempty"`
	CertificateProvider                         lpaRequestCertificateProvider      `json:"certificateProvider"`
	PeopleToNotify                              []lpaRequestPersonToNotify         `json:"peopleToNotify,omitempty"`
	HowAttorneysMakeDecisions                   lpadata.AttorneysAct               `json:"howAttorneysMakeDecisions,omitempty"`
	HowAttorneysMakeDecisionsDetails            string                             `json:"howAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysMakeDecisions        lpadata.AttorneysAct               `json:"howReplacementAttorneysMakeDecisions,omitempty"`
	HowReplacementAttorneysMakeDecisionsDetails string                             `json:"howReplacementAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysStepIn               lpadata.ReplacementAttorneysStepIn `json:"howReplacementAttorneysStepIn,omitempty"`
	HowReplacementAttorneysStepInDetails        string                             `json:"howReplacementAttorneysStepInDetails,omitempty"`
	Restrictions                                string                             `json:"restrictionsAndConditions"`
	WhenTheLpaCanBeUsed                         lpadata.CanBeUsedWhen              `json:"whenTheLpaCanBeUsed,omitempty"`
	LifeSustainingTreatmentOption               lpadata.LifeSustainingTreatment    `json:"lifeSustainingTreatmentOption,omitempty"`
	SignedAt                                    time.Time                          `json:"signedAt"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time                         `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
}

type lpaRequestIdentityCheck struct {
	CheckedAt time.Time `json:"checkedAt"`
	Type      string    `json:"type"`
}

type lpaRequestDonor struct {
	UID                       actoruid.UID             `json:"uid"`
	FirstNames                string                   `json:"firstNames"`
	LastName                  string                   `json:"lastName"`
	DateOfBirth               date.Date                `json:"dateOfBirth"`
	Email                     string                   `json:"email"`
	Address                   place.Address            `json:"address"`
	OtherNamesKnownBy         string                   `json:"otherNamesKnownBy,omitempty"`
	ContactLanguagePreference localize.Lang            `json:"contactLanguagePreference"`
	IdentityCheck             *lpaRequestIdentityCheck `json:"identityCheck,omitempty"`
}

type lpaRequestAttorney struct {
	UID         actoruid.UID    `json:"uid"`
	FirstNames  string          `json:"firstNames"`
	LastName    string          `json:"lastName"`
	DateOfBirth date.Date       `json:"dateOfBirth"`
	Email       string          `json:"email,omitempty"`
	Address     place.Address   `json:"address"`
	Status      string          `json:"status"`
	Channel     lpadata.Channel `json:"channel"`
}

type lpaRequestTrustCorporation struct {
	UID           actoruid.UID    `json:"uid"`
	Name          string          `json:"name"`
	CompanyNumber string          `json:"companyNumber"`
	Email         string          `json:"email,omitempty"`
	Address       place.Address   `json:"address"`
	Status        string          `json:"status"`
	Channel       lpadata.Channel `json:"channel"`
}

type lpaRequestCertificateProvider struct {
	UID        actoruid.UID    `json:"uid"`
	FirstNames string          `json:"firstNames"`
	LastName   string          `json:"lastName"`
	Email      string          `json:"email,omitempty"`
	Phone      string          `json:"phone,omitempty"`
	Address    place.Address   `json:"address"`
	Channel    lpadata.Channel `json:"channel"`
}

type lpaRequestPersonToNotify struct {
	UID        actoruid.UID  `json:"uid"`
	FirstNames string        `json:"firstNames"`
	LastName   string        `json:"lastName"`
	Address    place.Address `json:"address"`
}

func (c *Client) SendLpa(ctx context.Context, donor *donordata.Provided) error {
	body := lpaRequest{
		LpaType: donor.Type,
		Channel: lpadata.ChannelOnline,
		Donor: lpaRequestDonor{
			UID:                       donor.Donor.UID,
			FirstNames:                donor.Donor.FirstNames,
			LastName:                  donor.Donor.LastName,
			DateOfBirth:               donor.Donor.DateOfBirth,
			Email:                     donor.Donor.Email,
			Address:                   donor.Donor.Address,
			OtherNamesKnownBy:         donor.Donor.OtherNames,
			ContactLanguagePreference: donor.Donor.ContactLanguagePreference,
		},
		CertificateProvider: lpaRequestCertificateProvider{
			UID:        donor.CertificateProvider.UID,
			FirstNames: donor.CertificateProvider.FirstNames,
			LastName:   donor.CertificateProvider.LastName,
			Email:      donor.CertificateProvider.Email,
			Phone:      donor.CertificateProvider.Mobile,
			Address:    donor.CertificateProvider.Address,
			Channel:    donor.CertificateProvider.CarryOutBy,
		},
		Restrictions: donor.Restrictions,
		SignedAt:     donor.SignedAt,
	}

	if donor.DonorIdentityConfirmed() {
		body.Donor.IdentityCheck = &lpaRequestIdentityCheck{
			CheckedAt: donor.IdentityUserData.RetrievedAt,
			Type:      "one-login",
		}
	}

	if !donor.CertificateProviderNotRelatedConfirmedAt.IsZero() {
		body.CertificateProviderNotRelatedConfirmedAt = &donor.CertificateProviderNotRelatedConfirmedAt
	}

	switch donor.Type {
	case lpadata.LpaTypePropertyAndAffairs:
		body.WhenTheLpaCanBeUsed = donor.WhenCanTheLpaBeUsed
	case lpadata.LpaTypePersonalWelfare:
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
			UID:         attorney.UID,
			FirstNames:  attorney.FirstNames,
			LastName:    attorney.LastName,
			DateOfBirth: attorney.DateOfBirth,
			Email:       attorney.Email,
			Address:     attorney.Address,
			Status:      statusActive,
			Channel:     attorney.Channel(),
		})
	}

	if trustCorporation := donor.Attorneys.TrustCorporation; trustCorporation.Name != "" {
		body.TrustCorporations = append(body.TrustCorporations, lpaRequestTrustCorporation{
			UID:           trustCorporation.UID,
			Name:          trustCorporation.Name,
			CompanyNumber: trustCorporation.CompanyNumber,
			Email:         trustCorporation.Email,
			Address:       trustCorporation.Address,
			Status:        statusActive,
			Channel:       trustCorporation.Channel(),
		})
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		body.Attorneys = append(body.Attorneys, lpaRequestAttorney{
			UID:         attorney.UID,
			FirstNames:  attorney.FirstNames,
			LastName:    attorney.LastName,
			DateOfBirth: attorney.DateOfBirth,
			Email:       attorney.Email,
			Address:     attorney.Address,
			Status:      statusReplacement,
			Channel:     attorney.Channel(),
		})
	}

	if trustCorporation := donor.ReplacementAttorneys.TrustCorporation; trustCorporation.Name != "" {
		body.TrustCorporations = append(body.TrustCorporations, lpaRequestTrustCorporation{
			UID:           trustCorporation.UID,
			Name:          trustCorporation.Name,
			CompanyNumber: trustCorporation.CompanyNumber,
			Email:         trustCorporation.Email,
			Address:       trustCorporation.Address,
			Status:        statusReplacement,
			Channel:       trustCorporation.Channel(),
		})
	}

	for _, person := range donor.PeopleToNotify {
		body.PeopleToNotify = append(body.PeopleToNotify, lpaRequestPersonToNotify{
			UID:        person.UID,
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

	resp, err := c.do(ctx, donor.Donor.UID, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusCreated:
		return nil

	case http.StatusBadRequest:
		body, _ := io.ReadAll(resp.Body)

		var error abstractError
		_ = json.Unmarshal(body, &error)

		if error.Detail == "LPA with UID already exists" {
			// ignore the error as this call will be part of a resubmitted form
			return nil
		}

		return responseError{
			name: fmt.Sprintf("expected 201 response but got %d", resp.StatusCode),
			body: string(body),
		}

	default:
		body, _ := io.ReadAll(resp.Body)

		return responseError{
			name: fmt.Sprintf("expected 201 response but got %d", resp.StatusCode),
			body: string(body),
		}

	}
}

type lpaResponseAttorney struct {
	lpaRequestAttorney
	Mobile                    string        `json:"mobile"`
	SignedAt                  time.Time     `json:"signedAt"`
	ContactLanguagePreference localize.Lang `json:"contactLanguagePreference"`
}

type lpaResponseTrustCorporation struct {
	lpaRequestTrustCorporation
	Mobile                    string                              `json:"mobile"`
	Signatories               []lpadata.TrustCorporationSignatory `json:"signatories"`
	ContactLanguagePreference localize.Lang                       `json:"contactLanguagePreference"`
}

type lpaResponse struct {
	LpaType                                     lpadata.LpaType                    `json:"lpaType"`
	Donor                                       lpaRequestDonor                    `json:"donor"`
	Channel                                     lpadata.Channel                    `json:"channel"`
	Attorneys                                   []lpaResponseAttorney              `json:"attorneys"`
	TrustCorporations                           []lpaResponseTrustCorporation      `json:"trustCorporations,omitempty"`
	CertificateProvider                         lpadata.CertificateProvider        `json:"certificateProvider"`
	PeopleToNotify                              []lpaRequestPersonToNotify         `json:"peopleToNotify,omitempty"`
	HowAttorneysMakeDecisions                   lpadata.AttorneysAct               `json:"howAttorneysMakeDecisions,omitempty"`
	HowAttorneysMakeDecisionsDetails            string                             `json:"howAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysMakeDecisions        lpadata.AttorneysAct               `json:"howReplacementAttorneysMakeDecisions,omitempty"`
	HowReplacementAttorneysMakeDecisionsDetails string                             `json:"howReplacementAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysStepIn               lpadata.ReplacementAttorneysStepIn `json:"howReplacementAttorneysStepIn,omitempty"`
	HowReplacementAttorneysStepInDetails        string                             `json:"howReplacementAttorneysStepInDetails,omitempty"`
	Restrictions                                string                             `json:"restrictionsAndConditions"`
	WhenTheLpaCanBeUsed                         lpadata.CanBeUsedWhen              `json:"whenTheLpaCanBeUsed,omitempty"`
	LifeSustainingTreatmentOption               lpadata.LifeSustainingTreatment    `json:"lifeSustainingTreatmentOption,omitempty"`
	SignedAt                                    time.Time                          `json:"signedAt"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time                         `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
	UID                                         string                             `json:"uid"`
	Status                                      string                             `json:"status"`
	RegistrationDate                            time.Time                          `json:"registrationDate"`
	UpdatedAt                                   time.Time                          `json:"updatedAt"`
}

func lpaResponseToLpa(l lpaResponse) *lpadata.Lpa {
	var attorneys, replacementAttorneys []lpadata.Attorney
	for _, a := range l.Attorneys {
		at := lpadata.Attorney{
			UID:                       a.UID,
			FirstNames:                a.FirstNames,
			LastName:                  a.LastName,
			DateOfBirth:               a.DateOfBirth,
			Email:                     a.Email,
			Address:                   a.Address,
			Mobile:                    a.Mobile,
			SignedAt:                  a.SignedAt,
			ContactLanguagePreference: a.ContactLanguagePreference,
			Channel:                   a.Channel,
		}

		if a.Status == "replacement" {
			replacementAttorneys = append(replacementAttorneys, at)
		} else if a.Status == "active" {
			attorneys = append(attorneys, at)
		}
	}

	var trustCorporation, replacementTrustCorporation lpadata.TrustCorporation
	for _, t := range l.TrustCorporations {
		tc := lpadata.TrustCorporation{
			UID:                       t.UID,
			Name:                      t.Name,
			CompanyNumber:             t.CompanyNumber,
			Email:                     t.Email,
			Address:                   t.Address,
			Mobile:                    t.Mobile,
			Signatories:               t.Signatories,
			ContactLanguagePreference: t.ContactLanguagePreference,
			Channel:                   t.Channel,
		}

		if t.Status == "replacement" {
			replacementTrustCorporation = tc
		} else if t.Status == "active" {
			trustCorporation = tc
		}
	}

	var peopleToNotify []lpadata.PersonToNotify
	for _, p := range l.PeopleToNotify {
		peopleToNotify = append(peopleToNotify, lpadata.PersonToNotify{
			UID:        p.UID,
			FirstNames: p.FirstNames,
			LastName:   p.LastName,
			Address:    p.Address,
		})
	}

	var confirmedAt time.Time
	if v := l.CertificateProviderNotRelatedConfirmedAt; v != nil {
		confirmedAt = *v
	}

	var identityCheck lpadata.IdentityCheck
	if l.Donor.IdentityCheck != nil {
		identityCheck.CheckedAt = l.Donor.IdentityCheck.CheckedAt
		identityCheck.Type = l.Donor.IdentityCheck.Type
	}

	return &lpadata.Lpa{
		LpaUID:       l.UID,
		RegisteredAt: l.RegistrationDate,
		UpdatedAt:    l.UpdatedAt,
		Type:         l.LpaType,
		Donor: lpadata.Donor{
			UID:                       l.Donor.UID,
			FirstNames:                l.Donor.FirstNames,
			LastName:                  l.Donor.LastName,
			DateOfBirth:               l.Donor.DateOfBirth,
			Email:                     l.Donor.Email,
			Address:                   l.Donor.Address,
			OtherNames:                l.Donor.OtherNamesKnownBy,
			Channel:                   l.Channel,
			ContactLanguagePreference: l.Donor.ContactLanguagePreference,
			IdentityCheck:             identityCheck,
		},
		Attorneys: lpadata.Attorneys{
			Attorneys:        attorneys,
			TrustCorporation: trustCorporation,
		},
		ReplacementAttorneys: lpadata.Attorneys{
			Attorneys:        replacementAttorneys,
			TrustCorporation: replacementTrustCorporation,
		},
		CertificateProvider: l.CertificateProvider,
		PeopleToNotify:      peopleToNotify,
		AttorneyDecisions: lpadata.AttorneyDecisions{
			How:     l.HowAttorneysMakeDecisions,
			Details: l.HowAttorneysMakeDecisionsDetails,
		},
		ReplacementAttorneyDecisions: lpadata.AttorneyDecisions{
			How:     l.HowReplacementAttorneysMakeDecisions,
			Details: l.HowReplacementAttorneysMakeDecisionsDetails,
		},
		HowShouldReplacementAttorneysStepIn:        l.HowReplacementAttorneysStepIn,
		HowShouldReplacementAttorneysStepInDetails: l.HowReplacementAttorneysStepInDetails,
		Restrictions:                  l.Restrictions,
		WhenCanTheLpaBeUsed:           l.WhenTheLpaCanBeUsed,
		LifeSustainingTreatmentOption: l.LifeSustainingTreatmentOption,
		// TODO: add authorised signatory and independent witness when these become available
		SignedAt:                                 l.SignedAt,
		CertificateProviderNotRelatedConfirmedAt: confirmedAt,
		CannotRegister:                           l.Status == "cannot-register",
	}
}

func FromDonorProvidedDetails(l *donordata.Provided) *lpadata.Lpa {
	attorneys := lpadata.Attorneys{}
	for _, a := range l.Attorneys.Attorneys {
		attorneys.Attorneys = append(attorneys.Attorneys, lpadata.Attorney{
			UID:         a.UID,
			FirstNames:  a.FirstNames,
			LastName:    a.LastName,
			DateOfBirth: a.DateOfBirth,
			Email:       a.Email,
			Address:     a.Address,
		})
	}

	if c := l.Attorneys.TrustCorporation; c.Name != "" {
		attorneys.TrustCorporation = lpadata.TrustCorporation{
			UID:           c.UID,
			Name:          c.Name,
			CompanyNumber: c.CompanyNumber,
			Email:         c.Email,
			Address:       c.Address,
		}
	}

	var replacementAttorneys lpadata.Attorneys
	for _, a := range l.ReplacementAttorneys.Attorneys {
		replacementAttorneys.Attorneys = append(replacementAttorneys.Attorneys, lpadata.Attorney{
			UID:         a.UID,
			FirstNames:  a.FirstNames,
			LastName:    a.LastName,
			DateOfBirth: a.DateOfBirth,
			Email:       a.Email,
			Address:     a.Address,
		})
	}

	if c := l.ReplacementAttorneys.TrustCorporation; c.Name != "" {
		replacementAttorneys.TrustCorporation = lpadata.TrustCorporation{
			UID:           c.UID,
			Name:          c.Name,
			CompanyNumber: c.CompanyNumber,
			Email:         c.Email,
			Address:       c.Address,
		}
	}

	var identityCheck lpadata.IdentityCheck
	if l.DonorIdentityConfirmed() {
		identityCheck.CheckedAt = l.IdentityUserData.RetrievedAt
		identityCheck.Type = "one-login"
	}

	var peopleToNotify []lpadata.PersonToNotify
	for _, p := range l.PeopleToNotify {
		peopleToNotify = append(peopleToNotify, lpadata.PersonToNotify{
			UID:        p.UID,
			FirstNames: p.FirstNames,
			LastName:   p.LastName,
			Address:    p.Address,
		})
	}

	var voucher lpadata.Voucher
	if v := l.Voucher; v.Allowed {
		voucher = lpadata.Voucher{
			UID:        v.UID,
			FirstNames: v.FirstNames,
			LastName:   v.LastName,
			Email:      v.Email,
		}
	}

	var authorisedSignatory actor.Actor
	if v := l.AuthorisedSignatory; v.FirstNames != "" {
		authorisedSignatory = actor.Actor{
			Type:       actor.TypeAuthorisedSignatory,
			UID:        v.UID,
			FirstNames: v.FirstNames,
			LastName:   v.LastName,
		}
	}

	var independentWitness lpadata.IndependentWitness
	if v := l.IndependentWitness; v.FirstNames != "" {
		independentWitness = lpadata.IndependentWitness{
			UID:        v.UID,
			FirstNames: v.FirstNames,
			LastName:   v.LastName,
			Mobile:     v.Mobile,
			Address:    v.Address,
		}
	}

	return &lpadata.Lpa{
		LpaID:     l.LpaID,
		LpaUID:    l.LpaUID,
		UpdatedAt: l.UpdatedAt,
		Type:      l.Type,
		Donor: lpadata.Donor{
			UID:                       l.Donor.UID,
			FirstNames:                l.Donor.FirstNames,
			LastName:                  l.Donor.LastName,
			Email:                     l.Donor.Email,
			OtherNames:                l.Donor.OtherNames,
			DateOfBirth:               l.Donor.DateOfBirth,
			Address:                   l.Donor.Address,
			Channel:                   l.Donor.Channel,
			ContactLanguagePreference: l.Donor.ContactLanguagePreference,
			IdentityCheck:             identityCheck,
		},
		Attorneys:            attorneys,
		ReplacementAttorneys: replacementAttorneys,
		CertificateProvider: lpadata.CertificateProvider{
			UID:        l.CertificateProvider.UID,
			FirstNames: l.CertificateProvider.FirstNames,
			LastName:   l.CertificateProvider.LastName,
			Email:      l.CertificateProvider.Email,
			Phone:      l.CertificateProvider.Mobile,
			Address:    l.CertificateProvider.Address,
			Channel:    l.CertificateProvider.CarryOutBy,
		},
		PeopleToNotify: peopleToNotify,
		AttorneyDecisions: lpadata.AttorneyDecisions{
			How:     l.AttorneyDecisions.How,
			Details: l.AttorneyDecisions.Details,
		},
		ReplacementAttorneyDecisions: lpadata.AttorneyDecisions{
			How:     l.ReplacementAttorneyDecisions.How,
			Details: l.ReplacementAttorneyDecisions.Details,
		},
		HowShouldReplacementAttorneysStepIn:        l.HowShouldReplacementAttorneysStepIn,
		HowShouldReplacementAttorneysStepInDetails: l.HowShouldReplacementAttorneysStepInDetails,
		Restrictions:                             l.Restrictions,
		WhenCanTheLpaBeUsed:                      l.WhenCanTheLpaBeUsed,
		LifeSustainingTreatmentOption:            l.LifeSustainingTreatmentOption,
		SignedAt:                                 l.SignedAt,
		CertificateProviderNotRelatedConfirmedAt: l.CertificateProviderNotRelatedConfirmedAt,
		Correspondent: lpadata.Correspondent{
			FirstNames: l.Correspondent.FirstNames,
			LastName:   l.Correspondent.LastName,
			Email:      l.Correspondent.Email,
		},
		AuthorisedSignatory: authorisedSignatory,
		IndependentWitness:  independentWitness,
		Voucher:             voucher,
	}
}

func (c *Client) Lpa(ctx context.Context, lpaUID string) (*lpadata.Lpa, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/lpas/"+lpaUID, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(ctx, actoruid.Service, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var v lpaResponse
		if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
			return nil, err
		}

		return lpaResponseToLpa(v), nil

	case http.StatusNotFound:
		return nil, ErrNotFound
	default:
		body, _ := io.ReadAll(resp.Body)

		return nil, responseError{
			name: fmt.Sprintf("expected 200 response but got %d", resp.StatusCode),
			body: string(body),
		}
	}
}

type lpasRequest struct {
	UIDs []string `json:"uids"`
}

type lpasResponse struct {
	Lpas []lpaResponse `json:"lpas"`
}

func (c *Client) Lpas(ctx context.Context, lpaUIDs []string) ([]*lpadata.Lpa, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(lpasRequest{UIDs: lpaUIDs}); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/lpas", &buf)
	if err != nil {
		return nil, err
	}

	resp, err := c.do(ctx, actoruid.Service, req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)

		return nil, responseError{
			name: fmt.Sprintf("expected 200 response but got %d", resp.StatusCode),
			body: string(body),
		}
	}

	var v lpasResponse
	if err := json.NewDecoder(resp.Body).Decode(&v); err != nil {
		return nil, err
	}

	var lpas []*lpadata.Lpa
	for _, l := range v.Lpas {
		lpas = append(lpas, lpaResponseToLpa(l))
	}

	return lpas, nil
}
