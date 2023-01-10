package page

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
)

type identityWithOneLoginCallbackData struct {
	App             AppData
	Errors          map[string]string
	FullName        string
	ConfirmedAt     time.Time
	CouldNotConfirm bool
}

func IdentityWithOneLoginCallback(tmpl template.Template, authRedirectClient authRedirectClient, sessionStore sessions.Store, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if lpa.OneLoginUserData.OK {
				return appData.Lang.Redirect(w, r, lpa, Paths.ReadYourLpa)
			} else {
				return appData.Lang.Redirect(w, r, lpa, Paths.SelectYourIdentityOptions1)
			}
		}

		data := &identityWithOneLoginCallbackData{App: appData}

		if lpa.OneLoginUserData.OK {
			data.FullName = lpa.OneLoginUserData.FullName
			data.ConfirmedAt = lpa.OneLoginUserData.RetrievedAt
		} else {
			params, err := sessionStore.Get(r, "params")
			if err != nil {
				return err
			}

			nonce, ok := params.Values["nonce"].(string)
			if !ok {
				return fmt.Errorf("nonce missing from session")
			}

			jwt, err := authRedirectClient.Exchange(r.Context(), r.FormValue("code"), nonce)
			if err != nil {
				return err
			}

			userInfo, err := authRedirectClient.UserInfo(jwt)
			if err != nil {
				return err
			}

			if userInfo.CoreIdentityJWT == "" {
				data.CouldNotConfirm = true
			} else {
				lpa.OneLoginUserData = identity.UserData{
					OK:          true,
					RetrievedAt: time.Now(),
					FullName:    userInfo.CoreIdentityJWT, // we will parse this later
				}

				if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
					return err
				}

				data.FullName = lpa.OneLoginUserData.FullName
				data.ConfirmedAt = lpa.OneLoginUserData.RetrievedAt
			}
		}

		return tmpl(w, data)
	}
}
