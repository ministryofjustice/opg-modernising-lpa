package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
)

type certificateProviderDetailsData struct {
	App  appcontext.Data
	Form *certificateProviderDetailsForm
}

func CertificateProviderDetails(tmpl template.Template, service CertificateProviderService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &certificateProviderDetailsData{
			App:  appData,
			Form: newCertificateProviderDetailsForm(appData.Localizer),
		}

		data.Form.FirstNames.SetInput(provided.CertificateProvider.FirstNames)
		data.Form.LastName.SetInput(provided.CertificateProvider.LastName)
		data.Form.HasNonUKMobile.SetInput(provided.CertificateProvider.HasNonUKMobile)

		if provided.CertificateProvider.HasNonUKMobile {
			data.Form.NonUKMobile.SetInput(provided.CertificateProvider.Mobile)
		} else {
			data.Form.Mobile.SetInput(provided.CertificateProvider.Mobile)
		}

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				nameHasChanged := provided.CertificateProvider.NameHasChanged(data.Form.FirstNames.Value, data.Form.LastName.Value)

				provided.CertificateProvider.FirstNames = data.Form.FirstNames.Value
				provided.CertificateProvider.LastName = data.Form.LastName.Value
				provided.CertificateProvider.HasNonUKMobile = data.Form.HasNonUKMobile.Value

				if data.Form.HasNonUKMobile.Value {
					provided.CertificateProvider.Mobile = data.Form.NonUKMobile.Value
				} else {
					provided.CertificateProvider.Mobile = data.Form.Mobile.Value
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
	newforms.Form
	FirstNames     *newforms.String
	LastName       *newforms.String
	HasNonUKMobile *newforms.Bool
	Mobile         *newforms.String
	NonUKMobile    *newforms.String
}

func newCertificateProviderDetailsForm(l Localizer) *certificateProviderDetailsForm {
	return &certificateProviderDetailsForm{
		FirstNames: newforms.NewString("first-names", l.T("firstNames")).
			NotEmpty(),
		LastName: newforms.NewString("last-name", l.T("lastName")).
			NotEmpty(),
		HasNonUKMobile: newforms.NewBool("has-non-uk-mobile", l.T("theyDoNotHaveAUkMobileNumber")),
		Mobile: newforms.NewString("mobile", l.T("ukMobileNumber")).
			NotEmpty().
			Mobile(),
		NonUKMobile: newforms.NewString("non-uk-mobile", l.T("mobilePhoneNumber")).
			NotEmpty().
			NonUKMobile(),
	}
}

func (f *certificateProviderDetailsForm) Parse(r *http.Request) bool {
	ok := f.ParsePostForm(r,
		f.FirstNames,
		f.LastName,
		f.HasNonUKMobile,
	)

	if f.HasNonUKMobile.Value {
		ok = f.ParsePostForm(r, f.NonUKMobile) && ok
	} else {
		ok = f.ParsePostForm(r, f.Mobile) && ok
	}

	return ok
}
