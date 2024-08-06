package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howWouldCertificateProviderPreferToCarryOutTheirRoleData struct {
	App                 appcontext.Data
	Errors              validation.List
	CertificateProvider donordata.CertificateProvider
	Form                *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
	Options             lpadata.ChannelOptions
}

func HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:                 appData,
			CertificateProvider: donor.CertificateProvider,
			Form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: donor.CertificateProvider.CarryOutBy,
				Email:      donor.CertificateProvider.Email,
			},
			Options: lpadata.ChannelValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowWouldCertificateProviderPreferToCarryOutTheirRole(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.CertificateProvider.CarryOutBy = data.Form.CarryOutBy
				donor.CertificateProvider.Email = data.Form.Email

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.CertificateProviderAddress.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type howWouldCertificateProviderPreferToCarryOutTheirRoleForm struct {
	CarryOutBy lpadata.Channel
	Email      string
	Error      error
}

func readHowWouldCertificateProviderPreferToCarryOutTheirRole(r *http.Request) *howWouldCertificateProviderPreferToCarryOutTheirRoleForm {
	channel, err := lpadata.ParseChannel(page.PostFormString(r, "carry-out-by"))

	email := page.PostFormString(r, "email")
	if channel.IsPaper() {
		email = ""
	}

	return &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
		CarryOutBy: channel,
		Email:      email,
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
