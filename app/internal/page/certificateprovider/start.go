package certificateprovider

import (
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type startData struct {
	App    page.AppData
	Errors validation.List
	Start  string
}

func Start(tmpl template.Template, lpaStore page.LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		sessionID := r.FormValue("sessionId")
		lpaID := r.FormValue("lpaId")

		_, err := lpaStore.Get(page.ContextWithSessionData(r.Context(), &page.SessionData{
			SessionID: sessionID,
			LpaID:     lpaID,
		}))
		if err != nil {
			return err
		}

		data := &startData{
			App:   appData,
			Start: page.Paths.CertificateProviderLogin + "?" + url.Values{"lpaId": {lpaID}, "sessionId": {sessionID}}.Encode(),
		}

		return tmpl(w, data)
	}
}
