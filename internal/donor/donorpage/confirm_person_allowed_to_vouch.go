package donorpage

import (
	"net/http"
	"slices"
	"strings"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmPersonAllowedToVouchData struct {
	App          appcontext.Data
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
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		matches := provided.Voucher.Matches(provided)

		data := &confirmPersonAllowedToVouchData{
			App:          appData,
			Form:         form.NewYesNoForm(form.YesNoUnknown),
			Matches:      matches,
			MatchSurname: strings.EqualFold(provided.Voucher.LastName, provided.Donor.LastName) && !slices.Contains(matches, actor.TypeDonor),
			FullName:     provided.Voucher.FullName(),
		}

		if r.Method == http.MethodPost {
			data.Form = form.ReadYesNoForm(r, "yesIfPersonIsAllowedToVouchForYou")
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				var redirect donor.Path
				if data.Form.YesNo.IsYes() {
					provided.Voucher.Allowed = true
					redirect = page.Paths.CheckYourDetails
				} else {
					provided.Voucher = donordata.Voucher{}
					redirect = page.Paths.EnterVoucher
				}

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return redirect.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
