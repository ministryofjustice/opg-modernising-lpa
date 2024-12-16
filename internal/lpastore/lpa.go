package lpastore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
)

type abstractError struct {
	Detail string `json:"detail"`
}

type CreateLpa struct {
	LpaType                                     lpadata.LpaType                    `json:"lpaType"`
	Channel                                     lpadata.Channel                    `json:"channel"`
	Donor                                       lpadata.Donor                      `json:"donor"`
	Attorneys                                   []lpadata.Attorney                 `json:"attorneys"`
	TrustCorporations                           []lpadata.TrustCorporation         `json:"trustCorporations,omitempty"`
	CertificateProvider                         lpadata.CertificateProvider        `json:"certificateProvider"`
	PeopleToNotify                              []lpadata.PersonToNotify           `json:"peopleToNotify,omitempty"`
	IndependentWitness                          *lpadata.IndependentWitness        `json:"independentWitness,omitempty"`
	AuthorisedSignatory                         *lpadata.AuthorisedSignatory       `json:"authorisedSignatory,omitempty"`
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
	WitnessedByCertificateProviderAt            time.Time                          `json:"witnessedByCertificateProviderAt"`
	WitnessedByIndependentWitnessAt             *time.Time                         `json:"witnessedByIndependentWitnessAt,omitempty"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time                         `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
}

func CreateLpaFromDonorProvided(donor *donordata.Provided) CreateLpa {
	body := CreateLpa{
		LpaType: donor.Type,
		Channel: lpadata.ChannelOnline,
		Donor: lpadata.Donor{
			UID:                       donor.Donor.UID,
			FirstNames:                donor.Donor.FirstNames,
			LastName:                  donor.Donor.LastName,
			DateOfBirth:               donor.Donor.DateOfBirth,
			Email:                     donor.Donor.Email,
			Mobile:                    donor.Donor.Mobile,
			Address:                   donor.Donor.Address,
			OtherNamesKnownBy:         donor.Donor.OtherNames,
			ContactLanguagePreference: donor.Donor.ContactLanguagePreference,
		},
		CertificateProvider: lpadata.CertificateProvider{
			UID:        donor.CertificateProvider.UID,
			FirstNames: donor.CertificateProvider.FirstNames,
			LastName:   donor.CertificateProvider.LastName,
			Email:      donor.CertificateProvider.Email,
			Phone:      donor.CertificateProvider.Mobile,
			Address:    donor.CertificateProvider.Address,
			Channel:    donor.CertificateProvider.CarryOutBy,
		},
		Restrictions:                     donor.Restrictions,
		SignedAt:                         donor.SignedAt,
		WitnessedByCertificateProviderAt: donor.WitnessedByCertificateProviderAt,
	}

	if donor.DonorIdentityConfirmed() {
		body.Donor.IdentityCheck = &lpadata.IdentityCheck{
			CheckedAt: donor.IdentityUserData.CheckedAt,
			Type:      "one-login",
		}
	}

	if !donor.CertificateProviderNotRelatedConfirmedAt.IsZero() {
		body.CertificateProviderNotRelatedConfirmedAt = &donor.CertificateProviderNotRelatedConfirmedAt
	}

	if !donor.WitnessedByIndependentWitnessAt.IsZero() {
		body.WitnessedByIndependentWitnessAt = &donor.WitnessedByIndependentWitnessAt
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
		body.Attorneys = append(body.Attorneys, lpadata.Attorney{
			UID:             attorney.UID,
			FirstNames:      attorney.FirstNames,
			LastName:        attorney.LastName,
			DateOfBirth:     attorney.DateOfBirth,
			Email:           attorney.Email,
			Address:         attorney.Address,
			Status:          lpadata.AttorneyStatusActive,
			AppointmentType: lpadata.AppointmentTypeOriginal,
			Channel:         attorney.Channel(),
		})
	}

	if trustCorporation := donor.Attorneys.TrustCorporation; trustCorporation.Name != "" {
		body.TrustCorporations = append(body.TrustCorporations, lpadata.TrustCorporation{
			UID:             trustCorporation.UID,
			Name:            trustCorporation.Name,
			CompanyNumber:   trustCorporation.CompanyNumber,
			Email:           trustCorporation.Email,
			Address:         trustCorporation.Address,
			Status:          lpadata.AttorneyStatusActive,
			AppointmentType: lpadata.AppointmentTypeOriginal,
			Channel:         trustCorporation.Channel(),
		})
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		body.Attorneys = append(body.Attorneys, lpadata.Attorney{
			UID:             attorney.UID,
			FirstNames:      attorney.FirstNames,
			LastName:        attorney.LastName,
			DateOfBirth:     attorney.DateOfBirth,
			Email:           attorney.Email,
			Address:         attorney.Address,
			Status:          lpadata.AttorneyStatusInactive,
			AppointmentType: lpadata.AppointmentTypeReplacement,
			Channel:         attorney.Channel(),
		})
	}

	if trustCorporation := donor.ReplacementAttorneys.TrustCorporation; trustCorporation.Name != "" {
		body.TrustCorporations = append(body.TrustCorporations, lpadata.TrustCorporation{
			UID:             trustCorporation.UID,
			Name:            trustCorporation.Name,
			CompanyNumber:   trustCorporation.CompanyNumber,
			Email:           trustCorporation.Email,
			Address:         trustCorporation.Address,
			Status:          lpadata.AttorneyStatusInactive,
			AppointmentType: lpadata.AppointmentTypeReplacement,
			Channel:         trustCorporation.Channel(),
		})
	}

	for _, person := range donor.PeopleToNotify {
		body.PeopleToNotify = append(body.PeopleToNotify, lpadata.PersonToNotify{
			UID:        person.UID,
			FirstNames: person.FirstNames,
			LastName:   person.LastName,
			Address:    person.Address,
		})
	}

	if !donor.IndependentWitness.UID.IsZero() {
		body.IndependentWitness = &lpadata.IndependentWitness{
			UID:        donor.IndependentWitness.UID,
			FirstNames: donor.IndependentWitness.FirstNames,
			LastName:   donor.IndependentWitness.LastName,
			Phone:      donor.IndependentWitness.Mobile,
			Address:    donor.IndependentWitness.Address,
		}
	}

	if !donor.AuthorisedSignatory.UID.IsZero() {
		body.AuthorisedSignatory = &lpadata.AuthorisedSignatory{
			UID:        donor.AuthorisedSignatory.UID,
			FirstNames: donor.AuthorisedSignatory.FirstNames,
			LastName:   donor.AuthorisedSignatory.LastName,
		}
	}

	return body
}

