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
	PK        string
	SK        string
	PaymentID string
	// This is the Sirius UID - better name?
	LpaUID     string
	FeeType    string
	AmountPaid int
	UpdatedAt  time.Time
	//TODO add S3 refs once we're saving docs
}

func (r *reducedFeeStore) Put(ctx context.Context, lpa *page.Lpa) error {
	lpa.UpdatedAt = r.now()
	return r.dynamoClient.Put(ctx, lpa)
}

func (r *reducedFeeStore) Create(ctx context.Context, lpa *page.Lpa) error {
	if err := r.dynamoClient.Create(ctx, &reducedFee{
		PK:        "LPAUID#" + lpa.UID,
		SK:        "#PAYMENT#" + lpa.PaymentDetails.PaymentId,
		PaymentID: lpa.PaymentDetails.PaymentId,
		LpaUID:    lpa.UID,
		FeeType:   lpa.FeeType.String(),
		UpdatedAt: r.now(),
	}); err != nil {
		return err
	}

	return nil
}
