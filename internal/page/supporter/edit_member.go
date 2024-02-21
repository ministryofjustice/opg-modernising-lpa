package supporter

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type editMemberData struct {
	App    page.AppData
	Errors validation.List
	Form   *editMemberForm
	Member *actor.Member
}

func EditMember(tmpl template.Template, memberStore MemberStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		member, err := memberStore.GetByID(r.Context(), r.URL.Query().Get("id"))
		if err != nil {
			return err
		}

		data := &editMemberData{
			App: appData,
			Form: &editMemberForm{
				FirstNames: member.FirstNames,
				LastName:   member.LastName,
			},
			Member: member,
		}

		if r.Method == http.MethodPost {
			data.Form = readEditMemberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				query := url.Values{}
				if data.Form.FirstNames != member.FirstNames || data.Form.LastName != member.LastName {
					member.FirstNames = data.Form.FirstNames
					member.LastName = data.Form.LastName

					query.Add("nameUpdated", member.FullName())
				}

				if err := memberStore.Put(r.Context(), member); err != nil {
					return err
				}

				return page.Paths.Supporter.ManageTeamMembers.RedirectQuery(w, r, appData, query)
			}
		}

		return tmpl(w, data)
	}
}

type editMemberForm struct {
	FirstNames string
	LastName   string
}

func readEditMemberForm(r *http.Request) *editMemberForm {
	return &editMemberForm{
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
	}
}

func (f *editMemberForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	return errors
}
