package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
)

type howDoYouKnowYourCertificateProviderData struct {
	App                 AppData
	Errors              map[string]string
	CertificateProvider CertificateProvider
	Form                *howDoYouKnowYourCertificateProviderForm
}

func HowDoYouKnowYourCertificateProvider(tmpl template.Template, dataStore DataStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var lpa Lpa
		if err := dataStore.Get(r.Context(), appData.SessionID, &lpa); err != nil {
			return err
		}

		data := &howDoYouKnowYourCertificateProviderData{
			App:                 appData,
			CertificateProvider: lpa.CertificateProvider,
			Form: &howDoYouKnowYourCertificateProviderForm{
				Description: lpa.CertificateProvider.RelationshipDescription,
				How:         lpa.CertificateProvider.Relationship,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readHowDoYouKnowYourCertificateProviderForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				lpa.CertificateProvider.Relationship = data.Form.How
				lpa.CertificateProvider.RelationshipDescription = data.Form.Description
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

type howDoYouKnowYourCertificateProviderForm struct {
	Description string
	How         []string
}

func readHowDoYouKnowYourCertificateProviderForm(r *http.Request) *howDoYouKnowYourCertificateProviderForm {
	r.ParseForm()

	return &howDoYouKnowYourCertificateProviderForm{
		Description: postFormString(r, "description"),
		How:         r.PostForm["how"],
	}
}

func (f *howDoYouKnowYourCertificateProviderForm) Validate() map[string]string {
	errors := map[string]string{}

	if len(f.How) == 0 {
		errors["how"] = "selectHowYouKnowCertificateProvider"
	}

	for _, value := range f.How {
		if value == "other" && f.Description == "" {
			errors["description"] = "enterDescription"
		}
		if value != "friend" && value != "neighbour" && value != "colleague" && value != "health-professional" && value != "legal-professional" && value != "other" {
			errors["how"] = "selectHowYouKnowCertificateProvider"
			break
		}
	}

	return errors
}
