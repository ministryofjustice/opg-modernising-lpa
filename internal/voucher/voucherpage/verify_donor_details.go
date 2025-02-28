package voucherpage

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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

func VerifyDonorDetails(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, voucherStore VoucherStore, fail vouchFailer, donorStore DonorStore) Handler {
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
					return fmt.Errorf("error updating voucher: %w", err)
				}

				donor, err := donorStore.GetAny(r.Context())
				if err != nil {
					return fmt.Errorf("error getting donor: %w", err)
				}

				donor.VouchAttempts++

				if data.Form.YesNo.IsYes() {
					donor.DetailsVerifiedByVoucher = true
				}

				if err = donorStore.Put(r.Context(), donor); err != nil {
					return fmt.Errorf("error updating donor: %w", err)
				}

				if data.Form.YesNo.IsNo() {
					if err := fail(r.Context(), provided, lpa); err != nil {
						return fmt.Errorf("error failing voucher: %w", err)
					}

					return page.PathVoucherDonorDetailsDoNotMatch.RedirectQuery(w, r, appData, url.Values{
						"donorFullName":   {lpa.Donor.FullName()},
						"donorFirstNames": {lpa.Donor.FirstNames},
					})
				}

				return voucher.PathTaskList.Redirect(w, r, appData, appData.LpaID)
			}
		}

		return tmpl(w, data)
	}
}
