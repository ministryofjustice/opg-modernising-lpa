package donorpage

import (
	"errors"
	"net/http"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
)

type progressNotification struct {
	Heading string
	Body    string
}

type progressData struct {
	App               appcontext.Data
	Errors            validation.List
	Donor             *donordata.Provided
	Progress          task.Progress
	InfoNotifications []progressNotification
}

func Progress(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, progressTracker ProgressTracker, certificateProviderStore CertificateProviderStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, certificateProviderErr := certificateProviderStore.GetAny(r.Context())
		if certificateProviderErr != nil && !errors.Is(certificateProviderErr, dynamo.NotFoundError{}) {
			return certificateProviderErr
		}

		data := &progressData{
			App:      appData,
			Donor:    donor,
			Progress: progressTracker.Progress(lpa),
		}

		if donor.IdentityUserData.Status.IsUnknown() && donor.Tasks.ConfirmYourIdentity.IsPending() {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: "youHaveChosenToConfirmYourIdentityAtPostOffice",
				Body:    "whenYouHaveConfirmedAtPostOfficeReturnToTaskList",
			})
		}

		if lpa.Submitted && (lpa.CertificateProvider.SignedAt == nil || lpa.CertificateProvider.SignedAt.IsZero()) {
			if errors.Is(certificateProviderErr, dynamo.NotFoundError{}) {
				data.InfoNotifications = append(data.InfoNotifications, progressNotification{
					Heading: "youveSubmittedYourLpaToOpg",
					Body:    "opgIsCheckingYourLpa",
				})
			}
		}

		if donor.Tasks.PayForLpa.IsMoreEvidenceRequired() {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: "weNeedMoreEvidenceToMakeADecisionAboutYourLPAFee",
				Body: appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.MoreEvidenceRequiredAt)},
				),
			})
		}

		if !donor.Tasks.ConfirmYourIdentity.IsCompleted() && donor.Voucher.FirstNames != "" {
			if donor.VoucherInvitedAt.IsZero() && !donor.Tasks.PayForLpa.IsCompleted() {
				data.InfoNotifications = append(data.InfoNotifications, progressNotification{
					Heading: "youMustPayForYourLPA",
					Body: appData.Localizer.Format(
						"returnToTaskListToPayForLPAWeWillThenContactVoucher",
						map[string]any{"VoucherFullName": donor.Voucher.FullName()},
					),
				})
			} else if !donor.VoucherInvitedAt.IsZero() {
				data.InfoNotifications = append(data.InfoNotifications, progressNotification{
					Heading: appData.Localizer.Format(
						"weHaveContactedVoucherToConfirmYourIdentity",
						map[string]any{"VoucherFullName": donor.Voucher.FullName()},
					),
					Body: "youDoNotNeedToTakeAnyAction",
				})
			}
		}

		if !donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
			!donor.FailedVoucher.FailedAt.IsZero() &&
			!donor.WantVoucher.IsNo() &&
			donor.Voucher.FirstNames == "" {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: appData.Localizer.Format(
					"voucherHasBeenUnableToConfirmYourIdentity",
					map[string]any{"VoucherFullName": donor.FailedVoucher.FullName()},
				),
				Body: appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.FailedVoucher.FailedAt)},
				),
			})
		}

		if lpa.Status.IsDoNotRegister() && !donor.DoNotRegisterAt.IsZero() {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: appData.Localizer.T("thereIsAProblemWithYourLpa"),
				Body: appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.DoNotRegisterAt)},
				),
			})
		}

		if donor.IdentityUserData.Status.IsFailed() {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: appData.Localizer.T("youHaveBeenUnableToConfirmYourIdentity"),
				Body: appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.IdentityUserData.CheckedAt)},
				),
			})
		}

		if certificateProviderErr == nil && certificateProvider.IdentityUserData.Status.IsFailed() {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: appData.Localizer.Format(
					"certificateProviderHasBeenUnableToConfirmIdentity",
					map[string]any{"CertificateProviderFullName": lpa.CertificateProvider.FullName()},
				),
				Body: appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.IdentityUserData.CheckedAt)},
				),
			})
		}

		return tmpl(w, data)
	}
}
