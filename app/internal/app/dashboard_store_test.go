package app

import (
	"context"
	"testing"
	"time"

	"github.com/ministryofjustice/opg-modernising-lpa/app/internal/actor"
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
		[]lpaLink{
			{PK: "LPA#123", SK: "#SUB#an-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor},
			{PK: "LPA#456", SK: "#SUB#an-id", DonorKey: "#DONOR#another-id", ActorType: actor.TypeCertificateProvider},
			{PK: "LPA#789", SK: "#SUB#an-id", DonorKey: "#DONOR#different-id", ActorType: actor.TypeAttorney},
			{PK: "LPA#0", SK: "#SUB#an-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor},
		}, nil)
	dataStore.ExpectGetAllByKeys(ctx, []dynamo.Key{
		{PK: "LPA#123", SK: "#DONOR#an-id"},
		{PK: "LPA#456", SK: "#DONOR#another-id"},
		{PK: "LPA#789", SK: "#DONOR#different-id"},
		{PK: "LPA#0", SK: "#DONOR#an-id"},
	}, []*page.Lpa{lpa123, lpa456, lpa789, lpa0}, nil)

	dashboardStore := &dashboardStore{dynamoClient: dataStore}

	donor, attorney, certificateProvider, err := dashboardStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Equal(t, []*page.Lpa{lpa123, lpa0}, donor)
	assert.Equal(t, []*page.Lpa{lpa789}, attorney)
	assert.Equal(t, []*page.Lpa{lpa456}, certificateProvider)
}

func TestDashboardStoreGetAllWhenNone(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dataStore := newMockDataStore(t)
	dataStore.ExpectGetAllByGsi(ctx, "ActorIndex", "#SUB#an-id",
		[]map[string]any{}, nil)

	dashboardStore := &dashboardStore{dynamoClient: dataStore}

	donor, attorney, certificateProvider, err := dashboardStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Nil(t, donor)
	assert.Nil(t, attorney)
	assert.Nil(t, certificateProvider)
}
