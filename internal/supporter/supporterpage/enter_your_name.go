package supporterpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

type enterYourNameData struct {
	App  appcontext.Data
	Form *enterYourNameForm
}

func EnterYourName(tmpl template.Template, memberStore MemberStore) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := &enterYourNameData{
			App:  appData,
			Form: newEnterYourNameForm(appData.Localizer),
		}

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			if _, err := memberStore.Create(r.Context(), data.Form.FirstNames.Value, data.Form.LastName.Value); err != nil {
				return err
			}

			return page.PathSupporterEnterOrganisationName.Redirect(w, r, appData)
		}

		return tmpl(w, data)
	}
}

type enterYourNameForm struct {
	newforms.Form
	FirstNames *newforms.String
	LastName   *newforms.String
}

func newEnterYourNameForm(l Localizer) *enterYourNameForm {
	return &enterYourNameForm{
		FirstNames: newforms.NewString("first-names", l.T("firstNames")).
			NotEmpty().
			MaxLength(53),
		LastName: newforms.NewString("last-name", l.T("lastName")).
			NotEmpty().
			MaxLength(61),
	}
}

func (f *enterYourNameForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r,
		f.FirstNames,
		f.LastName,
	)
}
