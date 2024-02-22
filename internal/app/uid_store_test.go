package app

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/search"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestUidStoreSet(t *testing.T) {
	testcases := map[string]struct {
		organisationID string
		sk             string
	}{
		"donor": {
			sk: "#DONOR#session-id",
		},
		"organisation": {
			organisationID: "org-id",
			sk:             "ORGANISATION#org-id",
		},
	}

	for name, tc := range testcases {
		t.Run(name, func(t *testing.T) {
			values, _ := attributevalue.MarshalMap(map[string]any{
				":uid": "uid",
				":now": testNow,
			})

			returnValues, _ := attributevalue.MarshalMap(actor.DonorProvidedDetails{
				Donor: actor.Donor{
					FirstNames: "x",
					LastName:   "y",
				},
			})

			dynamoClient := newMockDynamoUpdateClient(t)
			dynamoClient.EXPECT().
				UpdateReturn(ctx, "LPA#lpa-id", tc.sk, values,
					"set LpaUID = :uid, UpdatedAt = :now").
				Return(returnValues, nil)

			searchClient := newMockSearchClient(t)
			searchClient.EXPECT().
				Index(ctx, search.Lpa{
					PK:            "LPA#lpa-id",
					SK:            tc.sk,
					DonorFullName: "x y",
				}).
				Return(nil)

			uidStore := NewUidStore(dynamoClient, func() (SearchClient, error) { return searchClient, nil }, testNowFn)

			assert.Nil(t, uidStore.Set(ctx, "lpa-id", "session-id", tc.organisationID, "uid"))
		})
	}
}

func TestUidStoreSetWhenSearchClientSet(t *testing.T) {
	returnValues, _ := attributevalue.MarshalMap(actor.DonorProvidedDetails{
		Donor: actor.Donor{
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
		Return(nil)

	uidStore := NewUidStore(dynamoClient, func() (SearchClient, error) { return nil, expectedError }, testNowFn)
	uidStore.searchClient = searchClient

	assert.Nil(t, uidStore.Set(ctx, "lpa-id", "session-id", "", "uid"))
}

func TestUidStoreSetWhenDynamoClientError(t *testing.T) {
	values, _ := attributevalue.MarshalMap(map[string]any{
		":uid": "uid",
		":now": testNow,
	})

	dynamoClient := newMockDynamoUpdateClient(t)
	dynamoClient.EXPECT().
		UpdateReturn(ctx, "LPA#lpa-id", "#DONOR#session-id", values,
			"set LpaUID = :uid, UpdatedAt = :now").
		Return(nil, expectedError)

	uidStore := NewUidStore(dynamoClient, func() (SearchClient, error) { return nil, nil }, testNowFn)

	assert.ErrorIs(t, uidStore.Set(ctx, "lpa-id", "session-id", "", "uid"), expectedError)
}

func TestUidStoreSetWhenSearchClientFactoryErrors(t *testing.T) {
	returnValues, _ := attributevalue.MarshalMap(actor.DonorProvidedDetails{
		Donor: actor.Donor{
			FirstNames: "x",
			LastName:   "y",
		},
	})

	dynamoClient := newMockDynamoUpdateClient(t)
	dynamoClient.EXPECT().
		UpdateReturn(ctx, mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(returnValues, nil)

	uidStore := NewUidStore(dynamoClient, func() (SearchClient, error) { return nil, expectedError }, testNowFn)
	err := uidStore.Set(ctx, "lpa-id", "session-id", "", "uid")
	assert.ErrorIs(t, err, expectedError)
}

func TestUidStoreSetWhenSearchIndexErrors(t *testing.T) {
	returnValues, _ := attributevalue.MarshalMap(actor.DonorProvidedDetails{
		Donor: actor.Donor{
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

	uidStore := NewUidStore(dynamoClient, func() (SearchClient, error) { return searchClient, nil }, testNowFn)
	err := uidStore.Set(ctx, "lpa-id", "session-id", "", "uid")
	assert.ErrorIs(t, err, expectedError)
}
