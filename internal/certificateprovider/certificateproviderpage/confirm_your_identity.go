package certificateproviderpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourIdentityData struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpadata.Lpa
}

func ConfirmYourIdentity(tmpl template.Template, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		if r.Method == http.MethodPost {
			if provided.Tasks.ConfirmYourIdentity.IsNotStarted() {
				provided.Tasks.ConfirmYourIdentity = task.IdentityStateInProgress

				if err := certificateProviderStore.Put(r.Context(), provided); err != nil {
					return fmt.Errorf("error updating certificate provider: %w", err)
				}
			}

			return certificateprovider.PathIdentityWithOneLogin.Redirect(w, r, appData, provided.LpaID)
		}

		return tmpl(w, &confirmYourIdentityData{App: appData, Lpa: lpa})
	}
}
