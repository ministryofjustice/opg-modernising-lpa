package lpastore

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"slices"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/place"
)

type abstractError struct {
	Detail string `json:"detail"`
}

type lpaRequest struct {
	LpaType                                     actor.LpaType                    `json:"lpaType"`
	Channel                                     actor.Channel                    `json:"channel"`
	Donor                                       lpaRequestDonor                  `json:"donor"`
	Attorneys                                   []lpaRequestAttorney             `json:"attorneys"`
	TrustCorporations                           []lpaRequestTrustCorporation     `json:"trustCorporations,omitempty"`
	CertificateProvider                         lpaRequestCertificateProvider    `json:"certificateProvider"`
	PeopleToNotify                              []lpaRequestPersonToNotify       `json:"peopleToNotify,omitempty"`
	HowAttorneysMakeDecisions                   donordata.AttorneysAct           `json:"howAttorneysMakeDecisions,omitempty"`
	HowAttorneysMakeDecisionsDetails            string                           `json:"howAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysMakeDecisions        donordata.AttorneysAct           `json:"howReplacementAttorneysMakeDecisions,omitempty"`
	HowReplacementAttorneysMakeDecisionsDetails string                           `json:"howReplacementAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysStepIn               actor.ReplacementAttorneysStepIn `json:"howReplacementAttorneysStepIn,omitempty"`
	HowReplacementAttorneysStepInDetails        string                           `json:"howReplacementAttorneysStepInDetails,omitempty"`
	Restrictions                                string                           `json:"restrictionsAndConditions"`
	WhenTheLpaCanBeUsed                         donordata.CanBeUsedWhen          `json:"whenTheLpaCanBeUsed,omitempty"`
	LifeSustainingTreatmentOption               actor.LifeSustainingTreatment    `json:"lifeSustainingTreatmentOption,omitempty"`
	SignedAt                                    time.Time                        `json:"signedAt"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time                       `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
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
	UID         actoruid.UID  `json:"uid"`
	FirstNames  string        `json:"firstNames"`
	LastName    string        `json:"lastName"`
	DateOfBirth date.Date     `json:"dateOfBirth"`
	Email       string        `json:"email,omitempty"`
	Address     place.Address `json:"address"`
	Status      string        `json:"status"`
	Channel     actor.Channel `json:"channel"`
}

type lpaRequestTrustCorporation struct {
	UID           actoruid.UID  `json:"uid"`
	Name          string        `json:"name"`
	CompanyNumber string        `json:"companyNumber"`
	Email         string        `json:"email,omitempty"`
	Address       place.Address `json:"address"`
	Status        string        `json:"status"`
	Channel       actor.Channel `json:"channel"`
}

type lpaRequestCertificateProvider struct {
	UID        actoruid.UID  `json:"uid"`
	FirstNames string        `json:"firstNames"`
	LastName   string        `json:"lastName"`
	Email      string        `json:"email,omitempty"`
	Phone      string        `json:"phone,omitempty"`
	Address    place.Address `json:"address"`
	Channel    actor.Channel `json:"channel"`
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
		Channel: actor.ChannelOnline,
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
			CheckedAt: donor.DonorIdentityUserData.RetrievedAt,
			Type:      "one-login",
		}
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
	Mobile                    string                      `json:"mobile"`
	Signatories               []TrustCorporationSignatory `json:"signatories"`
	ContactLanguagePreference localize.Lang               `json:"contactLanguagePreference"`
}

type Donor struct {
	UID                       actoruid.UID
	FirstNames                string
	LastName                  string
	Email                     string
	OtherNames                string
	DateOfBirth               date.Date
	Address                   place.Address
	Channel                   actor.Channel
	ContactLanguagePreference localize.Lang
	IdentityCheck             IdentityCheck
}

func (d Donor) FullName() string {
	return d.FirstNames + " " + d.LastName
}

type Attorney struct {
	UID                       actoruid.UID
	FirstNames                string
	LastName                  string
	DateOfBirth               date.Date
	Email                     string
	Address                   place.Address
	Mobile                    string
	SignedAt                  time.Time
	ContactLanguagePreference localize.Lang
	Channel                   actor.Channel
}

