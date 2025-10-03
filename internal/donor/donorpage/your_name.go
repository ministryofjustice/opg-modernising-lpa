package donorpage

import (
	"fmt"
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

type yourNameData struct {
	App              appcontext.Data
	Errors           validation.List
	Form             *yourNameForm
	CanTaskList      bool
	MakingAnotherLPA bool
}

func YourName(tmpl template.Template, donorStore DonorStore, sessionStore SessionStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &yourNameData{
			App: appData,
			Form: &yourNameForm{
				FirstNames: provided.Donor.FirstNames,
				LastName:   provided.Donor.LastName,
				OtherNames: provided.Donor.OtherNames,
			},
			CanTaskList:      !provided.Type.Empty(),
			MakingAnotherLPA: r.FormValue("makingAnotherLPA") == "1",
		}

		data.App.CanGoBack = data.CanTaskList || data.MakingAnotherLPA

		if r.Method == http.MethodPost {
			data.Form = readYourNameForm(r)
			data.Errors = data.Form.Validate()
			nameHasChanged := provided.Donor.NameHasChanged(data.Form.FirstNames, data.Form.LastName, data.Form.OtherNames)

			if data.Errors.None() {
				if !nameHasChanged {
					if data.MakingAnotherLPA {
						return donor.PathMakeANewLPA.Redirect(w, r, appData, provided)
					}

					return donor.PathYourDateOfBirth.Redirect(w, r, appData, provided)
				}

				if appData.SupporterData == nil {
					loginSession, err := sessionStore.Login(r)
					if err != nil {
						return err
					}
					if loginSession.Email == "" {
						return fmt.Errorf("no email in login session")
					}

					provided.Donor.Email = loginSession.Email
				}

				provided.Donor.FirstNames = data.Form.FirstNames
				provided.Donor.LastName = data.Form.LastName
				provided.Donor.OtherNames = data.Form.OtherNames

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				next := donor.PathYourDateOfBirth
				if data.MakingAnotherLPA {
					next = donor.PathWeHaveUpdatedYourDetails
				}

				if nameHasChanged && !donorMatches(provided, provided.Donor.FirstNames, provided.Donor.LastName).IsNone() {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"warningFrom": {appData.Page},
						"next":        {next.Format(provided.LpaID)},
						"actor":       {actor.TypeDonor.String()},
					})
				}

				if data.MakingAnotherLPA {
					return next.RedirectQuery(w, r, appData, provided, url.Values{"detail": {"name"}})
				}

				return next.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}

type yourNameForm struct {
	FirstNames string
	LastName   string
	OtherNames string
}

func readYourNameForm(r *http.Request) *yourNameForm {
	return &yourNameForm{
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
		OtherNames: page.PostFormString(r, "other-names"),
	}
}

func (f *yourNameForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	errors.String("other-names", "otherNamesYouAreKnownBy", f.OtherNames,
		validation.StringTooLong(50))

	return errors
}
