package page

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
)

type identityWithYotiCallbackData struct {
	App         AppData
	Errors      map[string]string
	FullName    string
	ConfirmedAt time.Time
}

func IdentityWithYotiCallback(tmpl template.Template, yotiClient YotiClient, lpaStore LpaStore) Handler {
	return func(appData AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context(), appData.SessionID)
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			return appData.Lang.Redirect(w, r, lpa.IdentityOptions.NextPath(Yoti, appData.Paths), http.StatusFound)
		}

		data := &identityWithYotiCallbackData{App: appData}

		if lpa.YotiUserData.OK {
			data.FullName = lpa.YotiUserData.FullName
			data.ConfirmedAt = lpa.YotiUserData.RetrievedAt
		} else {
			user, err := yotiClient.User(r.FormValue("token"))
			if err != nil {
				return err
			}

			lpa.YotiUserData = user
			if err := lpaStore.Put(r.Context(), appData.SessionID, lpa); err != nil {
				return err
			}

			data.FullName = user.FullName
			data.ConfirmedAt = user.RetrievedAt
		}

		return tmpl(w, data)
	}
}
