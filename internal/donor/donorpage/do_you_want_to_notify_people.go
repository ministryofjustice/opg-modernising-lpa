package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type doYouWantToNotifyPeopleData struct {
	App             appcontext.Data
	Errors          validation.List
	Form            *form.YesNoForm
	Donor           *donordata.Provided
	HowWorkTogether string
}

func DoYouWantToNotifyPeople(tmpl template.Template, service PeopleToNotifyService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if len(provided.PeopleToNotify) > 0 {
			return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
		}

		data := &doYouWantToNotifyPeopleData{
			App:   appData,
			Donor: provided,
			Form:  form.NewYesNoForm(provided.DoYouWantToNotifyPeople),
		}

		switch provided.AttorneyDecisions.How {
		case lpadata.Jointly:
			data.HowWorkTogether = "jointlyDescription"
		case lpadata.JointlyAndSeverally:
			data.HowWorkTogether = "jointlyAndSeverallyDescription"
		case lpadata.JointlyForSomeSeverallyForOthers:
			data.HowWorkTogether = "jointlyForSomeSeverallyForOthersDescription"
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToNotifySomeoneAboutYourLpa")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				redirectPath := donor.PathChoosePeopleToNotify
				if data.Form.YesNo.IsNo() {
					redirectPath = donor.PathTaskList
				}

				if err := service.WantPeopleToNotify(r.Context(), provided, data.Form.YesNo); err != nil {
					return err
				}

				return redirectPath.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
