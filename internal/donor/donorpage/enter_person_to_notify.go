package donorpage

import (
	"net/http"
	"net/url"

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

type enterPersonToNotifyData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *enterPersonToNotifyForm
}

func EnterPersonToNotify(tmpl template.Template, donorStore DonorStore, newUID func() actoruid.UID) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if len(provided.PeopleToNotify) > 4 {
			return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
		}

		addAnother := r.FormValue("addAnother") == "1"
		personToNotify, personFound := provided.PeopleToNotify.Get(actoruid.FromRequest(r))

		if r.Method == http.MethodGet && len(provided.PeopleToNotify) > 0 && personFound == false && addAnother == false {
			return donor.PathChoosePeopleToNotifySummary.Redirect(w, r, appData, provided)
		}

		data := &enterPersonToNotifyData{
			App: appData,
			Form: &enterPersonToNotifyForm{
				FirstNames: personToNotify.FirstNames,
				LastName:   personToNotify.LastName,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterPersonToNotifyForm(r)
			data.Errors = data.Form.Validate()

			nameMatches := personToNotifyMatches(provided, personToNotify.UID, data.Form.FirstNames, data.Form.LastName)
			redirectToWarning := false

			if !nameMatches.IsNone() && personToNotify.NameHasChanged(data.Form.FirstNames, data.Form.LastName) {
				redirectToWarning = true
			}

			if data.Errors.None() {
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

				if redirectToWarning {
					return donor.PathWarningInterruption.RedirectQuery(w, r, appData, provided, url.Values{
						"id":          {personToNotify.UID.String()},
						"warningFrom": {appData.Page},
						"next": {donor.PathEnterPersonToNotifyAddress.FormatQuery(
							provided.LpaID,
							url.Values{"id": {personToNotify.UID.String()}}),
						},
						"actor": {actor.TypePersonToNotify.String()},
					})
				}

				return donor.PathEnterPersonToNotifyAddress.RedirectQuery(w, r, appData, provided, url.Values{"id": {personToNotify.UID.String()}})
			}
		}

		return tmpl(w, data)
	}
}

type enterPersonToNotifyForm struct {
	FirstNames string
	LastName   string
}

func readEnterPersonToNotifyForm(r *http.Request) *enterPersonToNotifyForm {
	return &enterPersonToNotifyForm{
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
	}
}

func (f *enterPersonToNotifyForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	return errors
}
