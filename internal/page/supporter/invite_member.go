package supporter

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type inviteMemberData struct {
	App     page.AppData
	Errors  validation.List
	Form    *inviteMemberForm
	Options actor.PermissionOptions
}

func InviteMember(tmpl template.Template, organisationStore OrganisationStore, notifyClient NotifyClient, randomString func(int) string, appPublicURL string) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		data := &inviteMemberData{
			App:     appData,
			Form:    &inviteMemberForm{},
			Options: actor.PermissionValues,
		}

		if r.Method == http.MethodPost {
			data.Form = readInviteMemberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				session, err := page.SessionDataFromContext(r.Context())
				if err != nil {
					return err
				}

				inviteCode := randomString(12)
				if err := organisationStore.CreateMemberInvite(
					r.Context(),
					organisation,
					data.Form.FirstNames,
					data.Form.LastName,
					data.Form.Email,
					inviteCode,
					data.Form.Permission,
				); err != nil {
					return err
				}

				if err := notifyClient.SendEmail(r.Context(), data.Form.Email, notify.OrganisationMemberInviteEmail{
					OrganisationName:      organisation.Name,
					InviterEmail:          session.Email,
					InviteCode:            inviteCode,
					JoinAnOrganisationURL: appPublicURL + page.Paths.Supporter.Start.Format(),
				}); err != nil {
					return err
				}

				return page.Paths.Supporter.InviteMemberConfirmation.RedirectQuery(w, r, appData, url.Values{"email": {data.Form.Email}})
			}
		}

		return tmpl(w, data)
	}
}

type inviteMemberForm struct {
	FirstNames string
	LastName   string
	Email      string
	Permission actor.Permission
}

func readInviteMemberForm(r *http.Request) *inviteMemberForm {
	form := &inviteMemberForm{
		Email:      page.PostFormString(r, "email"),
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
	}

	form.Permission, _ = actor.ParsePermission(page.PostFormString(r, "permission"))

	return form
}

func (f *inviteMemberForm) Validate() validation.List {
	var errors validation.List

	errors.String("first-names", "firstNames", f.FirstNames,
		validation.Empty(),
		validation.StringTooLong(53))

	errors.String("last-name", "lastName", f.LastName,
		validation.Empty(),
		validation.StringTooLong(61))

	errors.String("email", "email", f.Email,
		validation.Empty(),
		validation.Email())

	errors.Options("permission", "makeThisPersonAnAdmin", []string{f.Permission.String()}, validation.Select(actor.None.String(), actor.Admin.String()))

	return errors
}
