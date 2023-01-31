package page

import (
	"net/http"

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

	errors.String("carry-out-by", "howYourCertificateProviderWouldPreferToCarryOutTheirRole", f.CarryOutBy,
		validation.Select("email", "paper")) // selectHowWouldCertificateProviderPreferToCarryOutTheirRole

	if f.CarryOutBy == "email" {
		errors.String("email", "certificateProvidersEmail", f.Email,
			validation.Empty(),
			validation.Email())
	}

	return errors
}
