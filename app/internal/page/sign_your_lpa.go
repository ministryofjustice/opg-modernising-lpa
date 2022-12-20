package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type signYourLpaData struct {
	App    AppData
	Errors map[string]string
	Lpa    *Lpa
	Form   *signYourLpaForm
}

type signYourLpaForm struct {
	CPWitnessedSigning bool
	WantToApply        bool
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
				CPWitnessedSigning: lpa.DonorConfirmedCPWitnessedSigning,
				WantToApply:        lpa.WantToApplyForLpa,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readSignYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.DonorConfirmedCPWitnessedSigning = data.Form.CPWitnessedSigning
				lpa.WantToApplyForLpa = data.Form.WantToApply
				lpa.Tasks.ConfirmYourIdentityAndSign = TaskCompleted

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
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

	return &signYourLpaForm{
		CPWitnessedSigning: postFormString(r, "cp-witnessed") == "1",
		WantToApply:        postFormString(r, "want-to-apply") == "1",
	}
}

func (f *signYourLpaForm) Validate() map[string]string {
	errors := map[string]string{}

	if !f.CPWitnessedSigning {
		errors["cp-witnessed"] = "selectCPHasWitnessedSigning"
	}

	if !f.WantToApply {
		errors["want-to-apply"] = "selectWantToApply"
	}

	return errors
}
