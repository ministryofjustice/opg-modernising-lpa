package supporterpage

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/invitecode"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/supporterdata"
)

type inviteMemberData struct {
	App     appcontext.Data
	Form    *inviteMemberForm
	Options supporterdata.PermissionOptions
}

func InviteMember(tmpl template.Template, memberStore MemberStore, notifyClient NotifyClient, generate invitecode.Generator, appPublicURL string) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, organisation *supporterdata.Organisation, _ *supporterdata.Member) error {
		data := &inviteMemberData{
			App:     appData,
			Form:    newInviteMemberForm(appData.Localizer),
			Options: supporterdata.PermissionValues,
		}

		if r.Method == http.MethodPost && data.Form.Parse(r) {
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

		return tmpl(w, data)
	}
}

type inviteMemberForm struct {
	newforms.Form
	FirstNames *newforms.String
	LastName   *newforms.String
	Email      *newforms.String
	Permission *newforms.Enum[supporterdata.Permission, supporterdata.PermissionOptions, *supporterdata.Permission]
}

func newInviteMemberForm(l Localizer) *inviteMemberForm {
	return &inviteMemberForm{
		Email: newforms.NewString("email", l.T("email")).
			NotEmpty().
			Email(),
		FirstNames: newforms.NewString("first-names", l.T("firstNames")).
			NotEmpty().
			MaxLength(53),
		LastName: newforms.NewString("last-name", l.T("lastName")).
			NotEmpty().
			MaxLength(61),
		Permission: newforms.NewEnum[supporterdata.Permission]("permission", l.T("makeThisPersonAnAdmin"), supporterdata.PermissionValues).
			Selected(),
	}
}

func (f *inviteMemberForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r,
		f.FirstNames,
		f.LastName,
		f.Email,
		f.Permission,
	)
}
