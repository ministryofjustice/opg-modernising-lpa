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

func Start(tmpl template.Template, lpaStore LpaStore, dataStore page.DataStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		shareCode := r.FormValue("share-code")

		var v page.ShareCodeData
		if err := dataStore.Get(r.Context(), "SHARECODE#"+shareCode, "#METADATA#"+shareCode, &v); err != nil {
			return err
		}

		_, err := lpaStore.Get(page.ContextWithSessionData(r.Context(), &page.SessionData{
			SessionID: v.SessionID,
			LpaID:     v.LpaID,
		}))
		if err != nil {
			return err
		}

		query := url.Values{
			"lpaId":     {v.LpaID},
			"sessionId": {v.SessionID},
		}
		if v.Identity {
			query.Add("identity", "1")
		}

		data := &startData{
			App:   appData,
			Start: page.Paths.CertificateProviderLogin + "?" + query.Encode(),
		}

		return tmpl(w, data)
	}
}
