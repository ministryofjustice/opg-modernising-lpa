package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

//go:generate enumerator -type EvidenceDelivery -linecomment -empty
type EvidenceDelivery uint8

const (
	Upload EvidenceDelivery = iota + 1 // upload
	Post                               // post
)

type howWouldYouLikeToSendEvidenceData struct {
	App     page.AppData
	Errors  validation.List
	Options EvidenceDeliveryOptions
}

func HowWouldYouLikeToSendEvidence(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &howWouldYouLikeToSendEvidenceData{
			App:     appData,
			Options: EvidenceDeliveryValues,
		}

		if r.Method == http.MethodPost {
			form := readHowWouldYouLikeToSendEvidenceForm(r)
			data.Errors = form.Validate()

			if data.Errors.None() {
				if form.EvidenceDelivery.IsUpload() {
					return appData.Redirect(w, r, lpa, page.Paths.UploadEvidence.Format(lpa.ID))
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.HowToEmailOrPostEvidence.Format(lpa.ID))
				}
			}
		}

		return tmpl(w, data)
	}
}

type evidenceDeliveryForm struct {
	EvidenceDelivery EvidenceDelivery
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
	evidenceDelivery, err := ParseEvidenceDelivery(form.PostFormString(r, "evidence-delivery"))

	return &evidenceDeliveryForm{
		EvidenceDelivery: evidenceDelivery,
		Error:            err,
		ErrorLabel:       "howYouWouldLikeToSendUsYourEvidence",
	}
}
