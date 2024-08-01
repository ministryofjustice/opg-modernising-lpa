package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howWouldYouLikeToSendEvidenceData struct {
	App     page.AppData
	Errors  validation.List
	Options pay.EvidenceDeliveryOptions
}

func HowWouldYouLikeToSendEvidence(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &howWouldYouLikeToSendEvidenceData{
			App:     appData,
			Options: pay.EvidenceDeliveryValues,
		}

		if r.Method == http.MethodPost {
			form := readHowWouldYouLikeToSendEvidenceForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				donor.EvidenceDelivery = form.EvidenceDelivery

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if donor.EvidenceDelivery.IsUpload() {
					return page.Paths.UploadEvidence.Redirect(w, r, appData, donor)
				} else {
					return page.Paths.SendUsYourEvidenceByPost.Redirect(w, r, appData, donor)
				}
			}
		}

		return tmpl(w, data)
	}
}

type evidenceDeliveryForm struct {
	EvidenceDelivery pay.EvidenceDelivery
	Error            error
	ErrorLabel       string
}

func (f *evidenceDeliveryForm) Validate() validation.List {
	var errors validation.List

	errors.Error("evidence-delivery", f.ErrorLabel, f.Error,
		validation.Selected())

	return errors
}

func readHowWouldYouLikeToSendEvidenceForm(r *http.Request) *evidenceDeliveryForm {
	evidenceDelivery, err := pay.ParseEvidenceDelivery(form.PostFormString(r, "evidence-delivery"))

	return &evidenceDeliveryForm{
		EvidenceDelivery: evidenceDelivery,
		Error:            err,
		ErrorLabel:       "howYouWouldLikeToSendUsYourEvidence",
	}
}
