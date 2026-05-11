package certificateproviderpage

import (
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/certificateprovider/certificateproviderdata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/event"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/localize"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore/lpadata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/newforms"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/notify"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/scheduled/scheduleddata"
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
	eventClient EventClient,
) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, certificateProvider *certificateproviderdata.Provided, lpa *lpadata.Lpa) error {
		if !certificateProvider.SignedAt.IsZero() {
			return certificateprovider.PathCertificateProvided.Redirect(w, r, appData, lpa.LpaID)
		}

		data := &provideCertificateData{
			App:                 appData,
			CertificateProvider: certificateProvider,
			Lpa:                 lpa,
			Form:                newProvideCertificateForm(appData.Localizer, lpa.Language, lpa.CertificateProvider.FullName()),
		}

		data.Form.AgreeToStatement.SetInput(!certificateProvider.SignedAt.IsZero())

		if r.Method == http.MethodPost {
			if data.Form.Parse(r) {
				if data.Form.Submittable.Value == "cannot-submit" {
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
						Action:       scheduleddata.ActionRemindCertificateProviderToConfirmIdentity,
						TargetLpaKey: certificateProvider.PK,
						LpaUID:       lpa.LpaUID,
					}, scheduled.Event{
						At:           lpa.SignedAt.AddDate(0, 21, 1),
						Action:       scheduleddata.ActionRemindCertificateProviderToConfirmIdentity,
						TargetLpaKey: certificateProvider.PK,
						LpaUID:       lpa.LpaUID,
					}); err != nil {
						return fmt.Errorf("error scheduling certificate provider prompt: %w", err)
					}
				}

				if err := certificateProviderStore.Put(r.Context(), certificateProvider); err != nil {
					return fmt.Errorf("error updating certificate provider: %w", err)
				}

				if lpa.Donor.Channel.IsPaper() {
					if err := eventClient.SendLetterRequested(r.Context(), event.LetterRequested{
						UID:        lpa.LpaUID,
						LetterType: "ADVISE_DONOR_CERTIFICATE_HAS_BEEN_PROVIDED",
						ActorType:  actor.TypeDonor,
						ActorUID:   lpa.Donor.UID,
					}); err != nil {
						return fmt.Errorf("error sending letter requested event: %w", err)
					}
				}

				return certificateprovider.PathCertificateProvided.Redirect(w, r, appData, certificateProvider.LpaID)
			}
		}

		return tmpl(w, data)
	}
}

type provideCertificateForm struct {
	Submittable      *newforms.String
	AgreeToStatement *newforms.Bool
	Errors           []newforms.Field

	lpaLanguage localize.Lang
}

func newProvideCertificateForm(l Localizer, lang localize.Lang, certificateProviderFullName string) *provideCertificateForm {
	return &provideCertificateForm{
		Submittable:      newforms.NewString("submittable", ""),
		AgreeToStatement: newforms.NewBool("agree-to-statement", l.Format("iAgreeToTheseStatements", map[string]any{"FullName": certificateProviderFullName})),
		lpaLanguage:      lang,
	}
}

func (f *provideCertificateForm) Parse(r *http.Request) bool {
	f.Errors = newforms.ParsePostForm(r,
		f.Submittable,
		f.AgreeToStatement,
	)

	if f.Submittable.Value != "cannot-submit" && !f.AgreeToStatement.Value {
		f.AgreeToStatement.Error = newforms.SelectError{
			Field: newforms.Field{
				Label: "toSignAsCertificateProvider",
			},
		}
		f.Errors = append(f.Errors, f.AgreeToStatement.Field)
	}

	if f.Submittable.Value == "wrong-language" && f.AgreeToStatement.Value {
		f.AgreeToStatement.Error = toSignCertificateYouMustViewInLanguageError{LpaLanguage: f.lpaLanguage}
		f.Errors = append(f.Errors, f.AgreeToStatement.Field)
	}

	return len(f.Errors) == 0
}

type toSignCertificateYouMustViewInLanguageError struct {
	LpaLanguage localize.Lang
}

func (e toSignCertificateYouMustViewInLanguageError) Format(l newforms.Localizer) string {
	return l.Format("toSignCertificateYouMustViewInLanguage", map[string]any{
		"InLang": l.T("in:" + e.LpaLanguage.String()),
	})
}
