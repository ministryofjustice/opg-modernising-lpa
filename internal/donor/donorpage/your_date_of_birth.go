package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourDateOfBirthData struct {
	App              page.AppData
	Errors           validation.List
	Form             *yourDateOfBirthForm
	DobWarning       string
	CanTaskList      bool
	MakingAnotherLPA bool
}

func YourDateOfBirth(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &yourDateOfBirthData{
			App: appData,
			Form: &yourDateOfBirthForm{
				Dob: donor.Donor.DateOfBirth,
			},
			CanTaskList:      !donor.Type.Empty(),
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
				if donor.Donor.DateOfBirth == data.Form.Dob {
					if data.MakingAnotherLPA {
						return page.Paths.MakeANewLPA.Redirect(w, r, appData, donor)
					}

					return page.Paths.YourAddress.Redirect(w, r, appData, donor)
				}

				donor.Donor.DateOfBirth = data.Form.Dob
				donor.HasSentApplicationUpdatedEvent = false

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if data.MakingAnotherLPA {
					return page.Paths.WeHaveUpdatedYourDetails.RedirectQuery(w, r, appData, donor, url.Values{"detail": {"dateOfBirth"}})
				}

				return page.Paths.YourAddress.Redirect(w, r, appData, donor)
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
