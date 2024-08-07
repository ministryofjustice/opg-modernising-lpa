package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type witnessingYourSignatureData struct {
	App    appcontext.Data
	Errors validation.List
	Donor  *donordata.Provided
}

func WitnessingYourSignature(tmpl template.Template, witnessCodeSender WitnessCodeSender, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if r.Method == http.MethodPost {
			if err := witnessCodeSender.SendToCertificateProvider(r.Context(), provided, appData.Localizer); err != nil {
				return err
			}

			if provided.Donor.CanSign.IsYes() {
				return donor.PathWitnessingAsCertificateProvider.Redirect(w, r, appData, provided)
			} else {
				lpa, err := donorStore.Get(r.Context())
				if err != nil {
					return err
				}

				if err := witnessCodeSender.SendToIndependentWitness(r.Context(), lpa, appData.Localizer); err != nil {
					return err
				}

				return donor.PathWitnessingAsIndependentWitness.Redirect(w, r, appData, lpa)
			}
		}

		data := &witnessingYourSignatureData{
			App:   appData,
			Donor: provided,
		}

		return tmpl(w, data)
	}
}
