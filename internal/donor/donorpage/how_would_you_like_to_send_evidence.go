package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
)

type howWouldYouLikeToSendEvidenceData struct {
	App  appcontext.Data
	Form *newforms.EnumForm[pay.EvidenceDelivery, pay.EvidenceDeliveryOptions, *pay.EvidenceDelivery]
}

func HowWouldYouLikeToSendEvidence(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howWouldYouLikeToSendEvidenceData{
			App:  appData,
			Form: newforms.NewEnumForm[pay.EvidenceDelivery]("howYouWouldLikeToSendUsYourEvidence", pay.EvidenceDeliveryValues),
		}

		if r.Method == http.MethodPost && data.Form.Parse(r) {
			provided.EvidenceDelivery = data.Form.Enum.Value

			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			if provided.EvidenceDelivery.IsUpload() {
				return donor.PathUploadEvidence.Redirect(w, r, appData, provided)
			} else {
				return donor.PathSendUsYourEvidenceByPost.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
