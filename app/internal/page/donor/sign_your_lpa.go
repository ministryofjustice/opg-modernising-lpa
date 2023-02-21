package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type signYourLpaData struct {
	App                  page.AppData
	Errors               validation.List
	Lpa                  *page.Lpa
	Form                 *signYourLpaForm
	WantToSignFormValue  string
	WantToApplyFormValue string
}

const (
	WantToSignLpa     = "want-to-sign"
	WantToApplyForLpa = "want-to-apply"
)

func SignYourLpa(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
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
			lpa.WantToSignLpa = data.Form.WantToSign
			if err = lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			if data.Errors.None() {
				lpa.Tasks.ConfirmYourIdentityAndSign = page.TaskCompleted
				if err = lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}
				return appData.Redirect(w, r, lpa, page.Paths.WitnessingYourSignature)
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

	form := &signYourLpaForm{}

	for _, checkBox := range r.PostForm["sign-lpa"] {
		if checkBox == WantToSignLpa {
			form.WantToSign = true
		}

		if checkBox == WantToApplyForLpa {
			form.WantToApply = true
		}
	}

	return form
}

func (f *signYourLpaForm) Validate() validation.List {
	var errors validation.List

	if !(f.WantToApply && f.WantToSign) {
		errors.Add("sign-lpa", validation.CustomError{Label: "bothBoxesToSignAndApply"})
	}

	return errors
}
