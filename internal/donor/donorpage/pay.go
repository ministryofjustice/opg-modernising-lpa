package donorpage

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/appcontext"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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
	return func(appData appcontext.Data, w http.ResponseWriter, r *http.Request, donor *donordata.Provided) error {
		if donor.FeeType.IsNoFee() || donor.FeeType.IsHardshipFee() || donor.Tasks.PayForLpa.IsMoreEvidenceRequired() {
			donor.Tasks.PayForLpa = task.PaymentStatePending
			if err := donorStore.Put(r.Context(), donor); err != nil {
				return err
			}

			if donor.EvidenceDelivery.IsPost() {
				return page.Paths.WhatHappensNextPostEvidence.Redirect(w, r, appData, donor)
			}

			return page.Paths.EvidenceSuccessfullyUploaded.Redirect(w, r, appData, donor)
		}

		createPaymentBody := pay.CreatePaymentBody{
			Amount:      donor.FeeAmount().Pence(),
			Reference:   randomString(12),
			Description: "Property and Finance LPA",
			ReturnURL:   appPublicURL + appData.Lang.URL(page.Paths.PaymentConfirmation.Format(donor.LpaID)),
			Email:       donor.Donor.Email,
			Language:    appData.Lang.String(),
		}

		resp, err := payClient.CreatePayment(r.Context(), donor.LpaUID, createPaymentBody)
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
		return page.Paths.PaymentConfirmation.Redirect(w, r, appData, donor)
	}
}
