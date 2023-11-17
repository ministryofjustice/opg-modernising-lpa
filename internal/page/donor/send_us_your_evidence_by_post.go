package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
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

func SendUsYourEvidenceByPost(tmpl template.Template, payer Payer, eventClient EventClient) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *actor.Lpa) error {
		data := &sendUsYourEvidenceByPostData{
			App:     appData,
			FeeType: lpa.FeeType,
		}

		if r.Method == http.MethodPost {
			if err := eventClient.SendReducedFeeRequested(r.Context(), event.ReducedFeeRequested{
				UID:              lpa.UID,
				RequestType:      lpa.FeeType.String(),
				EvidenceDelivery: lpa.EvidenceDelivery.String(),
			}); err != nil {
				return err
			}

			return payer.Pay(appData, w, r, lpa)
		}

		return tmpl(w, data)
	}
}
