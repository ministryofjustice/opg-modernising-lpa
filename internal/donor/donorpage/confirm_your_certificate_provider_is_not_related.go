package donorpage

import (
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/form"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type confirmYourCertificateProviderIsNotRelatedData struct {
	App    appcontext.Data
	Errors validation.List
	Form   *form.YesNoForm
	Donor  *donordata.Provided
}

func ConfirmYourCertificateProviderIsNotRelated(tmpl template.Template, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		data := &confirmYourCertificateProviderIsNotRelatedData{
			App:   appData,
			Form:  form.NewYesNoForm(form.YesNoUnknown),
			Donor: provided,
		}

		// To prevent going back from 'choose-new' and submitting without having
		// picked a new certificate provider
		if !provided.Tasks.CertificateProvider.Completed() {
			return donor.PathTaskList.Redirect(w, r, appData, provided)
		}

		if r.Method == http.MethodPost {
			if r.PostFormValue("action") == "choose-new" {
				provided.CertificateProvider = donordata.CertificateProvider{}
				provided.Tasks.CertificateProvider = task.StateNotStarted
				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathCertificateProviderDetails.Redirect(w, r, appData, provided)
			}

			data.Form = form.ReadYesNoForm(r, "theBoxToConfirmYourCertificateProviderIsNotRelated")
			data.Errors = data.Form.Validate()

			if data.Errors.None() && data.Form.YesNo.IsYes() {
				provided.CertificateProviderNotRelatedConfirmedAt = now()
				provided.Tasks.CheckYourLpa = task.StateInProgress

				if err := donorStore.Put(r.Context(), provided); err != nil {
					return err
				}

				return donor.PathCheckYourLpa.Redirect(w, r, appData, provided)
			}
		}

		return tmpl(w, data)
	}
}
