package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App                 page.AppData
	Errors              validation.List
	Lpa                 *lpastore.Lpa
	CertificateProvider *certificateproviderdata.Provided
}

func Guidance(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided) error {
		data := &guidanceData{
			App:                 appData,
			CertificateProvider: certificateProvider,
		}

		if lpaStoreResolvingService != nil {
			lpa, err := lpaStoreResolvingService.Get(r.Context())
			if err != nil {
				return err
			}
			data.Lpa = lpa
		}

		return tmpl(w, data)
	}
}