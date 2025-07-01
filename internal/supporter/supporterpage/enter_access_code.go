package supporterpage

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/supporter"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type enterAccessCode struct {
	App    appcontext.Data
	Errors validation.List
	Form   *enterAccessCodeForm
}

func EnterAccessCode(logger Logger, tmpl template.Template, memberStore MemberStore, sessionStore SessionStore) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		data := &enterAccessCode{
			App: appData,
			Form: &enterAccessCodeForm{
				FieldName: form.FieldNames.AccessCode,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterAccessCodeForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				invite, err := memberStore.InvitedMember(r.Context())
				if err != nil {
					return fmt.Errorf("get invited member: %w", err)
				}

				if invite.AccessCode != accesscodedata.HashedFromString(data.Form.AccessCode) {
					data.Errors.Add(form.FieldNames.AccessCode, validation.IncorrectError{Label: "accessCode"})
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
		}

		return tmpl(w, data)
	}
}

type enterAccessCodeForm struct {
	AccessCode    string
	AccessCodeRaw string
	FieldName     string
}

func readEnterAccessCodeForm(r *http.Request) *enterAccessCodeForm {
	return &enterAccessCodeForm{
		AccessCode:    page.PostFormReferenceNumber(r, form.FieldNames.AccessCode),
		AccessCodeRaw: page.PostFormString(r, form.FieldNames.AccessCode),
		FieldName:     form.FieldNames.AccessCode,
	}
}

func (f *enterAccessCodeForm) Validate() validation.List {
	var errors validation.List

	errors.String(form.FieldNames.AccessCode, "yourAccessCode", f.AccessCode,
		validation.Empty())

	errors.String(form.FieldNames.AccessCode, "theAccessCodeYouEnter", f.AccessCode,
		validation.StringLength(8))

	return errors
}
