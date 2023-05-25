package page

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func SignOut(logger Logger, sessionStore sesh.Store, oneLoginClient OneLoginClient, appPublicURL string) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		var postLogoutURL, idToken string

		if session, err := sesh.Donor(sessionStore, r); err == nil && session != nil {
			postLogoutURL = Paths.Start
			idToken = session.IDToken
		} else if session, err := sesh.CertificateProvider(sessionStore, r); err == nil && session != nil {
			postLogoutURL = Paths.CertificateProviderStart
			idToken = session.IDToken
		} else if session, err := sesh.Attorney(sessionStore, r); err == nil && session != nil {
			postLogoutURL = Paths.Attorney.Start
			idToken = session.IDToken
		}

		if err := sesh.ClearSession(sessionStore, r, w); err != nil {
			logger.Print(fmt.Sprintf("unable to expire session: %s", err.Error()))
		}

		http.Redirect(w, r, oneLoginClient.EndSessionURL(idToken, appPublicURL+postLogoutURL), http.StatusFound)
		return nil
	}
}
