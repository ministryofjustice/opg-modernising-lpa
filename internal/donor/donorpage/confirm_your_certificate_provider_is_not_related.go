package donorpage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourCertificateProviderIsNotRelatedData struct {
	App    page.AppData
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func ConfirmYourCertificateProviderIsNotRelated(tmpl template.Template, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		data := &confirmYourCertificateProviderIsNotRelatedData{
			App:   appData,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
			Donor: donor,
		}

		// To prevent going back from 'choose-new' and submitting without having
		// picked a new certificate provider
		if !donor.Tasks.CertificateProvider.Completed() {
			return page.Paths.TaskList.Redirect(w, r, appData, donor)
		}

		if r.Method == http.MethodPost {
			if r.PostFormValue("action") == "choose-new" {
				donor.CertificateProvider = donordata.CertificateProvider{}
				donor.Tasks.CertificateProvider = actor.TaskNotStarted
				if err := donorStore.Put(r.Context(), donor); err != nil {
					return err
				}

				return page.Paths.CertificateProviderDetails.Redirect(w, r, appData, donor)
			}

			data.Form = form.ReadYesNoForm(r, "theBoxToConfirmYourCertificateProviderIsNotRelated")
			data.Errors = data.Form.Validate()

			if data.Errors.None() && data.Form.YesNo.IsYes() {
				donor.CertificateProviderNotRelatedConfirmedAt = now()
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
