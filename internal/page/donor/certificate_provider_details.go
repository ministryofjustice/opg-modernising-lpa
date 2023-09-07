package donor

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type certificateProviderDetailsData struct {
	App                 page.AppData
	Errors              validation.List
	Form                *certificateProviderDetailsForm
	NameWarning         *actor.SameNameWarning
	SameLastnameAsDonor bool
}

func CertificateProviderDetails(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &certificateProviderDetailsData{
			App: appData,
			Form: &certificateProviderDetailsForm{
				FirstNames:     lpa.CertificateProvider.FirstNames,
				LastName:       lpa.CertificateProvider.LastName,
				HasNonUKMobile: lpa.CertificateProvider.HasNonUKMobile,
			},
		}

		if lpa.CertificateProvider.HasNonUKMobile {
			data.Form.NonUKMobile = lpa.CertificateProvider.Mobile
		} else {
			data.Form.Mobile = lpa.CertificateProvider.Mobile
		}

		if r.Method == http.MethodPost {
			data.Form = readCertificateProviderDetailsForm(r)
			data.Errors = data.Form.Validate()

			sameNameWarning := actor.NewSameNameWarning(
				actor.TypeCertificateProvider,
				certificateProviderMatches(lpa, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreNameWarning != sameNameWarning.String() {
				data.NameWarning = sameNameWarning
			}

			if lpa.Donor.LastName == data.Form.LastName && !data.Form.IgnoreSimilarNameWarning && sameNameWarning == nil {
				data.SameLastnameAsDonor = true
			}

			if data.Errors.None() && data.NameWarning == nil && !data.SameLastnameAsDonor {
				lpa.CertificateProvider.FirstNames = data.Form.FirstNames
				lpa.CertificateProvider.LastName = data.Form.LastName
				lpa.CertificateProvider.HasNonUKMobile = data.Form.HasNonUKMobile
				if data.Form.HasNonUKMobile {
					lpa.CertificateProvider.Mobile = data.Form.NonUKMobile
				} else {
					lpa.CertificateProvider.Mobile = data.Form.Mobile
				}

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole.Format(lpa.ID))
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

func certificateProviderMatches(lpa *page.Lpa, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	if strings.EqualFold(lpa.Donor.FirstNames, firstNames) && strings.EqualFold(lpa.Donor.LastName, lastName) {
		return actor.TypeDonor
	}

	for _, attorney := range lpa.Attorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeAttorney
		}
	}

	for _, attorney := range lpa.ReplacementAttorneys.Attorneys {
		if strings.EqualFold(attorney.FirstNames, firstNames) && strings.EqualFold(attorney.LastName, lastName) {
			return actor.TypeReplacementAttorney
		}
	}

	if strings.EqualFold(lpa.Signatory.FirstNames, firstNames) && strings.EqualFold(lpa.Signatory.LastName, lastName) {
		return actor.TypeSignatory
	}

	if strings.EqualFold(lpa.IndependentWitness.FirstNames, firstNames) && strings.EqualFold(lpa.IndependentWitness.LastName, lastName) {
		return actor.TypeIndependentWitness
	}

	return actor.TypeNone
}
