package attorneypage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type wouldLikeSecondSignatoryData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
}

func WouldLikeSecondSignatory(tmpl template.Template, attorneyStore AttorneyStore, lpaStoreResolvingService LpaStoreResolvingService, lpaStoreClient LpaStoreClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided) error {
		if attorneyProvidedDetails.Signed() {
			return page.Paths.Attorney.WhatHappensNext.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
		}

		data := &wouldLikeSecondSignatoryData{
			App:  appData,
			Form: form.NewYesNoForm(attorneyProvidedDetails.WouldLikeSecondSignatory),
		}

		if r.Method == http.MethodPost {
			form := form.ReadYesNoForm(r, "yesIfWouldLikeSecondSignatory")
			data.Errors = form.Validate()

			if data.Errors.None() {
				attorneyProvidedDetails.WouldLikeSecondSignatory = form.YesNo

				if form.YesNo.IsYes() {
					if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
						return err
					}

					return page.Paths.Attorney.Sign.RedirectQuery(w, r, appData, attorneyProvidedDetails.LpaID, url.Values{"second": {""}})
				}

				lpa, err := lpaStoreResolvingService.Get(r.Context())
				if err != nil {
					return err
				}

				hasSigned := (appData.IsReplacementAttorney() &&
					len(lpa.ReplacementAttorneys.TrustCorporation.Signatories) > 0) ||
					(!appData.IsReplacementAttorney() &&
						len(lpa.Attorneys.TrustCorporation.Signatories) > 0)

				if !hasSigned {
					if err := lpaStoreClient.SendAttorney(r.Context(), lpa, attorneyProvidedDetails); err != nil {
						return err
					}
				}

				if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
					return err
				}

				return page.Paths.Attorney.WhatHappensNext.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
