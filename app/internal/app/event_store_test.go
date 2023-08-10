package app

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestCreateReducedFee(t *testing.T) {
	now := time.Date(2000, time.March, 4, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Create", ctx, &reducedFee{
			PK:           "LPAUID#lpa-uid",
			SK:           "#DATE#952128000",
			PaymentID:    "payment-id",
			LpaUID:       "lpa-uid",
			FeeType:      "HalfFee",
			Amount:       123,
			UpdatedAt:    now,
			EvidenceKeys: []string{"http://evidence-key"},
		}).
		Return(expectedError)

	store := &eventStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	assert.Equal(t, expectedError, store.CreateReducedFee(ctx, &page.Lpa{
		UID: "lpa-uid",
		PaymentDetails: page.PaymentDetails{
			PaymentId: "payment-id",
			Amount:    123,
		},
		FeeType:     page.HalfFee,
		EvidenceKey: "http://evidence-key",
	}))
}

func TestCreatePreviousApplication(t *testing.T) {
	now := time.Date(2000, time.March, 4, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Put", ctx, &previousApplication{
			PK:                        "LPA#lpa-uid",
			SK:                        "#PREVIOUSAPPLICATION",
			UpdatedAt:                 now,
			LpaUID:                    "lpa-uid",
			ApplicationReason:         page.RemakeOfInvalidApplication.String(),
			PreviousApplicationNumber: "123",
		}).
		Return(expectedError)

	store := &eventStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	assert.Equal(t, expectedError, store.CreatePreviousApplication(ctx, &page.Lpa{
		UID:                       "lpa-uid",
		ApplicationReason:         page.RemakeOfInvalidApplication,
		PreviousApplicationNumber: "123",
	}))
}
