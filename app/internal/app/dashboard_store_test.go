package app

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestDashboardStoreGetAll(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	lpa0 := &page.Lpa{ID: "0", UpdatedAt: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)}
	lpa123 := &page.Lpa{ID: "123", UpdatedAt: time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC)}
	lpa456 := &page.Lpa{ID: "456"}
	lpa789 := &page.Lpa{ID: "789"}

	dataStore := newMockDataStore(t)
	dataStore.ExpectGetAllByGsi(ctx, "ActorIndex", "#SUB#an-id",
		[]map[string]any{
			{"PK": "LPA#123", "SK": "#SUB#an-id", "Data": "#DONOR#an-id|DONOR"},
			{"PK": "LPA#456", "SK": "#SUB#an-id", "Data": "#DONOR#another-id|CERTIFICATE_PROVIDER"},
			{"PK": "LPA#789", "SK": "#SUB#an-id", "Data": "#DONOR#different-id|ATTORNEY"},
			{"PK": "LPA#0", "SK": "#SUB#an-id", "Data": "#DONOR#an-id|DONOR"},
		}, nil)
	dataStore.ExpectGetAllByKeys(ctx, []dynamo.Key{
		{PK: "LPA#123", SK: "#DONOR#an-id"},
		{PK: "LPA#456", SK: "#DONOR#another-id"},
		{PK: "LPA#789", SK: "#DONOR#different-id"},
		{PK: "LPA#0", SK: "#DONOR#an-id"},
	}, []struct {
		PK   string
		Data *page.Lpa
	}{
		{PK: "LPA#123", Data: lpa123},
		{PK: "LPA#456", Data: lpa456},
		{PK: "LPA#789", Data: lpa789},
		{PK: "LPA#0", Data: lpa0},
	}, nil)

	dashboardStore := &dashboardStore{dataStore: dataStore}

	donor, attorney, certificateProvider, err := dashboardStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, []*page.Lpa{lpa123, lpa0}, donor)
	assert.Equal(t, []*page.Lpa{lpa789}, attorney)
	assert.Equal(t, []*page.Lpa{lpa456}, certificateProvider)
}
