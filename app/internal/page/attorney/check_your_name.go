package attorney

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type checkYourNameData struct {
	App      page.AppData
	Form     *checkYourNameForm
	Errors   validation.List
	Lpa      *page.Lpa
	Attorney actor.Attorney
}

func CheckYourName(tmpl template.Template, donorStore DonorStore, attorneyStore AttorneyStore, notifyClient NotifyClient) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, attorneyProvidedDetails *actor.AttorneyProvidedDetails) error {
		lpa, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		attorneys := lpa.Attorneys
		if appData.IsReplacementAttorney() {
			attorneys = lpa.ReplacementAttorneys
		}

		attorney, ok := attorneys.Get(appData.AttorneyID)
		if !ok {
			return appData.Redirect(w, r, lpa, page.Paths.Attorney.Start)
		}

		data := &checkYourNameData{
			App: appData,
			Form: &checkYourNameForm{
				IsNameCorrect: attorneyProvidedDetails.IsNameCorrect,
				CorrectedName: attorneyProvidedDetails.CorrectedName,
			},
			Lpa:      lpa,
			Attorney: attorney,
		}

		if r.Method == http.MethodPost {
			data.Form = readCheckYourNameForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				previousCorrectedName := attorneyProvidedDetails.CorrectedName
				attorneyProvidedDetails.IsNameCorrect = data.Form.IsNameCorrect
				attorneyProvidedDetails.CorrectedName = data.Form.CorrectedName

				if attorneyProvidedDetails.Tasks.ConfirmYourDetails == actor.TaskNotStarted {
					attorneyProvidedDetails.Tasks.ConfirmYourDetails = actor.TaskInProgress
				}

				if data.Form.CorrectedName != "" && data.Form.CorrectedName != previousCorrectedName {
					attorneyProvidedDetails.CorrectedName = data.Form.CorrectedName
					_, err := notifyClient.Email(r.Context(), notify.Email{
						EmailAddress:    lpa.Donor.Email,
						TemplateID:      notifyClient.TemplateID(notify.AttorneyNameChangeEmail),
						Personalisation: map[string]string{"declaredName": attorneyProvidedDetails.CorrectedName},
					})
					if err != nil {
						return err
					}
				}

				if err := attorneyStore.Put(r.Context(), attorneyProvidedDetails); err != nil {
					return err
				}

				appData.Redirect(w, r, lpa, page.Paths.Attorney.DateOfBirth)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type checkYourNameForm struct {
	IsNameCorrect string
	CorrectedName string
}

func readCheckYourNameForm(r *http.Request) *checkYourNameForm {

	return &checkYourNameForm{
		IsNameCorrect: page.PostFormString(r, "is-name-correct"),
		CorrectedName: page.PostFormString(r, "corrected-name"),
	}
}

func (f *checkYourNameForm) Validate() validation.List {
	errors := validation.List{}

	errors.String("is-name-correct", "confirmIfTheNameIsCorrect", f.IsNameCorrect,
		validation.Select("yes", "no").CustomError())

	if f.IsNameCorrect == "no" && f.CorrectedName == "" {
		errors.String("corrected-name", "yourFullName", f.CorrectedName,
			validation.Empty())
	}

	return errors
}
