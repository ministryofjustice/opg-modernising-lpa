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
				FirstNames:    member.FirstNames,
				LastName:      member.LastName,
				Status:        member.Status,
				StatusOptions: actor.StatusValues,
			},
			Member: member,
		}

		if r.Method == http.MethodPost {
			data.Form = readEditMemberForm(r, appData.IsAdmin())
			data.Errors = data.Form.Validate(appData.IsAdmin())

			if data.Errors.None() {
				query := url.Values{}
				if data.Form.FirstNames != member.FirstNames || data.Form.LastName != member.LastName {
					member.FirstNames = data.Form.FirstNames
					member.LastName = data.Form.LastName

					query.Add("nameUpdated", member.FullName())

					if member.Email == appData.LoginSessionEmail {
						query.Add("selfUpdated", "1")
					}
				}

				if appData.IsAdmin() && data.Form.Status != member.Status {
					query.Add("statusUpdated", data.Form.Status.String()+":"+member.Email)
					member.Status = data.Form.Status
				}

				if err := memberStore.Put(r.Context(), member); err != nil {
					return err
				}

				redirect := page.Paths.Supporter.ManageTeamMembers
				if !appData.IsAdmin() {
					redirect = page.Paths.Supporter.Dashboard
				}

				return redirect.RedirectQuery(w, r, appData, query)
			}
		}

		return tmpl(w, data)
	}
}

type editMemberForm struct {
	FirstNames    string
	LastName      string
	Status        actor.Status
	StatusOptions actor.StatusOptions
}

func readEditMemberForm(r *http.Request, isAdmin bool) *editMemberForm {
	f := &editMemberForm{
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
	}

	if isAdmin {
		f.Status, _ = actor.ParseStatus(page.PostFormString(r, "status"))
	}

	return f
}

func (f *editMemberForm) Validate(isAdmin bool) validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	if isAdmin {
		errors.Options("status", "status", []string{f.Status.String()}, validation.Select(actor.Active.String(), actor.Suspended.String()))
	}

	return errors
}
