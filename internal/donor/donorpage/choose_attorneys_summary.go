package donorpage

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type chooseAttorneysSummaryData struct {
	App       appcontext.Data
	Errors    validation.List
	Form      *chooseAttorneysSummaryForm
	Donor     *donordata.Provided
	Options   donordata.YesNoMaybeOptions
	CanChoose bool
}

func ChooseAttorneysSummary(tmpl template.Template, reuseStore ReuseStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if provided.Attorneys.Len() == 0 {
			return donor.PathChooseAttorneys.Redirect(w, r, appData, provided)
		}

		attorneys, err := reuseStore.Attorneys(r.Context(), provided)
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}

		data := &chooseAttorneysSummaryData{
			App:       appData,
			Donor:     provided,
			Options:   donordata.YesNoMaybeValues,
			CanChoose: len(attorneys) > 0,
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysSummaryForm(r, "yesToAddAnotherAttorney")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirectUrl := donor.PathTaskList
				if provided.Attorneys.Len() > 1 {
					redirectUrl = donor.PathHowShouldAttorneysMakeDecisions
				}

				if data.Form.Option.IsYes() {
					return donor.PathEnterAttorney.RedirectQuery(w, r, appData, provided, url.Values{"addAnother": {"1"}, "id": {newUID().String()}})
				} else if data.Form.Option.IsMaybe() {
					return donor.PathChooseAttorneys.Redirect(w, r, appData, provided)
				} else {
					return redirectUrl.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}

type chooseAttorneysSummaryForm struct {
	errorLabel string
	Option     donordata.YesNoMaybe
}

func readChooseAttorneysSummaryForm(r *http.Request, errorLabel string) *chooseAttorneysSummaryForm {
	option, _ := donordata.ParseYesNoMaybe(page.PostFormString(r, "option"))

	return &chooseAttorneysSummaryForm{
		errorLabel: errorLabel,
		Option:     option,
	}
}

func (f *chooseAttorneysSummaryForm) Validate() validation.List {
	var errors validation.List

	errors.Enum("option", f.errorLabel, f.Option,
		validation.Selected())

	return errors
}
