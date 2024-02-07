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
	App    page.AppData
	Errors validation.List
	Form   *inviteMemberForm
}

func InviteMember(tmpl template.Template, organisationStore OrganisationStore, notifyClient NotifyClient, randomString func(int) string) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation) error {
		data := &inviteMemberData{
			App: appData,
			Form: &inviteMemberForm{
				Options: PermissionValues,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readInviteMemberForm(r)
			data.Errors = data.Form.Validate()

			if !data.Errors.Any() {
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

				if err := notifyClient.SendEmail(r.Context(), data.Form.Email, notify.MemberInviteEmail{
					OrganisationName: organisation.Name,
					InviteCode:       inviteCode,
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
	FirstNames      string
	LastName        string
	Email           string
	Permission      Permission
	Options         PermissionOptions
	PermissionError error
}

func readInviteMemberForm(r *http.Request) *inviteMemberForm {
	form := &inviteMemberForm{
		Email:      page.PostFormString(r, "email"),
		FirstNames: page.PostFormString(r, "first-names"),
		LastName:   page.PostFormString(r, "last-name"),
		Options:    PermissionValues,
	}

	form.Permission, form.PermissionError = ParsePermission(page.PostFormString(r, "permission"))

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

	errors.Error("permission", "makeThisPersonAnAdmin", f.PermissionError)

	return errors
}
