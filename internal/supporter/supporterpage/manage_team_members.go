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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sharecode/sharecodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type manageTeamMembersData struct {
	App            appcontext.Data
	Errors         validation.List
	Organisation   *supporterdata.Organisation
	InvitedMembers []*supporterdata.MemberInvite
	Members        []*supporterdata.Member
	Form           *inviteMemberForm
}

func ManageTeamMembers(tmpl template.Template, memberStore MemberStore, generate func() (sharecodedata.PlainText, sharecodedata.Hashed), notifyClient NotifyClient, appPublicURL string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
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

				plainCode, hashedCode := generate()
				if err := memberStore.CreateMemberInvite(
					r.Context(),
					organisation,
					data.Form.FirstNames,
					data.Form.LastName,
					data.Form.Email,
					hashedCode,
					data.Form.Permission,
				); err != nil {
					return err
				}

				if err := notifyClient.SendEmail(r.Context(), notify.ToCustomEmail(localize.En, data.Form.Email), notify.OrganisationMemberInviteEmail{
					OrganisationName:      organisation.Name,
					InviterEmail:          appData.LoginSessionEmail,
					InviteCode:            plainCode.Plain(),
					JoinAnOrganisationURL: appPublicURL + page.PathSupporterStart.Format(),
				}); err != nil {
					return err
				}

				return supporter.PathManageTeamMembers.RedirectQuery(w, r, appData, url.Values{"inviteSent": {data.Form.Email}})
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
