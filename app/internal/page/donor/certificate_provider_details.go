package donor

import (
	"net/http"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
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

func CertificateProviderDetails(tmpl template.Template, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &certificateProviderDetailsData{
			App: appData,
			Form: &certificateProviderDetailsForm{
				FirstNames: lpa.CertificateProvider.FirstNames,
				LastName:   lpa.CertificateProvider.LastName,
				Mobile:     lpa.CertificateProvider.Mobile,
				Dob:        lpa.CertificateProvider.DateOfBirth,
			},
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
				lpa.CertificateProvider.DateOfBirth = data.Form.Dob
				lpa.CertificateProvider.Mobile = data.Form.Mobile

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, page.Paths.HowWouldCertificateProviderPreferToCarryOutTheirRole)
			}
		}

		return tmpl(w, data)
	}
}

type certificateProviderDetailsForm struct {
	FirstNames               string
	LastName                 string
	Dob                      date.Date
	Mobile                   string
	IgnoreNameWarning        string
	IgnoreSimilarNameWarning bool
}

func readCertificateProviderDetailsForm(r *http.Request) *certificateProviderDetailsForm {
	return &certificateProviderDetailsForm{
		FirstNames:               page.PostFormString(r, "first-names"),
		LastName:                 page.PostFormString(r, "last-name"),
		Dob:                      date.New(page.PostFormString(r, "date-of-birth-year"), page.PostFormString(r, "date-of-birth-month"), page.PostFormString(r, "date-of-birth-day")),
		Mobile:                   page.PostFormString(r, "mobile"),
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

	errors.Date("date-of-birth", "dateOfBirth", d.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	errors.String("mobile", "mobile", strings.ReplaceAll(d.Mobile, " ", ""),
		validation.Empty(),
		validation.Mobile())

	return errors
}

func certificateProviderMatches(lpa *page.Lpa, firstNames, lastName string) actor.Type {
	if lpa.Donor.FirstNames == firstNames && lpa.Donor.LastName == lastName {
		return actor.TypeDonor
	}

	for _, attorney := range lpa.Attorneys {
		if attorney.FirstNames == firstNames && attorney.LastName == lastName {
			return actor.TypeAttorney
		}
	}

	for _, attorney := range lpa.ReplacementAttorneys {
		if attorney.FirstNames == firstNames && attorney.LastName == lastName {
			return actor.TypeReplacementAttorney
		}
	}

	return actor.TypeNone
}
