package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type choosePeopleToNotifyData struct {
	App         AppData
	Errors      validation.List
	Form        *choosePeopleToNotifyForm
	NameWarning *actor.SameNameWarning
}

func ChoosePeopleToNotify(tmpl template.Template, lpaStore LpaStore, randomString func(int) string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if len(lpa.PeopleToNotify) > 4 {
			return appData.Redirect(w, r, lpa, Paths.ChoosePeopleToNotifySummary)
		}

		addAnother := r.FormValue("addAnother") == "1"
		personToNotify, personFound := lpa.PeopleToNotify.Get(r.URL.Query().Get("id"))

		if r.Method == http.MethodGet && len(lpa.PeopleToNotify) > 0 && personFound == false && addAnother == false {
			return appData.Redirect(w, r, lpa, Paths.ChoosePeopleToNotifySummary)
		}

		data := &choosePeopleToNotifyData{
			App: appData,
			Form: &choosePeopleToNotifyForm{
				FirstNames: personToNotify.FirstNames,
				LastName:   personToNotify.LastName,
				Email:      personToNotify.Email,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readChoosePeopleToNotifyForm(r)
			data.Errors = data.Form.Validate()

			nameWarning := actor.NewSameNameWarning(
				actor.TypePersonToNotify,
				personToNotifyMatches(lpa, personToNotify.ID, data.Form.FirstNames, data.Form.LastName),
				data.Form.FirstNames,
				data.Form.LastName,
			)

			if data.Errors.Any() || data.Form.IgnoreNameWarning != nameWarning.String() {
				data.NameWarning = nameWarning
			}

			if data.Errors.None() && data.NameWarning == nil {
				if personFound == false {
					personToNotify = actor.PersonToNotify{
						FirstNames: data.Form.FirstNames,
						LastName:   data.Form.LastName,
						Email:      data.Form.Email,
						ID:         randomString(8),
					}

					lpa.PeopleToNotify = append(lpa.PeopleToNotify, personToNotify)
				} else {
					personToNotify.FirstNames = data.Form.FirstNames
					personToNotify.LastName = data.Form.LastName
					personToNotify.Email = data.Form.Email

					lpa.PeopleToNotify.Put(personToNotify)
				}

				lpa.Tasks.PeopleToNotify = TaskInProgress

				if err := lpaStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				from := r.FormValue("from")

				if from == "" {
					from = fmt.Sprintf("%s?id=%s", appData.Paths.ChoosePeopleToNotifyAddress, personToNotify.ID)
				}

				return appData.Redirect(w, r, lpa, from)
			}
		}

		return tmpl(w, data)
	}
}

type choosePeopleToNotifyForm struct {
	FirstNames        string
	LastName          string
	Email             string
	IgnoreNameWarning string
}

func readChoosePeopleToNotifyForm(r *http.Request) *choosePeopleToNotifyForm {
	return &choosePeopleToNotifyForm{
		FirstNames:        postFormString(r, "first-names"),
		LastName:          postFormString(r, "last-name"),
		Email:             postFormString(r, "email"),
		IgnoreNameWarning: postFormString(r, "ignore-name-warning"),
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

	errors.String("email", "email", f.Email,
		validation.Empty(),
		validation.Email())

	return errors
}

func personToNotifyMatches(lpa *Lpa, id, firstNames, lastName string) actor.Type {
	if lpa.You.FirstNames == firstNames && lpa.You.LastName == lastName {
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

	for _, person := range lpa.PeopleToNotify {
		if person.ID != id && person.FirstNames == firstNames && person.LastName == lastName {
			return actor.TypePersonToNotify
		}
	}

	return actor.TypeNone
}
