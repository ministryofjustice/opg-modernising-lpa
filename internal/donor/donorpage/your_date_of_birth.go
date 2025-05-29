package donorpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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
			dobWarning := dateOfBirthWarning(data.Form.Dob, actor.TypeDonor)

			if data.Errors.None() {
				if provided.Donor.DateOfBirth == data.Form.Dob {
					if data.MakingAnotherLPA {
						return donor.PathMakeANewLPA.Redirect(w, r, appData, provided)
					}

					return donor.PathDoYouLiveInTheUK.Redirect(w, r, appData, provided)
				}

				provided.Donor.DateOfBirth = data.Form.Dob
				provided.HasSentApplicationUpdatedEvent = false

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.Donor.IsUnder18() {
					from := r.FormValue("from")
					r.Form = url.Values{}
					return donor.PathYouHaveToldUsYouAreUnder18.RedirectQuery(w, r, appData, provided, url.Values{
						"next": {from},
					})
				}

				next := donor.PathDoYouLiveInTheUK
				if data.MakingAnotherLPA {
					next = donor.PathWeHaveUpdatedYourDetails
				}

				if dobWarning != "" {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"warningFrom": {appData.Page},
						"next":        {next.Format(provided.LpaID)},
						"actor":       {actor.TypeDonor.String()},
					})
				}

				if data.MakingAnotherLPA {
					return next.RedirectQuery(w, r, appData, provided, url.Values{"detail": {"dateOfBirth"}})
				}

				return next.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type yourDateOfBirthForm struct {
	Dob date.Date
}

func readYourDateOfBirthForm(r *http.Request) *yourDateOfBirthForm {
	d := &yourDateOfBirthForm{}

	d.Dob = date.New(
		page.PostFormString(r, "date-of-birth-year"),
		page.PostFormString(r, "date-of-birth-month"),
		page.PostFormString(r, "date-of-birth-day"))

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
