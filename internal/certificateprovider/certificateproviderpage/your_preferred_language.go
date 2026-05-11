package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourPreferredLanguageData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *newforms.LanguageForm
	Lpa    *lpadata.Lpa
}

func YourPreferredLanguage(tmpl template.Template, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		data := &yourPreferredLanguageData{
			App:  appData,
			Form: newforms.NewLanguageForm(appData.Localizer.T("whichLanguageYouWouldLikeUsToUseWhenWeContactYou")),
			Lpa:  lpa,
		}

		data.Form.Language.SetInput(certificateProvider.ContactLanguagePreference)

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				certificateProvider.ContactLanguagePreference = data.Form.Language.Value
				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				return certificateprovider.PathConfirmYourDetails.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
