package app

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/donor/donordata"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestUidStoreSet(t *testing.T) {
	testcases := map[string]struct {
		organisationID string
		sk             dynamo.SK
	}{
		"donor": {
			sk: dynamo.DonorKey("session-id"),
		},
		"organisation": {
			organisationID: "org-id",
			sk:             dynamo.OrganisationKey("org-id"),
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			values, _ := attributevalue.MarshalMap(map[string]any{
				":uid": "uid",
				":now": testNow,
			})

			returnValues, _ := attributevalue.MarshalMap(donordata.Provided{
				Donor: donordata.Donor{
					FirstNames: "x",
					LastName:   "y",
				},
			})

			dynamoClient := newMockDynamoUpdateClient(t)
			dynamoClient.EXPECT().
				UpdateReturn(ctx, dynamo.LpaKey("lpa-id"), tc.sk, values,
					"set LpaUID = :uid, UpdatedAt = :now").
				Return(returnValues, nil)

			searchClient := newMockSearchClient(t)
			searchClient.EXPECT().
				Index(ctx, search.Lpa{
					PK:    dynamo.LpaKey("lpa-id").PK(),
					SK:    tc.sk.SK(),
					Donor: search.LpaDonor{FirstNames: "x", LastName: "y"},
				}).
				Return(nil)

			uidStore := NewUidStore(dynamoClient, searchClient, testNowFn)

			assert.Nil(t, uidStore.Set(ctx, "lpa-id", "session-id", tc.organisationID, "uid"))
		})
	}
}

func TestUidStoreSetWhenDynamoClientError(t *testing.T) {
	values, _ := attributevalue.MarshalMap(map[string]any{
		":uid": "uid",
		":now": testNow,
	})

	dynamoClient := newMockDynamoUpdateClient(t)
	dynamoClient.EXPECT().
		UpdateReturn(ctx, dynamo.LpaKey("lpa-id"), dynamo.DonorKey("session-id"), values,
			"set LpaUID = :uid, UpdatedAt = :now").
		Return(nil, expectedError)

	uidStore := NewUidStore(dynamoClient, nil, testNowFn)

	assert.ErrorIs(t, uidStore.Set(ctx, "lpa-id", "session-id", "", "uid"), expectedError)
}

func TestUidStoreSetWhenSearchIndexErrors(t *testing.T) {
	returnValues, _ := attributevalue.MarshalMap(donordata.Provided{
		Donor: donordata.Donor{
			FirstNames: "x",
			LastName:   "y",
		},
	})

	dynamoClient := newMockDynamoUpdateClient(t)
	dynamoClient.EXPECT().
		UpdateReturn(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(returnValues, nil)

	searchClient := newMockSearchClient(t)
	searchClient.EXPECT().
		Index(ctx, mock.Anything).
		Return(expectedError)

	uidStore := NewUidStore(dynamoClient, searchClient, testNowFn)
	err := uidStore.Set(ctx, "lpa-id", "session-id", "", "uid")
	assert.ErrorIs(t, err, expectedError)
}
