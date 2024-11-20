package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howWouldYouLikeToSendEvidenceData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.SelectForm[pay.EvidenceDelivery, pay.EvidenceDeliveryOptions, *pay.EvidenceDelivery]
}

func HowWouldYouLikeToSendEvidence(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &howWouldYouLikeToSendEvidenceData{
			App:  appData,
			Form: form.NewEmptySelectForm[pay.EvidenceDelivery](pay.EvidenceDeliveryValues, "howYouWouldLikeToSendUsYourEvidence"),
		}

		if r.Method == http.MethodPost {
			data.Form.Read(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				provided.EvidenceDelivery = data.Form.Selected

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				if provided.EvidenceDelivery.IsUpload() {
					return donor.PathUploadEvidence.Redirect(w, r, appData, provided)
				} else {
					return donor.PathSendUsYourEvidenceByPost.Redirect(w, r, appData, provided)
				}
			}
		}

		return tmpl(w, data)
	}
}
