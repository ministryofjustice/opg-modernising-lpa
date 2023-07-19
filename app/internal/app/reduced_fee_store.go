package app

import (
	"context"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
)

type reducedFeeStore struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

type reducedFee struct {
	PK           string
	SK           string
	PaymentID    string
	LpaUID       string
	FeeType      string
	Amount       int
	UpdatedAt    time.Time
	EvidenceKeys []string
}

func (r *reducedFeeStore) Create(ctx context.Context, lpa *page.Lpa) error {
	if err := r.dynamoClient.Create(ctx, &reducedFee{
		PK:        "LPAUID#" + lpa.UID,
		SK:        "#PAYMENT#" + lpa.PaymentDetails.PaymentId,
		PaymentID: lpa.PaymentDetails.PaymentId,
		LpaUID:    lpa.UID,
		FeeType:   lpa.FeeType.String(),
		Amount:    lpa.PaymentDetails.Amount,
		UpdatedAt: r.now(),
		//TODO just reference multiple keys when we support multi-file uploads
		EvidenceKeys: []string{lpa.EvidenceKey},
	}); err != nil {
		return err
	}

	return nil
}
