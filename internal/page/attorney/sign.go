package attorney

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

func canSign(ctx context.Context, certificateProviderStore CertificateProviderStore, donor *lpastore.Lpa) (bool, error) {
	ctx = page.ContextWithSessionData(ctx, &page.SessionData{LpaID: donor.LpaID})

	certificateProvider, err := certificateProviderStore.GetAny(ctx)
	if err != nil {
		if errors.Is(err, dynamo.NotFoundError{}) {
			certificateProvider = &actor.CertificateProviderProvidedDetails{}
		} else {
			return false, err
		}
	}

	return !donor.SignedAt.IsZero() && certificateProvider.Signed(donor.SignedAt), nil
}

type signData struct {
	App                         page.AppData
	Errors                      validation.List
	LpaID                       string
	Attorney                    lpastore.Attorney
	TrustCorporation            lpastore.TrustCorporation
	IsReplacement               bool
	IsSecondSignatory           bool
	LpaCanBeUsedWhenHasCapacity bool
	Form                        *signForm
}

func Sign(
	tmpl template.Template,
	lpaStoreResolvingService LpaStoreResolvingService,
	certificateProviderStore CertificateProviderStore,
	attorneyStore AttorneyStore,
	lpaStoreClient LpaStoreClient,
	now func() time.Time,
) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		signatoryIndex := 0
		if r.URL.Query().Has("second") {
			signatoryIndex = 1
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		if ok, _ := canSign(r.Context(), certificateProviderStore, lpa); !ok {
			return page.Paths.Attorney.TaskList.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
		}

		data := &signData{
			App:                         appData,
			LpaID:                       lpa.LpaID,
			IsReplacement:               appData.IsReplacementAttorney(),
			IsSecondSignatory:           signatoryIndex == 1,
			LpaCanBeUsedWhenHasCapacity: lpa.WhenCanTheLpaBeUsed.IsHasCapacity(),
			Form: &signForm{
				Confirm: !attorneyProvidedDetails.Confirmed.IsZero(),
			},
		}

		if appData.IsTrustCorporation() {
			signatory := attorneyProvidedDetails.AuthorisedSignatories[signatoryIndex]

			data.Form = &signForm{
				FirstNames:        signatory.FirstNames,
				LastName:          signatory.LastName,
				ProfessionalTitle: signatory.ProfessionalTitle,
				Confirm:           !signatory.Confirmed.IsZero(),
			}

			if appData.IsReplacementAttorney() {
				data.TrustCorporation = lpa.ReplacementAttorneys.TrustCorporation
			} else {
				data.TrustCorporation = lpa.Attorneys.TrustCorporation
			}
		} else {
			attorneys := lpa.Attorneys
			if appData.IsReplacementAttorney() {
				attorneys = lpa.ReplacementAttorneys
			}

			attorney, ok := attorneys.Get(appData.AttorneyUID)
			if !ok {
				return page.Paths.Attorney.Start.Redirect(w, r, appData)
			}

			data.Attorney = attorney
		}

		if r.Method == http.MethodPost {
			data.Form = readSignForm(r)
			data.Errors = data.Form.Validate(appData.IsTrustCorporation(), appData.IsReplacementAttorney())

			if data.Errors.None() {
				if signatoryIndex == 1 {
					attorneyProvidedDetails.Tasks.SignTheLpaSecond = actor.TaskCompleted
				} else {
					attorneyProvidedDetails.Tasks.SignTheLpa = actor.TaskCompleted
				}

				if appData.IsTrustCorporation() {
					attorneyProvidedDetails.AuthorisedSignatories[signatoryIndex] = actor.TrustCorporationSignatory{
						FirstNames:        data.Form.FirstNames,
						LastName:          data.Form.LastName,
						ProfessionalTitle: data.Form.ProfessionalTitle,
						Confirmed:         now(),
					}
				} else {
					attorneyProvidedDetails.Confirmed = now()
				}

				if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
					return err
				}

				if appData.IsTrustCorporation() && signatoryIndex == 0 {
					return page.Paths.Attorney.WouldLikeSecondSignatory.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
				} else {
					if err := lpaStoreClient.SendAttorney(r.Context(), lpa, attorneyProvidedDetails); err != nil {
						return err
					}

					return page.Paths.Attorney.WhatHappensNext.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
				}
			}
		}

		return tmpl(w, data)
	}
}

type signForm struct {
	FirstNames        string
	LastName          string
	ProfessionalTitle string
	Confirm           bool
}

func readSignForm(r *http.Request) *signForm {
	return &signForm{
		FirstNames:        page.PostFormString(r, "first-names"),
		LastName:          page.PostFormString(r, "last-name"),
		ProfessionalTitle: page.PostFormString(r, "professional-title"),
		Confirm:           page.PostFormString(r, "confirm") == "1",
	}
}

func (f *signForm) Validate(isTrustCorporation, isReplacement bool) validation.List {
	var errors validation.List

	if isTrustCorporation {
		errors.String("first-names", "firstNames", f.FirstNames,
			validation.Empty())
		errors.String("last-name", "lastName", f.LastName,
			validation.Empty())
		errors.String("professional-title", "professionalTitle", f.ProfessionalTitle,
			validation.Empty())
		errors.Bool("confirm", "youMustSelectTheBoxToSignAttorney", f.Confirm,
			validation.Selected().CustomError())
	} else if isReplacement {
		errors.Bool("confirm", "youMustSelectTheBoxToSignReplacementAttorney", f.Confirm,
			validation.Selected().CustomError())
	} else {
		errors.Bool("confirm", "youMustSelectTheBoxToSignAttorney", f.Confirm,
			validation.Selected().CustomError())
	}

	return errors
}
