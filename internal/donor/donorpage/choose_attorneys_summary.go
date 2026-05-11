package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
)

type chooseAttorneysSummaryData struct {
	App       appcontext.Data
	Form      *donordata.YesNoMaybeForm
	Donor     *donordata.Provided
	CanChoose bool
}

func ChooseAttorneysSummary(tmpl template.Template, service AttorneyService, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if provided.Attorneys.Len() == 0 {
			return donor.PathChooseAttorneys.Redirect(w, r, appData, provided)
		}

		attorneys, err := service.Reusable(r.Context(), provided)
		if err != nil {
			return err
		}

		data := &chooseAttorneysSummaryData{
			App:       appData,
			Donor:     provided,
			Form:      donordata.NewYesNoMaybeForm(appData.Localizer.T("yesToAddAnotherAttorney")),
			CanChoose: len(attorneys) > 0,
		}

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				redirectUrl := donor.PathTaskList
				if provided.Attorneys.Len() > 1 {
					redirectUrl = donor.PathHowShouldAttorneysMakeDecisions
				}

				if data.Form.Enum.Value.IsYes() {
					return donor.PathEnterAttorney.RedirectQuery(w, r, appData, provided, url.Values{"addAnother": {"1"}, "id": {newUID().String()}})
				} else if data.Form.Enum.Value.IsMaybe() {
					return donor.PathChooseAttorneys.Redirect(w, r, appData, provided)
				} else {
					return redirectUrl.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
