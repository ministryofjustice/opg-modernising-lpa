package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterOrganisationNameData struct {
	App    page.AppData
	Errors validation.List
	Form   *enterOrganisationNameForm
}

func EnterOrganisationName(tmpl template.Template, organisationStore OrganisationStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		data := &enterOrganisationNameData{
			App:  appData,
			Form: &enterOrganisationNameForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterOrganisationNameForm(r)
			data.Errors = data.Form.Validate()

			if !data.Errors.Any() {
				if err := organisationStore.Create(r.Context(), data.Form.Name); err != nil {
					return err
				}

				return page.Paths.Supporter.OrganisationCreated.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}

type enterOrganisationNameForm struct {
	Name string
}

func readEnterOrganisationNameForm(r *http.Request) *enterOrganisationNameForm {
	return &enterOrganisationNameForm{
		Name: page.PostFormString(r, "name"),
	}
}

func (f *enterOrganisationNameForm) Validate() validation.List {
	var errors validation.List

	errors.String("name", "fullOrganisationOrCompanyName", f.Name,
		validation.Empty(),
		validation.StringTooLong(100))

	return errors
}
