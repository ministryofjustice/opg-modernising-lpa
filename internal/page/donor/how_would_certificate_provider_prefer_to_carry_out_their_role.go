package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
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

func HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpl template.Template, donorStore DonorStore, notifyClient NotifyClient, shareCodeSender ShareCodeSender) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &howWouldCertificateProviderPreferToCarryOutTheirRoleData{
			App:                 appData,
			CertificateProvider: donor.CertificateProvider,
			Form: &howWouldCertificateProviderPreferToCarryOutTheirRoleForm{
				CarryOutBy: donor.CertificateProvider.CarryOutBy,
				Email:      donor.CertificateProvider.Email,
			},
			Options: actor.CertificateProviderCarryOutByValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readHowWouldCertificateProviderPreferToCarryOutTheirRole(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				shouldSendNotifications := donor.CertificateProvider.Email != data.Form.Email && data.Form.CarryOutBy.IsOnline() &&
					donor.Tasks.CheckYourLpa.Completed() && !donor.Tasks.ConfirmYourIdentityAndSign.Completed()
				donor.CertificateProvider.CarryOutBy = data.Form.CarryOutBy
				donor.CertificateProvider.Email = data.Form.Email

				if shouldSendNotifications {
					if err := shareCodeSender.SendCertificateProvider(r.Context(), notify.CertificateProviderInviteEmail, appData, true, donor); err != nil {
						return err
					}

					if _, err := notifyClient.Sms(r.Context(), notify.Sms{
						PhoneNumber: donor.CertificateProvider.Mobile,
						TemplateID:  notifyClient.TemplateID(notify.CertificateProviderDigitalLpaDetailsChangedNotSeenLpaSMS),
						Personalisation: map[string]string{
							"donorFullName": donor.Donor.FullName(),
							"lpaType":       appData.Localizer.T(donor.Type.LegalTermTransKey()),
						},
					}); err != nil {
						return err
					}
				}

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
