package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type unableToConfirmIdentityData struct {
	App    appcontext.Data
	Donor  lpadata.Donor
	Errors validation.List
}

func UnableToConfirmIdentity(tmpl template.Template, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		if r.Method == http.MethodPost {
			certificateProvider.Tasks.ConfirmYourIdentity = task.IdentityStateCompleted

			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			return certificateprovider.PathReadTheLpa.Redirect(w, r, appData, certificateProvider.LpaID)
		}

		return tmpl(w, &unableToConfirmIdentityData{
			App:   appData,
			Donor: lpa.Donor,
		})
	}
}
