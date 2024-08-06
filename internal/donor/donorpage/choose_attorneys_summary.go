package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAttorneysSummaryData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func ChooseAttorneysSummary(tmpl template.Template, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if provided.Attorneys.Len() == 0 {
			return donor.PathChooseAttorneys.RedirectQuery(w, r, appData, provided, url.Values{"id": {newUID().String()}})
		}

		data := &chooseAttorneysSummaryData{
			App:   appData,
			Donor: provided,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToAddAnotherAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirectUrl := donor.PathTaskList
				if provided.Attorneys.Len() > 1 {
					redirectUrl = donor.PathHowShouldAttorneysMakeDecisions
				}

				if data.Form.YesNo.IsYes() {
					return donor.PathChooseAttorneys.RedirectQuery(w, r, appData, provided, url.Values{"addAnother": {"1"}, "id": {newUID().String()}})
				} else {
					return redirectUrl.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
