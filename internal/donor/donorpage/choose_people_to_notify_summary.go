package donorpage

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type choosePeopleToNotifySummaryData struct {
	App       appcontext.Data
	Errors    validation.List
	Form      *donordata.YesNoMaybeForm
	Donor     *donordata.Provided
	Options   donordata.YesNoMaybeOptions
	CanChoose bool
}

func ChoosePeopleToNotifySummary(tmpl template.Template, reuseStore ReuseStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if len(provided.PeopleToNotify) == 0 {
			return donor.PathDoYouWantToNotifyPeople.Redirect(w, r, appData, provided)
		}

		peopleToNotify, err := reuseStore.PeopleToNotify(r.Context(), provided)
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return err
		}

		data := &choosePeopleToNotifySummaryData{
			App:       appData,
			Donor:     provided,
			Options:   donordata.YesNoMaybeValues,
			CanChoose: len(peopleToNotify) > 0,
		}

		if r.Method == http.MethodPost {
			data.Form = donordata.ReadYesNoMaybeForm(r, "yesToAddAnotherPersonToNotify")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				switch data.Form.Option {
				case donordata.Yes:
					return donor.PathEnterPersonToNotify.RedirectQuery(w, r, appData, provided, url.Values{"addAnother": {"1"}})

				case donordata.Maybe:
					return donor.PathChoosePeopleToNotify.RedirectQuery(w, r, appData, provided, url.Values{"addAnother": {"1"}})

				case donordata.No:
					return donor.PathTaskList.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
