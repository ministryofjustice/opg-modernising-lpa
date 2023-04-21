package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howWouldCertificateProviderPreferToCarryOutTheirRoleData struct {
	App                        page.AppData
	Errors                     validation.List
	CertificateProviderDetails page.CertificateProviderDetails
	Form                       *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
}

func HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:                        appData,
			CertificateProviderDetails: lpa.CertificateProviderDetails,
			Form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpa.CertificateProviderDetails.CarryOutBy,
				Email:      lpa.CertificateProviderDetails.Email,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readHowWouldCertificateProviderPreferToCarryOutTheirRole(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				lpa.CertificateProviderDetails.CarryOutBy = data.Form.CarryOutBy
				lpa.CertificateProviderDetails.Email = data.Form.Email

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				if lpa.CertificateProviderDetails.CarryOutBy == "paper" {
					return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderAddress)
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.HowDoYouKnowYourCertificateProvider)
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
