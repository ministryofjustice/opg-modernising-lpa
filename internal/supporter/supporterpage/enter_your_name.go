package supporterpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterYourNameData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *enterYourNameForm
}

func EnterYourName(tmpl template.Template, memberStore MemberStore) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := &enterYourNameData{
			App:  appData,
			Form: &enterYourNameForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterYourNameForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if _, err := memberStore.Create(r.Context(), data.Form.FirstNames, data.Form.LastName); err != nil {
					return err
				}

				return supporter.PathEnterOrganisationName.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}

type enterYourNameForm struct {
	FirstNames string
	LastName   string
}

func readEnterYourNameForm(r *http.Request) *enterYourNameForm {
	return &enterYourNameForm{
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
	}
}

func (f *enterYourNameForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	return errors
}
