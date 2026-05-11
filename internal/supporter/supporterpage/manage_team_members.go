package supporterpage

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/invitecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
)

type manageTeamMembersData struct {
	App            appcontext.Data
	Organisation   *supporterdata.Organisation
	InvitedMembers []*supporterdata.MemberInvite
	Members        []*supporterdata.Member
	Form           *inviteMemberForm
}

func ManageTeamMembers(tmpl template.Template, memberStore MemberStore, generate invitecode.Generator, notifyClient NotifyClient, appPublicURL string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		data := &manageTeamMembersData{
			App:          appData,
			Organisation: organisation,
			Form:         newInviteMemberForm(appData.Localizer),
		}

		if r.Method == http.MethodPost {
			if !data.Form.Parse(r) {
				return errors.New("unable to resend invite")
			}

			if err := memberStore.DeleteMemberInvite(r.Context(), organisation.ID, data.Form.Email.Value); err != nil {
				return err
			}

			plainCode, hashedCode := generate()
			if err := memberStore.CreateMemberInvite(
				r.Context(),
				organisation,
				data.Form.FirstNames.Value,
				data.Form.LastName.Value,
				data.Form.Email.Value,
				hashedCode,
				data.Form.Permission.Value,
			); err != nil {
				return err
			}

			if err := notifyClient.SendEmail(r.Context(), notify.ToCustomEmail(localize.En, data.Form.Email.Value), notify.OrganisationMemberInviteEmail{
				OrganisationName:      organisation.Name,
				InviterEmail:          appData.LoginSessionEmail,
				InviteCode:            plainCode.Plain(),
				JoinAnOrganisationURL: appPublicURL + page.PathSupporterStart.Format(),
			}); err != nil {
				return err
			}

			return supporter.PathManageTeamMembers.RedirectQuery(w, r, appData, url.Values{"inviteSent": {data.Form.Email.Value}})
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
