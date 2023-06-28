package attorney

import (
	"context"
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

func canSign(ctx context.Context, certificateProviderStore CertificateProviderStore, lpa *page.Lpa) (bool, error) {
	ctx = page.ContextWithSessionData(ctx, &page.SessionData{LpaID: lpa.ID})

	certificateProvider, err := certificateProviderStore.GetAny(ctx)
	if err != nil {
		if errors.Is(err, dynamo.NotFoundError{}) {
			certificateProvider = &actor.CertificateProviderProvidedDetails{}
		} else {
			return false, err
		}
	}

	progress := lpa.Progress(certificateProvider)

	return progress.LpaSigned.Completed() && progress.CertificateProviderDeclared.Completed(), nil
}

type signData struct {
	App                        page.AppData
	Errors                     validation.List
	Attorney                   actor.Attorney
	IsReplacement              bool
	LpaCanBeUsedWhenRegistered bool
	Form                       *signForm
}

func Sign(tmpl template.Template, donorStore DonorStore, certificateProviderStore CertificateProviderStore, attorneyStore AttorneyStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		if ok, _ := canSign(r.Context(), certificateProviderStore, lpa); !ok {
			return appData.Redirect(w, r, lpa, page.Paths.Attorney.TaskList.Format(attorneyProvidedDetails.LpaID))
		}

		attorneys := lpa.Attorneys
		if appData.IsReplacementAttorney() {
			attorneys = lpa.ReplacementAttorneys
		}

		attorney, ok := attorneys.Get(appData.AttorneyID)
		if !ok {
			return appData.Redirect(w, r, lpa, page.Paths.Attorney.Start.Format())
		}

		data := &signData{
			App:                        appData,
			Attorney:                   attorney,
			IsReplacement:              appData.IsReplacementAttorney(),
			LpaCanBeUsedWhenRegistered: lpa.WhenCanTheLpaBeUsed == page.CanBeUsedWhenRegistered,
			Form: &signForm{
				Confirm: attorneyProvidedDetails.Confirmed,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readSignForm(r)
			data.Errors = data.Form.Validate(appData.IsReplacementAttorney())

			if data.Errors.None() {
				attorneyProvidedDetails.Confirmed = true
				attorneyProvidedDetails.Tasks.SignTheLpa = actor.TaskCompleted

				if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.Attorney.WhatHappensNext.Format(attorneyProvidedDetails.LpaID))
			}
		}

		return tmpl(w, data)
	}
}

type signForm struct {
	Confirm bool
}

func readSignForm(r *http.Request) *signForm {
	return &signForm{
		Confirm: page.PostFormString(r, "confirm") == "1",
	}
}

func (f *signForm) Validate(isReplacement bool) validation.List {
	var errors validation.List

	if isReplacement {
		errors.Bool("confirm", "youMustSelectTheBoxToSignReplacementAttorney", f.Confirm,
			validation.Selected().CustomError())
	} else {
		errors.Bool("confirm", "youMustSelectTheBoxToSignAttorney", f.Confirm,
			validation.Selected().CustomError())
	}

	return errors
}
