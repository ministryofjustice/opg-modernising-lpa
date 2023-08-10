package app

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateReducedFee(t *testing.T) {
	now := time.Date(2000, time.March, 4, 0, 0, 0, 0, time.UTC)
	ctx := context.Background()
	lpa := &page.Lpa{
		UID: "lpa-uid",
		PaymentDetails: page.PaymentDetails{
			PaymentId: "payment-id",
			Amount:    123,
		},
		FeeType:     page.HalfFee,
		EvidenceKey: "http://evidence-key",
	}

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
		Return(nil)

	store := &eventStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	assert.Nil(t, store.CreateReducedFee(ctx, lpa))
}

func TestCreateReducedFeeOnDynamoError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Create", mock.Anything, mock.Anything).
		Return(expectedError)

	store := &eventStore{
		dynamoClient: dynamoClient,
		now:          time.Now,
	}

	assert.Equal(t, expectedError, store.CreateReducedFee(context.Background(), &page.Lpa{}))
}
