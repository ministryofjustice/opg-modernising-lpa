package lpastore

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

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
	Restrictions                                string                           `json:"restrictionsAndConditions"`
	WhenTheLpaCanBeUsed                         actor.CanBeUsedWhen              `json:"whenTheLpaCanBeUsed,omitempty"`
	LifeSustainingTreatmentOption               actor.LifeSustainingTreatment    `json:"lifeSustainingTreatmentOption,omitempty"`
	SignedAt                                    time.Time                        `json:"signedAt"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time                       `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
}

type lpaRequestDonor struct {
	UID               actoruid.UID  `json:"uid"`
	FirstNames        string        `json:"firstNames"`
	LastName          string        `json:"lastName"`
	DateOfBirth       date.Date     `json:"dateOfBirth"`
	Email             string        `json:"email"`
	Address           place.Address `json:"address"`
	OtherNamesKnownBy string        `json:"otherNamesKnownBy,omitempty"`
}

type lpaRequestAttorney struct {
	UID         actoruid.UID  `json:"uid"`
	FirstNames  string        `json:"firstNames"`
	LastName    string        `json:"lastName"`
	DateOfBirth date.Date     `json:"dateOfBirth"`
	Email       string        `json:"email"`
	Address     place.Address `json:"address"`
	Status      string        `json:"status"`
}

type lpaRequestTrustCorporation struct {
	UID           actoruid.UID  `json:"uid"`
	Name          string        `json:"name"`
	CompanyNumber string        `json:"companyNumber"`
	Email         string        `json:"email"`
	Address       place.Address `json:"address"`
	Status        string        `json:"status"`
}

type lpaRequestCertificateProvider struct {
	UID        actoruid.UID                        `json:"uid"`
	FirstNames string                              `json:"firstNames"`
	LastName   string                              `json:"lastName"`
	Email      string                              `json:"email,omitempty"`
	Phone      string                              `json:"phone"`
	Address    place.Address                       `json:"address"`
	Channel    actor.CertificateProviderCarryOutBy `json:"channel"`
}

type lpaRequestPersonToNotify struct {
	UID        actoruid.UID  `json:"uid"`
	FirstNames string        `json:"firstNames"`
	LastName   string        `json:"lastName"`
	Address    place.Address `json:"address"`
}