func (a Attorney) FullName() string {
	return a.FirstNames + " " + a.LastName
}

type TrustCorporation struct {
	UID                       actoruid.UID
	Name                      string
	CompanyNumber             string
	Email                     string
	Address                   place.Address
	Mobile                    string
	Signatories               []TrustCorporationSignatory
	ContactLanguagePreference localize.Lang
	Channel                   actor.Channel
}

type TrustCorporationSignatory struct {
	FirstNames        string    `json:"firstNames"`
	LastName          string    `json:"lastName"`
	ProfessionalTitle string    `json:"professionalTitle"`
	SignedAt          time.Time `json:"signedAt"`
}

type CertificateProvider struct {
	UID                       actoruid.UID  `json:"uid"`
	FirstNames                string        `json:"firstNames"`
	LastName                  string        `json:"lastName"`
	Email                     string        `json:"email,omitempty"`
	Phone                     string        `json:"phone,omitempty"`
	Address                   place.Address `json:"address"`
	Channel                   actor.Channel `json:"channel"`
	SignedAt                  time.Time     `json:"signedAt"`
	ContactLanguagePreference localize.Lang `json:"contactLanguagePreference"`
	IdentityCheck             IdentityCheck `json:"identityCheck"`
	// Relationship is not stored in the lpa-store so is defaulted to
	// Professional. We require it to determine whether to show the home address
	// page to a certificate provider.
	Relationship donordata.CertificateProviderRelationship
}

func (c CertificateProvider) FullName() string {
	return c.FirstNames + " " + c.LastName
}

type IdentityCheck struct {
	CheckedAt time.Time `json:"checkedAt"`
	Type      string    `json:"type"`
}

