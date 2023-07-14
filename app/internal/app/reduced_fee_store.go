package app

import (
	"context"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
)

type reducedFeeStore struct {
	dataStore DynamoClient
	now       func() time.Time
}

func (r *reducedFeeStore) Put(ctx context.Context, lpa *page.Lpa) error {
	lpa.UpdatedAt = r.now()
	return r.dataStore.Put(ctx, lpa)
}
