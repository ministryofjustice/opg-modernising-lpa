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
	App         page.AppData
	Errors      validation.List
	FirstNames  string
	LastName    string
	DateOfBirth date.Date
	ConfirmedAt time.Time
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

		userData, err := oneLoginClient.ParseIdentityClaim(r.Context(), userInfo)
		if err != nil {
			return err
		}

		donor.DonorIdentityUserData = userData

		if userData.Status.IsFailed() {
			donor.Tasks.ConfirmYourIdentityAndSign = actor.IdentityTaskProblem
		} else {
			donor.Tasks.ConfirmYourIdentityAndSign = actor.IdentityTaskInProgress
		}

		if err := donorStore.Put(r.Context(), donor); err != nil {
			return err
		}

		switch donor.DonorIdentityUserData.Status {
		case identity.StatusFailed:
			return page.Paths.RegisterWithCourtOfProtection.Redirect(w, r, appData, donor)
		case identity.StatusInsufficientEvidence:
			return page.Paths.UnableToConfirmIdentity.Redirect(w, r, appData, donor)
		}

		data.FirstNames = userData.FirstNames
		data.LastName = userData.LastName
		data.DateOfBirth = userData.DateOfBirth
		data.ConfirmedAt = userData.RetrievedAt

		return tmpl(w, data)
	}
}
