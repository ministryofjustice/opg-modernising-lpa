package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
)

func WhatYouCanDoNowExpired(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &whatYouCanDoNowData{
			App: appData,
			Form: &whatYouCanDoNowForm{
				Options:        donordata.NoVoucherDecisionValues,
				CanHaveVoucher: provided.CanHaveVoucher(),
			},
			VouchAttempts: provided.VouchAttempts,
		}

		if r.Method == http.MethodPost {
			data.Form = readWhatYouCanDoNowForm(r, provided)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				next := handleDoNext(data.Form.DoNext, provided)

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return next.Redirect(w, r, appData, provided)
			}
		}

		data.BannerContent = "yourConfirmedIdentityHasExpired"
		data.NewVoucherLabel = "iHaveSomeoneWhoCanVouch"
		data.ProveOwnIdentityLabel = "iWillReturnToOneLogin"

		if provided.WantVoucher.IsYes() || provided.WantVoucher.IsUnknown() && provided.VouchAttempts > 0 {
			data.ProveOwnIdentityLabel = "iWillGetOrFindID"

			switch provided.VouchAttempts {
			case 0, 1:
				data.BannerContent = "yourVouchedForIdentityHasExpired"
			default:
				data.BannerContent = "yourVouchedForIdentityHasExpiredSecondAttempt"
			}
		}

		return tmpl(w, data)
	}
}
