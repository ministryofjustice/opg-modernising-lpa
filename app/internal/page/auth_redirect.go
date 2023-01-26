package page

import (
	"net/http"

	"github.com/gorilla/sessions"
)

func AuthRedirect(logger Logger, oneLoginClient OneLoginClient, store sessions.Store, secure bool) http.HandlerFunc {
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

		if s, ok := params.Values["state"].(string); !ok || s != r.FormValue("state") {
			logger.Print("state missing from session or incorrect")
			return
		}

		nonce, ok := params.Values["nonce"].(string)
		if !ok {
			logger.Print("nonce missing from session")
			return
		}

		locale, ok := params.Values["locale"].(string)
		if !ok {
			logger.Print("locale missing from session")
			return
		}

		identity, _ := params.Values["identity"].(bool)
		lpaID, _ := params.Values["lpa-id"].(string)

		lang := En
		if locale == "cy" {
			lang = Cy
		}

		appData := AppData{Lang: lang, LpaID: lpaID}

		if identity {
			appData.Redirect(w, r, nil, Paths.IdentityWithOneLoginCallback+"?"+r.URL.RawQuery)
		} else {
			accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), nonce)
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
			session.Values = map[interface{}]interface{}{
				"sub":   userInfo.Sub,
				"email": userInfo.Email,
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
