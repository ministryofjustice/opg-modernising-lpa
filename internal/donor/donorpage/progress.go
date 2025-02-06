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
	Voucher              *voucherdata.Provided
	Progress             task.Progress
	InfoNotifications    []progressNotification
	SuccessNotifications []progressNotification
}

func (d *progressData) addNotification(heading, body string, success bool) {
	notification := progressNotification{Heading: heading, Body: body}

	if success {
		d.SuccessNotifications = append(d.SuccessNotifications, notification)
	} else {
		d.InfoNotifications = append(d.InfoNotifications, notification)
	}
}

type progressNotification struct {
	Heading string
	Body    string
}

type notificationRule struct {
	Condition func() bool
	Heading   func() string
	Body      func() string
	Success   bool
	SetSeen   func() error
}

func Progress(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, progressTracker ProgressTracker, certificateProviderStore CertificateProviderStore, voucherStore VoucherStore, donorStore DonorStore, now func() time.Time) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		certificateProvider, certificateProviderErr := certificateProviderStore.GetAny(r.Context())
		if certificateProviderErr != nil && !errors.Is(certificateProviderErr, dynamo.NotFoundError{}) {
			return certificateProviderErr
		}

		var voucher *voucherdata.Provided
		if donor.Voucher.FirstNames != "" {
			voucher, err = voucherStore.GetAny(r.Context())
			if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
				return err
			}
		}

		data := &progressData{
			App:      appData,
			Donor:    donor,
			Progress: progressTracker.Progress(lpa),
		}

		notificationRules := []notificationRule{
			{
				Condition: func() bool {
					return !donor.WithdrawnAt.IsZero()
				},
				Heading: func() string {
					return "lpaRevoked"
				},
				Body: func() string {
					return appData.Localizer.Format(
						"weContactedYouOnAboutLPARevokedOPGWillNotRegister",
						map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.WithdrawnAt)},
					)
				},
			},
			{
				Condition: func() bool {
					return donor.IdentityUserData.Status.IsUnknown() &&
						donor.Tasks.ConfirmYourIdentity.IsPending()
				},
				Heading: func() string {
					return "youHaveChosenToConfirmYourIdentityAtPostOffice"
				},
				Body: func() string {
					return "whenYouHaveConfirmedAtPostOfficeReturnToTaskList"
				},
			},
			{
				Condition: func() bool {
					return lpa.Submitted &&
						(lpa.CertificateProvider.SignedAt == nil || lpa.CertificateProvider.SignedAt.IsZero()) &&
						errors.Is(certificateProviderErr, dynamo.NotFoundError{})
				},
				Heading: func() string {
					return "youveSubmittedYourLpaToOpg"
				},
				Body: func() string {
					return "opgIsCheckingYourLpa"
				},
			},
			{
				Condition: func() bool {
					return donor.Tasks.PayForLpa.IsMoreEvidenceRequired()
				},
				Heading: func() string {
					return "weNeedMoreEvidenceToMakeADecisionAboutYourLPAFee"
				},
				Body: func() string {
					return appData.Localizer.Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.MoreEvidenceRequiredAt)},
					)
				},
			},
			{
				Condition: func() bool {
					return !donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
						donor.Voucher.FirstNames != "" &&
						donor.VoucherInvitedAt.IsZero() &&
						!donor.Tasks.PayForLpa.IsCompleted()
				},
				Heading: func() string {
					return "youMustPayForYourLPA"
				},
				Body: func() string {
					return appData.Localizer.Format(
						"returnToTaskListToPayForLPAWeWillThenContactVoucher",
						map[string]any{"VoucherFullName": donor.Voucher.FullName()},
					)
				},
			},
			{
				Condition: func() bool {
					return !donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
						donor.Voucher.FirstNames != "" &&
						!donor.VoucherInvitedAt.IsZero()
				},
				Heading: func() string {
					return appData.Localizer.Format(
						"weHaveContactedVoucherToConfirmYourIdentity",
						map[string]any{"VoucherFullName": donor.Voucher.FullName()},
					)
				},
				Body: func() string {
					return "youDoNotNeedToTakeAnyAction"
				},
			},
			{
				Condition: func() bool {
					return !donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
						!donor.FailedVoucher.FailedAt.IsZero() &&
						!donor.WantVoucher.IsNo() &&
						donor.Voucher.FirstNames == ""
				},
				Heading: func() string {
					return appData.Localizer.Format(
						"voucherHasBeenUnableToConfirmYourIdentity",
						map[string]any{"VoucherFullName": donor.FailedVoucher.FullName()},
					)
				},
				Body: func() string {
					return appData.Localizer.Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.FailedVoucher.FailedAt)},
					)
				},
			},
			{
				Condition: func() bool {
					return lpa.Status.IsDoNotRegister() &&
						!donor.DoNotRegisterAt.IsZero()
				},
				Heading: func() string {
					return "thereIsAProblemWithYourLpa"
				},
				Body: func() string {
					return appData.Localizer.Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.DoNotRegisterAt)},
					)
				},
			},
			{
				Condition: func() bool {
					return donor.IdentityUserData.Status.IsFailed()
				},
				Heading: func() string {
					return "youHaveBeenUnableToConfirmYourIdentity"
				},
				Body: func() string {
					return appData.Localizer.Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.IdentityUserData.CheckedAt)},
					)
				},
			},
			{
				Condition: func() bool {
					return certificateProviderErr == nil &&
						certificateProvider.IdentityUserData.Status.IsFailed()
				},
				Heading: func() string {
					return appData.Localizer.Format(
						"certificateProviderHasBeenUnableToConfirmIdentity",
						map[string]any{"CertificateProviderFullName": lpa.CertificateProvider.FullName()},
					)
				},
				Body: func() string {
					return appData.Localizer.Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": appData.Localizer.FormatDate(certificateProvider.IdentityUserData.CheckedAt)},
					)
				},
			},
			{
				Condition: func() bool {
					return !donor.HasSeenSuccessfulVouchBanner &&
						donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
						!donor.Tasks.SignTheLpa.IsCompleted() &&
						voucher != nil &&
						!voucher.SignedAt.IsZero()
				},
				Heading: func() string {
					return appData.Localizer.Format(
						"voucherHasConfirmedYourIdentity",
						map[string]any{"VoucherFullName": voucher.FullName()},
					)
				},
				Body: func() string {
					return "returnToYourTaskListForInformationAboutWhatToDoNext"
				},
				Success: true,
				SetSeen: func() error {
					donor.HasSeenSuccessfulVouchBanner = true
					return donorStore.Put(r.Context(), donor)
				},
			},
			{
				Condition: func() bool {
					return !donor.HasSeenSuccessfulVouchBanner &&
						donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
						donor.Tasks.SignTheLpa.IsCompleted() &&
						voucher != nil &&
						!voucher.SignedAt.IsZero()
				},
				Heading: func() string {
					return appData.Localizer.Format(
						"voucherHasConfirmedYourIdentity",
						map[string]any{"VoucherFullName": voucher.FullName()},
					)
				},
				Body: func() string {
					return "youDoNotNeedToTakeAnyAction"
				},
				Success: true,
				SetSeen: func() error {
					donor.HasSeenSuccessfulVouchBanner = true
					return donorStore.Put(r.Context(), donor)
				},
			},
			{
				Condition: func() bool {
					return donor.Tasks.PayForLpa.IsPending() &&
						donor.FeeAmount() == 0
				},
				Heading: func() string {
					return "weAreReviewingTheEvidenceYouSent"
				},
				Body: func() string {
					return "ifYourEvidenceIsApprovedWillShowPaid"
				},
			},
			{
				Condition: func() bool {
					return !donor.HasSeenReducedFeeApprovalNotification &&
						!donor.ReducedFeeApprovedAt.IsZero() &&
						donor.Tasks.PayForLpa.IsCompleted()
				},
				Heading: func() string {
					return "weHaveApprovedYourLPAFeeRequest"
				},
				Body: func() string {
					return "yourLPAIsNowPaid"
				},
				Success: true,
				SetSeen: func() error {
					donor.HasSeenReducedFeeApprovalNotification = true
					return donorStore.Put(r.Context(), donor)
				},
			},
			{
				Condition: func() bool {
					return donor.RegisteringWithCourtOfProtection &&
						donor.Tasks.PayForLpa.IsCompleted() &&
						!donor.WitnessedByCertificateProviderAt.IsZero()
				},
				Heading: func() string {
					return "yourLpaMustBeReviewedByCourtOfProtection"
				},
				Body: func() string {
					return "opgIsCompletingChecksSoYouCanSubmitToCourtOfProtection"
				},
			},
			{
				Condition: func() bool {
					return donor.RegisteringWithCourtOfProtection &&
						donor.Tasks.PayForLpa.IsCompleted() &&
						donor.WitnessedByCertificateProviderAt.IsZero()
				},
				Heading: func() string {
					return "yourLpaMustBeReviewedByCourtOfProtection"
				},
				Body: func() string {
					return "returnToYourTaskListToSignThenOpgWillCheck"
				},
			},
			{
				Condition: func() bool {
					return donor.RegisteringWithCourtOfProtection &&
						!donor.Tasks.PayForLpa.IsCompleted() &&
						!donor.WitnessedByCertificateProviderAt.IsZero()
				},
				Heading: func() string {
					return "yourLpaMustBeReviewedByCourtOfProtection"
				},
				Body: func() string {
					return "whenYouHavePaidOpgWillCheck"
				},
			},
			{
				Condition: func() bool {
					return now().After(donor.IdentityDeadline()) &&
						donor.Tasks.SignTheLpa.IsCompleted() &&
						!donor.Tasks.ConfirmYourIdentity.IsCompleted()
				},
				Heading: func() string {
					return "yourLPACannotBeRegisteredByOPG"
				},
				Body: func() string {
					return "youDidNotConfirmYourIdentityWithinSixMonthsOfSigning"
				},
			},
			{
				Condition: func() bool {
					return donor.IdentityUserData.Status.IsExpired() &&
						!donor.Tasks.SignTheLpa.IsCompleted()
				},
				Heading: func() string {
					return "youMustConfirmYourIdentityAgain"
				},
				Body: func() string {
					return "youDidNotSignYourLPAWithinSixMonthsOfConfirmingYourIdentity"
				},
			},
			{
				Condition: func() bool {
					return lpa.Status.IsStatutoryWaitingPeriod()
				},
				Heading: func() string {
					return "yourLpaIsAwaitingRegistration"
				},
				Body: func() string {
					return "theOpgWillRegisterYourLpaAtEndOfWaitingPeriod"
				},
			},
			{
				Condition: func() bool {
					return donor.Tasks.ConfirmYourIdentity.IsPending() &&
						donor.ContinueWithMismatchedIdentity
				},
				Heading: func() string {
					return "confirmationOfIdentityPending"
				},
				Body: func() string {
					return "youDoNotNeedToTakeAnyAction"
				},
			},
			{
				Condition: func() bool {
					return !lpa.Status.IsRegistered() &&
						!donor.PriorityCorrespondenceSentAt.IsZero()
				},
				Heading: func() string {
					return "thereIsAProblemWithYourLpa"
				},
				Body: func() string {
					return appData.Localizer.Format(
						"weContactedYouOnWithGuidanceAboutWhatToDoNext",
						map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.PriorityCorrespondenceSentAt)},
					)
				},
			},
		}

		for _, rule := range notificationRules {
			if rule.Condition() {
				if rule.Heading() == "lpaRevoked" {
					data.InfoNotifications = nil
					data.SuccessNotifications = nil

					data.addNotification(rule.Heading(), rule.Body(), rule.Success)
					break
				}

				data.addNotification(rule.Heading(), rule.Body(), rule.Success)

				if rule.Success {
					if err := rule.SetSeen(); err != nil {
						return fmt.Errorf("failed to update donor: %v", err)
					}
				}
			}
		}

		return tmpl(w, data)
	}
}
