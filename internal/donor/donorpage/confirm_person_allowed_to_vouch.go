package donorpage

import (
	"net/http"
	"slices"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmPersonAllowedToVouchData struct {
	App          page.AppData
	Errors       validation.List
	Form         *form.YesNoForm
	Matches      []actor.Type
	MatchSurname bool
	FullName     string
}

func (d confirmPersonAllowedToVouchData) MultipleMatches() bool {
	count := len(d.Matches)
	if d.MatchSurname {
		count++
	}

	return count > 1
}

func ConfirmPersonAllowedToVouch(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.DonorProvidedDetails) error {
		matches := donor.Voucher.Matches(donor)

		data := &confirmPersonAllowedToVouchData{
			App:          appData,
			Form:         form.NewYesNoForm(form.YesNoUnknown),
			Matches:      matches,
			MatchSurname: strings.EqualFold(donor.Voucher.LastName, donor.Donor.LastName) && !slices.Contains(matches, actor.TypeDonor),
			FullName:     donor.Voucher.FullName(),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesIfPersonIsAllowedToVouchForYou")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				var redirect page.LpaPath
				if data.Form.YesNo.IsYes() {
					donor.Voucher.Allowed = true
					redirect = page.Paths.CheckYourDetails
				} else {
					donor.Voucher = donordata.Voucher{}
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
