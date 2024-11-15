package voucherpage

import (
	"fmt"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type confirmYourIdentityData struct {
	App    appcontext.Data
	Errors validation.List
	Lpa    *lpadata.Lpa
}

func ConfirmYourIdentity(tmpl template.Template, voucherStore VoucherStore, lpaStoreResolvingService LpaStoreResolvingService) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		if r.Method == http.MethodPost {
			if provided.Tasks.ConfirmYourIdentity.IsNotStarted() {
				provided.Tasks.ConfirmYourIdentity = task.IdentityStateInProgress

				if err := voucherStore.Put(r.Context(), provided); err != nil {
					return fmt.Errorf("error updating voucher: %w", err)
				}
			}

			return voucher.PathIdentityWithOneLogin.Redirect(w, r, appData, provided.LpaID)
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return fmt.Errorf("error retrieving lpa: %w", err)
		}

		return tmpl(w, &confirmYourIdentityData{App: appData, Lpa: lpa})
	}
}
