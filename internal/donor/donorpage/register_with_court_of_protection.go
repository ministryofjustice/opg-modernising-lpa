package donorpage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type registerWithCourtOfProtectionData struct {
	App    page.AppData
	Errors validation.List
	Form   *form.YesNoForm
}

func RegisterWithCourtOfProtection(tmpl template.Template, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &registerWithCourtOfProtectionData{
			App:  appData,
			Form: form.NewYesNoForm(form.YesNoUnknown),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "whatYouWouldLikeToDo")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.YesNo.IsYes() {
					return page.Paths.WithdrawThisLpa.Redirect(w, r, appData, donor)
				} else {
					donor.RegisteringWithCourtOfProtection = true
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.Dashboard.Redirect(w, r, appData)
			}
		}

		return tmpl(w, data)
	}
}
