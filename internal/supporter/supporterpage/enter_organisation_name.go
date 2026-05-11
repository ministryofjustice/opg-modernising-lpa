package supporterpage

import (
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
)

type enterOrganisationNameData struct {
	App  appcontext.Data
	Form *organisationNameForm
}

func EnterOrganisationName(logger Logger, tmpl template.Template, organisationStore OrganisationStore, memberStore MemberStore, sessionStore SessionStore) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := &enterOrganisationNameData{
			App:  appData,
			Form: newOrganisationNameForm(appData.Localizer.T("fullOrganisationOrCompanyName")),
		}

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			member, err := memberStore.GetAny(r.Context())
			if err != nil {
				return err
			}

			organisation, err := organisationStore.Create(r.Context(), member, data.Form.Name.Value)
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

		return tmpl(w, data)
	}
}

type organisationNameForm struct {
	newforms.Form
	Name *newforms.String
}

func newOrganisationNameForm(label string) *organisationNameForm {
	return &organisationNameForm{
		Name: newforms.NewString("name", label).
			NotEmpty().
			MaxLength(100),
	}
}

func (f *organisationNameForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.Name)
}
