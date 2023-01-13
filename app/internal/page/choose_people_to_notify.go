package page

import (
	"fmt"
	"net/http"
	"net/mail"

	"github.com/ministryofjustice/opg-go-common/template"
)

type choosePeopleToNotifyData struct {
	App    AppData
	Errors map[string]string
	Form   *choosePeopleToNotifyForm
}

type choosePeopleToNotifyForm struct {
	FirstNames string
	LastName   string
	Email      string
}

func ChoosePeopleToNotify(tmpl template.Template, lpaStore LpaStore, randomString func(int) string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		if len(lpa.PeopleToNotify) > 4 {
			return appData.Lang.Redirect(w, r, lpa, Paths.ChoosePeopleToNotifySummary)
		}

		addAnother := r.FormValue("addAnother") == "1"
		personToNotify, personFound := lpa.GetPersonToNotify(r.URL.Query().Get("id"))

		if r.Method == http.MethodGet && len(lpa.PeopleToNotify) > 0 && personFound == false && addAnother == false {
			return appData.Lang.Redirect(w, r, lpa, Paths.ChoosePeopleToNotifySummary)
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

			if len(data.Errors) == 0 {
				if personFound == false {
					personToNotify = PersonToNotify{
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

					lpa.PutPersonToNotify(personToNotify)
				}

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				from := r.FormValue("from")

				if from == "" {
					from = fmt.Sprintf("%s?id=%s", appData.Paths.ChoosePeopleToNotifyAddress, personToNotify.ID)
				}

				return appData.Lang.Redirect(w, r, lpa, from)
			}
		}

		return tmpl(w, data)
	}
}

func readChoosePeopleToNotifyForm(r *http.Request) *choosePeopleToNotifyForm {
	d := &choosePeopleToNotifyForm{}
	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.Email = postFormString(r, "email")

	return d
}

func (d *choosePeopleToNotifyForm) Validate() map[string]string {
	errors := map[string]string{}

	if d.FirstNames == "" {
		errors["first-names"] = "enterTheirFirstNames"
	}
	if len(d.FirstNames) > 53 {
		errors["first-names"] = "firstNamesTooLong"
	}

	if d.LastName == "" {
		errors["last-name"] = "enterTheirLastName"
	}
	if len(d.LastName) > 61 {
		errors["last-name"] = "lastNameTooLong"
	}

	if d.Email == "" {
		errors["email"] = "enterTheirEmail"
	} else if _, err := mail.ParseAddress(fmt.Sprintf("<%s>", d.Email)); err != nil {
		errors["email"] = "theirEmailIncorrectFormat"
	}

	return errors
}
