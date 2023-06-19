package page

import (
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/sesh"
)

func YotiRedirect(logger Logger, store sesh.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		yotiSession, err := sesh.Yoti(store, r)
		if err != nil {
			logger.Print(err)
			return
		}

		lang := localize.En
		if yotiSession.Locale == "cy" {
			lang = localize.Cy
		}

		appData := AppData{Lang: lang, LpaID: yotiSession.LpaID}

		redirect := Paths.IdentityWithYotiCallback.Format(yotiSession.LpaID)
		if yotiSession.CertificateProvider {
			redirect = Paths.CertificateProvider.IdentityWithYotiCallback.Format(yotiSession.LpaID)
		}
		appData.Redirect(w, r, nil, redirect+"?"+r.URL.RawQuery)
	}
}
