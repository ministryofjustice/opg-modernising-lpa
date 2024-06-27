package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type whatIsVouchingData struct {
	App    page.AppData
	Errors validation.List
	Form   *form.YesNoForm
}

func WhatIsVouching(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &whatIsVouchingData{
			App:  appData,
			Form: form.NewYesNoForm(donor.WantVoucher),
		}

		if r.Method == http.MethodPost {
			f := form.ReadYesNoForm(r, "yesIfHaveSomeoneCanVouchForYou")
			data.Errors = f.Validate()

			if data.Errors.None() {
				donor.WantVoucher = f.YesNo
				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if donor.WantVoucher.IsYes() {
					return page.Paths.EnterVoucher.Redirect(w, r, appData, donor)
				} else {
					return page.Paths.WhatYouCanDoNow.Redirect(w, r, appData, donor)
				}
			}
		}

		return tmpl(w, data)
	}
}
