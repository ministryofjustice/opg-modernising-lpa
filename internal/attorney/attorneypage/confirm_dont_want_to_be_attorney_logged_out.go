package attorneypage

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/accesscode/accesscodedata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmDontWantToBeAttorneyDataLoggedOut struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpadata.Lpa
}

func ConfirmDontWantToBeAttorneyLoggedOut(tmpl template.Template, accessCodeStore AccessCodeStore, lpaStoreResolvingService LpaStoreResolvingService, sessionStore SessionStore, notifyClient NotifyClient, lpaStoreClient LpaStoreClient) page.Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request) error {
		session, err := sessionStore.LpaData(r)
		if err != nil {
			return err
		}

		ctx := appcontext.ContextWithSession(r.Context(), &appcontext.Session{LpaID: session.LpaID})

		lpa, err := lpaStoreResolvingService.Get(ctx)
		if err != nil {
			return err
		}

		data := &confirmDontWantToBeAttorneyDataLoggedOut{
			App: appData,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			code := accesscodedata.HashedFromQuery(r.URL.Query())

			accessCode, err := accessCodeStore.Get(r.Context(), actor.TypeAttorney, code)
			if err != nil {
				return err
			}

			fullName, _, actorType := lpa.Attorney(accessCode.ActorUID)
			if actorType.IsNone() {
				return errors.New("attorney not found")
			}

			email := notify.AttorneyOptedOutEmail{
				Greeting:           notifyClient.EmailGreeting(lpa),
				AttorneyFullName:   fullName,
				LpaType:            appData.Localizer.T(lpa.Type.String()),
				LpaReferenceNumber: lpa.LpaUID,
			}

			if err := notifyClient.SendActorEmail(ctx, notify.ToLpaDonor(lpa), lpa.LpaUID, email); err != nil {
				return err
			}

			if err := lpaStoreClient.SendAttorneyOptOut(r.Context(), lpa.LpaUID, accessCode.ActorUID, actorType); err != nil {
				return err
			}

			if err := accessCodeStore.Delete(r.Context(), accessCode); err != nil {
				return err
			}

			return page.PathAttorneyYouHaveDecidedNotToBeAttorney.RedirectQuery(w, r, appData, url.Values{"donorFullName": {lpa.Donor.FullName()}})
		}

		return tmpl(w, data)
	}
}
