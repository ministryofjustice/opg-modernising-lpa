package supporterpage

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type editMemberData struct {
	App        appcontext.Data
	Errors     validation.List
	Form       *editMemberForm
	Member     *supporterdata.Member
	CanEditAll bool
}

func EditMember(logger Logger, tmpl template.Template, memberStore MemberStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, member *supporterdata.Member) error {
		memberID := r.FormValue("id")
		isLoggedInMember := member.ID == memberID
		if !isLoggedInMember {
			memberByID, err := memberStore.GetByID(r.Context(), memberID)
			if err != nil {
				return err
			}

			member = memberByID
		}

		canEditAll := appData.IsAdmin() && !isLoggedInMember

		data := &editMemberData{
			App: appData,
			Form: &editMemberForm{
				FirstNames:        member.FirstNames,
				LastName:          member.LastName,
				Permission:        member.Permission,
				PermissionOptions: supporterdata.PermissionValues,
				Status:            member.Status,
				StatusOptions:     supporterdata.StatusValues,
			},
			Member:     member,
			CanEditAll: canEditAll,
		}

		if r.Method == http.MethodPost {
			data.Form = readEditMemberForm(r, canEditAll)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				query := url.Values{}
				changed := false

				if data.Form.FirstNames != member.FirstNames || data.Form.LastName != member.LastName {
					changed = true
					member.FirstNames = data.Form.FirstNames
					member.LastName = data.Form.LastName

					query.Add("nameUpdated", member.FullName())

					if isLoggedInMember {
						query.Add("selfUpdated", "1")
					}
				}

				if canEditAll {
					if data.Form.Permission != member.Permission {
						changed = true
						logger.InfoContext(r.Context(), "member permission changed", slog.String("member_id", member.ID), slog.String("permission_old", member.Permission.String()), slog.String("permission_new", data.Form.Permission.String()))
						member.Permission = data.Form.Permission
					}

					if data.Form.Status != member.Status {
						changed = true
						logger.InfoContext(r.Context(), "member status changed", slog.String("member_id", member.ID), slog.String("status_old", member.Status.String()), slog.String("status_new", data.Form.Status.String()))
						query.Add("statusUpdated", data.Form.Status.String())
						query.Add("statusEmail", member.Email)
						member.Status = data.Form.Status
					}
				}

				if changed {
					if err := memberStore.Put(r.Context(), member); err != nil {
						return err
					}
				}

				redirect := page.Paths.Supporter.Dashboard
				if appData.IsAdmin() {
					redirect = page.Paths.Supporter.ManageTeamMembers
				}

				return redirect.RedirectQuery(w, r, appData, query)
			}
		}

		return tmpl(w, data)
	}
}

type editMemberForm struct {
	FirstNames        string
	LastName          string
	Permission        supporterdata.Permission
	PermissionOptions supporterdata.PermissionOptions
	PermissionError   error
	Status            supporterdata.Status
	StatusOptions     supporterdata.StatusOptions
	StatusError       error
	canEditAll        bool
}

func readEditMemberForm(r *http.Request, canEditAll bool) *editMemberForm {
	f := &editMemberForm{
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
		canEditAll: canEditAll,
	}

	if canEditAll {
		f.Permission, f.PermissionError = supporterdata.ParsePermission(page.PostFormString(r, "permission"))
		f.Status, f.StatusError = supporterdata.ParseStatus(page.PostFormString(r, "status"))
	}

	return f
}

func (f *editMemberForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	if f.canEditAll {
		errors.Options("permission", "makeThisPersonAnAdmin", []string{f.Permission.String()},
			validation.Select(supporterdata.PermissionNone.String(), supporterdata.PermissionAdmin.String()))

		errors.Error("status", "status", f.StatusError,
			validation.Selected())
	}

	return errors
}
