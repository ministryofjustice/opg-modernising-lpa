package donorpage

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type choosePeopleToNotifySummaryData struct {
	App       appcontext.Data
	Errors    validation.List
	Form      *donordata.YesNoMaybeForm
	Donor     *donordata.Provided
	CanChoose bool
}

func ChoosePeopleToNotifySummary(tmpl template.Template, service PeopleToNotifyService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if len(provided.PeopleToNotify) == 0 {
			return donor.PathDoYouWantToNotifyPeople.Redirect(w, r, appData, provided)
		}

		peopleToNotify, err := service.Reusable(r.Context(), provided)
		if err != nil {
			return fmt.Errorf("retrieving reusable people to notify: %w", err)
		}

		data := &choosePeopleToNotifySummaryData{
			App:       appData,
			Donor:     provided,
			Form:      donordata.NewYesNoMaybeForm(appData.Localizer.T("yesToAddAnotherPersonToNotify")),
			CanChoose: len(peopleToNotify) > 0,
		}

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				switch data.Form.Enum.Value {
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
