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

type removeCorrespondentData struct {
	App    appcontext.Data
	Name   string
	Errors validation.List
	Form   *form.YesNoForm
}

func RemoveCorrespondent(tmpl template.Template, service CorrespondentService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &removeCorrespondentData{
			App:  appData,
			Name: provided.Correspondent.FullName(),
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesToRemoveCorrespondent")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo == form.Yes {
					if err := service.Delete(r.Context(), provided); err != nil {
						return err
					}

					return donor.PathAddCorrespondent.Redirect(w, r, appData, provided)
				}

				return donor.PathCorrespondentSummary.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
