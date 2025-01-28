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

type progressNotification struct {
	Heading string
	Body    string
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

		data := &progressData{
			App:      appData,
			Donor:    donor,
			Progress: progressTracker.Progress(lpa),
		}

		if !donor.WithdrawnAt.IsZero() {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: "lpaRevoked",
				Body: appData.Localizer.Format(
					"weContactedYouOnAboutLPARevokedOPGWillNotRegister",
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.WithdrawnAt)},
				),
			})

			return tmpl(w, data)
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
					map[string]any{"ContactedDate": appData.Localizer.FormatDate(certificateProvider.IdentityUserData.CheckedAt)},
				),
			})
		}

		if !donor.HasSeenSuccessfulVouchBanner &&
			donor.Tasks.ConfirmYourIdentity.IsCompleted() &&
			donor.Voucher.FirstNames != "" {
			voucher, err := voucherStore.GetAny(r.Context())
			if err != nil && !errors.Is(err, dynamo.NotFoundError{}) {
				return err
			}

			if voucher != nil && !voucher.SignedAt.IsZero() {
				body := "returnToYourTaskListForInformationAboutWhatToDoNext"
				if donor.Tasks.SignTheLpa.IsCompleted() {
					body = "youDoNotNeedToTakeAnyAction"
				}

				data.SuccessNotifications = append(data.SuccessNotifications, progressNotification{
					Heading: appData.Localizer.Format(
						"voucherHasConfirmedYourIdentity",
						map[string]any{"VoucherFullName": voucher.FullName()},
					),
					Body: body,
				})

				donor.HasSeenSuccessfulVouchBanner = true

				if err = donorStore.Put(r.Context(), donor); err != nil {
					return err
				}
			}
		}

		if donor.Tasks.PayForLpa.IsPending() && donor.FeeAmount() == 0 {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: appData.Localizer.T("weAreReviewingTheEvidenceYouSent"),
				Body:    appData.Localizer.T("ifYourEvidenceIsApprovedWillShowPaid"),
			})
		}

		if !donor.HasSeenReducedFeeApprovalNotification &&
			!donor.ReducedFeeApprovedAt.IsZero() &&
			donor.Tasks.PayForLpa.IsCompleted() {
			data.SuccessNotifications = append(data.SuccessNotifications, progressNotification{
				Heading: "weHaveApprovedYourLPAFeeRequest",
				Body:    "yourLPAIsNowPaid",
			})

			donor.HasSeenReducedFeeApprovalNotification = true

			if err = donorStore.Put(r.Context(), donor); err != nil {
				return fmt.Errorf("failed to update donor: %v", err)
			}
		}

		if donor.RegisteringWithCourtOfProtection {
			if donor.Tasks.PayForLpa.IsCompleted() && !donor.WitnessedByCertificateProviderAt.IsZero() {
				data.InfoNotifications = append(data.InfoNotifications, progressNotification{
					Heading: appData.Localizer.T("yourLpaMustBeReviewedByCourtOfProtection"),
					Body:    appData.Localizer.T("opgIsCompletingChecksSoYouCanSubmitToCourtOfProtection"),
				})
			} else if donor.Tasks.PayForLpa.IsCompleted() {
				data.InfoNotifications = append(data.InfoNotifications, progressNotification{
					Heading: appData.Localizer.T("yourLpaMustBeReviewedByCourtOfProtection"),
					Body:    appData.Localizer.T("returnToYourTaskListToSignThenOpgWillCheck"),
				})
			} else if !donor.WitnessedByCertificateProviderAt.IsZero() {
				data.InfoNotifications = append(data.InfoNotifications, progressNotification{
					Heading: appData.Localizer.T("yourLpaMustBeReviewedByCourtOfProtection"),
					Body:    appData.Localizer.T("whenYouHavePaidOpgWillCheck"),
				})
			}
		}

		if now().After(donor.IdentityDeadline()) && donor.Tasks.SignTheLpa.IsCompleted() && !donor.Tasks.ConfirmYourIdentity.IsCompleted() {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: appData.Localizer.T("yourLPACannotBeRegisteredByOPG"),
				Body:    appData.Localizer.T("youDidNotConfirmYourIdentityWithinSixMonthsOfSigning"),
			})
		}

		if donor.IdentityUserData.Status.IsExpired() && !donor.Tasks.SignTheLpa.IsCompleted() {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: appData.Localizer.T("youMustConfirmYourIdentityAgain"),
				Body:    appData.Localizer.T("youDidNotSignYourLPAWithinSixMonthsOfConfirmingYourIdentity"),
			})
		}

		if lpa.Status.IsStatutoryWaitingPeriod() {
			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: appData.Localizer.T("yourLpaIsAwaitingRegistration"),
				Body:    appData.Localizer.T("theOpgWillRegisterYourLpaAtEndOfWaitingPeriod"),
			})
		}

		return tmpl(w, data)
	}
}
