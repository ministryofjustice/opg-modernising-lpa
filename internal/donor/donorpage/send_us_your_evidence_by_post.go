package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type sendUsYourEvidenceByPostData struct {
	App     page.AppData
	Errors  validation.List
	FeeType pay.FeeType
}

func SendUsYourEvidenceByPost(tmpl template.Template, payer Handler, eventClient EventClient) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.DonorProvidedDetails) error {
		data := &sendUsYourEvidenceByPostData{
			App:     appData,
			FeeType: donor.FeeType,
		}

		if r.Method == http.MethodPost {
			if err := eventClient.SendReducedFeeRequested(r.Context(), event.ReducedFeeRequested{
				UID:              donor.LpaUID,
				RequestType:      donor.FeeType.String(),
				EvidenceDelivery: donor.EvidenceDelivery.String(),
			}); err != nil {
				return err
			}

			return payer(appData, w, r, donor)
		}

		return tmpl(w, data)
	}
}
