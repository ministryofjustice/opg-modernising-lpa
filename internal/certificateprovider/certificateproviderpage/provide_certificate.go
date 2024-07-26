package certificateproviderpage

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
) Handler {
	return func(appData page.AppData, w http.ResponseWriter, r *http.Request, certificateProvider *actor.CertificateProviderProvidedDetails) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		if lpa.SignedAt.IsZero() {
			return page.Paths.CertificateProvider.TaskList.Redirect(w, r, appData, lpa.LpaID)
		}

		if !certificateProvider.SignedAt.IsZero() {
			return page.Paths.CertificateProvider.CertificateProvided.Redirect(w, r, appData, lpa.LpaID)
		}

		data := &provideCertificateData{
			App:                 appData,
			CertificateProvider: certificateProvider,
			Lpa:                 lpa,
			Form: &provideCertificateForm{
				AgreeToStatement: !certificateProvider.SignedAt.IsZero(),
			},
		}

		if r.Method == http.MethodPost {
			data.Form = readProvideCertificateForm(r)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.Submittable == "cannot-submit" {
					return page.Paths.CertificateProvider.ConfirmDontWantToBeCertificateProvider.Redirect(w, r, appData, certificateProvider.LpaID)
				}

				certificateProvider.SignedAt = now()
				certificateProvider.Tasks.ProvideTheCertificate = actor.TaskCompleted

				if lpa.CertificateProvider.SignedAt.IsZero() {
					if err := lpaStoreClient.SendCertificateProvider(r.Context(), certificateProvider, lpa); err != nil {
						return err
					}
				} else {
					certificateProvider.SignedAt = lpa.CertificateProvider.SignedAt
				}

				if err := notifyClient.SendActorEmail(r.Context(), certificateProvider.Email, lpa.LpaUID, notify.CertificateProviderCertificateProvidedEmail{
					DonorFullNamePossessive:     appData.Localizer.Possessive(lpa.Donor.FullName()),
					DonorFirstNamesPossessive:   appData.Localizer.Possessive(lpa.Donor.FirstNames),
					LpaType:                     localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
					CertificateProviderFullName: lpa.CertificateProvider.FullName(),
					CertificateProvidedDateTime: appData.Localizer.FormatDateTime(certificateProvider.SignedAt),
				}); err != nil {
					return fmt.Errorf("email failed: %w", err)
				}

				if err := shareCodeSender.SendAttorneys(r.Context(), appData, lpa); err != nil {
					return err
				}

				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
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
