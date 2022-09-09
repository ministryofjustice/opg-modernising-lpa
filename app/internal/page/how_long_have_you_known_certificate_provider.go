package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type howLongHaveYouKnownCertificateProviderData struct {
	App                 AppData
	Errors              map[string]string
	CertificateProvider CertificateProvider
	HowLong             string
}

func HowLongHaveYouKnownCertificateProvider(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		data := &howLongHaveYouKnownCertificateProviderData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
			HowLong:             lpa.CertificateProvider.RelationshipLength,
		}

		if r.Method == http.MethodPost {
			form := readHowLongHaveYouKnownCertificateProviderForm(r)
			data.Errors = form.Validate()

			if len(data.Errors) == 0 {
				lpa.Tasks.CertificateProvider = TaskCompleted
				lpa.CertificateProvider.RelationshipLength = form.HowLong
				if err := dataStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}
				appData.Lang.Redirect(w, r, taskListPath, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type howLongHaveYouKnownCertificateProviderForm struct {
	HowLong string
}

func readHowLongHaveYouKnownCertificateProviderForm(r *http.Request) *howLongHaveYouKnownCertificateProviderForm {
	return &howLongHaveYouKnownCertificateProviderForm{
		HowLong: postFormString(r, "how-long"),
	}
}

func (f *howLongHaveYouKnownCertificateProviderForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.HowLong != "gte-2-years" && f.HowLong != "lt-2-years" {
		errors["how-long"] = "selectHowLongHaveYouKnownCertificateProvider"
	}
	if f.HowLong == "lt-2-years" {
		errors["how-long"] = "mustHaveKnownCertificateProviderTwoYears"
	}

	return errors
}
