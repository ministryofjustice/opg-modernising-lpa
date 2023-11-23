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
	Donor  *actor.DonorProvidedDetails
}

func ConfirmYourCertificateProviderIsNotRelated(tmpl template.Template, donorStore DonorStore) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *actor.DonorProvidedDetails) error {
		data := &confirmYourCertificateProviderIsNotRelatedData{
			App:   appData,
			Yes:   form.Yes,
			Donor: donor,
		}

		// To prevent going back from 'choose-new' and submitting without having
		// picked a new certificate provider
		if !donor.Tasks.CertificateProvider.Completed() {
			return page.Paths.TaskList.Redirect(w, r, appData, donor)
		}

		if r.Method == http.MethodPost {
			if r.PostFormValue("action") == "choose-new" {
				donor.CertificateProvider = actor.CertificateProvider{}
				donor.Tasks.CertificateProvider = actor.TaskNotStarted
				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.CertificateProviderDetails.Redirect(w, r, appData, donor)
			}

			form := form.ReadYesNoForm(r, "theBoxToConfirmYourCertificateProviderIsNotRelated")
			data.Errors = form.Validate()

			if data.Errors.None() && form.YesNo.IsYes() {
				donor.Tasks.CheckYourLpa = actor.TaskInProgress

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.CheckYourLpa.Redirect(w, r, appData, donor)
			}
		}

		return tmpl(w, data)
	}
}
