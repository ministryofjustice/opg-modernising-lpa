package voucherpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type confirmYourNameData struct {
	App        appcontext.Data
	Errors     validation.List
	Lpa        *lpadata.Lpa
	FirstNames string
	LastName   string
}

func ConfirmYourName(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		if r.Method == http.MethodPost {
			if provided.Tasks.ConfirmYourName.NotStarted() {
				provided.Tasks.ConfirmYourName = task.StateInProgress

				if err := voucherStore.Put(r.Context(), provided); err != nil {
					return err
				}
			}

			return voucher.PathTaskList.Redirect(w, r, appData, appData.LpaID)
		}

		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		firstNames := provided.FirstNames
		if firstNames == "" {
			firstNames = lpa.Voucher.FirstNames
		}

		lastName := provided.LastName
		if lastName == "" {
			lastName = lpa.Voucher.LastName
		}

		return tmpl(w, &confirmYourNameData{
			App:        appData,
			Lpa:        lpa,
			FirstNames: firstNames,
			LastName:   lastName,
		})
	}
}
