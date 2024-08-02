package supporter

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type LoginCallbackOneLoginClient interface {
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

func LoginCallback(logger Logger, oneLoginClient LoginCallbackOneLoginClient, sessionStore SessionStore, organisationStore OrganisationStore, now func() time.Time, memberStore MemberStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		if error := r.FormValue("error"); error != "" {
			logger.InfoContext(r.Context(), "login error",
				slog.String("error", error),
				slog.String("error_description", r.FormValue("error_description")))
			return errors.New("access denied")
		}

		oneLoginSession, err := sessionStore.OneLogin(r)
		if err != nil {
			return err
		}

		idToken, accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return err
		}

		loginSession := &sesh.LoginSession{
			IDToken: idToken,
			Sub:     "supporter-" + userInfo.Sub,
			Email:   userInfo.Email,
		}

		logger.InfoContext(r.Context(), "login", slog.String("session_id", loginSession.SessionID()))

		sessionData := &appcontext.Session{SessionID: loginSession.SessionID(), Email: loginSession.Email}
		ctx := appcontext.ContextWithSession(r.Context(), sessionData)

		member, err := memberStore.GetAny(ctx)
		if errors.Is(err, dynamo.NotFoundError{}) {
			invites, err := memberStore.InvitedMembersByEmail(ctx)
			if err != nil {
				return err
			}

			if err := sessionStore.SetLogin(r, w, loginSession); err != nil {
				return err
			}

			if len(invites) > 0 {
				return page.Paths.Supporter.EnterReferenceNumber.Redirect(w, r, appData)
			}

			return page.Paths.Supporter.EnterYourName.Redirect(w, r, appData)
		} else if err != nil {
			return err
		}

		organisation, err := organisationStore.Get(ctx)
		if errors.Is(err, dynamo.NotFoundError{}) {
			if err := sessionStore.SetLogin(r, w, loginSession); err != nil {
				return err
			}

			return page.Paths.Supporter.EnterOrganisationName.Redirect(w, r, appData)
		} else if err != nil {
			return err
		}

		loginSession.OrganisationID = organisation.ID
		loginSession.OrganisationName = organisation.Name
		if err := sessionStore.SetLogin(r, w, loginSession); err != nil {
			return err
		}

		sessionData.OrganisationID = organisation.ID
		ctx = appcontext.ContextWithSession(r.Context(), sessionData)

		member.LastLoggedInAt = now()
		member.Email = loginSession.Email

		if err := memberStore.Put(ctx, member); err != nil {
			return err
		}

		return page.Paths.Supporter.Dashboard.Redirect(w, r, appData)
	}
}
