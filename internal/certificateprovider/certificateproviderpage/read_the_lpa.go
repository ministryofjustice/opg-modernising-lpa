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

type readTheLpaData struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpadata.Lpa
}

func ReadTheLpa(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		if r.Method == http.MethodPost {
			if !lpa.SignedForDonor() || !lpa.Paid {
				return certificateprovider.PathTaskList.Redirect(w, r, appData, lpa.LpaID)
			}

			certificateProvider.Tasks.ReadTheLpa = task.StateCompleted
			if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
				return err
			}

			return certificateprovider.PathWhatHappensNext.Redirect(w, r, appData, lpa.LpaID)
		}

		data := &readTheLpaData{
			App: appData,
			Lpa: lpa,
		}

		return tmpl(w, data)
	}
}
