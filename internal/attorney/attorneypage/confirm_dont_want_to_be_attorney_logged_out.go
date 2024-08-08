package attorneypage

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor/actoruid"
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

func ConfirmDontWantToBeAttorneyLoggedOut(tmpl template.Template, shareCodeStore ShareCodeStore, lpaStoreResolvingService LpaStoreResolvingService, sessionStore SessionStore, notifyClient NotifyClient, appPublicURL string) page.Handler {
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
			shareCode, err := shareCodeStore.Get(r.Context(), actor.TypeAttorney, r.URL.Query().Get("referenceNumber"))
			if err != nil {
				return err
			}

			attorneyFullName, err := findAttorneyFullName(lpa, shareCode.ActorUID, shareCode.IsTrustCorporation, shareCode.IsReplacementAttorney)
			if err != nil {
				return err
			}

			email := notify.AttorneyOptedOutEmail{
				Greeting:          notifyClient.EmailGreeting(lpa),
				AttorneyFullName:  attorneyFullName,
				DonorFullName:     lpa.Donor.FullName(),
				LpaType:           appData.Localizer.T(lpa.Type.String()),
				LpaUID:            lpa.LpaUID,
				DonorStartPageURL: appPublicURL + page.PathStart.Format(),
			}

			if err := notifyClient.SendActorEmail(ctx, lpa.CorrespondentEmail(), lpa.LpaUID, email); err != nil {
				return err
			}

			if err := shareCodeStore.Delete(r.Context(), shareCode); err != nil {
				return err
			}

			return page.PathAttorneyYouHaveDecidedNotToBeAttorney.RedirectQuery(w, r, appData, url.Values{"donorFullName": {lpa.Donor.FullName()}})
		}

		return tmpl(w, data)
	}
}

func findAttorneyFullName(lpa *lpadata.Lpa, uid actoruid.UID, isTrustCorporation, isReplacement bool) (string, error) {
	if t := lpa.ReplacementAttorneys.TrustCorporation; t.UID == uid {
		return t.Name, nil
	}

	if t := lpa.Attorneys.TrustCorporation; t.UID == uid {
		return t.Name, nil
	}

	if a, ok := lpa.ReplacementAttorneys.Get(uid); ok {
		return a.FullName(), nil
	}

	if a, ok := lpa.Attorneys.Get(uid); ok {
		return a.FullName(), nil
	}

	return "", errors.New("attorney not found")
}
