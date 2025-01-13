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
	"github.com/ministryofjustice/opg-modernising-lpa/internal/voucher/voucherdata"
)

type progressData struct {
	App               appcontext.Data
	Errors            validation.List
	Donor             *donordata.Provided
	Voucher           *voucherdata.Provided
	Progress          task.Progress
	InfoNotifications []task.ProgressNotification
}

func Progress(tmpl template.Template, lpaStoreResolvingService LpaStoreResolvingService, progressTracker ProgressTracker, certificateProviderStore CertificateProviderStore, voucherStore VoucherStore, donorStore DonorStore) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		lpa, err := lpaStoreResolvingService.Get(r.Context())
		if err != nil {
			return err
		}

		data := &progressData{
			App:      appData,
			Donor:    donor,
			Progress: progressTracker.Progress(lpa),
		}

		if donor.IdentityUserData.Status.IsUnknown() && donor.Tasks.ConfirmYourIdentity.IsPending() {
			data.InfoNotifications = append(data.InfoNotifications, task.ProgressNotification{
				Heading: "youHaveChosenToConfirmYourIdentityAtPostOffice",
				Body:    "whenYouHaveConfirmedAtPostOfficeReturnToTaskList",
			})
		}

		if lpa.Submitted && (lpa.CertificateProvider.SignedAt == nil || lpa.CertificateProvider.SignedAt.IsZero()) {
			_, err := certificateProviderStore.GetAny(r.Context())
			if errors.Is(err, dynamo.NotFoundError{}) {
				data.InfoNotifications = append(data.InfoNotifications, task.ProgressNotification{
					Heading: "youveSubmittedYourLpaToOpg",
					Body:    "opgIsCheckingYourLpa",
				})
			} else if err != nil {
				return err
			}
		}

		if donor.Tasks.PayForLpa.IsMoreEvidenceRequired() {
			body := appData.Localizer.Format(
				"weContactedYouOnWithGuidanceAboutWhatToDoNext",
				map[string]any{"ContactedDate": appData.Localizer.FormatDate(donor.MoreEvidenceRequiredAt)},
			)

			data.InfoNotifications = append(data.InfoNotifications, task.ProgressNotification{
				Heading: "weNeedMoreEvidenceToMakeADecisionAboutYourLPAFee",
				Body:    body,
			})
		}

		if !donor.Tasks.ConfirmYourIdentity.IsCompleted() && donor.Voucher.FirstNames != "" {
			if donor.VoucherInvitedAt.IsZero() && !donor.Tasks.PayForLpa.IsCompleted() {
				data.InfoNotifications = append(data.InfoNotifications, task.ProgressNotification{
					Heading: "youMustPayForYourLPA",
					Body: appData.Localizer.Format(
						"returnToTaskListToPayForLPAWeWillThenContactVoucher",
						map[string]any{"VoucherFullName": donor.Voucher.FullName()},
					),
				})
			} else if !donor.VoucherInvitedAt.IsZero() {
				data.InfoNotifications = append(data.InfoNotifications, task.ProgressNotification{
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
			data.InfoNotifications = append(data.InfoNotifications, task.ProgressNotification{
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

		if donor.Tasks.ConfirmYourIdentity.IsCompleted() && donor.Voucher.FirstNames != "" && !donor.ViewedProgressNotifications.HasSeen(task.NotificationSuccessfulVouch) {
			voucher, err := voucherStore.GetAny(r.Context())
			if err != nil {
				return err
			}

			if !voucher.SignedAt.IsZero() {
				body := "returnToYourTaskListForInformationAboutWhatToDoNext"
				if donor.Tasks.SignTheLpa.IsCompleted() {
					body = "youDoNotNeedToTakeAnyAction"
				}

				data.InfoNotifications = append(data.InfoNotifications, task.ProgressNotification{
					Heading: appData.Localizer.Format(
						"voucherHasConfirmedYourIdentity",
						map[string]any{"VoucherFullName": voucher.FullName()},
					),
					Body: body,
				})

				donor.ViewedProgressNotifications = append(donor.ViewedProgressNotifications, task.NotificationSuccessfulVouch)

				if err = donorStore.Put(r.Context(), donor); err != nil {
					return err
				}
			}
		}

		return tmpl(w, data)
	}
}
