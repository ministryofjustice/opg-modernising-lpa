package voucherpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type verifyDonorDetailsData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Lpa    *lpadata.Lpa
}

func VerifyDonorDetails(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *voucherdata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &verifyDonorDetailsData{
			App:  appData,
			Form: form.NewYesNoForm(form.YesNoUnknown),
			Lpa:  lpa,
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesIfTheseDetailsMatch")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.DonorDetailsMatch = data.Form.YesNo
				provided.Tasks.VerifyDonorDetails = task.StateCompleted
				if err := voucherStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if data.Form.YesNo.IsNo() {
					return voucher.PathDonorDetailsDoNotMatch.Redirect(w, r, appData, appData.LpaID)
				}

				return voucher.PathTaskList.Redirect(w, r, appData, appData.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
