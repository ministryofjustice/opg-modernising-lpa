package page

import (
	"net/http"

	"golang.org/x/exp/slices"

	"github.com/ministryofjustice/opg-go-common/template"
)

type signYourLpaData struct {
	App    AppData
	Errors map[string]string
	Lpa    *Lpa
	Form   *signYourLpaForm
}

const (
	CertificateProviderHasWitnessed = "cp-witnessed"
	WantToApplyForLpa               = "want-to-apply"
)

type signYourLpaForm struct {
	DonorSignatures []string
}

func SignYourLpa(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		data := &signYourLpaData{
			App: appData,
			Lpa: lpa,
			Form: &signYourLpaForm{
				DonorSignatures: lpa.DonorSignatures,
			},
		}

		if r.Method == http.MethodPost {
			r.ParseForm()

			data.Form = readSignYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.DonorSignatures = data.Form.DonorSignatures
				lpa.Tasks.ConfirmYourIdentityAndSign = TaskCompleted

				if err = lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				return appData.Lang.Redirect(w, r, appData.Paths.WitnessingYourSignature, http.StatusFound)
			}
		}

		return tmpl(w, data)
	}
}

func readSignYourLpaForm(r *http.Request) *signYourLpaForm {
	r.ParseForm()

	form := &signYourLpaForm{
		DonorSignatures: []string{},
	}

	for _, checkBox := range r.PostForm["sign-lpa"] {
		if checkBox == CertificateProviderHasWitnessed {
			form.DonorSignatures = append(form.DonorSignatures, CertificateProviderHasWitnessed)
		}

		if checkBox == WantToApplyForLpa {
			form.DonorSignatures = append(form.DonorSignatures, WantToApplyForLpa)
		}
	}

	return form
}

func (f *signYourLpaForm) Validate() map[string]string {
	errors := map[string]string{}

	if !slices.Contains(f.DonorSignatures, CertificateProviderHasWitnessed) {
		errors["sign-lpa"] = "selectBothBoxes"
	}

	if !slices.Contains(f.DonorSignatures, WantToApplyForLpa) {
		errors["sign-lpa"] = "selectBothBoxes"
	}

	return errors
}
