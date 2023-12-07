package certificateprovider

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type provideCertificateData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider *actor.CertificateProviderProvidedDetails
	Donor               *actor.DonorProvidedDetails
	Form                *provideCertificateForm
}

func ProvideCertificate(tmpl template.Template, donorStore DonorStore, now func() time.Time, certificateProviderStore CertificateProviderStore, notifyClient NotifyClient, shareCodeSender ShareCodeSender) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		donor, err := donorStore.GetAny(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		if donor.SignedAt.IsZero() {
			return page.Paths.CertificateProvider.TaskList.Redirect(w, r, appData, donor.LpaID)
		}

		data := &provideCertificateData{
			App:                 appData,
			CertificateProvider: certificateProvider,
			Donor:               donor,
			Form: &provideCertificateForm{
				AgreeToStatement: certificateProvider.Certificate.AgreeToStatement,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readProvideCertificateForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				certificateProvider.Certificate.AgreeToStatement = true
				certificateProvider.Certificate.Agreed = now()
				certificateProvider.Tasks.ProvideTheCertificate = actor.TaskCompleted
				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				if _, err := notifyClient.SendEmail(r.Context(), donor.CertificateProvider.Email, notify.CertificateProviderCertificateProvidedEmail{
					DonorFullNamePossessive:     appData.Localizer.Possessive(donor.Donor.FullName()),
					DonorFirstNamesPossessive:   appData.Localizer.Possessive(donor.Donor.FirstNames),
					LpaType:                     appData.Localizer.T(donor.Type.LegalTermTransKey()),
					CertificateProviderFullName: donor.CertificateProvider.FullName(),
					CertificateProvidedDateTime: appData.Localizer.FormatDateTime(certificateProvider.Certificate.Agreed),
				}); err != nil {
					return fmt.Errorf("email failed: %w", err)
				}

				if err := shareCodeSender.SendAttorneys(r.Context(), appData, donor); err != nil {
					return err
				}

				return page.Paths.CertificateProvider.CertificateProvided.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

type provideCertificateForm struct {
	AgreeToStatement bool
}

func readProvideCertificateForm(r *http.Request) *provideCertificateForm {
	return &provideCertificateForm{
		AgreeToStatement: page.PostFormString(r, "agree-to-statement") == "1",
	}
}

func (f *provideCertificateForm) Validate() validation.List {
	var errors validation.List

	errors.Bool("agree-to-statement", "toSignAsCertificateProvider", f.AgreeToStatement,
		validation.Selected())

	return errors
}
