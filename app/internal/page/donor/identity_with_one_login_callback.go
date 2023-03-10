package donor

import (
	"net/http"
	"time"

	"github.com/gorilla/sessions"
	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type identityWithOneLoginCallbackData struct {
	App             page.AppData
	Errors          validation.List
	FirstNames      string
	LastName        string
	DateOfBirth     date.Date
	ConfirmedAt     time.Time
	CouldNotConfirm bool
}

func IdentityWithOneLoginCallback(tmpl template.Template, oneLoginClient OneLoginClient, sessionStore sessions.Store, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if lpa.DonorIdentityConfirmed() {
				return appData.Redirect(w, r, lpa, page.Paths.ReadYourLpa)
			} else {
				return appData.Redirect(w, r, lpa, page.Paths.SelectYourIdentityOptions1)
			}
		}

		data := &identityWithOneLoginCallbackData{App: appData}

		if lpa.DonorIdentityConfirmed() {
			data.FirstNames = lpa.DonorIdentityUserData.FirstNames
			data.LastName = lpa.DonorIdentityUserData.LastName
			data.DateOfBirth = lpa.DonorIdentityUserData.DateOfBirth
			data.ConfirmedAt = lpa.DonorIdentityUserData.RetrievedAt

			return tmpl(w, data)
		}

		if r.FormValue("error") == "access_denied" {
			data.CouldNotConfirm = true

			return tmpl(w, data)
		}

		oneLoginSession, err := sesh.OneLogin(sessionStore, r)
		if err != nil {
			return err
		}

		accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
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

		lpa.DonorIdentityUserData = userData

		if lpa.DonorIdentityConfirmed() {
			if err := lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			data.FirstNames = userData.FirstNames
			data.LastName = userData.LastName
			data.DateOfBirth = userData.DateOfBirth
			data.ConfirmedAt = userData.RetrievedAt
		} else {
			data.CouldNotConfirm = true
		}

		return tmpl(w, data)
	}
}
