package donorpage

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

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
	NameWarning      *actor.SameNameWarning
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

			nameWarning := actor.NewSameNameWarning(
				actor.TypeDonor,
				donorMatches(provided, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() ||
				data.Form.IgnoreNameWarning != nameWarning.String() &&
					provided.Donor.FullName() != fmt.Sprintf("%s %s", data.Form.FirstNames, data.Form.LastName) {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.NameWarning == nil {
				if !provided.NamesChanged(data.Form.FirstNames, data.Form.LastName, data.Form.OtherNames) {
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
				provided.HasSentApplicationUpdatedEvent = false

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if data.MakingAnotherLPA {
					return donor.PathWeHaveUpdatedYourDetails.RedirectQuery(w, r, appData, provided, url.Values{"detail": {"name"}})
				}

				return donor.PathYourDateOfBirth.Redirect(w, r, appData, provided)
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
	return &yourNameForm{
		FirstNames:        page.PostFormString(r, "first-names"),
		LastName:          page.PostFormString(r, "last-name"),
		OtherNames:        page.PostFormString(r, "other-names"),
		IgnoreNameWarning: page.PostFormString(r, "ignore-name-warning"),
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

func donorMatches(donor *donordata.Provided, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
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

	if strings.EqualFold(donor.CertificateProvider.FirstNames, firstNames) && strings.EqualFold(donor.CertificateProvider.LastName, lastName) {
		return actor.TypeCertificateProvider
	}

	for _, person := range donor.PeopleToNotify {
		if strings.EqualFold(person.FirstNames, firstNames) && strings.EqualFold(person.LastName, lastName) {
			return actor.TypePersonToNotify
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
