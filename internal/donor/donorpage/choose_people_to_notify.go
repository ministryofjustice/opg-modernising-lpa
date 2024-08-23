package donorpage

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type choosePeopleToNotifyData struct {
	App         appcontext.Data
	Errors      validation.List
	Form        *choosePeopleToNotifyForm
	NameWarning *actor.SameNameWarning
}

func ChoosePeopleToNotify(tmpl template.Template, donorStore DonorStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if len(provided.PeopleToNotify) > 4 {
			return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
		}

		addAnother := r.FormValue("addAnother") == "1"
		personToNotify, personFound := provided.PeopleToNotify.Get(actoruid.FromRequest(r))

		if r.Method == http.MethodGet && len(provided.PeopleToNotify) > 0 && personFound == false && addAnother == false {
			return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
		}

		data := &choosePeopleToNotifyData{
			App: appData,
			Form: &choosePeopleToNotifyForm{
				FirstNames: personToNotify.FirstNames,
				LastName:   personToNotify.LastName,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readChoosePeopleToNotifyForm(r)
			data.Errors = data.Form.Validate()

			nameWarning := actor.NewSameNameWarning(
				actor.TypePersonToNotify,
				personToNotifyMatches(provided, personToNotify.UID, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String() {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.NameWarning == nil {
				if personFound == false {
					personToNotify = donordata.PersonToNotify{
						UID:        newUID(),
						FirstNames: data.Form.FirstNames,
						LastName:   data.Form.LastName,
					}

					provided.PeopleToNotify = append(provided.PeopleToNotify, personToNotify)
				} else {
					personToNotify.FirstNames = data.Form.FirstNames
					personToNotify.LastName = data.Form.LastName

					provided.PeopleToNotify.Put(personToNotify)
				}

				if !provided.Tasks.PeopleToNotify.IsCompleted() {
					provided.Tasks.PeopleToNotify = task.StateInProgress
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathChoosePeopleToNotifyAddress.RedirectQuery(w, r, appData, provided, url.Values{"id": {personToNotify.UID.String()}})
			}
		}

		return tmpl(w, data)
	}
}

type choosePeopleToNotifyForm struct {
	FirstNames        string
	LastName          string
	IgnoreNameWarning string
}

func readChoosePeopleToNotifyForm(r *http.Request) *choosePeopleToNotifyForm {
	return &choosePeopleToNotifyForm{
		FirstNames:        page.PostFormString(r, "first-names"),
		LastName:          page.PostFormString(r, "last-name"),
		IgnoreNameWarning: page.PostFormString(r, "ignore-name-warning"),
	}
}

func (f *choosePeopleToNotifyForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	return errors
}

func personToNotifyMatches(donor *donordata.Provided, uid actoruid.UID, firstNames, lastName string) actor.Type {
	if firstNames == "" && lastName == "" {
		return actor.TypeNone
	}

	for person := range donor.Actors() {
		if !(person.Type.IsPersonToNotify() && person.UID == uid) &&
			!person.Type.IsCertificateProvider() &&
			!person.Type.IsAuthorisedSignatory() &&
			!person.Type.IsIndependentWitness() &&
			strings.EqualFold(person.FirstNames, firstNames) &&
			strings.EqualFold(person.LastName, lastName) {
			return person.Type
		}
	}

	return actor.TypeNone
}
