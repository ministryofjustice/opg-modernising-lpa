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

func HowDoYouKnowYourCertificateProvider(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
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

				requireLength := false

				if lpa.CertificateProvider.Relationship != "legal-professional" && lpa.CertificateProvider.Relationship != "health-professional" {
					requireLength = true
				}

				if requireLength {
					lpa.Tasks.CertificateProvider = TaskInProgress
				} else {
					lpa.CertificateProvider.RelationshipLength = ""
					lpa.Tasks.CertificateProvider = TaskCompleted
				}

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				if requireLength {
					appData.Lang.Redirect(w, r, howLongHaveYouKnownCertificateProviderPath, http.StatusFound)
				} else {
					appData.Lang.Redirect(w, r, checkYourLpaPath, http.StatusFound)
				}
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type howDoYouKnowYourCertificateProviderForm struct {
	Description string
	How         string
}

func readHowDoYouKnowYourCertificateProviderForm(r *http.Request) *howDoYouKnowYourCertificateProviderForm {
	r.ParseForm()

	return &howDoYouKnowYourCertificateProviderForm{
		Description: postFormString(r, "description"),
		How:         postFormString(r, "how"),
	}
}

func (f *howDoYouKnowYourCertificateProviderForm) Validate() map[string]string {
	errors := map[string]string{}

	if f.How == "" {
		errors["how"] = "selectHowYouKnowCertificateProvider"
	}

	if f.How == "other" && f.Description == "" {
		errors["description"] = "enterDescription"
	}

	if f.How != "friend" && f.How != "neighbour" && f.How != "colleague" && f.How != "health-professional" && f.How != "legal-professional" && f.How != "other" {
		errors["how"] = "selectHowYouKnowCertificateProvider"
	}

	return errors
}
