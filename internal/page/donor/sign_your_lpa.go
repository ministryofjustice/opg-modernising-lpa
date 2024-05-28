package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type signYourLpaData struct {
	App                  page.AppData
	Errors               validation.List
	Donor                *actor.DonorProvidedDetails
	Form                 *signYourLpaForm
	WantToSignFormValue  string
	WantToApplyFormValue string
}

const (
	WantToSignLpa     = "want-to-sign"
	WantToApplyForLpa = "want-to-apply"
)

func SignYourLpa(tmpl template.Template, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &signYourLpaData{
			App:   appData,
			Donor: donor,
			Form: &signYourLpaForm{
				WantToApply: donor.WantToApplyForLpa,
				WantToSign:  donor.WantToSignLpa,
			},
			WantToSignFormValue:  WantToSignLpa,
			WantToApplyFormValue: WantToApplyForLpa,
		}

		if r.Method == http.MethodPost {
			data.Form = readSignYourLpaForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.WantToApplyForLpa = data.Form.WantToApply
				donor.WantToSignLpa = data.Form.WantToSign
				donor.SignedAt = now()

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.WitnessingYourSignature.Redirect(w, r, appData, donor)
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
		errors.Add("sign-lpa", validation.SelectError{Label: "bothBoxesToSignAndApply"})
	}

	return errors
}
