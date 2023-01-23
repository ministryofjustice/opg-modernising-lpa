package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type signYourLpaData struct {
	App                  AppData
	Errors               map[string]string
	Lpa                  *Lpa
	Form                 *signYourLpaForm
	CPWitnessedFormValue string
	WantFormValue        string
}

const (
	CertificateProviderHasWitnessed = "cp-witnessed"
	WantToApplyForLpa               = "want-to-apply"
)

type signYourLpaForm struct {
	WantToApply bool
	CPWitnessed bool
}

func SignYourLpa(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &signYourLpaData{
			App: appData,
			Lpa: lpa,
			Form: &signYourLpaForm{
				WantToApply: lpa.WantToApplyForLpa,
				CPWitnessed: lpa.CPWitnessedDonorSign,
			},
			CPWitnessedFormValue: CertificateProviderHasWitnessed,
			WantFormValue:        WantToApplyForLpa,
		}

		if r.Method == http.MethodPost {
			r.ParseForm()

			data.Form = readSignYourLpaForm(r)
			data.Errors = data.Form.Validate()

			lpa.WantToApplyForLpa = data.Form.WantToApply
			lpa.CPWitnessedDonorSign = data.Form.CPWitnessed
			if err = lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			if len(data.Errors) == 0 {
				lpa.Tasks.ConfirmYourIdentityAndSign = TaskCompleted
				if err = lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}
				return appData.Redirect(w, r, lpa, Paths.WitnessingYourSignature)
			}
		}

		return tmpl(w, data)
	}
}

func readSignYourLpaForm(r *http.Request) *signYourLpaForm {
	r.ParseForm()

	f := &signYourLpaForm{}

	for _, checkBox := range r.PostForm["sign-lpa"] {
		if checkBox == CertificateProviderHasWitnessed {
			f.CPWitnessed = true
		}

		if checkBox == WantToApplyForLpa {
			f.WantToApply = true
		}
	}

	return f
}

func (f *signYourLpaForm) Validate() map[string]string {
	errors := map[string]string{}

	if !f.WantToApply {
		errors["sign-lpa"] = "selectBothBoxes"
	}

	if !f.CPWitnessed {
		errors["sign-lpa"] = "selectBothBoxes"
	}

	return errors
}
