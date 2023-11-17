package donor

import (
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourCertificateProviderIsNotRelatedData struct {
	App    page.AppData
	Errors validation.List
	Yes    form.YesNo
	Lpa    *page.Lpa
}

func ConfirmYourCertificateProviderIsNotRelated(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, lpa *page.Lpa) error {
		data := &confirmYourCertificateProviderIsNotRelatedData{
			App: appData,
			Yes: form.Yes,
			Lpa: lpa,
		}

		if r.Method == http.MethodPost {
			form := form.ReadYesNoForm(r, "theBoxToConfirmYourCertificateProviderIsNotRelated")
			data.Errors = form.Validate()

			if data.Errors.None() && form.YesNo.IsYes() {
				lpa.Tasks.CheckYourLpa = actor.TaskInProgress

				if err := donorStore.Put(r.Context(), lpa); err != nil {
					return err
				}

				return page.Paths.CheckYourLpa.Redirect(w, r, appData, lpa)
			}
		}

		return tmpl(w, data)
	}
}
