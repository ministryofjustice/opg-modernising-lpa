package certificateprovider

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type provideCertificateData struct {
	App                 page.AppData
	Errors              validation.List
	CertificateProvider *actor.CertificateProviderProvidedDetails
	Lpa                 *lpastore.Lpa
	Form                *provideCertificateForm
}

func ProvideCertificate(
	tmpl template.Template,
	lpaStoreResolvingService LpaStoreResolvingService,
	certificateProviderStore CertificateProviderStore,
	notifyClient NotifyClient,
	shareCodeSender ShareCodeSender,
	lpaStoreClient LpaStoreClient,
	now func() time.Time,
) page.Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		if lpa.SignedAt.IsZero() {
			return page.Paths.CertificateProvider.TaskList.Redirect(w, r, appData, lpa.LpaID)
		}

		certificateProvider, err := certificateProviderStore.Get(r.Context())
		if err != nil {
			return err
		}

		data := &provideCertificateData{
			App:                 appData,
			CertificateProvider: certificateProvider,
			Lpa:                 lpa,
			Form: &provideCertificateForm{
				AgreeToStatement: certificateProvider.Certificate.AgreeToStatement,
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readProvideCertificateForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.Submittable == "cannot-submit" {
					return page.Paths.CertificateProvider.ConfirmDontWantToBeCertificateProvider.Redirect(w, r, appData, certificateProvider.LpaID)
				}

				certificateProvider.Certificate.AgreeToStatement = true
				certificateProvider.Certificate.Agreed = now()
				certificateProvider.Tasks.ProvideTheCertificate = actor.TaskCompleted
				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return err
				}

				if err := lpaStoreClient.SendCertificateProvider(r.Context(), certificateProvider, lpa); err != nil {
					return err
				}

				if err := notifyClient.SendActorEmail(r.Context(), certificateProvider.Email, lpa.LpaUID, notify.CertificateProviderCertificateProvidedEmail{
					DonorFullNamePossessive:     appData.Localizer.Possessive(lpa.Donor.FullName()),
					DonorFirstNamesPossessive:   appData.Localizer.Possessive(lpa.Donor.FirstNames),
					LpaType:                     localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
					CertificateProviderFullName: lpa.CertificateProvider.FullName(),
					CertificateProvidedDateTime: appData.Localizer.FormatDateTime(certificateProvider.Certificate.Agreed),
				}); err != nil {
					return fmt.Errorf("email failed: %w", err)
				}

				if err := shareCodeSender.SendAttorneys(r.Context(), appData, lpa); err != nil {
					return err
				}

				return page.Paths.CertificateProvider.CertificateProvided.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

type provideCertificateForm struct {
	Submittable      string
	AgreeToStatement bool
}

func readProvideCertificateForm(r *http.Request) *provideCertificateForm {
	return &provideCertificateForm{
		Submittable:      r.FormValue("submittable"),
		AgreeToStatement: page.PostFormString(r, "agree-to-statement") == "1",
	}
}

func (f *provideCertificateForm) Validate() validation.List {
	var errors validation.List

	if f.Submittable != "cannot-submit" {
		errors.Bool("agree-to-statement", "toSignAsCertificateProvider", f.AgreeToStatement,
			validation.Selected())
	}

	return errors
}
