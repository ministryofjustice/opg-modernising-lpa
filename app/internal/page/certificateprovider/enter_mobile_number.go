package certificateprovider

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

func EnterMobileNumber(tmpl template.Template, lpaStore LpaStore, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &mobileNumberData{
			App: appData,
			Lpa: lpa,
			Form: &mobileNumberForm{
				Mobile: certificateProvider.Mobile,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readMobileNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				certificateProvider.Mobile = data.Form.Mobile

				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderYourAddress)
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
		validation.Empty(),
		validation.Mobile())

	return errors
}
