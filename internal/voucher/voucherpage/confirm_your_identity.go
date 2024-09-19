package voucherpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type confirmYourIdentity struct {
	App                  appcontext.Data
	Errors               validation.List
	LowConfidenceEnabled bool
}

func ConfirmYourIdentity(tmpl template.Template, lowConfidenceEnabled bool) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, _ *voucherdata.Provided) error {
		return tmpl(w, &confirmYourIdentity{
			App:                  appData,
			LowConfidenceEnabled: lowConfidenceEnabled,
		})
	}
}