func (c *Client) SendLpa(ctx context.Context, uid string, body CreateLpa) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(body); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, c.baseURL+"/lpas/"+uid, &buf)
	if err != nil {
		return err
	}

	resp, err := c.do(ctx, body.Donor.UID, req)
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

type lpaResponse struct {
	LpaType                                     lpadata.LpaType                    `json:"lpaType"`
	Donor                                       lpadata.Donor                      `json:"donor"`
	Channel                                     lpadata.Channel                    `json:"channel"`
	Attorneys                                   []lpadata.Attorney                 `json:"attorneys"`
	TrustCorporations                           []lpadata.TrustCorporation         `json:"trustCorporations"`
	CertificateProvider                         lpadata.CertificateProvider        `json:"certificateProvider"`
	PeopleToNotify                              []lpadata.PersonToNotify           `json:"peopleToNotify"`
	IndependentWitness                          *lpadata.IndependentWitness        `json:"independentWitness"`
	AuthorisedSignatory                         *lpadata.AuthorisedSignatory       `json:"authorisedSignatory"`
	HowAttorneysMakeDecisions                   lpadata.AttorneysAct               `json:"howAttorneysMakeDecisions"`
	HowAttorneysMakeDecisionsDetails            string                             `json:"howAttorneysMakeDecisionsDetails"`
	HowReplacementAttorneysMakeDecisions        lpadata.AttorneysAct               `json:"howReplacementAttorneysMakeDecisions"`
	HowReplacementAttorneysMakeDecisionsDetails string                             `json:"howReplacementAttorneysMakeDecisionsDetails"`
	HowReplacementAttorneysStepIn               lpadata.ReplacementAttorneysStepIn `json:"howReplacementAttorneysStepIn"`
	HowReplacementAttorneysStepInDetails        string                             `json:"howReplacementAttorneysStepInDetails"`
	Restrictions                                string                             `json:"restrictionsAndConditions"`
	WhenTheLpaCanBeUsed                         lpadata.CanBeUsedWhen              `json:"whenTheLpaCanBeUsed"`
	LifeSustainingTreatmentOption               lpadata.LifeSustainingTreatment    `json:"lifeSustainingTreatmentOption"`
	SignedAt                                    time.Time                          `json:"signedAt"`
	WitnessedByCertificateProviderAt            time.Time                          `json:"witnessedByCertificateProviderAt"`
	WitnessedByIndependentWitnessAt             *time.Time                         `json:"witnessedByIndependentWitnessAt"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time                         `json:"certificateProviderNotRelatedConfirmedAt"`
	UID                                         string                             `json:"uid"`
	Status                                      lpadata.Status                     `json:"status"`
	RegistrationDate                            time.Time                          `json:"registrationDate"`
	UpdatedAt                                   time.Time                          `json:"updatedAt"`
}

func lpaResponseToLpa(l lpaResponse) *lpadata.Lpa {
	data := &lpadata.Lpa{
		LpaUID:       l.UID,
		Status:       l.Status,
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
			OtherNamesKnownBy:         l.Donor.OtherNamesKnownBy,
			Channel:                   l.Channel,
			ContactLanguagePreference: l.Donor.ContactLanguagePreference,
		},
		CertificateProvider: l.CertificateProvider,
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
		Restrictions:                     l.Restrictions,
		WhenCanTheLpaBeUsed:              l.WhenTheLpaCanBeUsed,
		LifeSustainingTreatmentOption:    l.LifeSustainingTreatmentOption,
		SignedAt:                         l.SignedAt,
		WitnessedByCertificateProviderAt: l.WitnessedByCertificateProviderAt,
	}

	if l.LpaType.IsPersonalWelfare() {
		data.WhenCanTheLpaBeUsed = lpadata.CanBeUsedWhenCapacityLost
	}

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
			Removed:                   a.Status.IsRemoved(),
		}

		if a.AppointmentType.IsReplacement() {
			data.ReplacementAttorneys.Attorneys = append(data.ReplacementAttorneys.Attorneys, at)
		} else {
			data.Attorneys.Attorneys = append(data.Attorneys.Attorneys, at)
		}
	}

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
			Removed:                   t.Status.IsRemoved(),
		}

		if t.AppointmentType.IsReplacement() {
			data.ReplacementAttorneys.TrustCorporation = tc
		} else {
			data.Attorneys.TrustCorporation = tc
		}
	}

	for _, p := range l.PeopleToNotify {
		data.PeopleToNotify = append(data.PeopleToNotify, lpadata.PersonToNotify{
			UID:        p.UID,
			FirstNames: p.FirstNames,
			LastName:   p.LastName,
			Address:    p.Address,
		})
	}

	if l.AuthorisedSignatory != nil {
		data.AuthorisedSignatory = lpadata.AuthorisedSignatory{
			UID:        l.AuthorisedSignatory.UID,
			FirstNames: l.AuthorisedSignatory.FirstNames,
			LastName:   l.AuthorisedSignatory.LastName,
		}
	}

	if l.IndependentWitness != nil {
		data.IndependentWitness = lpadata.IndependentWitness{
			UID:        l.IndependentWitness.UID,
			FirstNames: l.IndependentWitness.FirstNames,
			LastName:   l.IndependentWitness.LastName,
			Phone:      l.IndependentWitness.Phone,
			Address:    l.IndependentWitness.Address,
		}
	}

	if v := l.CertificateProviderNotRelatedConfirmedAt; v != nil {
		data.CertificateProviderNotRelatedConfirmedAt = *v
	}

	if v := l.WitnessedByIndependentWitnessAt; v != nil {
		data.WitnessedByIndependentWitnessAt = *v
	}

	if l.Donor.IdentityCheck != nil {
		data.Donor.IdentityCheck = &lpadata.IdentityCheck{
			CheckedAt: l.Donor.IdentityCheck.CheckedAt,
			Type:      l.Donor.IdentityCheck.Type,
		}
	}

	return data
}

func LpaFromDonorProvided(l *donordata.Provided) *lpadata.Lpa {
	data := &lpadata.Lpa{
		LpaID:     l.LpaID,
		LpaUID:    l.LpaUID,
		UpdatedAt: l.UpdatedAt,
		Type:      l.Type,
		Donor: lpadata.Donor{
			UID:                       l.Donor.UID,
			FirstNames:                l.Donor.FirstNames,
			LastName:                  l.Donor.LastName,
			Email:                     l.Donor.Email,
			OtherNamesKnownBy:         l.Donor.OtherNames,
			DateOfBirth:               l.Donor.DateOfBirth,
			Address:                   l.Donor.Address,
			Channel:                   l.Donor.Channel,
			ContactLanguagePreference: l.Donor.ContactLanguagePreference,
		},
		CertificateProvider: lpadata.CertificateProvider{
			UID:        l.CertificateProvider.UID,
			FirstNames: l.CertificateProvider.FirstNames,
			LastName:   l.CertificateProvider.LastName,
			Email:      l.CertificateProvider.Email,
			Phone:      l.CertificateProvider.Mobile,
			Address:    l.CertificateProvider.Address,
			Channel:    l.CertificateProvider.CarryOutBy,
		},
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
		WitnessedByCertificateProviderAt:         l.WitnessedByCertificateProviderAt,
		WitnessedByIndependentWitnessAt:          l.WitnessedByIndependentWitnessAt,
		CertificateProviderNotRelatedConfirmedAt: l.CertificateProviderNotRelatedConfirmedAt,
		Correspondent: lpadata.Correspondent{
			UID:        l.Correspondent.UID,
			FirstNames: l.Correspondent.FirstNames,
			LastName:   l.Correspondent.LastName,
			Email:      l.Correspondent.Email,
			Address:    l.Correspondent.Address,
		},
	}

	for _, a := range l.Attorneys.Attorneys {
		data.Attorneys.Attorneys = append(data.Attorneys.Attorneys, lpadata.Attorney{
			UID:         a.UID,
			FirstNames:  a.FirstNames,
			LastName:    a.LastName,
			DateOfBirth: a.DateOfBirth,
			Email:       a.Email,
			Address:     a.Address,
		})
	}

	if c := l.Attorneys.TrustCorporation; c.Name != "" {
		data.Attorneys.TrustCorporation = lpadata.TrustCorporation{
			UID:           c.UID,
			Name:          c.Name,
			CompanyNumber: c.CompanyNumber,
			Email:         c.Email,
			Address:       c.Address,
		}
	}

	for _, a := range l.ReplacementAttorneys.Attorneys {
		data.ReplacementAttorneys.Attorneys = append(data.ReplacementAttorneys.Attorneys, lpadata.Attorney{
			UID:         a.UID,
			FirstNames:  a.FirstNames,
			LastName:    a.LastName,
			DateOfBirth: a.DateOfBirth,
			Email:       a.Email,
			Address:     a.Address,
		})
	}

	if c := l.ReplacementAttorneys.TrustCorporation; c.Name != "" {
		data.ReplacementAttorneys.TrustCorporation = lpadata.TrustCorporation{
			UID:           c.UID,
			Name:          c.Name,
			CompanyNumber: c.CompanyNumber,
			Email:         c.Email,
			Address:       c.Address,
		}
	}

	if l.DonorIdentityConfirmed() {
		data.Donor.IdentityCheck = &lpadata.IdentityCheck{
			CheckedAt: l.IdentityUserData.CheckedAt,
			Type:      "one-login",
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

	if v := l.Voucher; v.Allowed {
		data.Voucher = lpadata.Voucher{
			UID:        v.UID,
			FirstNames: v.FirstNames,
			LastName:   v.LastName,
			Email:      v.Email,
		}
	}

	if v := l.AuthorisedSignatory; !v.UID.IsZero() {
		data.AuthorisedSignatory = lpadata.AuthorisedSignatory{
			UID:        v.UID,
			FirstNames: v.FirstNames,
			LastName:   v.LastName,
		}
	}

	if v := l.IndependentWitness; !v.UID.IsZero() {
		data.IndependentWitness = lpadata.IndependentWitness{
			UID:        v.UID,
			FirstNames: v.FirstNames,
			LastName:   v.LastName,
			Phone:      v.Mobile,
			Address:    v.Address,
		}
	}

	return data
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
