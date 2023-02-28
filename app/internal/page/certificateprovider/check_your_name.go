package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type checkYourNameData struct {
	App    page.AppData
	Form   *checkYourNameForm
	Errors validation.List
	Lpa    *page.Lpa
}

func CheckYourName(tmpl template.Template, lpaStore LpaStore, dataStore page.DataStore) page.Handler {
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
				appData.Redirect(w, r, lpa, page.Paths.CertificateProviderDetails)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type checkYourNameForm struct {
	NameCorrect   bool
	CorrectedName string
}

func readCheckYourNameForm(r *http.Request) *checkYourNameForm {

	return &checkYourNameForm{
		NameCorrect:   page.PostFormString(r, "name-correct") == "1",
		CorrectedName: page.PostFormString(r, "corrected-name"),
	}
}

func (f *checkYourNameForm) Validate() validation.List {
	errors := validation.List{}

	return errors
}
