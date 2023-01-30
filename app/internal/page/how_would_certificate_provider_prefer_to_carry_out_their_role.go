package page

import (
	"fmt"
	"net/http"
	"net/mail"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howWouldCertificateProviderPreferToCarryOutTheirRoleData struct {
	App                 AppData
	Errors              validation.List
	CertificateProvider CertificateProvider
	Form                *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
}

func HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
			Form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpa.CertificateProvider.CarryOutBy,
				Email:      lpa.CertificateProvider.Email,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readHowWouldCertificateProviderPreferToCarryOutTheirRole(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.CertificateProvider.CarryOutBy = data.Form.CarryOutBy
				lpa.CertificateProvider.Email = data.Form.Email

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if lpa.CertificateProvider.CarryOutBy == "paper" {
					return appData.Redirect(w, r, lpa, Paths.CertificateProviderAddress)
				} else {
					return appData.Redirect(w, r, lpa, Paths.HowDoYouKnowYourCertificateProvider)
				}
			}
		}

		return tmpl(w, data)
	}
}

type howWouldCertificateProviderPreferToCarryOutTheirRoleForm struct {
	CarryOutBy string
	Email      string
}

func readHowWouldCertificateProviderPreferToCarryOutTheirRole(r *http.Request) *howWouldCertificateProviderPreferToCarryOutTheirRoleForm {
	return &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
		CarryOutBy: postFormString(r, "carry-out-by"),
		Email:      postFormString(r, "email"),
	}
}

func (f *howWouldCertificateProviderPreferToCarryOutTheirRoleForm) Validate() validation.List {
	var errors validation.List

	if f.CarryOutBy != "email" && f.CarryOutBy != "paper" {
		errors.Add("carry-out-by", "selectHowWouldCertificateProviderPreferToCarryOutTheirRole")
	}

	if f.CarryOutBy == "email" {
		if f.Email == "" {
			errors.Add("email", "enterEmail")
		} else if _, err := mail.ParseAddress(fmt.Sprintf("<%s>", f.Email)); err != nil {
			errors.Add("email", "emailIncorrectFormat")
		}
	}

	return errors
}
