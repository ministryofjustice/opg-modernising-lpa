package donorpage

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type previousApplicationNumberData struct {
	App    page.AppData
	Errors validation.List
	Form   *previousApplicationNumberForm
}

func PreviousApplicationNumber(tmpl template.Template, donorStore DonorStore, eventClient EventClient) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.DonorProvidedDetails) error {
		data := &previousApplicationNumberData{
			App: appData,
			Form: &previousApplicationNumberForm{
				PreviousApplicationNumber: donor.PreviousApplicationNumber,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readPreviousApplicationNumberForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				donor.PreviousApplicationNumber = data.Form.PreviousApplicationNumber

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				if err := eventClient.SendPreviousApplicationLinked(r.Context(), event.PreviousApplicationLinked{
					UID:                       donor.LpaUID,
					PreviousApplicationNumber: donor.PreviousApplicationNumber,
				}); err != nil {
					return err
				}

				if donor.PreviousApplicationNumber[0] == '7' {
					return page.Paths.PreviousFee.Redirect(w, r, appData, donor)
				} else {
					return page.Paths.EvidenceSuccessfullyUploaded.Redirect(w, r, appData, donor)
				}
			}
		}

		return tmpl(w, data)
	}
}

type previousApplicationNumberForm struct {
	PreviousApplicationNumber string
}

func readPreviousApplicationNumberForm(r *http.Request) *previousApplicationNumberForm {
	return &previousApplicationNumberForm{
		PreviousApplicationNumber: page.PostFormString(r, "previous-application-number"),
	}
}

func (f *previousApplicationNumberForm) Validate() validation.List {
	var errors validation.List

	errors.String("previous-application-number", "previousApplicationNumber", f.PreviousApplicationNumber,
		validation.Empty(),
		validation.ReferenceNumber())

	return errors
}
