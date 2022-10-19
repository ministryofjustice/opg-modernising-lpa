package page

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
)

type chooseAttorneysData struct {
	App         AppData
	Errors      map[string]string
	Form        *chooseAttorneysForm
	ShowDetails bool
}

func ChooseAttorneys(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		attorneyId := r.URL.Query().Get("id")
		attorney, err := lpa.GetAttorney(attorneyId)

		data := &chooseAttorneysData{
			App: appData,
			Form: &chooseAttorneysForm{
				FirstNames: attorney.FirstNames,
				LastName:   attorney.LastName,
				Email:      attorney.Email,
			},
			ShowDetails: err != nil,
		}

		if !attorney.DateOfBirth.IsZero() {
			data.Form.Dob = readDate(attorney.DateOfBirth)
		}

		if r.Method == http.MethodPost {
			data.Form = readChooseAttorneysForm(r)
			data.Errors = data.Form.Validate()

			if len(data.Errors) == 0 {
				if err != nil {
					attorney = Attorney{
						FirstNames:  data.Form.FirstNames,
						LastName:    data.Form.LastName,
						Email:       data.Form.Email,
						DateOfBirth: data.Form.DateOfBirth,
					}

					_ = append(lpa.Attorneys, attorney)
				} else {
					attorney.FirstNames = data.Form.FirstNames
					attorney.LastName = data.Form.LastName
					attorney.Email = data.Form.Email
					attorney.DateOfBirth = data.Form.DateOfBirth

					lpa.PutAttorney(attorney)
				}

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}
				appData.Lang.Redirect(w, r, chooseAttorneysAddressPath, http.StatusFound)
				return nil
			}
		}

		return tmpl(w, data)
	}
}

type chooseAttorneysForm struct {
	FirstNames       string
	LastName         string
	Email            string
	Dob              Date
	DateOfBirth      time.Time
	DateOfBirthError error
}

func readChooseAttorneysForm(r *http.Request) *chooseAttorneysForm {
	d := &chooseAttorneysForm{}
	d.FirstNames = postFormString(r, "first-names")
	d.LastName = postFormString(r, "last-name")
	d.Email = postFormString(r, "email")
	d.Dob = Date{
		Day:   postFormString(r, "date-of-birth-day"),
		Month: postFormString(r, "date-of-birth-month"),
		Year:  postFormString(r, "date-of-birth-year"),
	}

	d.DateOfBirth, d.DateOfBirthError = time.Parse("2006-1-2", d.Dob.Year+"-"+d.Dob.Month+"-"+d.Dob.Day)

	return d
}

func (d *chooseAttorneysForm) Validate() map[string]string {
	errors := map[string]string{}

	if d.FirstNames == "" {
		errors["first-names"] = "enterFirstNames"
	}
	if d.LastName == "" {
		errors["last-name"] = "enterLastName"
	}
	if d.Email == "" {
		errors["email"] = "enterEmail"
	}
	if d.Dob.Day == "" {
		errors["date-of-birth"] = "dateOfBirthDay"
	}
	if d.Dob.Month == "" {
		errors["date-of-birth"] = "dateOfBirthMonth"
	}
	if d.Dob.Year == "" {
		errors["date-of-birth"] = "dateOfBirthYear"
	}
	if _, ok := errors["date-of-birth"]; !ok && d.DateOfBirthError != nil {
		errors["date-of-birth"] = "dateOfBirthMustBeReal"
	}

	return errors
}
