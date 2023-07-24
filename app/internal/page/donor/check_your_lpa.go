package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type checkYourLpaData struct {
	App       page.AppData
	Errors    validation.List
	Lpa       *page.Lpa
	Form      *checkYourLpaForm
	Completed bool
}

func CheckYourLpa(tmpl template.Template, donorStore DonorStore, shareCodeSender ShareCodeSender) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &checkYourLpaData{
			App: appData,
			Lpa: lpa,
			Form: &checkYourLpaForm{
				CheckedAndHappy: lpa.CheckedAndHappy,
			},
			Completed: lpa.Tasks.CheckYourLpa.Completed(),
		}

		if r.Method == http.MethodPost {
			data.Form = readCheckYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.CheckedAndHappy = data.Form.CheckedAndHappy
				lpa.Tasks.CheckYourLpa = actor.TaskCompleted

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if err := shareCodeSender.SendCertificateProvider(r.Context(), notify.CertificateProviderInviteEmail, appData, true, lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.LpaDetailsSaved.Format(lpa.ID))
			}
		}

		return tmpl(w, data)
	}
}

type checkYourLpaForm struct {
	CheckedAndHappy bool
}

func readCheckYourLpaForm(r *http.Request) *checkYourLpaForm {
	return &checkYourLpaForm{
		CheckedAndHappy: page.PostFormString(r, "checked-and-happy") == "1",
	}
}

func (f *checkYourLpaForm) Validate() validation.List {
	var errors validation.List

	errors.Bool("checked-and-happy", "theBoxIfYouHaveCheckedAndHappyToShareLpa", f.CheckedAndHappy,
		validation.Selected())

	return errors
}