type lpaResponse struct {
	LpaType                                     actor.LpaType                    `json:"lpaType"`
	Donor                                       lpaRequestDonor                  `json:"donor"`
	Channel                                     actor.Channel                    `json:"channel"`
	Attorneys                                   []lpaResponseAttorney            `json:"attorneys"`
	TrustCorporations                           []lpaResponseTrustCorporation    `json:"trustCorporations,omitempty"`
	CertificateProvider                         CertificateProvider              `json:"certificateProvider"`
	PeopleToNotify                              []lpaRequestPersonToNotify       `json:"peopleToNotify,omitempty"`
	HowAttorneysMakeDecisions                   donordata.AttorneysAct           `json:"howAttorneysMakeDecisions,omitempty"`
	HowAttorneysMakeDecisionsDetails            string                           `json:"howAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysMakeDecisions        donordata.AttorneysAct           `json:"howReplacementAttorneysMakeDecisions,omitempty"`
	HowReplacementAttorneysMakeDecisionsDetails string                           `json:"howReplacementAttorneysMakeDecisionsDetails,omitempty"`
	HowReplacementAttorneysStepIn               actor.ReplacementAttorneysStepIn `json:"howReplacementAttorneysStepIn,omitempty"`
	HowReplacementAttorneysStepInDetails        string                           `json:"howReplacementAttorneysStepInDetails,omitempty"`
	Restrictions                                string                           `json:"restrictionsAndConditions"`
	WhenTheLpaCanBeUsed                         donordata.CanBeUsedWhen          `json:"whenTheLpaCanBeUsed,omitempty"`
	LifeSustainingTreatmentOption               actor.LifeSustainingTreatment    `json:"lifeSustainingTreatmentOption,omitempty"`
	SignedAt                                    time.Time                        `json:"signedAt"`
	CertificateProviderNotRelatedConfirmedAt    *time.Time                       `json:"certificateProviderNotRelatedConfirmedAt,omitempty"`
	UID                                         string                           `json:"uid"`
	Status                                      string                           `json:"status"`
	RegistrationDate                            time.Time                        `json:"registrationDate"`
	UpdatedAt                                   time.Time                        `json:"updatedAt"`
}

type Attorneys struct {
	Attorneys        []Attorney
	TrustCorporation TrustCorporation
}

func (a Attorneys) Len() int {
	if a.TrustCorporation.Name != "" {
		return 1 + len(a.Attorneys)
	}

	return len(a.Attorneys)
}

func (a Attorneys) Index(uid actoruid.UID) int {
	return slices.IndexFunc(a.Attorneys, func(a Attorney) bool { return a.UID == uid })
}

func (a Attorneys) Get(uid actoruid.UID) (Attorney, bool) {
	idx := a.Index(uid)
	if idx == -1 {
		return Attorney{}, false
	}

	return a.Attorneys[idx], true
}

func (a Attorneys) FullNames() []string {
	var names []string

	if a.TrustCorporation.Name != "" {
		names = append(names, a.TrustCorporation.Name)
	}

	for _, a := range a.Attorneys {
		names = append(names, fmt.Sprintf("%s %s", a.FirstNames, a.LastName))
	}

	return names
}

type Lpa struct {
	LpaKey                                     dynamo.LpaKeyType
	LpaOwnerKey                                dynamo.LpaOwnerKeyType
	LpaID                                      string
	LpaUID                                     string
	RegisteredAt                               time.Time
	WithdrawnAt                                time.Time
	PerfectAt                                  time.Time
	UpdatedAt                                  time.Time
	Type                                       actor.LpaType
	Donor                                      Donor
	Attorneys                                  Attorneys
	ReplacementAttorneys                       Attorneys
	CertificateProvider                        CertificateProvider
	PeopleToNotify                             actor.PeopleToNotify
	AttorneyDecisions                          donordata.AttorneyDecisions
	ReplacementAttorneyDecisions               donordata.AttorneyDecisions
	HowShouldReplacementAttorneysStepIn        actor.ReplacementAttorneysStepIn
	HowShouldReplacementAttorneysStepInDetails string
	Restrictions                               string
	WhenCanTheLpaBeUsed                        donordata.CanBeUsedWhen
	LifeSustainingTreatmentOption              actor.LifeSustainingTreatment
	// SignedAt is the date the Donor signed their LPA (and signifies it has been
	// witnessed by their CertificateProvider)
	SignedAt                                 time.Time
	CertificateProviderNotRelatedConfirmedAt time.Time
	Submitted                                bool
	Paid                                     bool
	IsOrganisationDonor                      bool
	Drafted                                  bool
	CannotRegister                           bool
}

func (l Lpa) AllAttorneysSigned() bool {
	if l.Attorneys.Len() == 0 {
		return false
	}

	for _, attorneys := range []Attorneys{l.Attorneys, l.ReplacementAttorneys} {
		for _, a := range attorneys.Attorneys {
			if a.SignedAt.IsZero() {
				return false
			}
		}

		if t := attorneys.TrustCorporation; t.Name != "" {
			if len(t.Signatories) == 0 {
				return false
			}

			for _, s := range t.Signatories {
				if s.SignedAt.IsZero() {
					return false
				}
			}
		}
	}

	return true
}

func lpaResponseToLpa(l lpaResponse) *Lpa {
	var attorneys, replacementAttorneys []Attorney
	for _, a := range l.Attorneys {
		at := Attorney{
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

	var trustCorporation, replacementTrustCorporation TrustCorporation
	for _, t := range l.TrustCorporations {
		tc := TrustCorporation{
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

	var identityCheck IdentityCheck
	if l.Donor.IdentityCheck != nil {
		identityCheck.CheckedAt = l.Donor.IdentityCheck.CheckedAt
		identityCheck.Type = l.Donor.IdentityCheck.Type
	}

	return &Lpa{
		LpaUID:       l.UID,
		RegisteredAt: l.RegistrationDate,
		UpdatedAt:    l.UpdatedAt,
		Type:         l.LpaType,
		Donor: Donor{
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
		Attorneys: Attorneys{
			Attorneys:        attorneys,
			TrustCorporation: trustCorporation,
		},
		ReplacementAttorneys: Attorneys{
			Attorneys:        replacementAttorneys,
			TrustCorporation: replacementTrustCorporation,
		},
		CertificateProvider: l.CertificateProvider,
		PeopleToNotify:      peopleToNotify,
		AttorneyDecisions: donordata.AttorneyDecisions{
			How:     l.HowAttorneysMakeDecisions,
			Details: l.HowAttorneysMakeDecisionsDetails,
		},
		ReplacementAttorneyDecisions: donordata.AttorneyDecisions{
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
		CannotRegister:                           l.Status == "cannot-register",
	}
}

func FromDonorProvidedDetails(l *actor.DonorProvidedDetails) *Lpa {
	attorneys := Attorneys{}
	for _, a := range l.Attorneys.Attorneys {
		attorneys.Attorneys = append(attorneys.Attorneys, Attorney{
			UID:         a.UID,
			FirstNames:  a.FirstNames,
			LastName:    a.LastName,
			DateOfBirth: a.DateOfBirth,
			Email:       a.Email,
			Address:     a.Address,
		})
	}

	if c := l.Attorneys.TrustCorporation; c.Name != "" {
		attorneys.TrustCorporation = TrustCorporation{
			UID:           c.UID,
			Name:          c.Name,
			CompanyNumber: c.CompanyNumber,
			Email:         c.Email,
			Address:       c.Address,
		}
	}

	var replacementAttorneys Attorneys
	for _, a := range l.ReplacementAttorneys.Attorneys {
		replacementAttorneys.Attorneys = append(replacementAttorneys.Attorneys, Attorney{
			UID:         a.UID,
			FirstNames:  a.FirstNames,
			LastName:    a.LastName,
			DateOfBirth: a.DateOfBirth,
			Email:       a.Email,
			Address:     a.Address,
		})
	}

	if c := l.ReplacementAttorneys.TrustCorporation; c.Name != "" {
		replacementAttorneys.TrustCorporation = TrustCorporation{
			UID:           c.UID,
			Name:          c.Name,
			CompanyNumber: c.CompanyNumber,
			Email:         c.Email,
			Address:       c.Address,
		}
	}

	var identityCheck IdentityCheck
	if l.DonorIdentityConfirmed() {
		identityCheck.CheckedAt = l.DonorIdentityUserData.RetrievedAt
		identityCheck.Type = "one-login"
	}

	return &Lpa{
		LpaID:     l.LpaID,
		LpaUID:    l.LpaUID,
		UpdatedAt: l.UpdatedAt,
		Type:      l.Type,
		Donor: Donor{
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
		CertificateProvider: CertificateProvider{
			UID:        l.CertificateProvider.UID,
			FirstNames: l.CertificateProvider.FirstNames,
			LastName:   l.CertificateProvider.LastName,
			Email:      l.CertificateProvider.Email,
			Phone:      l.CertificateProvider.Mobile,
			Address:    l.CertificateProvider.Address,
			Channel:    l.CertificateProvider.CarryOutBy,
		},
		PeopleToNotify:                             l.PeopleToNotify,
		AttorneyDecisions:                          l.AttorneyDecisions,
		ReplacementAttorneyDecisions:               l.ReplacementAttorneyDecisions,
		HowShouldReplacementAttorneysStepIn:        l.HowShouldReplacementAttorneysStepIn,
		HowShouldReplacementAttorneysStepInDetails: l.HowShouldReplacementAttorneysStepInDetails,
		Restrictions:                               l.Restrictions,
		WhenCanTheLpaBeUsed:                        l.WhenCanTheLpaBeUsed,
		LifeSustainingTreatmentOption:              l.LifeSustainingTreatmentOption,
		SignedAt:                                   l.SignedAt,
		CertificateProviderNotRelatedConfirmedAt:   l.CertificateProviderNotRelatedConfirmedAt,
	}
}

func (c *Client) Lpa(ctx context.Context, lpaUID string) (*Lpa, error) {
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

func (c *Client) Lpas(ctx context.Context, lpaUIDs []string) ([]*Lpa, error) {
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

	var lpas []*Lpa
	for _, l := range v.Lpas {
		lpas = append(lpas, lpaResponseToLpa(l))
	}

	return lpas, nil
}