func (c *Client) SendLpa(ctx context.Context, donor *actor.DonorProvidedDetails) error {
	body := lpaRequest{
		LpaType: donor.Type,
		Donor: lpaRequestDonor{
			UID:               donor.Donor.UID,
			FirstNames:        donor.Donor.FirstNames,
			LastName:          donor.Donor.LastName,
			DateOfBirth:       donor.Donor.DateOfBirth,
			Email:             donor.Donor.Email,
			Address:           donor.Donor.Address,
			OtherNamesKnownBy: donor.Donor.OtherNames,
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

	if !donor.CertificateProviderNotRelatedConfirmedAt.IsZero() {
		body.CertificateProviderNotRelatedConfirmedAt = &donor.CertificateProviderNotRelatedConfirmedAt
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
			UID:         attorney.UID,
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
			UID:           trustCorporation.UID,
			Name:          trustCorporation.Name,
			CompanyNumber: trustCorporation.CompanyNumber,
			Email:         trustCorporation.Email,
			Address:       trustCorporation.Address,
			Status:        statusActive,
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

	return c.do(ctx, donor.Donor.UID, req, nil)
}

type lpaResponse struct {
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
	Restrictions                                string                           `json:"restrictionsAndConditions"`
	WhenTheLpaCanBeUsed                         actor.CanBeUsedWhen              `json:"whenTheLpaCanBeUsed,omitempty"`
	LifeSustainingTreatmentOption               actor.LifeSustainingTreatment    `json:"lifeSustainingTreatmentOption,omitempty"`
	SignedAt                                    time.Time                        `json:"signedAt"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time                       `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
	UID                                         string                           `json:"uid"`
	Status                                      string                           `json:"status"`
	RegistrationDate                            date.Date                        `json:"registrationDate"`
	UpdatedAt                                   date.Date                        `json:"updatedAt"`
}

func (l *lpaResponse) ToDonorProvidedDetails() *actor.DonorProvidedDetails {
	var attorneys, replacementAttorneys []actor.Attorney
	for _, a := range l.Attorneys {
		at := actor.Attorney{
			UID:         a.UID,
			FirstNames:  a.FirstNames,
			LastName:    a.LastName,
			DateOfBirth: a.DateOfBirth,
			Email:       a.Email,
			Address:     a.Address,
		}

		if a.Status == "replacement" {
			replacementAttorneys = append(replacementAttorneys, at)
		} else if a.Status == "active" {
			attorneys = append(attorneys, at)
		}
	}

	var trustCorporation, replacementTrustCorporation actor.TrustCorporation
	for _, t := range l.TrustCorporations {
		tc := actor.TrustCorporation{
			UID:           t.UID,
			Name:          t.Name,
			CompanyNumber: t.CompanyNumber,
			Email:         t.Email,
			Address:       t.Address,
		}

		if t.Status == "replacement" {
			replacementTrustCorporation = tc
		} else if t.Status == "active" {
			trustCorporation = tc
		}
	}

	var peopleToNotify []actor.PersonToNotify
	for _, p := range l.PeopleToNotify {
		peopleToNotify = append(peopleToNotify, actor.PersonToNotify{
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

	return &actor.DonorProvidedDetails{
		LpaUID:       l.UID,
		RegisteredAt: l.RegistrationDate.Time(),
		UpdatedAt:    l.UpdatedAt.Time(),
		Type:         l.LpaType,
		Donor: actor.Donor{
			UID:         l.Donor.UID,
			FirstNames:  l.Donor.FirstNames,
			LastName:    l.Donor.LastName,
			DateOfBirth: l.Donor.DateOfBirth,
			Email:       l.Donor.Email,
			Address:     l.Donor.Address,
			OtherNames:  l.Donor.OtherNamesKnownBy,
		},
		Attorneys: actor.Attorneys{
			Attorneys:        attorneys,
			TrustCorporation: trustCorporation,
		},
		ReplacementAttorneys: actor.Attorneys{
			Attorneys:        replacementAttorneys,
			TrustCorporation: replacementTrustCorporation,
		},
		CertificateProvider: actor.CertificateProvider{
			UID:        l.CertificateProvider.UID,
			FirstNames: l.CertificateProvider.FirstNames,
			LastName:   l.CertificateProvider.LastName,
			Email:      l.CertificateProvider.Email,
			Address:    l.CertificateProvider.Address,
			Mobile:     l.CertificateProvider.Phone,
			CarryOutBy: l.CertificateProvider.Channel,
		},
		PeopleToNotify: peopleToNotify,
		AttorneyDecisions: actor.AttorneyDecisions{
			How:     l.HowAttorneysMakeDecisions,
			Details: l.HowAttorneysMakeDecisionsDetails,
		},
		ReplacementAttorneyDecisions: actor.AttorneyDecisions{
			How:     l.HowReplacementAttorneysMakeDecisions,
			Details: l.HowReplacementAttorneysMakeDecisionsDetails,
		},
		HowShouldReplacementAttorneysStepIn:        l.HowReplacementAttorneysStepIn,
		HowShouldReplacementAttorneysStepInDetails: l.HowReplacementAttorneysStepInDetails,
		Restrictions:                             l.Restrictions,
		WhenCanTheLpaBeUsed:                      l.WhenTheLpaCanBeUsed,
		LifeSustainingTreatmentOption:            l.LifeSustainingTreatmentOption,
		SignedAt:                                 l.SignedAt,
		CertificateProviderNotRelatedConfirmedAt: confirmedAt,
	}
}

func (c *Client) Lpa(ctx context.Context, lpaUID string) (*actor.DonorProvidedDetails, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/lpas/"+lpaUID, nil)
	if err != nil {
		return nil, err
	}

	var v lpaResponse
	if err := c.do(ctx, actoruid.Service, req, &v); err != nil {
		return nil, err
	}

	return v.ToDonorProvidedDetails(), nil
}
