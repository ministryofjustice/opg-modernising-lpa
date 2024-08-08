package certificateproviderpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type guidanceData struct {
	App                 appcontext.Data
	Errors              validation.List
	Lpa                 *lpadata.Lpa
	CertificateProvider *certificateproviderdata.Provided
}

func Guidance(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided) error {
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
