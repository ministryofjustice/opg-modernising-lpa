package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type howWouldCertificateProviderPreferToCarryOutTheirRoleData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
	Form                *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
}

func HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
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

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderAddress)
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
		CarryOutBy: page.PostFormString(r, "carry-out-by"),
		Email:      page.PostFormString(r, "email"),
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
