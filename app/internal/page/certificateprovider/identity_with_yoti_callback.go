package certificateprovider

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type identityWithYotiCallbackData struct {
	App         page.AppData
	Errors      validation.List
	FullName    string
	ConfirmedAt time.Time
}

func IdentityWithYotiCallback(tmpl template.Template, yotiClient YotiClient, lpaStore LpaStore) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStore.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			return appData.Redirect(w, r, lpa, page.Paths.CertificateProviderReadTheLpa)
		}

		data := &identityWithYotiCallbackData{App: appData}

		if lpa.CertificateProviderYotiUserData.OK {
			data.FullName = lpa.CertificateProviderYotiUserData.FullName
			data.ConfirmedAt = lpa.CertificateProviderYotiUserData.RetrievedAt
		} else {
			user, err := yotiClient.User(r.FormValue("token"))
			if err != nil {
				return err
			}

			lpa.CertificateProviderYotiUserData = user
			if err := lpaStore.Put(r.Context(), lpa); err != nil {
				return err
			}

			data.FullName = user.FullName
			data.ConfirmedAt = user.RetrievedAt
		}

		return tmpl(w, data)
	}
}
