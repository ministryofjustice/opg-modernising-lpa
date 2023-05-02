package attorney

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type mobileNumberData struct {
	App    page.AppData
	Lpa    *page.Lpa
	Form   *mobileNumberForm
	Errors validation.List
}

type mobileNumberForm struct {
	Mobile string
}

func MobileNumber(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		attorneyProvidedDetails := getProvidedDetails(appData, lpa)

		data := &mobileNumberData{
			App: appData,
			Lpa: lpa,
			Form: &mobileNumberForm{
				Mobile: attorneyProvidedDetails.Mobile,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readMobileNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				attorneyProvidedDetails.Mobile = data.Form.Mobile
				setProvidedDetails(appData, lpa, attorneyProvidedDetails)

				tasks := getTasks(appData, lpa)
				tasks.ConfirmYourDetails = page.TaskCompleted
				setTasks(appData, lpa, tasks)

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.Attorney.Sign)
			}
		}

		return tmpl(w, data)
	}
}

func readMobileNumberForm(r *http.Request) *mobileNumberForm {
	return &mobileNumberForm{
		Mobile: page.PostFormString(r, "mobile"),
	}
}

func (f *mobileNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("mobile", "mobile", strings.ReplaceAll(f.Mobile, " ", ""),
		validation.Mobile())

	return errors
}
