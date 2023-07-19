package app

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreate(t *testing.T) {
	now := time.Now()
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
			SK:           "#PAYMENT#payment-id",
			PaymentID:    "payment-id",
			LpaUID:       "lpa-uid",
			FeeType:      "HalfFee",
			Amount:       123,
			UpdatedAt:    now,
			EvidenceKeys: []string{"http://evidence-key"},
		}).
		Return(nil)

	reducedFeeStore := reducedFeeStore{
		dynamoClient: dynamoClient,
		now:          func() time.Time { return now },
	}

	assert.Nil(t, reducedFeeStore.Create(ctx, lpa))
}

func TestCreateOnDynamoError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		On("Create", mock.Anything, mock.Anything).
		Return(expectedError)

	reducedFeeStore := reducedFeeStore{
		dynamoClient: dynamoClient,
		now:          time.Now,
	}

	assert.Equal(t, expectedError, reducedFeeStore.Create(context.Background(), &page.Lpa{}))
}
