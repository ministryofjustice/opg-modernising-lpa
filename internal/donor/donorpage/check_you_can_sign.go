package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

type checkYouCanSignData struct {
	App         appcontext.Data
	Form        *newforms.YesNoForm
	CanTaskList bool
}

func CheckYouCanSign(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &checkYouCanSignData{
			App:         appData,
			Form:        newforms.NewYesNoForm(appData.Localizer.T("yesIfYouWillBeAbleToSignYourself")),
			CanTaskList: !provided.Type.Empty(),
		}

		data.Form.YesNo.SetInput(provided.Donor.CanSign)

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				provided.Donor.CanSign = data.Form.YesNo.Value

				if provided.Donor.CanSign.IsYes() {
					provided.AuthorisedSignatory = donordata.AuthorisedSignatory{}
					provided.IndependentWitness = donordata.IndependentWitness{}
					provided.Tasks.ChooseYourSignatory = task.StateNotStarted
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				redirect := donor.PathYourPreferredLanguage
				if provided.Donor.CanSign.IsNo() {
					redirect = donor.PathNeedHelpSigningConfirmation
				}

				return redirect.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
