package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type lifeSustainingTreatmentData struct {
	App  appcontext.Data
	Form *newforms.EnumForm[lpadata.LifeSustainingTreatment, lpadata.LifeSustainingTreatmentOptions, *lpadata.LifeSustainingTreatment]
}

func LifeSustainingTreatment(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &lifeSustainingTreatmentData{
			App:  appData,
			Form: newforms.NewEnumForm[lpadata.LifeSustainingTreatment](appData.Localizer.T("ifTheDonorGivesConsentToLifeSustainingTreatment"), lpadata.LifeSustainingTreatmentValues),
		}

		data.Form.Enum.SetInput(provided.LifeSustainingTreatmentOption)

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			provided.LifeSustainingTreatmentOption = data.Form.Selected
			provided.Tasks.LifeSustainingTreatment = task.StateCompleted
			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			return donor.PathTaskList.Redirect(w, r, appData, provided)
		}

		return tmpl(w, data)
	}
}
