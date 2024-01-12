package donor

import (
	"fmt"
	"net/http"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type yourNameData struct {
	App         page.AppData
	Errors      validation.List
	Form        *yourNameForm
	NameWarning *actor.SameNameWarning
}

func YourName(tmpl template.Template, donorStore DonorStore, sessionStore sessions.Store) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &yourNameData{
			App: appData,
			Form: &yourNameForm{
				FirstNames: donor.Donor.FirstNames,
				LastName:   donor.Donor.LastName,
				OtherNames: donor.Donor.OtherNames,
			},
		}

		if r.Method == http.MethodGet && data.Form.FirstNames == "" {
			if latestDonor, _ := donorStore.Latest(r.Context()); latestDonor != nil {
				data.Form = &yourNameForm{
					FirstNames: latestDonor.Donor.FirstNames,
					LastName:   latestDonor.Donor.LastName,
					OtherNames: latestDonor.Donor.OtherNames,
				}
			}
		}

		if r.Method == http.MethodPost {
			loginSession, err := sesh.Login(sessionStore, r)
			if err != nil {
				return err
			}
			if loginSession.Email == "" {
				return fmt.Errorf("no email in login session")
			}

			data.Form = readYourNameForm(r)
			data.Errors = data.Form.Validate()

			nameWarning := actor.NewSameNameWarning(
				actor.TypeDonor,
				donorMatches(donor, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() ||
				data.Form.IgnoreNameWarning != nameWarning.String() &&
					donor.Donor.FullName() != fmt.Sprintf("%s %s", data.Form.FirstNames, data.Form.LastName) {
				data.NameWarning = nameWarning
			}

			if !data.Errors.Any() && data.NameWarning == nil {
				if donor.Donor.FirstNames != data.Form.FirstNames || donor.Donor.LastName != data.Form.LastName {
					donor.Donor.FirstNames = data.Form.FirstNames
					donor.Donor.LastName = data.Form.LastName

					donor.HasSentApplicationUpdatedEvent = false
				}

				donor.Donor.OtherNames = data.Form.OtherNames
				donor.Donor.Email = loginSession.Email

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.MakeANewLPA.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}

type yourNameForm struct {
	FirstNames        string
	LastName          string
	OtherNames        string
	IgnoreNameWarning string
}

func readYourNameForm(r *http.Request) *yourNameForm {
	d := &yourNameForm{}

	d.FirstNames = page.PostFormString(r, "first-names")
	d.LastName = page.PostFormString(r, "last-name")
	d.OtherNames = page.PostFormString(r, "other-names")
	d.IgnoreNameWarning = page.PostFormString(r, "ignore-name-warning")

	return d
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
