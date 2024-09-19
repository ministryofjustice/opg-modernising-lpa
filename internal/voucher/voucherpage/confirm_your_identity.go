package voucherpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type confirmYourIdentityData struct {
	App                  appcontext.Data
	Errors               validation.List
	LowConfidenceEnabled bool
	Lpa                  *lpadata.Lpa
}

func ConfirmYourIdentity(tmpl template.Template, lowConfidenceEnabled bool, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, _ *voucherdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		return tmpl(w, &confirmYourIdentityData{
			App:                  appData,
			LowConfidenceEnabled: lowConfidenceEnabled,
			Lpa:                  lpa,
		})
	}
}
