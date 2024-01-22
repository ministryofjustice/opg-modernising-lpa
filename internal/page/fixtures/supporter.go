package fixtures

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/random"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
)

func Supporter(sessionStore sesh.Store) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		var (
			redirect = r.FormValue("redirect")

			supporterSub = random.String(16)
		)

		if err := sesh.SetLoginSession(sessionStore, r, w, &sesh.LoginSession{Sub: supporterSub, Email: testEmail}); err != nil {
			return err
		}

		http.Redirect(w, r, "/supporter/"+redirect, http.StatusFound)
		return nil
	}
}
