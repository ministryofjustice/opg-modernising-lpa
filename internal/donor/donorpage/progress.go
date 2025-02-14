package donorpage

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-go-common/template"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/validation"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type progressData struct {
	App                  appcontext.Data
	Errors               validation.List
	Donor                *donordata.Provided
	Progress             task.Progress
	InfoNotifications    []progressNotification
	SuccessNotifications []progressNotification
}

func (d *progressData) addInfo(heading, body string) {
	d.InfoNotifications = append(d.InfoNotifications, progressNotification{Heading: heading, Body: body})
}

func (d *progressData) addSuccess(heading, body string) {
	d.SuccessNotifications = append(d.SuccessNotifications, progressNotification{Heading: heading, Body: body})
}

type progressNotification struct {
	Heading string
	Body    string
}

func Progress(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, progressTracker ProgressTracker, certificateProviderStore CertificateProviderStore, voucherStore VoucherStore, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return fmt.Errorf("error getting lpa: %w", err)
		}

		certificateProvider, err := certificateProviderStore.GetAny(r.Context())
		if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
			return fmt.Errorf("error getting certificate provider: %w", err)
		}

		var voucher *voucherdata.Provided
		if donor.Voucher.FirstNames != "" {
			voucher, err = voucherStore.GetAny(r.Context())
			if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
				return fmt.Errorf("error getting voucher: %w", err)
			}
		}

		data := &progressData{
			App:      appData,
			Donor:    donor,
			Progress: progressTracker.Progress(lpa),
		}

		if !donor.WithdrawnAt.IsZero() {
			data.addInfo("lpaRevoked",
				appData.Localizer.Format(
					"weContactedYouOnAboutLPARevokedOPGWillNotRegister",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.WithdrawnAt)},
				),
			)

			return tmpl(w, data)
		}

		if lpa.Submitted &&
			(lpa.CertificateProvider.SignedAt == nil || lpa.CertificateProvider.SignedAt.IsZero()) &&
			certificateProvider == nil {
			data.addInfo("youveSubmittedYourLpaToOpg", "opgIsCheckingYourLpa")
		}

		if donor.IdentityUserData.Status.IsUnknown() &&
			donor.Tasks.ConfirmYourIdentity.IsPending() {
			data.addInfo("youHaveChosenToConfirmYourIdentityAtPostOffice", "whenYouHaveConfirmedAtPostOfficeReturnToTaskList")
		}

		if donor.Tasks.PayForLpa.IsMoreEvidenceRequired() {
			data.addInfo(
				"weNeedMoreEvidenceToMakeADecisionAboutYourLPAFee",
				appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.MoreEvidenceRequiredAt)},
				))
		}

		if !donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
			donor.Voucher.FirstNames != "" &&
			donor.VoucherInvitedAt.IsZero() &&
			!donor.Tasks.PayForLpa.IsCompleted() {
			data.addInfo(
				"youMustPayForYourLPA",
				appData.Localizer.Format(
					"returnToTaskListToPayForLPAWeWillThenContactVoucher",
					map[string]any{"VoucherFullName": donor.Voucher.FullName()},
				),
			)
		}

		if !donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
			donor.Voucher.FirstNames != "" &&
			!donor.VoucherInvitedAt.IsZero() {
			data.addInfo(
				appData.Localizer.Format(
					"weHaveContactedVoucherToConfirmYourIdentity",
					map[string]any{"VoucherFullName": donor.Voucher.FullName()},
				),
				"youDoNotNeedToTakeAnyAction",
			)
		}

		if !donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
			!donor.FailedVoucher.FailedAt.IsZero() &&
			!donor.WantVoucher.IsNo() &&
			donor.Voucher.FirstNames == "" {
			data.addInfo(
				appData.Localizer.Format(
					"voucherHasBeenUnableToConfirmYourIdentity",
					map[string]any{"VoucherFullName": donor.FailedVoucher.FullName()},
				),
				appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.FailedVoucher.FailedAt)},
				),
			)
		}

		if lpa.Status.IsDoNotRegister() &&
			!donor.DoNotRegisterAt.IsZero() {
			data.addInfo(
				"thereIsAProblemWithYourLpa",
				appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.DoNotRegisterAt)},
				),
			)
		}

		if donor.IdentityUserData.Status.IsFailed() {
			data.addInfo(
				"youHaveBeenUnableToConfirmYourIdentity",
				appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.IdentityUserData.CheckedAt)},
				),
			)
		}

		if certificateProvider != nil && certificateProvider.IdentityUserData.Status.IsFailed() {
			data.addInfo(
				appData.Localizer.Format(
					"certificateProviderHasBeenUnableToConfirmIdentity",
					map[string]any{"CertificateProviderFullName": lpa.CertificateProvider.FullName()},
				),
				appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(certificateProvider.IdentityUserData.CheckedAt)},
				),
			)
		}

		if certificateProvider != nil && certificateProvider.Tasks.ConfirmYourIdentity.IsPending() {
			data.addInfo(
				appData.Localizer.Format(
					"certificateProviderConfirmationOfIdentityPending",
					map[string]any{"CertificateProviderFullName": lpa.CertificateProvider.FullName()},
				),
				"wellContactYouIfYouNeedToTakeAnyAction",
			)
		}

		if !donor.HasSeenSuccessfulVouchBanner &&
			donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
			!donor.Tasks.SignTheLpa.IsCompleted() &&
			voucher != nil &&
			!voucher.SignedAt.IsZero() {
			data.addSuccess(
				appData.Localizer.Format(
					"voucherHasConfirmedYourIdentity",
					map[string]any{"VoucherFullName": voucher.FullName()},
				),
				"returnToYourTaskListForInformationAboutWhatToDoNext",
			)

			donor.HasSeenSuccessfulVouchBanner = true
		}

		if !donor.HasSeenSuccessfulVouchBanner &&
			donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
			donor.Tasks.SignTheLpa.IsCompleted() &&
			voucher != nil &&
			!voucher.SignedAt.IsZero() {
			data.addSuccess(
				appData.Localizer.Format(
					"voucherHasConfirmedYourIdentity",
					map[string]any{"VoucherFullName": voucher.FullName()},
				),
				"youDoNotNeedToTakeAnyAction",
			)

			donor.HasSeenSuccessfulVouchBanner = true
		}

		if donor.Tasks.PayForLpa.IsPending() &&
			donor.FeeAmount() == 0 {
			data.addInfo("weAreReviewingTheEvidenceYouSent", "ifYourEvidenceIsApprovedWillShowPaid")
		}

		if !donor.HasSeenReducedFeeApprovalNotification &&
			!donor.ReducedFeeApprovedAt.IsZero() &&
			donor.Tasks.PayForLpa.IsCompleted() {
			data.addSuccess("weHaveApprovedYourLPAFeeRequest", "yourLPAIsNowPaid")

			donor.HasSeenReducedFeeApprovalNotification = true
		}

		if donor.RegisteringWithCourtOfProtection &&
			donor.Tasks.PayForLpa.IsCompleted() &&
			!donor.WitnessedByCertificateProviderAt.IsZero() {
			data.addInfo("yourLpaMustBeReviewedByCourtOfProtection", "opgIsCompletingChecksSoYouCanSubmitToCourtOfProtection")
		}

		if donor.RegisteringWithCourtOfProtection &&
			donor.Tasks.PayForLpa.IsCompleted() &&
			donor.WitnessedByCertificateProviderAt.IsZero() {
			data.addInfo("yourLpaMustBeReviewedByCourtOfProtection", "returnToYourTaskListToSignThenOpgWillCheck")
		}

		if donor.RegisteringWithCourtOfProtection &&
			!donor.Tasks.PayForLpa.IsCompleted() &&
			!donor.WitnessedByCertificateProviderAt.IsZero() {
			data.addInfo("yourLpaMustBeReviewedByCourtOfProtection", "whenYouHavePaidOpgWillCheck")
		}

		if now().After(donor.IdentityDeadline()) &&
			donor.Tasks.SignTheLpa.IsCompleted() &&
			!donor.Tasks.ConfirmYourIdentity.IsCompleted() {
			data.addInfo("yourLPACannotBeRegisteredByOPG", "youDidNotConfirmYourIdentityWithinSixMonthsOfSigning")
		}

		if donor.IdentityUserData.Status.IsExpired() &&
			!donor.Tasks.SignTheLpa.IsCompleted() {
			data.addInfo("youMustConfirmYourIdentityAgain", "youDidNotSignYourLPAWithinSixMonthsOfConfirmingYourIdentity")
		}

		if lpa.Status.IsStatutoryWaitingPeriod() {
			data.addInfo("yourLpaIsAwaitingRegistration", "theOpgWillRegisterYourLpaAtEndOfWaitingPeriod")
		}

		if donor.Tasks.ConfirmYourIdentity.IsPending() &&
			donor.ContinueWithMismatchedIdentity {
			data.addInfo("confirmationOfIdentityPending", "youDoNotNeedToTakeAnyAction")
		}

		if donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
			donor.ContinueWithMismatchedIdentity &&
			!donor.HasSeenIdentityMismatchResolvedNotification {
			data.addSuccess("yourIdentityHadBeenConfirmed", "youDoNotNeedToTakeAnyAction")

			donor.HasSeenIdentityMismatchResolvedNotification = true
		}

		if certificateProvider != nil &&
			certificateProvider.Tasks.ConfirmYourIdentity.IsCompleted() &&
			!certificateProvider.ImmaterialChangeConfirmedAt.IsZero() &&
			!donor.HasSeenCertificateProviderIdentityMismatchResolvedNotification {
			data.addSuccess(
				appData.Localizer.Format(
					"certificateProviderIdentityConfirmed",
					map[string]any{"CertificateProviderFullName": lpa.CertificateProvider.FullName()},
				),
				"youDoNotNeedToTakeAnyAction",
			)

			donor.HasSeenCertificateProviderIdentityMismatchResolvedNotification = true
		}

		if !lpa.Status.IsRegistered() &&
			!donor.PriorityCorrespondenceSentAt.IsZero() {
			data.addInfo(
				"thereIsAProblemWithYourLpa",
				appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.PriorityCorrespondenceSentAt)},
				),
			)
		}

		if donor.Tasks.ConfirmYourIdentity.IsProblem() &&
			donor.ContinueWithMismatchedIdentity &&
			!donor.MaterialChangeConfirmedAt.IsZero() {
			data.addInfo(
				"yourLPACannotBeRegisteredByOPG",
				appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.MaterialChangeConfirmedAt)},
				),
			)
		}

		if certificateProvider != nil &&
			certificateProvider.Tasks.ConfirmYourIdentity.IsProblem() &&
			!certificateProvider.MaterialChangeConfirmedAt.IsZero() {
			data.addInfo(
				"yourLPACannotBeRegisteredByOPG",
				appData.Localizer.Format(
					"weContactedYouOnWithGuidanceAboutWhatToDoNext",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(certificateProvider.MaterialChangeConfirmedAt)},
				),
			)
		}

		if err := donorStore.Put(r.Context(), donor); err != nil {
			return fmt.Errorf("failed to update donor: %v", err)
		}

		return tmpl(w, data)
	}
}
