package app

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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
			Updates: []*types.Update{{
				Key: map[string]types.AttributeValue{
					"PK": &types.AttributeValueMemberS{Value: "LPA#lpa-id"},
					"SK": &types.AttributeValueMemberS{Value: "SUB#a-donor"},
				},
				UpdateExpression: aws.String("SET #Field = :Value"),
				ExpressionAttributeValues: map[string]types.AttributeValue{
					":Value": &types.AttributeValueMemberS{Value: "uid"},
				},
				ExpressionAttributeNames: map[string]string{
					"#Field": "LpaUID",
				},
			}},
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

	err := uidStore.Set(ctx, &donordata.Provided{
		SK: dynamo.LpaOwnerKey(dynamo.DonorKey("a-donor")),
	}, "uid")
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
