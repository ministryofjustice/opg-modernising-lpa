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
			_, err := certificateProviderStore.GetAny(r.Context())
			if errors.Is(err, dynamo.NotFoundError{}) {
				data.InfoNotifications = append(data.InfoNotifications, progressNotification{
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
				map[string]any{"MoreEvidenceRequiredAt": appData.Localizer.FormatDateTime(donor.MoreEvidenceRequiredAt)},
			)

			data.InfoNotifications = append(data.InfoNotifications, progressNotification{
				Heading: "weNeedMoreEvidenceToMakeADecisionAboutYourLPAFee",
				Body:    body,
			})
		}

		if !donor.Tasks.ConfirmYourIdentity.IsCompleted() && donor.Voucher.FirstNames != "" {
			var notification progressNotification

			if donor.VoucherInvitedAt.IsZero() && !donor.Tasks.PayForLpa.IsCompleted() {
				notification.Heading = "youMustPayForYourLPA"
				notification.Body = appData.Localizer.Format(
					"returnToTaskListToPayForLPAWeWillThenContactVoucher",
					map[string]any{"VoucherFullName": donor.Voucher.FullName()},
				)
			} else if !donor.VoucherInvitedAt.IsZero() {
				notification.Heading = appData.Localizer.Format(
					"weHaveContactedVoucherToConfirmYourIdentity",
					map[string]any{"VoucherFullName": donor.Voucher.FullName()},
				)
				notification.Body = "youDoNotNeedToTakeAnyAction"
			}

			data.InfoNotifications = append(data.InfoNotifications, notification)
		}

		return tmpl(w, data)
	}
}
