package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type checkYourNameData struct {
	App    page.AppData
	Form   *checkYourNameForm
	Errors validation.List
	Lpa    *page.Lpa
}

func CheckYourName(tmpl template.Template, lpaStore LpaStore, notifyClient NotifyClient) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())

		if err != nil {
			return err
		}

		data := checkYourNameData{
			App:  appData,
			Form: &checkYourNameForm{},
			Lpa:  lpa,
		}

		if r.Method == http.MethodPost {
			data.Form = readCheckYourNameForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				if data.Form.CorrectedName != "" {
					lpa.CertificateProvider.DeclaredFullName = data.Form.CorrectedName

					if err := lpaStore.Put(r.Context(), lpa); err != nil {
						return err
					}

					_, err := notifyClient.Email(r.Context(), notify.Email{
						EmailAddress:    lpa.Donor.Email,
						TemplateID:      notifyClient.TemplateID(notify.CertificateProviderNameChangeEmail),
						Personalisation: map[string]string{"declaredName": lpa.CertificateProvider.DeclaredFullName},
					})

					if err != nil {
						return err
					}
				}

				appData.Redirect(w, r, lpa, page.Paths.CertificateProviderEnterDateOfBirth)
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

	errors.String("is-name-correct", "yesIfTheNameIsCorrect", f.IsNameCorrect,
		validation.Select("yes", "no"))

	if f.IsNameCorrect == "no" && f.CorrectedName == "" {
		errors.String("corrected-name", "yourFullName", f.CorrectedName,
			validation.Empty(),
		)
	}

	return errors
}
