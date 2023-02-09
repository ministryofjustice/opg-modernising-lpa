package page

import (
	"errors"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type certificateProviderLoginCallbackData struct {
	App             AppData
	Errors          validation.List
	FullName        string
	ConfirmedAt     time.Time
	CouldNotConfirm bool
}

func CertificateProviderLoginCallback(tmpl template.Template, oneLoginClient OneLoginClient, sessionStore sessions.Store, lpaStore LpaStore) Handler {
	cookieOptions := &sessions.Options{
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   true,
	}

	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		params, err := sessionStore.Get(r, "params")
		if err != nil {
			return err
		}

		oneLoginSession, ok := params.Values["one-login"].(*OneLoginSession)
		if !ok {
			return errors.New("one-login session missing")
		}
		if !oneLoginSession.Valid() {
			return errors.New("one-login session invalid")
		}

		ctx := contextWithSessionData(r.Context(), &sessionData{
			SessionID: oneLoginSession.SessionID,
			LpaID:     oneLoginSession.LpaID,
		})

		lpa, err := lpaStore.Get(ctx)
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if lpa.CertificateProviderUserData.OK {
				return appData.Redirect(w, r, lpa, Paths.CertificateProviderYourDetails)
			} else {
				return appData.Redirect(w, r, lpa, Paths.Start)
			}
		}

		data := &certificateProviderLoginCallbackData{App: appData}

		if lpa.CertificateProviderUserData.OK {
			data.FullName = lpa.CertificateProviderUserData.FullName
			data.ConfirmedAt = lpa.CertificateProviderUserData.RetrievedAt

			return tmpl(w, data)
		}

		if r.FormValue("error") == "access_denied" {
			data.CouldNotConfirm = true

			return tmpl(w, data)
		}

		accessToken, err := oneLoginClient.Exchange(ctx, r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(ctx, accessToken)
		if err != nil {
			return err
		}

		userData, err := oneLoginClient.ParseIdentityClaim(ctx, userInfo)
		if err != nil {
			return err
		}

		if !userData.OK {
			data.CouldNotConfirm = true
		} else {
			session := sessions.NewSession(sessionStore, "session")
			session.Values["certificate-provider"] = &CertificateProviderLoginSession{
				Sub:       userInfo.Sub,
				Email:     userInfo.Email,
				LpaID:     oneLoginSession.LpaID,
				SessionID: oneLoginSession.SessionID,
			}
			session.Options = cookieOptions
			if err := sessionStore.Save(r, w, session); err != nil {
				return err
			}

			data.FullName = userData.FullName
			data.ConfirmedAt = userData.RetrievedAt

			lpa.CertificateProviderUserData = userData

			if err := lpaStore.Put(ctx, lpa); err != nil {
				return err
			}
		}

		return tmpl(w, data)
	}
}
