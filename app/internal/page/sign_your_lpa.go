package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type signYourLpaData struct {
	App                  AppData
	Errors               validation.List
	Lpa                  *Lpa
	Form                 *signYourLpaForm
	WantToSignFormValue  string
	WantToApplyFormValue string
}

const (
	WantToSignLpa     = "want-to-sign"
	WantToApplyForLpa = "want-to-apply"
)

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
				WantToSign:  lpa.WantToSignLpa,
			},
			WantToSignFormValue:  WantToSignLpa,
			WantToApplyFormValue: WantToApplyForLpa,
		}

		if r.Method == http.MethodPost {
			r.ParseForm()

			data.Form = readSignYourLpaForm(r)
			data.Errors = data.Form.Validate()

			lpa.WantToApplyForLpa = data.Form.WantToApply
			lpa.WantToSignLpa = data.Form.WantToApply
			if err = lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			if data.Errors.None() {
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

type signYourLpaForm struct {
	WantToApply bool
	WantToSign  bool
}

func readSignYourLpaForm(r *http.Request) *signYourLpaForm {
	r.ParseForm()

	f := &signYourLpaForm{}

	for _, checkBox := range r.PostForm["sign-lpa"] {
		if checkBox == WantToSignLpa {
			f.WantToSign = true
		}

		if checkBox == WantToApplyForLpa {
			f.WantToApply = true
		}
	}

	return f
}

func (f *signYourLpaForm) Validate() validation.List {
	var errors validation.List

	errors.Bool("sign-lpa", "bothBoxesToSign", f.WantToApply && f.WantToSign,
		validation.Selected())

	return errors
}
