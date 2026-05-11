package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
)

type chooseSomeoneToVouchForYouData struct {
	App  appcontext.Data
	Form *newforms.YesNoForm
}

func ChooseSomeoneToVouchForYou(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &chooseSomeoneToVouchForYouData{
			App:  appData,
			Form: newforms.NewYesNoForm(appData.Localizer.T("yesIfHaveSomeoneCanVouchForYou")),
		}

		data.Form.YesNo.SetInput(provided.WantVoucher)

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
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
