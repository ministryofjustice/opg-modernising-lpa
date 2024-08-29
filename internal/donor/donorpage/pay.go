package donorpage

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/pay"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/sesh"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/task"
)

func Pay(
	logger Logger,
	sessionStore SessionStore,
	donorStore DonorStore,
	payClient PayClient,
	randomString func(int) string,
	appPublicURL string,
) Handler {
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, provided *donordata.Provided) error {
		if provided.FeeType.IsNoFee() || provided.FeeType.IsHardshipFee() || provided.Tasks.PayForLpa.IsMoreEvidenceRequired() {
			provided.Tasks.PayForLpa = task.PaymentStatePending

			if provided.EvidenceDelivery.IsUpload() {
				provided.ProgressSteps.Complete(task.FeeEvidenceSubmitted, time.Now())
			}

			if err := donorStore.Put(r.Context(), provided); err != nil {
				return err
			}

			if provided.EvidenceDelivery.IsPost() {
				return donor.PathWhatHappensNextPostEvidence.Redirect(w, r, appData, provided)
			}

			return donor.PathEvidenceSuccessfullyUploaded.Redirect(w, r, appData, provided)
		}

		createPaymentBody := pay.CreatePaymentBody{
			Amount:      provided.FeeAmount().Pence(),
			Reference:   randomString(12),
			Description: "Property and Finance LPA",
			ReturnURL:   appPublicURL + appData.Lang.URL(donor.PathPaymentConfirmation.Format(provided.LpaID)),
			Email:       provided.Donor.Email,
			Language:    appData.Lang.String(),
		}

		resp, err := payClient.CreatePayment(r.Context(), provided.LpaUID, createPaymentBody)
		if err != nil {
			return fmt.Errorf("error creating payment: %w", err)
		}

		if err = sessionStore.SetPayment(r, w, &sesh.PaymentSession{PaymentID: resp.PaymentID}); err != nil {
			return err
		}

		nextUrl := resp.Links["next_url"].Href
		// If URL matches expected domain for GOV UK PAY redirect there. If not,
		// redirect to the confirmation code and carry on with flow.
		if payClient.CanRedirect(nextUrl) {
			http.Redirect(w, r, nextUrl, http.StatusFound)
			return nil
		}

		logger.InfoContext(r.Context(), "skipping payment", slog.String("next_url", nextUrl))
		return donor.PathPaymentConfirmation.Redirect(w, r, appData, provided)
	}
}
