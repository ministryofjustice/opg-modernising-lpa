package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingYourSignatureData struct {
	App    page.AppData
	Errors validation.List
	Donor  *donordata.Provided
}

func WitnessingYourSignature(tmpl template.Template, witnessCodeSender WitnessCodeSender, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		if r.Method == http.MethodPost {
			if err := witnessCodeSender.SendToCertificateProvider(r.Context(), donor, appData.Localizer); err != nil {
				return err
			}

			if donor.Donor.CanSign.IsYes() {
				return page.Paths.WitnessingAsCertificateProvider.Redirect(w, r, appData, donor)
			} else {
				lpa, err := donorStore.Get(r.Context())
				if err != nil {
					return err
				}

				if err := witnessCodeSender.SendToIndependentWitness(r.Context(), lpa, appData.Localizer); err != nil {
					return err
				}

				return page.Paths.WitnessingAsIndependentWitness.Redirect(w, r, appData, lpa)
			}
		}

		data := &witnessingYourSignatureData{
			App:   appData,
			Donor: donor,
		}

		return tmpl(w, data)
	}
}
