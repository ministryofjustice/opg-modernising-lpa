package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whatIsVouchingData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
}

func WhatIsVouching(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &whatIsVouchingData{
			App:  appData,
			Form: form.NewYesNoForm(provided.WantVoucher),
		}

		if r.Method == http.MethodPost {
			f := form.ReadYesNoForm(r, "yesIfHaveSomeoneCanVouchForYou")
			data.Errors = f.Validate()

			if data.Errors.None() {
				provided.WantVoucher = f.YesNo
				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.WantVoucher.IsYes() {
					return donor.PathEnterVoucher.Redirect(w, r, appData, provided)
				} else {
					return donor.PathWhatYouCanDoNow.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
