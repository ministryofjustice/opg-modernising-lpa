package page

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type certificateProviderStartData struct {
	App    AppData
	Errors validation.List
	Start  string
}

func CertificateProviderStart(tmpl template.Template, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		sessionID := r.FormValue("sessionId")
		lpaID := r.FormValue("lpaId")

		_, err := lpaStore.Get(contextWithSessionData(r.Context(), &sessionData{
			SessionID: sessionID,
			LpaID:     lpaID,
		}))
		if err != nil {
			return err
		}

		data := &certificateProviderStartData{
			App:   appData,
			Start: Paths.CertificateProviderLogin + "?" + url.Values{"lpaId": {lpaID}, "sessionId": {sessionID}}.Encode(),
		}

		return tmpl(w, data)
	}
}
