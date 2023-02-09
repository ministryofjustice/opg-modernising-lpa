package page

import (
	"fmt"
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
		// TODO we should come up with a different way of linking LPAs. I think
		// adding a "shareId=>[sessionId, lpaId]" record to dynamo might work
		// better?
		sessionID := r.FormValue("sessionId")
		lpaID := r.FormValue("lpaId")

		lpa, err := lpaStore.Get(contextWithSessionData(r.Context(), &sessionData{
			SessionID: sessionID,
			LpaID:     lpaID,
		}))
		if err != nil {
			return err
		}
		if lpa == nil {
			return fmt.Errorf("lpa does not exist")
		}

		data := &certificateProviderStartData{
			App:   appData,
			Start: Paths.CertificateProviderLogin + "?" + url.Values{"lpaId": {lpaID}, "sessionId": {sessionID}}.Encode(),
		}

		return tmpl(w, data)
	}
}
