package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type addCorrespondentData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func AddCorrespondent(tmpl template.Template, service CorrespondentService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if provided.HasCorrespondent() {
			return donor.PathCorrespondentSummary.Redirect(w, r, appData, provided)
		}

		data := &addCorrespondentData{
			App:   appData,
			Donor: provided,
			Form:  form.NewYesNoForm(provided.AddCorrespondent),
		}

		if r.Method == http.MethodPost {
			f := form.ReadYesNoForm(r, "yesToAddCorrespondent")
			data.Errors = f.Validate()

			if data.Errors.None() {
				provided.AddCorrespondent = f.YesNo

				if provided.AddCorrespondent.IsNo() {
					if err := service.NotWanted(r.Context(), provided); err != nil {
						return err
					}

					if provided.SignedAt.IsZero() {
						return donor.PathTaskList.Redirect(w, r, appData, provided)
					} else {
						return donor.PathProgress.Redirect(w, r, appData, provided)
					}
				} else {
					if err := service.Put(r.Context(), provided); err != nil {
						return err
					}

					return donor.PathChooseCorrespondent.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
