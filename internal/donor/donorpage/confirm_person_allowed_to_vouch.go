package donorpage

import (
	"net/http"
	"slices"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/names"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmPersonAllowedToVouchData struct {
	App          appcontext.Data
	Errors       validation.List
	Form         *newforms.YesNoForm
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
			Form:         newforms.NewYesNoForm(appData.Localizer.T("yesIfPersonIsAllowedToVouchForYou")),
			Matches:      matches,
			MatchSurname: names.Equal(provided.Voucher.LastName, provided.Donor.LastName) && !slices.Contains(matches, actor.TypeDonor),
			FullName:     provided.Voucher.FullName(),
		}

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				var redirect donor.Path
				if data.Form.YesNo.Value.IsYes() {
					provided.Voucher.Allowed = true
					redirect = donor.PathCheckYourDetails
				} else {
					provided.Voucher = donordata.Voucher{}
					redirect = donor.PathEnterVoucher
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
