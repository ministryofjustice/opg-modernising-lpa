package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
)

func WhatYouCanDoNow(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &enterVoucherData{
			App: appData,
			Form: &enterVoucherForm{
				FirstNames: donor.Voucher.FirstNames,
				LastName:   donor.Voucher.LastName,
				Email:      donor.Voucher.Email,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readEnterVoucherForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.Voucher.Email = data.Form.Email

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.CheckYourDetails.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
