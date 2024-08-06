package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourPreferredLanguageData struct {
	App       appcontext.Data
	Errors    validation.List
	Form      *form.LanguagePreferenceForm
	Options   localize.LangOptions
	FieldName string
	Lpa       *lpastore.Lpa
}

func YourPreferredLanguage(tmpl template.Template, certificateProviderStore CertificateProviderStore, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &yourPreferredLanguageData{
			App: appData,
			Form: &form.LanguagePreferenceForm{
				Preference: certificateProvider.ContactLanguagePreference,
			},
			Options:   localize.LangValues,
			FieldName: form.FieldNames.LanguagePreference,
			Lpa:       lpa,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadLanguagePreferenceForm(r, "whichLanguageYouWouldLikeUsToUseWhenWeContactYou")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				certificateProvider.ContactLanguagePreference = data.Form.Preference
				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				return certificateprovider.PathConfirmYourDetails.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
