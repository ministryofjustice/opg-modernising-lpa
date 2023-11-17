package donor


import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howWouldCertificateProviderPreferToCarryOutTheirRoleData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider actor.CertificateProvider
	Form                *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
	Options             actor.CertificateProviderCarryOutByOptions
}

func HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.Lpa) error {
		data := &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
			Form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: lpa.CertificateProvider.CarryOutBy,
				Email:      lpa.CertificateProvider.Email,
			},
			Options: actor.CertificateProviderCarryOutByValues,
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

				return page.Paths.CertificateProviderAddress.Redirect(w, r, appData, lpa)
			}
		}

		return tmpl(w, data)
	}
}

type howWouldCertificateProviderPreferToCarryOutTheirRoleForm struct {
	CarryOutBy actor.CertificateProviderCarryOutBy
	Email      string
	Error      error
}

func readHowWouldCertificateProviderPreferToCarryOutTheirRole(r *http.Request) *howWouldCertificateProviderPreferToCarryOutTheirRoleForm {
	carryOutBy, err := actor.ParseCertificateProviderCarryOutBy(page.PostFormString(r, "carry-out-by"))

	return &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
		CarryOutBy: carryOutBy,
		Email:      page.PostFormString(r, "email"),
		Error:      err,
	}
}

func (f *howWouldCertificateProviderPreferToCarryOutTheirRoleForm) Validate() validation.List {
	var errors validation.List

	errors.Error("carry-out-by", "howYourCertificateProviderWouldPreferToCarryOutTheirRole", f.Error,
		validation.Selected())

	if f.CarryOutBy.IsOnline() {
		errors.String("email", "certificateProvidersEmail", f.Email,
			validation.Empty(),
			validation.Email())
	}

	return errors
}
