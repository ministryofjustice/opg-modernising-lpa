package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterOrganisationNameData struct {
	App    page.AppData
	Errors validation.List
	Form   *organisationNameForm
}

func EnterOrganisationName(tmpl template.Template, organisationStore OrganisationStore, sessionStore sesh.Store) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		data := &enterOrganisationNameData{
			App:  appData,
			Form: &organisationNameForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readOrganisationNameForm(r, "fullOrganisationOrCompanyName")
			data.Errors = data.Form.Validate()

			if !data.Errors.Any() {
				organisation, err := organisationStore.Create(r.Context(), data.Form.Name)
				if err != nil {
					return err
				}

				loginSession, err := sesh.Login(sessionStore, r)
				if err != nil {
					return page.Paths.Supporter.Start.Redirect(w, r, appData)
				}

				loginSession.OrganisationID = organisation.ID
				if err := sesh.SetLoginSession(sessionStore, r, w, loginSession); err != nil {
					return err
				}

				return page.Paths.Supporter.OrganisationCreated.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}

type organisationNameForm struct {
	Name  string
	Label string
}

func readOrganisationNameForm(r *http.Request, label string) *organisationNameForm {
	return &organisationNameForm{
		Name:  page.PostFormString(r, "name"),
		Label: label,
	}
}

func (f *organisationNameForm) Validate() validation.List {
	var errors validation.List

	errors.String("name", f.Label, f.Name,
		validation.Empty(),
		validation.StringTooLong(100))

	return errors
}
