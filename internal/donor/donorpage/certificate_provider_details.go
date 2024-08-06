package donorpage

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type certificateProviderDetailsData struct {
	App         appcontext.Data
	Errors      validation.List
	Form        *certificateProviderDetailsForm
	NameWarning *actor.SameNameWarning
}

func CertificateProviderDetails(tmpl template.Template, donorStore DonorStore, newUID func() actoruid.UID) Handler {
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

			sameNameWarning := actor.NewSameNameWarning(
				actor.TypeCertificateProvider,
				certificateProviderMatches(provided, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreNameWarning != sameNameWarning.String() {
				data.NameWarning = sameNameWarning
			}

			if data.Errors.None() && data.NameWarning == nil {
				if provided.CertificateProvider.UID.IsZero() {
					provided.CertificateProvider.UID = newUID()
				}

				provided.CertificateProvider.FirstNames = data.Form.FirstNames
				provided.CertificateProvider.LastName = data.Form.LastName
				provided.CertificateProvider.HasNonUKMobile = data.Form.HasNonUKMobile

				if data.Form.HasNonUKMobile {
					provided.CertificateProvider.Mobile = data.Form.NonUKMobile
				} else {
					provided.CertificateProvider.Mobile = data.Form.Mobile
				}

				if !provided.Tasks.CertificateProvider.Completed() {
					provided.Tasks.CertificateProvider = task.StateInProgress
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathHowDoYouKnowYourCertificateProvider.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type certificateProviderDetailsForm struct {
	FirstNames               string
	LastName                 string
	Mobile                   string
	HasNonUKMobile           bool
	NonUKMobile              string
	IgnoreNameWarning        string
	IgnoreSimilarNameWarning bool
}

func readCertificateProviderDetailsForm(r *http.Request) *certificateProviderDetailsForm {
	return &certificateProviderDetailsForm{
		FirstNames:               page.PostFormString(r, "first-names"),
		LastName:                 page.PostFormString(r, "last-name"),
		Mobile:                   page.PostFormString(r, "mobile"),
		HasNonUKMobile:           page.PostFormString(r, "has-non-uk-mobile") == "1",
		NonUKMobile:              page.PostFormString(r, "non-uk-mobile"),
		IgnoreNameWarning:        page.PostFormString(r, "ignore-name-warning"),
		IgnoreSimilarNameWarning: page.PostFormString(r, "ignore-similar-name-warning") == "yes",
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

func certificateProviderMatches(donor *donordata.Provided, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	if strings.EqualFold(donor.Donor.FirstNames, firstNames) && strings.EqualFold(donor.Donor.LastName, lastName) {
		return actor.TypeDonor
	}

	for _, attorney := range donor.Attorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeAttorney
		}
	}

	for _, attorney := range donor.ReplacementAttorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeReplacementAttorney
		}
	}

	if strings.EqualFold(donor.AuthorisedSignatory.FirstNames, firstNames) && strings.EqualFold(donor.AuthorisedSignatory.LastName, lastName) {
		return actor.TypeAuthorisedSignatory
	}

	if strings.EqualFold(donor.IndependentWitness.FirstNames, firstNames) && strings.EqualFold(donor.IndependentWitness.LastName, lastName) {
		return actor.TypeIndependentWitness
	}

	return actor.TypeNone
}
