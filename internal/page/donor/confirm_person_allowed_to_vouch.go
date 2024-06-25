package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmPersonAllowedToVouchData struct {
	App      page.AppData
	Errors   validation.List
	Form     *form.YesNoForm
	Matches  []actor.Type
	FullName string
}

func ConfirmPersonAllowedToVouch(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &confirmPersonAllowedToVouchData{
			App:      appData,
			Form:     form.NewYesNoForm(form.YesNoUnknown),
			Matches:  donor.Voucher.Matches(donor),
			FullName: donor.Voucher.FullName(),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesIfPersonIsAllowedToVouchForYou")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				var redirect page.LpaPath
				if data.Form.YesNo.IsYes() {
					donor.Voucher.Allowed = true
					redirect = page.Paths.TaskList
				} else {
					donor.Voucher = actor.Voucher{}
					redirect = page.Paths.EnterVoucher
				}

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return redirect.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
