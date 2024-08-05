package supporterpage

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type manageTeamMembersData struct {
	App            appcontext.Data
	Errors         validation.List
	Organisation   *actor.Organisation
	InvitedMembers []*actor.MemberInvite
	Members        []*actor.Member
	Form           *inviteMemberForm
}

func ManageTeamMembers(tmpl template.Template, memberStore MemberStore, randomString func(int) string, notifyClient NotifyClient, appPublicURL string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *actor.Organisation, _ *actor.Member) error {
		data := &manageTeamMembersData{
			App:          appData,
			Organisation: organisation,
		}

		if r.Method == http.MethodPost {
			data.Form = readInviteMemberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if err := memberStore.DeleteMemberInvite(r.Context(), organisation.ID, data.Form.Email); err != nil {
					return err
				}

				inviteCode := randomString(12)
				if err := memberStore.CreateMemberInvite(
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
					InviterEmail:          appData.LoginSessionEmail,
					InviteCode:            inviteCode,
					JoinAnOrganisationURL: appPublicURL + page.Paths.Supporter.Start.Format(),
				}); err != nil {
					return err
				}

				return page.Paths.Supporter.ManageTeamMembers.RedirectQuery(w, r, appData, url.Values{"inviteSent": {data.Form.Email}})
			}

			return errors.New("unable to resend invite")
		}

		invitedMembers, err := memberStore.InvitedMembers(r.Context())
		if err != nil {
			return err
		}

		members, err := memberStore.GetAll(r.Context())
		if err != nil {
			return err
		}

		data.InvitedMembers = invitedMembers
		data.Members = members

		return tmpl(w, data)
	}
}
