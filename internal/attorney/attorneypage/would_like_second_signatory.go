package attorneypage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type wouldLikeSecondSignatoryData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *newforms.YesNoForm
}

func WouldLikeSecondSignatory(tmpl template.Template, attorneyStore AttorneyStore, lpaStoreClient LpaStoreClient) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *attorneydata.Provided, lpa *lpadata.Lpa) error {
		if attorneyProvidedDetails.Signed() {
			return attorney.PathWhatHappensNext.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
		}

		data := &wouldLikeSecondSignatoryData{
			App:  appData,
			Form: newforms.NewYesNoForm(appData.Localizer.T("yesIfWouldLikeSecondSignatory")),
		}

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				attorneyProvidedDetails.WouldLikeSecondSignatory = data.Form.YesNo.Value

				if data.Form.YesNo.Value.IsYes() {
					if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
						return err
					}

					return attorney.PathSign.RedirectQuery(w, r, appData, attorneyProvidedDetails.LpaID, url.Values{"second": {""}})
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

				return attorney.PathWhatHappensNext.Redirect(w, r, appData, attorneyProvidedDetails.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
