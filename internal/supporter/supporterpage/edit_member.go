package supporterpage

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
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
			App:        appData,
			Form:       newEditMemberForm(appData.Localizer, canEditAll),
			Member:     member,
			CanEditAll: canEditAll,
		}

		data.Form.FirstNames.SetInput(member.FirstNames)
		data.Form.LastName.SetInput(member.LastName)
		if canEditAll {
			data.Form.Permission.SetInput(member.Permission)
			data.Form.Status.SetInput(member.Status)
		}

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				query := url.Values{}
				changed := false

				if data.Form.FirstNames.Value != member.FirstNames || data.Form.LastName.Value != member.LastName {
					changed = true
					member.FirstNames = data.Form.FirstNames.Value
					member.LastName = data.Form.LastName.Value

					query.Add("nameUpdated", member.FullName())

					if isLoggedInMember {
						query.Add("selfUpdated", "1")
					}
				}

				if canEditAll {
					if data.Form.Permission.Value != member.Permission {
						changed = true
						logger.InfoContext(r.Context(), "member permission changed",
							slog.String("member_id", member.ID),
							slog.String("permission_old", member.Permission.String()),
							slog.String("permission_new", data.Form.Permission.Value.String()))
						member.Permission = data.Form.Permission.Value
					}

					if data.Form.Status.Value != member.Status {
						changed = true
						logger.InfoContext(r.Context(), "member status changed",
							slog.String("member_id", member.ID),
							slog.String("status_old", member.Status.String()),
							slog.String("status_new", data.Form.Status.Value.String()))
						query.Add("statusUpdated", data.Form.Status.Value.String())
						query.Add("statusEmail", member.Email)
						member.Status = data.Form.Status.Value
					}
				}

				if changed {
					if err := memberStore.Put(r.Context(), member); err != nil {
						return err
					}
				}

				redirect := supporter.PathDashboard
				if appData.IsAdmin() {
					redirect = supporter.PathManageTeamMembers
				}

				return redirect.RedirectQuery(w, r, appData, query)
			}
		}

		return tmpl(w, data)
	}
}

type editMemberForm struct {
	FirstNames *newforms.String
	LastName   *newforms.String
	Permission *newforms.Enum[supporterdata.Permission, supporterdata.PermissionOptions, *supporterdata.Permission]
	Status     *newforms.Enum[supporterdata.Status, supporterdata.StatusOptions, *supporterdata.Status]
	Errors     []newforms.Field
	canEditAll bool
}

func newEditMemberForm(l Localizer, canEditAll bool) *editMemberForm {
	f := &editMemberForm{
		FirstNames: newforms.NewString("first-names", l.T("firstNames")).
			NotEmpty().
			MaxLength(53),
		LastName: newforms.NewString("last-name", l.T("lastName")).
			NotEmpty().
			MaxLength(61),
		canEditAll: canEditAll,
	}

	if canEditAll {
		f.Permission = newforms.NewEnum[supporterdata.Permission]("permission", l.T("permissions"), supporterdata.PermissionValues).
			OrDefault(supporterdata.PermissionNone)
		f.Status = newforms.NewEnum[supporterdata.Status]("status", l.T("status"), supporterdata.StatusValues).
			Selected()
	}

	return f
}

func (f *editMemberForm) Parse(r *http.Request) bool {
	if f.canEditAll {
		f.Errors = newforms.ParsePostForm(r,
			f.FirstNames,
			f.LastName,
			f.Permission,
			f.Status,
		)
	} else {
		f.Errors = newforms.ParsePostForm(r,
			f.FirstNames,
			f.LastName,
		)
	}

	return len(f.Errors) == 0
}
