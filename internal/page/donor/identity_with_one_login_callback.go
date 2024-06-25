package donor

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/date"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/identity"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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

func IdentityWithOneLoginCallback(tmpl template.Template, oneLoginClient OneLoginClient, sessionStore SessionStore, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		if r.Method == http.MethodPost {
			if donor.DonorIdentityConfirmed() {
				return page.Paths.ReadYourLpa.Redirect(w, r, appData, donor)
			} else {
				return page.Paths.ProveYourIdentity.Redirect(w, r, appData, donor)
			}
		}

		data := &identityWithOneLoginCallbackData{App: appData}

		if donor.DonorIdentityConfirmed() {
			data.FirstNames = donor.DonorIdentityUserData.FirstNames
			data.LastName = donor.DonorIdentityUserData.LastName
			data.DateOfBirth = donor.DonorIdentityUserData.DateOfBirth
			data.ConfirmedAt = donor.DonorIdentityUserData.RetrievedAt

			return tmpl(w, data)
		}

		if r.FormValue("error") == "access_denied" {
			data.CouldNotConfirm = true

			return tmpl(w, data)
		}

		oneLoginSession, err := sessionStore.OneLogin(r)
		if err != nil {
			return err
		}

		_, accessToken, err := oneLoginClient.Exchange(r.Context(), r.FormValue("code"), oneLoginSession.Nonce)
		if err != nil {
			return err
		}

		userInfo, err := oneLoginClient.UserInfo(r.Context(), accessToken)
		if err != nil {
			return err
		}

		if len(userInfo.ReturnCodes) > 0 {
			for _, c := range userInfo.ReturnCodes {
				if c.Code == "X" {
					donor.DonorIdentityUserData = identity.UserData{OK: false, InsufficientEvidence: true}
					if err := donorStore.Put(r.Context(), donor); err != nil {
						return err
					}

					return page.Paths.UnableToConfirmIdentity.Redirect(w, r, appData, donor)
				}
			}

			data.CouldNotConfirm = true
			return tmpl(w, data)
		}

		userData, err := oneLoginClient.ParseIdentityClaim(r.Context(), userInfo)
		if err != nil {
			return err
		}

		donor.DonorIdentityUserData = userData

		if donor.DonorIdentityConfirmed() {
			data.FirstNames = userData.FirstNames
			data.LastName = userData.LastName
			data.DateOfBirth = userData.DateOfBirth
			data.ConfirmedAt = userData.RetrievedAt

			if err := donorStore.Put(r.Context(), donor); err != nil {
				return err
			}
		} else {
			data.CouldNotConfirm = true
		}

		return tmpl(w, data)
	}
}
