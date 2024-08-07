package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourDateOfBirthData struct {
	App              appcontext.Data
	Errors           validation.List
	Form             *yourDateOfBirthForm
	DobWarning       string
	CanTaskList      bool
	MakingAnotherLPA bool
}

func YourDateOfBirth(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourDateOfBirthData{
			App: appData,
			Form: &yourDateOfBirthForm{
				Dob: provided.Donor.DateOfBirth,
			},
			CanTaskList:      !provided.Type.Empty(),
			MakingAnotherLPA: r.FormValue("makingAnotherLPA") == "1",
		}

		if r.Method == http.MethodPost {
			data.Form = readYourDateOfBirthForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			if data.Errors.Any() || data.Form.IgnoreDobWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if data.Errors.None() && data.DobWarning == "" {
				if provided.Donor.DateOfBirth == data.Form.Dob {
					if data.MakingAnotherLPA {
						return donor.PathMakeANewLPA.Redirect(w, r, appData, provided)
					}

					return donor.PathYourAddress.Redirect(w, r, appData, provided)
				}

				provided.Donor.DateOfBirth = data.Form.Dob
				provided.HasSentApplicationUpdatedEvent = false

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if data.MakingAnotherLPA {
					return donor.PathWeHaveUpdatedYourDetails.RedirectQuery(w, r, appData, provided, url.Values{"detail": {"dateOfBirth"}})
				}

				return donor.PathYourAddress.Redirect(w, r, appData, provided)
			}
		}

		data.DobWarning = data.Form.DobWarning()

		return tmpl(w, data)
	}
}

type yourDateOfBirthForm struct {
	Dob              date.Date
	IgnoreDobWarning string
}

func readYourDateOfBirthForm(r *http.Request) *yourDateOfBirthForm {
	d := &yourDateOfBirthForm{}

	d.Dob = date.New(
		page.PostFormString(r, "date-of-birth-year"),
		page.PostFormString(r, "date-of-birth-month"),
		page.PostFormString(r, "date-of-birth-day"))

	d.IgnoreDobWarning = page.PostFormString(r, "ignore-dob-warning")

	return d
}

func (f *yourDateOfBirthForm) Validate() validation.List {
	var errors validation.List

	errors.Date("date-of-birth", "dateOfBirth", f.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	return errors
}

func (f *yourDateOfBirthForm) DobWarning() string {
	var (
		today                = date.Today()
		hundredYearsEarlier  = today.AddDate(-100, 0, 0)
		eighteenYearsEarlier = today.AddDate(-18, 0, 0)
	)

	if !f.Dob.IsZero() {
		if f.Dob.Before(hundredYearsEarlier) {
			return "dateOfBirthIsOver100"
		}
		if f.Dob.Before(today) && f.Dob.After(eighteenYearsEarlier) {
			return "dateOfBirthIsUnder18"
		}
	}

	return ""
}
