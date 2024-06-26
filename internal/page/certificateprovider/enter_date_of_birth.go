package certificateprovider

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type dateOfBirthData struct {
	App        page.AppData
	Lpa        *lpastore.Lpa
	Form       *dateOfBirthForm
	Errors     validation.List
	DobWarning string
}

type dateOfBirthForm struct {
	Dob              date.Date
	IgnoreDobWarning string
}

func EnterDateOfBirth(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, certificateProviderStore CertificateProviderStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &dateOfBirthData{
			App: appData,
			Lpa: lpa,
			Form: &dateOfBirthForm{
				Dob: certificateProvider.DateOfBirth,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readDateOfBirthForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			if data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.None() && data.DobWarning == "" {
				certificateProvider.DateOfBirth = data.Form.Dob
				if !certificateProvider.Tasks.ConfirmYourDetails.Completed() {
					certificateProvider.Tasks.ConfirmYourDetails = actor.TaskInProgress
				}

				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				if lpa.CertificateProvider.Relationship.IsProfessionally() {
					return page.Paths.CertificateProvider.WhatIsYourHomeAddress.Redirect(w, r, appData, certificateProvider.LpaID)
				}

				return page.Paths.CertificateProvider.YourPreferredLanguage.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

func readDateOfBirthForm(r *http.Request) *dateOfBirthForm {
	return &dateOfBirthForm{
		Dob:              date.New(page.PostFormString(r, "date-of-birth-year"), page.PostFormString(r, "date-of-birth-month"), page.PostFormString(r, "date-of-birth-day")),
		IgnoreDobWarning: page.PostFormString(r, "ignore-dob-warning"),
	}
}

func (f *dateOfBirthForm) DobWarning() string {
	var (
		hundredYearsEarlier = date.Today().AddDate(-100, 0, 0)
	)

	if !f.Dob.IsZero() {
		if f.Dob.Before(hundredYearsEarlier) {
			return "dateOfBirthIsOver100"
		}
	}

	return ""
}

func (f *dateOfBirthForm) Validate() validation.List {
	var errors validation.List

	errors.Date("date-of-birth", "dateOfBirth", f.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	if f.Dob.After(date.Today().AddDate(-18, 0, 0)) {
		errors.Add("date-of-birth", validation.CustomError{Label: "youAreUnder18Error"})
	}

	return errors
}
