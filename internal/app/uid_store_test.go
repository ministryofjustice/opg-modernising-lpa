package app

import (
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUidStoreSet(t *testing.T) {
	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, search.Lpa{
			PK:    dynamo.LpaKey("lpa-id").PK(),
			SK:    dynamo.DonorKey("a-donor").SK(),
			Donor: search.LpaDonor{FirstNames: "x", LastName: "y"},
		}).
		Return(nil)

	dynamoClient := newMockDynamoUpdateClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(ctx, &dynamo.Transaction{
			Puts: []any{
				&donordata.Provided{
					PK:        dynamo.LpaKey("lpa-id"),
					SK:        dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
					Donor:     donordata.Donor{FirstNames: "x", LastName: "y"},
					LpaUID:    "uid",
					UpdatedAt: testNow,
				},
			},
			Creates: []any{dynamo.Keys{PK: dynamo.UIDKey("uid"), SK: dynamo.MetadataKey("")}},
		}).
		Return(nil)

	uidStore := NewUidStore(dynamoClient, searchClient, testNowFn)

	err := uidStore.Set(ctx, &donordata.Provided{
		PK:    dynamo.LpaKey("lpa-id"),
		SK:    dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
		Donor: donordata.Donor{FirstNames: "x", LastName: "y"},
	}, "uid")
	assert.Nil(t, err)
}

func TestUidStoreSetWhenDynamoClientError(t *testing.T) {
	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(mock.Anything, mock.Anything).
		Return(nil)

	dynamoClient := newMockDynamoUpdateClient(t)
	dynamoClient.EXPECT().
		WriteTransaction(mock.Anything, mock.Anything).
		Return(expectedError)

	uidStore := NewUidStore(dynamoClient, searchClient, testNowFn)

	err := uidStore.Set(ctx, &donordata.Provided{}, "uid")
	assert.ErrorIs(t, err, expectedError)
}

func TestUidStoreSetWhenSearchIndexErrors(t *testing.T) {
	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, mock.Anything).
		Return(expectedError)

	uidStore := NewUidStore(nil, searchClient, testNowFn)
	err := uidStore.Set(ctx, &donordata.Provided{}, "uid")
	assert.ErrorIs(t, err, expectedError)
}
