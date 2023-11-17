package attorney

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type wouldLikeSecondSignatoryData struct {
	App     page.AppData
	Errors  validation.List
	YesNo   form.YesNo
	Options form.YesNoOptions
}

func WouldLikeSecondSignatory(tmpl template.Template, attorneyStore AttorneyStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		data := &wouldLikeSecondSignatoryData{
			App:     appData,
			YesNo:   attorneyProvidedDetails.WouldLikeSecondSignatory,
			Options: form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			form := form.ReadYesNoForm(r, "yesIfWouldLikeSecondSignatory")
			data.Errors = form.Validate()

			if data.Errors.None() {
				attorneyProvidedDetails.WouldLikeSecondSignatory = form.YesNo

				if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
					return err
				}

				if form.YesNo.IsYes() {
					return page.Paths.Attorney.Sign.RedirectQuery(w, r, appData, attorneyProvidedDetails.LpaID, url.Values{"second": {""}})
				} else {
					return page.Paths.Attorney.WhatHappensNext.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
				}
			}
		}

		return tmpl(w, data)
	}
}
