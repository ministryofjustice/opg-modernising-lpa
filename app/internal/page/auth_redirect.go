package page

import (
	"encoding/gob"
	"net/http"

	"github.com/gorilla/sessions"
)

type OneLoginSession struct {
	State               string
	Nonce               string
	Locale              string
	Identity            bool
	CertificateProvider bool
	SessionID           string
	LpaID               string
}

func (s OneLoginSession) Valid() bool {
	ok := s.State != "" && s.Nonce != ""
	if s.CertificateProvider {
		ok = ok && s.SessionID != "" && s.LpaID != ""
	}

	return ok
}

type DonorLoginSession struct {
	Sub   string
	Email string
}

func (s DonorLoginSession) Valid() bool {
	return s.Sub != ""
}

type CertificateProviderLoginSession struct {
	Sub       string
	Email     string
	LpaID     string
	SessionID string // this is the donor's sessionID
}

func (s CertificateProviderLoginSession) Valid() bool {
	return s.Sub != ""
}

func AuthRedirect(logger Logger, oneLoginClient OneLoginClient, store sessions.Store, secure bool) http.HandlerFunc {
	gob.Register(&OneLoginSession{})
	gob.Register(&DonorLoginSession{})
	gob.Register(&CertificateProviderLoginSession{})

	cookieOptions := &sessions.Options{
		Path:     "/",
		MaxAge:   24 * 60 * 60,
		SameSite: http.SameSiteLaxMode,
		HttpOnly: true,
		Secure:   secure,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		params, err := store.Get(r, "params")
		if err != nil {
			logger.Print(err)
			return
		}

		oneLoginSession, ok := params.Values["one-login"].(*OneLoginSession)
		if !ok || !oneLoginSession.Valid() {
			logger.Print("valid one-login session missing")
			return
		}

		if oneLoginSession.State != r.FormValue("state") {
			logger.Print("state incorrect")
			return
		}

		lang := En
		if oneLoginSession.Locale == "cy" {
			lang = Cy
		}

		appData := AppData{Lang: lang, LpaID: oneLoginSession.LpaID}

		if oneLoginSession.CertificateProvider {
			appData.Redirect(w, r, nil, Paths.CertificateProviderLoginCallback+"?"+r.URL.RawQuery)
		} else if oneLoginSession.Identity {
			appData.Redirect(w, r, nil, Paths.IdentityWithOneLoginCallback+"?"+r.URL.RawQuery)
		} else {
			accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
			if err != nil {
				logger.Print(err)
				return
			}

			userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
			if err != nil {
				logger.Print(err)
				return
			}

			session := sessions.NewSession(store, "session")
			session.Values["donor"] = &DonorLoginSession{
				Sub:   userInfo.Sub,
				Email: userInfo.Email,
			}
			session.Options = cookieOptions
			if err := store.Save(r, w, session); err != nil {
				logger.Print(err)
				return
			}

			appData.Redirect(w, r, nil, Paths.Dashboard)
		}
	}
}
