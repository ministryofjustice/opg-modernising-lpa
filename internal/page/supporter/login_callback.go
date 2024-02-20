package supporter

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/onelogin"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

type LoginCallbackOneLoginClient interface {
	Exchange(ctx context.Context, code, nonce string) (idToken, accessToken string, err error)
	UserInfo(ctx context.Context, accessToken string) (onelogin.UserInfo, error)
}

func LoginCallback(oneLoginClient LoginCallbackOneLoginClient, sessionStore sesh.Store, organisationStore OrganisationStore, now func() time.Time, memberStore MemberStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		oneLoginSession, err := sesh.OneLogin(sessionStore, r)
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

		sessionData := &page.SessionData{SessionID: loginSession.SessionID(), Email: loginSession.Email}
		ctx := page.ContextWithSessionData(r.Context(), sessionData)

		_, err = memberStore.InvitedMember(ctx)
		if err == nil {
			if err := sesh.SetLoginSession(sessionStore, r, w, loginSession); err != nil {
				return err
			}

			return page.Paths.Supporter.EnterReferenceNumber.Redirect(w, r, appData)
		}

		organisation, err := organisationStore.Get(ctx)
		if err == nil {
			loginSession.OrganisationID = organisation.ID
			loginSession.OrganisationName = organisation.Name
			if err := sesh.SetLoginSession(sessionStore, r, w, loginSession); err != nil {
				return err
			}

			sessionData.OrganisationID = organisation.ID
			ctx = page.ContextWithSessionData(r.Context(), sessionData)

			member, err := memberStore.Self(ctx)
			if err != nil {
				return err
			}

			member.LastLoggedInAt = now()
			member.Email = loginSession.Email

			if err := memberStore.PutMember(ctx, member); err != nil {
				return err
			}

			return page.Paths.Supporter.Dashboard.Redirect(w, r, appData)
		}

		if errors.Is(err, dynamo.NotFoundError{}) {
			if err := sesh.SetLoginSession(sessionStore, r, w, loginSession); err != nil {
				return err
			}
		} else {
			return err
		}

		return page.Paths.Supporter.EnterOrganisationName.Redirect(w, r, appData)
	}
}
