package supporterpage

import (
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterOrganisationNameData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *organisationNameForm
}

func EnterOrganisationName(logger Logger, tmpl template.Template, organisationStore OrganisationStore, memberStore MemberStore, sessionStore SessionStore) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := &enterOrganisationNameData{
			App:  appData,
			Form: &organisationNameForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readOrganisationNameForm(r, "fullOrganisationOrCompanyName")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				member, err := memberStore.GetAny(r.Context())
				if err != nil {
					return err
				}

				organisation, err := organisationStore.Create(r.Context(), member, data.Form.Name)
				if err != nil {
					return err
				}
				logger.InfoContext(r.Context(), "organisation created", slog.String("organisation_id", organisation.ID))

				loginSession, err := sessionStore.Login(r)
				if err != nil {
					return page.PathSupporterStart.Redirect(w, r, appData)
				}

				loginSession.OrganisationID = organisation.ID
				loginSession.OrganisationName = organisation.Name
				if err := sessionStore.SetLogin(r, w, loginSession); err != nil {
					return err
				}

				return supporter.PathOrganisationCreated.Redirect(w, r, appData)
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
