package page

import (
	"fmt"
	"net/http"
	"net/mail"

	"github.com/ministryofjustice/opg-go-common/template"
)

type howWouldCertificateProviderPreferToCarryOutTheirRoleData struct {
	App                 AppData
	Errors              map[string]string
	CertificateProvider CertificateProvider
	Form                *howWouldCertificateProviderPreferToCarryOutTheirRoleForm
}

func HowWouldCertificateProviderPreferToCarryOutTheirRole(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		if len(lpa.PeopleToNotify) > 0 {
			return appData.Lang.Redirect(w, r, lpa, Paths.ChoosePeopleToNotifySummary)
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

			if len(data.Errors) == 0 {
				lpa.CertificateProvider.CarryOutBy = data.Form.CarryOutBy
				lpa.CertificateProvider.Email = data.Form.Email

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				if lpa.CertificateProvider.CarryOutBy == "paper" {
					return appData.Lang.Redirect(w, r, lpa, Paths.CertificateProviderAddress)
				} else {
					return appData.Lang.Redirect(w, r, lpa, Paths.HowDoYouKnowYourCertificateProvider)
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

func (f *howWouldCertificateProviderPreferToCarryOutTheirRoleForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.CarryOutBy != "email" && f.CarryOutBy != "paper" {
		errors["carry-out-by"] = "selectHowWouldCertificateProviderPreferToCarryOutTheirRole"
	}

	if f.CarryOutBy == "email" {
		if f.Email == "" {
			errors["email"] = "enterEmail"
		} else if _, err := mail.ParseAddress(fmt.Sprintf("<%s>", f.Email)); err != nil {
			errors["email"] = "emailIncorrectFormat"
		}
	}

	return errors
}
