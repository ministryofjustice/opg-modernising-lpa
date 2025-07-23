package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type certificateProviderDetailsData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *certificateProviderDetailsForm
}

func CertificateProviderDetails(tmpl template.Template, service CertificateProviderService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &certificateProviderDetailsData{
			App: appData,
			Form: &certificateProviderDetailsForm{
				FirstNames:     provided.CertificateProvider.FirstNames,
				LastName:       provided.CertificateProvider.LastName,
				HasNonUKMobile: provided.CertificateProvider.HasNonUKMobile,
			},
		}

		if provided.CertificateProvider.HasNonUKMobile {
			data.Form.NonUKMobile = provided.CertificateProvider.Mobile
		} else {
			data.Form.Mobile = provided.CertificateProvider.Mobile
		}

		if r.Method == http.MethodPost {
			data.Form = readCertificateProviderDetailsForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				nameHasChanged := provided.CertificateProvider.NameHasChanged(data.Form.FirstNames, data.Form.LastName)

				provided.CertificateProvider.FirstNames = data.Form.FirstNames
				provided.CertificateProvider.LastName = data.Form.LastName
				provided.CertificateProvider.HasNonUKMobile = data.Form.HasNonUKMobile

				if data.Form.HasNonUKMobile {
					provided.CertificateProvider.Mobile = data.Form.NonUKMobile
				} else {
					provided.CertificateProvider.Mobile = data.Form.Mobile
				}

				// Allow changing details for certificate provider on the page they
				// witness, without having to be notified.
				if !provided.SignedAt.IsZero() {
					provided.UpdateCheckedHash()
				}

				if err := service.Put(r.Context(), provided); err != nil {
					return err
				}

				actorType := certificateProviderMatches(provided, provided.CertificateProvider.FirstNames, provided.CertificateProvider.LastName)

				if nameHasChanged && !actorType.IsNone() {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"warningFrom": {appData.Page},
						"next":        {donor.PathHowDoYouKnowYourCertificateProvider.Format(provided.LpaID)},
						"actor":       {actor.TypeCertificateProvider.String()},
					})
				}

				return donor.PathHowDoYouKnowYourCertificateProvider.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type certificateProviderDetailsForm struct {
	FirstNames     string
	LastName       string
	Mobile         string
	HasNonUKMobile bool
	NonUKMobile    string
}

func readCertificateProviderDetailsForm(r *http.Request) *certificateProviderDetailsForm {
	return &certificateProviderDetailsForm{
		FirstNames:     page.PostFormString(r, "first-names"),
		LastName:       page.PostFormString(r, "last-name"),
		Mobile:         page.PostFormString(r, "mobile"),
		HasNonUKMobile: page.PostFormString(r, "has-non-uk-mobile") == "1",
		NonUKMobile:    page.PostFormString(r, "non-uk-mobile"),
	}
}

func (d *certificateProviderDetailsForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", d.FirstNames,
		validation.Empty())

	errors.String("last-name", "lastName", d.LastName,
		validation.Empty())

	if d.HasNonUKMobile {
		errors.String("non-uk-mobile", "yourCertificateProvidersMobileNumber", d.NonUKMobile,
			validation.Empty(),
			validation.NonUKMobile().ErrorLabel("enterAMobileNumberInTheCorrectFormat"))
	} else {
		errors.String("mobile", "yourCertificateProvidersUkMobileNumber", d.Mobile,
			validation.Empty(),
			validation.Mobile().ErrorLabel("enterAMobileNumberInTheCorrectFormat"))
	}

	return errors
}
