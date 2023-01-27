package page

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type identityWithOneLoginCallbackData struct {
	App             AppData
	Errors          validation.List
	FullName        string
	ConfirmedAt     time.Time
	CouldNotConfirm bool
}

func IdentityWithOneLoginCallback(tmpl template.Template, oneLoginClient OneLoginClient, sessionStore sessions.Store, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if lpa.OneLoginUserData.OK {
				return appData.Redirect(w, r, lpa, Paths.ReadYourLpa)
			} else {
				return appData.Redirect(w, r, lpa, Paths.SelectYourIdentityOptions1)
			}
		}

		data := &identityWithOneLoginCallbackData{App: appData}

		if lpa.OneLoginUserData.OK {
			data.FullName = lpa.OneLoginUserData.FullName
			data.ConfirmedAt = lpa.OneLoginUserData.RetrievedAt

			return tmpl(w, data)
		}

		if r.FormValue("error") == "access_denied" {
			data.CouldNotConfirm = true

			return tmpl(w, data)
		}

		params, err := sessionStore.Get(r, "params")
		if err != nil {
			return err
		}

		nonce, ok := params.Values["nonce"].(string)
		if !ok {
			return fmt.Errorf("nonce missing from session")
		}

		accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return err
		}

		userData, err := oneLoginClient.ParseIdentityClaim(r.Context(), userInfo)
		if err != nil {
			return err
		}

		if !userData.OK {
			data.CouldNotConfirm = true
		} else {
			data.FullName = userData.FullName
			data.ConfirmedAt = userData.RetrievedAt

			lpa.OneLoginUserData = userData

			if err := lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}
		}

		return tmpl(w, data)
	}
}
