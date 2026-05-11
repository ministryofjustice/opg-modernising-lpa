package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type canYouSignYourLpaData struct {
	App         appcontext.Data
	Form        *newforms.EnumForm[donordata.YesNoMaybe, donordata.YesNoMaybeOptions, *donordata.YesNoMaybe]
	CanTaskList bool
}

func CanYouSignYourLpa(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &canYouSignYourLpaData{
			App:         appData,
			Form:        newforms.NewEnumForm[donordata.YesNoMaybe](appData.Localizer.T("yesIfCanSign"), donordata.YesNoMaybeValues),
			CanTaskList: !provided.Type.Empty(),
		}

		data.Form.Enum.SetInput(provided.Donor.ThinksCanSign)

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				provided.Donor.ThinksCanSign = data.Form.Enum.Value

				if provided.Donor.ThinksCanSign.IsYes() {
					provided.Donor.CanSign = form.Yes
					provided.AuthorisedSignatory = donordata.AuthorisedSignatory{}
					provided.IndependentWitness = donordata.IndependentWitness{}
					provided.Tasks.ChooseYourSignatory = task.StateNotStarted
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.Donor.ThinksCanSign.IsYes() {
					return donor.PathYourPreferredLanguage.Redirect(w, r, appData, provided)
				}

				return donor.PathCheckYouCanSign.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
