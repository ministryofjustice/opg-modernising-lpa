package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/validation"
)

type canEvidenceBeUploadedData struct {
	App     page.AppData
	Errors  validation.List
	Options form.YesNoOptions
}

func CanEvidenceBeUploaded(tmpl template.Template) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &canEvidenceBeUploadedData{
			App:     appData,
			Options: form.YesNoValues,
		}

		if r.Method == http.MethodPost {
			form := form.ReadYesNoForm(r, "canEvidenceBeUploaded")
			data.Errors = form.Validate()

			if data.Errors.None() {
				if form.YesNo.IsYes() {
					return appData.Redirect(w, r, lpa, page.Paths.UploadInstructions.Format(lpa.ID))
				} else {
					return appData.Redirect(w, r, lpa, page.Paths.PrintEvidenceForm.Format(lpa.ID))
				}
			}
		}

		return tmpl(w, data)
	}
}
