package supporter

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
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
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		data := &inviteMemberData{
			App:  appData,
			Form: &inviteMemberForm{},
		}

		if r.Method == http.MethodPost {
			data.Form = readInviteMemberForm(r)
			data.Errors = data.Form.Validate()

			if !data.Errors.Any() {
				organisation, err := organisationStore.Get(r.Context())
				if err != nil {
					return err
				}

				inviteCode := randomString(12)
				if err := organisationStore.CreateMemberInvite(r.Context(), organisation, data.Form.Email, inviteCode); err != nil {
					return err
				}

				if _, err := notifyClient.SendEmail(r.Context(), data.Form.Email, notify.MemberInviteEmail{
					OrganisationName: organisation.Name,
					InviteCode:       inviteCode,
				}); err != nil {
					return err
				}

				return page.Paths.Supporter.Dashboard.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}

type inviteMemberForm struct {
	Email string
}

func readInviteMemberForm(r *http.Request) *inviteMemberForm {
	return &inviteMemberForm{
		Email: page.PostFormString(r, "email"),
	}
}

func (f *inviteMemberForm) Validate() validation.List {
	var errors validation.List

	errors.String("email", "email", f.Email,
		validation.Empty(),
		validation.Email())

	return errors
}
