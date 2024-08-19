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
	Tasks      voucherdata.Tasks
	FirstNames string
	LastName   string
}

func ConfirmYourName(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
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

		if r.Method == http.MethodPost {
			redirect := voucher.PathTaskList
			state := task.StateCompleted

			provided.FirstNames = firstNames
			provided.LastName = lastName

			if lastName == lpa.Donor.LastName {
				redirect = voucher.PathConfirmAllowedToVouch
				state = task.StateInProgress
			}

			provided.Tasks.ConfirmYourName = state
			if err := voucherStore.Put(r.Context(), provided); err != nil {
				return err
			}

			return redirect.Redirect(w, r, appData, appData.LpaID)
		}

		return tmpl(w, &confirmYourNameData{
			App:        appData,
			Lpa:        lpa,
			Tasks:      provided.Tasks,
			FirstNames: firstNames,
			LastName:   lastName,
		})
	}
}
