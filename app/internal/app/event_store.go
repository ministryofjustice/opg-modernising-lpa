package app

import (
	"context"
	"strconv"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
)

type eventStore struct {
	dynamoClient DynamoClient
	now          func() time.Time
}

type reducedFee struct {
	PK, SK       string
	UpdatedAt    time.Time
	PaymentID    string
	LpaUID       string
	FeeType      string
	Amount       int
	EvidenceKeys []string
}

func (r *eventStore) CreateReducedFee(ctx context.Context, lpa *page.Lpa) error {
	return r.dynamoClient.Create(ctx, &reducedFee{
		PK:        "LPAUID#" + lpa.UID,
		SK:        "#DATE#" + strconv.FormatInt(r.now().Unix(), 10),
		UpdatedAt: r.now(),
		PaymentID: lpa.PaymentDetails.PaymentId,
		LpaUID:    lpa.UID,
		FeeType:   lpa.FeeType.String(),
		Amount:    lpa.PaymentDetails.Amount,
		//TODO just reference multiple keys on lpa when we support multi-file uploads
		EvidenceKeys: []string{lpa.EvidenceKey},
	})
}

type previousApplication struct {
	PK, SK                    string
	UpdatedAt                 time.Time
	LpaUID                    string
	ApplicationReason         string
	PreviousApplicationNumber string
}

func (r *eventStore) CreatePreviousApplication(ctx context.Context, lpa *page.Lpa) error {
	return r.dynamoClient.Put(ctx, &previousApplication{
		PK:                        "LPA#" + lpa.UID,
		SK:                        "#PREVIOUSAPPLICATION",
		UpdatedAt:                 r.now(),
		LpaUID:                    lpa.UID,
		ApplicationReason:         lpa.ApplicationReason.String(),
		PreviousApplicationNumber: lpa.PreviousApplicationNumber,
	})
}
