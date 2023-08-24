package app

import (
	"context"
	"strconv"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
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
	reducedFee := &reducedFee{
		PK:        "LPAUID#" + lpa.UID,
		SK:        "#DATE#" + strconv.FormatInt(r.now().Unix(), 10),
		PaymentID: lpa.PaymentDetails.PaymentId,
		LpaUID:    lpa.UID,
		FeeType:   lpa.FeeType.String(),
		Amount:    lpa.PaymentDetails.Amount,
		UpdatedAt: r.now(),
		//TODO just reference multiple keys on lpa when we support multi-file uploads
		EvidenceKeys: []string{lpa.EvidenceKey},
	}

	return r.dynamoClient.Create(ctx, reducedFee)
}
