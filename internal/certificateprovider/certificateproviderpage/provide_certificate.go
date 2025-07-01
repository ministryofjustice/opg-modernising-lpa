package certificateproviderpage

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type provideCertificateData struct {
	App                 appcontext.Data
	Errors              validation.List
	CertificateProvider *certificateproviderdata.Provided
	Lpa                 *lpadata.Lpa
	Form                *provideCertificateForm
}

func ProvideCertificate(
	tmpl template.Template,
	certificateProviderStore CertificateProviderStore,
	notifyClient NotifyClient,
	accessCodeSender AccessCodeSender,
	lpaStoreClient LpaStoreClient,
	scheduledStore ScheduledStore,
	donorStore DonorStore,
	now func() time.Time,
	donorStartURL string,
) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		if !certificateProvider.SignedAt.IsZero() {
			return certificateprovider.PathCertificateProvided.Redirect(w, r, appData, lpa.LpaID)
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
			data.Form = readProvideCertificateForm(r, lpa.Language)
			data.Errors = data.Form.Validate()

			if data.Errors.None() {
				if data.Form.Submittable == "cannot-submit" {
					return certificateprovider.PathConfirmDontWantToBeCertificateProvider.Redirect(w, r, appData, certificateProvider.LpaID)
				}

				certificateProvider.SignedAt = now()
				certificateProvider.Tasks.ProvideTheCertificate = task.StateCompleted

				if lpa.CertificateProvider.SignedAt == nil || lpa.CertificateProvider.SignedAt.IsZero() {
					if err := lpaStoreClient.SendCertificateProvider(r.Context(), certificateProvider, lpa); err != nil {
						return fmt.Errorf("error sending certificate provider to lpa-store: %w", err)
					}
				} else {
					certificateProvider.SignedAt = *lpa.CertificateProvider.SignedAt
				}

				if err := notifyClient.SendActorEmail(r.Context(), notify.ToLpaCertificateProvider(certificateProvider, lpa), lpa.LpaUID, notify.CertificateProviderCertificateProvidedEmail{
					DonorFullNamePossessive:     appData.Localizer.Possessive(lpa.Donor.FullName()),
					DonorFirstNamesPossessive:   appData.Localizer.Possessive(lpa.Donor.FirstNames),
					LpaType:                     localize.LowerFirst(appData.Localizer.T(lpa.Type.String())),
					CertificateProviderFullName: lpa.CertificateProvider.FullName(),
					CertificateProvidedDateTime: appData.Localizer.FormatDateTime(certificateProvider.SignedAt),
				}); err != nil {
					return fmt.Errorf("email to certificate provider failed: %w", err)
				}

				if !certificateProvider.IdentityUserData.Status.IsConfirmed() {
					if err := notifyClient.SendActorEmail(r.Context(), notify.ToLpaDonor(lpa), lpa.LpaUID, notify.CertificateProviderFailedIdentityCheckEmail{
						Greeting:                    notifyClient.EmailGreeting(lpa),
						CertificateProviderFullName: lpa.CertificateProvider.FullName(),
						LpaType:                     appData.Localizer.T(lpa.Type.String()),
						LpaReferenceNumber:          lpa.LpaUID,
						DonorStartPageURL:           donorStartURL,
					}); err != nil {
						return fmt.Errorf("email to donor failed: %w", err)
					}
				}

				if err := accessCodeSender.SendAttorneys(r.Context(), appData, lpa); err != nil {
					return fmt.Errorf("error sending accesscode to attorneys: %w", err)
				}

				donor, err := donorStore.GetAny(r.Context())
				if err != nil {
					return fmt.Errorf("error getting donor: %w", err)
				}

				donor.AttorneysInvitedAt = now()

				if err := donorStore.Put(r.Context(), donor); err != nil {
					return fmt.Errorf("error putting donor: %w", err)
				}

				if !certificateProvider.Tasks.ConfirmYourIdentity.IsCompleted() {
					if err := scheduledStore.Create(r.Context(), scheduled.Event{
						At:           certificateProvider.SignedAt.AddDate(0, 3, 1),
						Action:       scheduled.ActionRemindCertificateProviderToConfirmIdentity,
						TargetLpaKey: certificateProvider.PK,
						LpaUID:       lpa.LpaUID,
					}, scheduled.Event{
						At:           lpa.SignedAt.AddDate(0, 21, 1),
						Action:       scheduled.ActionRemindCertificateProviderToConfirmIdentity,
						TargetLpaKey: certificateProvider.PK,
						LpaUID:       lpa.LpaUID,
					}); err != nil {
						return fmt.Errorf("error scheduling certificate provider prompt: %w", err)
					}
				}

				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return fmt.Errorf("error updating certificate provider: %w", err)
				}

				return certificateprovider.PathCertificateProvided.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

type provideCertificateForm struct {
	Submittable      string
	AgreeToStatement bool
	lpaLanguage      localize.Lang
}

func readProvideCertificateForm(r *http.Request, lang localize.Lang) *provideCertificateForm {
	return &provideCertificateForm{
		Submittable:      r.FormValue("submittable"),
		AgreeToStatement: page.PostFormString(r, "agree-to-statement") == "1",
		lpaLanguage:      lang,
	}
}

func (f *provideCertificateForm) Validate() validation.List {
	var errors validation.List

	if f.Submittable != "cannot-submit" {
		errors.Bool("agree-to-statement", "toSignAsCertificateProvider", f.AgreeToStatement,
			validation.Selected())
	}

	if f.Submittable == "wrong-language" && f.AgreeToStatement {
		errors.Add("agree-to-statement", toSignCertificateYouMustViewInLanguageError{LpaLanguage: f.lpaLanguage})
	}

	return errors
}

type toSignCertificateYouMustViewInLanguageError struct {
	LpaLanguage localize.Lang
}

func (e toSignCertificateYouMustViewInLanguageError) Format(l validation.Localizer) string {
	return l.Format("toSignCertificateYouMustViewInLanguage", map[string]any{
		"InLang": l.T("in:" + e.LpaLanguage.String()),
	})
}
