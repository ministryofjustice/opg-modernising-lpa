package page

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourDetailsData struct {
	App        AppData
	Errors     validation.List
	Form       *yourDetailsForm
	DobWarning string
}

func YourDetails(tmpl template.Template, lpaStore LpaStore, sessionStore sessions.Store) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &yourDetailsData{
			App: appData,
			Form: &yourDetailsForm{
				FirstNames: lpa.You.FirstNames,
				LastName:   lpa.You.LastName,
				OtherNames: lpa.You.OtherNames,
			},
		}

		if !lpa.You.DateOfBirth.IsZero() {
			data.Form.Dob = date.Read(lpa.You.DateOfBirth)
		}

		if r.Method == http.MethodPost {
			session, err := sessionStore.Get(r, "session")
			if err != nil {
				return err
			}

			email, ok := session.Values["email"].(string)
			if !ok {
				return fmt.Errorf("no email found in session")
			}

			data.Form = readYourDetailsForm(r)
			data.Errors = data.Form.Validate()
			dobWarning := data.Form.DobWarning()

			if data.Errors.Any() || data.Form.IgnoreWarning != dobWarning {
				data.DobWarning = dobWarning
			}

			if !data.Errors.Any() && data.DobWarning == "" {
				lpa.You.FirstNames = data.Form.FirstNames
				lpa.You.LastName = data.Form.LastName
				lpa.You.OtherNames = data.Form.OtherNames
				lpa.You.DateOfBirth = data.Form.Dob.T
				lpa.You.Email = email
				lpa.Tasks.YourDetails = TaskInProgress

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return appData.Redirect(w, r, lpa, Paths.YourAddress)
			}
		}

		return tmpl(w, data)
	}
}

type yourDetailsForm struct {
	FirstNames    string
	LastName      string
	OtherNames    string
	Dob           date.Date
	IgnoreWarning string
}

func readYourDetailsForm(r *http.Request) *yourDetailsForm {
	d := &yourDetailsForm{}

	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.OtherNames = postFormString(r, "other-names")

	d.Dob = date.FromParts(
		postFormString(r, "date-of-birth-year"),
		postFormString(r, "date-of-birth-month"),
		postFormString(r, "date-of-birth-day"))

	d.IgnoreWarning = postFormString(r, "ignore-warning")

	return d
}

func (f *yourDetailsForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	errors.String("other-names", "otherNamesLabel", f.OtherNames,
		validation.StringTooLong(50))

	errors.Date("date-of-birth", "dateOfBirth", f.Dob,
		validation.DateMissing(),
		validation.DateMustBeReal(),
		validation.DateMustBePast())

	return errors
}

func (f *yourDetailsForm) DobWarning() string {
	var (
		today                = time.Now().UTC().Round(24 * time.Hour)
		hundredYearsEarlier  = today.AddDate(-100, 0, 0)
		eighteenYearsEarlier = today.AddDate(-18, 0, 0)
	)

	if !f.Dob.T.IsZero() {
		if f.Dob.T.Before(hundredYearsEarlier) {
			return "dateOfBirthIsOver100"
		}
		if f.Dob.T.Before(today) && f.Dob.T.After(eighteenYearsEarlier) {
			return "dateOfBirthIsUnder18"
		}
	}

	return ""
}
