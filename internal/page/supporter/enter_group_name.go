package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterGroupNameData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterGroupNameForm
}

func EnterGroupName(tmpl template.Template, groupStore GroupStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		data := &enterGroupNameData{
			App:  appData,
			Form: &enterGroupNameForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterGroupNameForm(r)
			data.Errors = data.Form.Validate()

			if !data.Errors.Any() {
				if err := groupStore.Create(r.Context(), data.Form.Name); err != nil {
					return err
				}

				return page.Paths.Supporter.GroupCreated.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}

type enterGroupNameForm struct {
	Name string
}

func readEnterGroupNameForm(r *http.Request) *enterGroupNameForm {
	return &enterGroupNameForm{
		Name: page.PostFormString(r, "name"),
	}
}

func (f *enterGroupNameForm) Validate() validation.List {
	var errors validation.List

	errors.String("name", "fullOrganisationOrCompanyName", f.Name,
		validation.Empty(),
		validation.StringTooLong(100))

	return errors
}
