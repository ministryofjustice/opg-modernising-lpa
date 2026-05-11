package supporterpage

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter/invitecode"
)

type enterAccessCode struct {
	App  appcontext.Data
	Form *enterAccessCodeForm
}

func EnterAccessCode(logger Logger, tmpl template.Template, memberStore MemberStore, sessionStore SessionStore) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := &enterAccessCode{
			App:  appData,
			Form: newEnterAccessCodeForm(appData.Localizer),
		}

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			invite, err := memberStore.InvitedMember(r.Context())
			if err != nil {
				return fmt.Errorf("get invited member: %w", err)
			}

			if invite.InviteCode != invitecode.HashedFromString(data.Form.AccessCode.Value) {
				data.Form.AccessCode.Error = newforms.NewIncorrectError(appData.Localizer.T("accessCode"))
				data.Form.Errors = append(data.Form.Errors, data.Form.AccessCode.Field)
				return tmpl(w, data)
			}

			if invite.HasExpired() {
				return page.PathSupporterInviteExpired.Redirect(w, r, appData)
			}

			if err := memberStore.CreateFromInvite(r.Context(), invite); err != nil {
				return fmt.Errorf("create member from invite: %w", err)
			}

			loginSession, err := sessionStore.Login(r)
			if err != nil {
				return page.PathSupporterStart.Redirect(w, r, appData)
			}

			loginSession.OrganisationID = invite.OrganisationID
			loginSession.OrganisationName = invite.OrganisationName

			logger.InfoContext(r.Context(), "member invite redeemed", slog.String("organisation_id", loginSession.OrganisationID))

			if err := sessionStore.SetLogin(r, w, loginSession); err != nil {
				return fmt.Errorf("set login on session: %w", err)
			}

			return supporter.PathDashboard.Redirect(w, r, appData)
		}

		return tmpl(w, data)
	}
}

type enterAccessCodeForm struct {
	newforms.Form
	AccessCode *newforms.String
}

func newEnterAccessCodeForm(l Localizer) *enterAccessCodeForm {
	return &enterAccessCodeForm{
		AccessCode: newforms.NewString("access-code", l.T("yourAccessCode")).
			Replace(
				" ", "",
				"-", "",
			).
			NotEmpty().
			Length(8, newforms.CustomError(l.T("theAccessCodeYouEnter"))),
	}
}

func (f *enterAccessCodeForm) Parse(r *http.Request) bool {
	return f.ParsePostForm(r, f.AccessCode)
}
