package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type howWouldYouLikeToSendEvidenceData struct {
	App     page.AppData
	Errors  validation.List
	Options page.EvidenceDeliveryOptions
}

func HowWouldYouLikeToSendEvidence(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howWouldYouLikeToSendEvidenceData{
			App:     appData,
			Options: page.EvidenceDeliveryValues,
		}

		if r.Method == http.MethodPost {
			form := readHowWouldYouLikeToSendEvidenceForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				lpa.EvidenceDelivery = form.EvidenceDelivery

				if form.EvidenceDelivery.IsUpload() {
					return appData.Redirect(w, r, lpa, page.Paths.UploadEvidence.Format(lpa.ID))
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.SendUsYourEvidenceByPost.Format(lpa.ID))
				}
			}
		}

		return tmpl(w, data)
	}
}

type evidenceDeliveryForm struct {
	EvidenceDelivery page.EvidenceDelivery
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
	evidenceDelivery, err := page.ParseEvidenceDelivery(form.PostFormString(r, "evidence-delivery"))

	return &evidenceDeliveryForm{
		EvidenceDelivery: evidenceDelivery,
		Error:            err,
		ErrorLabel:       "howYouWouldLikeToSendUsYourEvidence",
	}
}
