package voucherpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type guidanceData struct {
	App      appcontext.Data
	Errors   validation.List
	Provided *voucherdata.Provided
	Lpa      *lpadata.Lpa
}

func Guidance(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		data := &guidanceData{
			App:      appData,
			Provided: provided,
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
