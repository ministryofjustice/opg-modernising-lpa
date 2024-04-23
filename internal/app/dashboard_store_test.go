package app

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/lpastore"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestDashboardStoreGetAll(t *testing.T) {
	sessionID := "an-id"
	aTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	lpa0 := &lpastore.Lpa{LpaID: "0", LpaUID: "M", UpdatedAt: aTime}
	lpa0Donor := &actor.DonorProvidedDetails{LpaID: "0", LpaUID: "M", UpdatedAt: aTime, SK: dynamo.DonorKey(sessionID), PK: dynamo.LpaKey("0")}
	lpa123 := &lpastore.Lpa{LpaID: "123", LpaUID: "M", UpdatedAt: aTime}
	lpa123Donor := &actor.DonorProvidedDetails{LpaID: "123", LpaUID: "M", UpdatedAt: aTime, SK: dynamo.DonorKey(sessionID), PK: dynamo.LpaKey("123")}
	lpa456 := &lpastore.Lpa{LpaID: "456", LpaUID: "M"}
	lpa456Donor := &actor.DonorProvidedDetails{LpaID: "456", LpaUID: "M", SK: dynamo.DonorKey("another-id"), PK: dynamo.LpaKey("456")}
	lpa456CertificateProvider := &actor.CertificateProviderProvidedDetails{
		LpaID: "456", Tasks: actor.CertificateProviderTasks{ConfirmYourDetails: actor.TaskCompleted}, SK: dynamo.CertificateProviderKey(sessionID),
	}
	lpa789 := &lpastore.Lpa{LpaID: "789", LpaUID: "M"}
	lpa789Donor := &actor.DonorProvidedDetails{LpaID: "789", LpaUID: "M", SK: dynamo.DonorKey("different-id"), PK: dynamo.LpaKey("789")}
	lpa789Attorney := &actor.AttorneyProvidedDetails{
		LpaID: "789", Tasks: actor.AttorneyTasks{ConfirmYourDetails: actor.TaskInProgress}, SK: dynamo.AttorneyKey(sessionID),
	}
	lpaNoUIDDonor := &actor.DonorProvidedDetails{LpaID: "999", UpdatedAt: aTime, SK: dynamo.DonorKey(sessionID), PK: dynamo.LpaKey("0")}
	lpaCertified := &lpastore.Lpa{LpaID: "signed-by-cp", LpaUID: "M"}
	lpaCertifiedDonor := &actor.DonorProvidedDetails{LpaID: "signed-by-cp", LpaUID: "M", SK: dynamo.DonorKey("another-id"), PK: dynamo.LpaKey("signed-by-cp")}
	lpaCertifiedCertificateProvider := &actor.CertificateProviderProvidedDetails{
		LpaID: "signed-by-cp", SK: dynamo.CertificateProviderKey(sessionID), Certificate: actor.Certificate{AgreeToStatement: true},
	}

	testCases := map[string][]map[string]types.AttributeValue{
		"details returned after lpas": {
			makeAttributeValueMap(lpa123Donor),
			makeAttributeValueMap(lpa456Donor),
			makeAttributeValueMap(lpa456CertificateProvider),
			makeAttributeValueMap(lpa789Donor),
			makeAttributeValueMap(lpa789Attorney),
			makeAttributeValueMap(lpa0Donor),
			makeAttributeValueMap(lpaCertifiedDonor),
			makeAttributeValueMap(lpaCertifiedCertificateProvider),
		},
		"details returned before lpas": {
			makeAttributeValueMap(lpaNoUIDDonor),
			makeAttributeValueMap(lpa456CertificateProvider),
			makeAttributeValueMap(lpa789Attorney),
			makeAttributeValueMap(lpaCertifiedCertificateProvider),
			makeAttributeValueMap(lpa123Donor),
			makeAttributeValueMap(lpa456Donor),
			makeAttributeValueMap(lpa789Donor),
			makeAttributeValueMap(lpa0Donor),
			makeAttributeValueMap(lpaCertifiedDonor),
		},
	}

	for name, attributeValues := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: sessionID})

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
				[]lpaLink{
					{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.DonorKey("an-id"), ActorType: actor.TypeDonor},
					{PK: dynamo.LpaKey("456"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.DonorKey("another-id"), ActorType: actor.TypeCertificateProvider},
					{PK: dynamo.LpaKey("789"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.DonorKey("different-id"), ActorType: actor.TypeAttorney},
					{PK: dynamo.LpaKey("0"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.DonorKey("an-id"), ActorType: actor.TypeDonor},
					{PK: dynamo.LpaKey("999"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.DonorKey("an-id"), ActorType: actor.TypeDonor},
					{PK: dynamo.LpaKey("signed-by-cp"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.DonorKey("another-id"), ActorType: actor.TypeCertificateProvider},
				}, nil)
			dynamoClient.ExpectAllByKeys(ctx, []dynamo.Key{
				{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("an-id")},
				{PK: dynamo.LpaKey("456"), SK: dynamo.DonorKey("another-id")},
				{PK: dynamo.LpaKey("456"), SK: dynamo.CertificateProviderKey("an-id")},
				{PK: dynamo.LpaKey("789"), SK: dynamo.DonorKey("different-id")},
				{PK: dynamo.LpaKey("789"), SK: dynamo.AttorneyKey("an-id")},
				{PK: dynamo.LpaKey("0"), SK: dynamo.DonorKey("an-id")},
				{PK: dynamo.LpaKey("999"), SK: dynamo.DonorKey("an-id")},
				{PK: dynamo.LpaKey("signed-by-cp"), SK: dynamo.DonorKey("another-id")},
				{PK: dynamo.LpaKey("signed-by-cp"), SK: dynamo.CertificateProviderKey("an-id")},
			}, attributeValues, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				ResolveList(ctx, []*actor.DonorProvidedDetails{lpa123Donor, lpa456Donor, lpa789Donor, lpa0Donor, lpaCertifiedDonor}).
				Return([]*lpastore.Lpa{lpa123, lpa456, lpa789, lpa0, lpaCertified}, nil)

			dashboardStore := &dashboardStore{dynamoClient: dynamoClient, lpaStoreResolvingService: lpaStoreResolvingService}

			donor, attorney, certificateProvider, err := dashboardStore.GetAll(ctx)
			assert.Nil(t, err)

			assert.Equal(t, []page.LpaAndActorTasks{{Lpa: lpa123}, {Lpa: lpa0}}, donor)
			assert.Equal(t, []page.LpaAndActorTasks{{Lpa: lpa456, CertificateProvider: lpa456CertificateProvider}}, certificateProvider)
			assert.Equal(t, []page.LpaAndActorTasks{{Lpa: lpa789, Attorney: lpa789Attorney}}, attorney)
		})
	}
}

func TestDashboardStoreGetAllSubmittedForAttorneys(t *testing.T) {
	sessionID := "an-id"
	aTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	lpaSubmitted := &lpastore.Lpa{LpaID: "submitted", LpaUID: "M", Submitted: true}
	lpaSubmittedDonor := &actor.DonorProvidedDetails{LpaID: "submitted", LpaUID: "M", SK: dynamo.DonorKey("another-id"), PK: dynamo.LpaKey("submitted"), SubmittedAt: aTime}
	lpaSubmittedAttorney := &actor.AttorneyProvidedDetails{LpaID: "submitted", SK: dynamo.AttorneyKey(sessionID)}
	lpaSubmittedReplacement := &lpastore.Lpa{LpaID: "submitted-replacement", LpaUID: "M", Submitted: true}
	lpaSubmittedReplacementDonor := &actor.DonorProvidedDetails{LpaID: "submitted-replacement", LpaUID: "M", SK: dynamo.DonorKey("another-id"), PK: dynamo.LpaKey("submitted-replacement"), SubmittedAt: aTime}
	lpaSubmittedReplacementAttorney := &actor.AttorneyProvidedDetails{LpaID: "submitted-replacement", SK: dynamo.AttorneyKey(sessionID), IsReplacement: true}
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: sessionID})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]lpaLink{
			{PK: dynamo.LpaKey("submitted"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.DonorKey("another-id"), ActorType: actor.TypeAttorney},
			{PK: dynamo.LpaKey("submitted-replacement"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.DonorKey("another-id"), ActorType: actor.TypeAttorney},
		}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Key{
		{PK: dynamo.LpaKey("submitted"), SK: dynamo.DonorKey("another-id")},
		{PK: dynamo.LpaKey("submitted"), SK: dynamo.AttorneyKey("an-id")},
		{PK: dynamo.LpaKey("submitted-replacement"), SK: dynamo.DonorKey("another-id")},
		{PK: dynamo.LpaKey("submitted-replacement"), SK: dynamo.AttorneyKey("an-id")},
	}, []map[string]types.AttributeValue{
		makeAttributeValueMap(lpaSubmittedDonor),
		makeAttributeValueMap(lpaSubmittedAttorney),
		makeAttributeValueMap(lpaSubmittedReplacementDonor),
		makeAttributeValueMap(lpaSubmittedReplacementAttorney),
	}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().
		ResolveList(ctx, []*actor.DonorProvidedDetails{lpaSubmittedDonor, lpaSubmittedReplacementDonor}).
		Return([]*lpastore.Lpa{lpaSubmitted, lpaSubmittedReplacement}, nil)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient, lpaStoreResolvingService: lpaStoreResolvingService}

	_, attorney, _, err := dashboardStore.GetAll(ctx)
	assert.Nil(t, err)

	assert.Equal(t, []page.LpaAndActorTasks{
		{Lpa: lpaSubmitted, Attorney: lpaSubmittedAttorney},
	}, attorney)
}

func makeAttributeValueMap(i interface{}) map[string]types.AttributeValue {
	result, _ := attributevalue.MarshalMap(i)
	return result
}

func TestDashboardStoreGetAllWhenResolveErrors(t *testing.T) {
	sessionID := "an-id"
	aTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	donor := &actor.DonorProvidedDetails{LpaID: "0", LpaUID: "M", UpdatedAt: aTime, SK: dynamo.DonorKey(sessionID), PK: dynamo.LpaKey("0")}

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: sessionID})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]lpaLink{
			{PK: dynamo.LpaKey("0"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.DonorKey("an-id"), ActorType: actor.TypeDonor},
		}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Key{
		{PK: dynamo.LpaKey("0"), SK: dynamo.DonorKey("an-id")},
	}, []map[string]types.AttributeValue{makeAttributeValueMap(donor)}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().ResolveList(ctx, mock.Anything).Return(nil, expectedError)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient, lpaStoreResolvingService: lpaStoreResolvingService}

	_, _, _, err := dashboardStore.GetAll(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDashboardStoreGetAllWhenNone(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]map[string]any{}, nil)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient}

	donor, attorney, certificateProvider, err := dashboardStore.GetAll(ctx)
	assert.Nil(t, err)
	assert.Nil(t, donor)
	assert.Nil(t, attorney)
	assert.Nil(t, certificateProvider)
}

func TestDashboardStoreGetAllWhenAllForActorErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]lpaLink{}, expectedError)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient}

	_, _, _, err := dashboardStore.GetAll(ctx)
	assert.Equal(t, err, expectedError)
}

func TestDashboardStoreGetAllWhenAllByKeysErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]lpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.DonorKey("an-id"), ActorType: actor.TypeDonor}}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Key{
		{PK: dynamo.LpaKey("123"), SK: dynamo.DonorKey("an-id")},
	}, nil, expectedError)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient}

	_, _, _, err := dashboardStore.GetAll(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDashboardStoreSubExists(t *testing.T) {
	testCases := map[string]struct {
		lpas           []lpaLink
		expectedExists bool
		actorType      actor.Type
	}{
		"lpas exist - correct actor": {
			lpas:           []lpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("a-sub-id"), DonorKey: dynamo.DonorKey("an-id"), ActorType: actor.TypeDonor}},
			expectedExists: true,
			actorType:      actor.TypeDonor,
		},
		"lpas exist - incorrect actor": {
			lpas:           []lpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("a-sub-id"), DonorKey: dynamo.DonorKey("an-id"), ActorType: actor.TypeDonor}},
			expectedExists: false,
			actorType:      actor.TypeAttorney,
		},
		"lpas do not exist": {
			expectedExists: false,
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.ExpectAllBySK(context.Background(), dynamo.SubKey("a-sub-id"),
				tc.lpas, nil)

			dashboardStore := &dashboardStore{dynamoClient: dynamoClient}
			exists, err := dashboardStore.SubExistsForActorType(context.Background(), "a-sub-id", tc.actorType)

			assert.Nil(t, err)
			assert.Equal(t, tc.expectedExists, exists)
		})
	}
}

func TestDashboardStoreSubExistsWhenDynamoError(t *testing.T) {
	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(context.Background(), dynamo.SubKey("a-sub-id"),
		[]lpaLink{}, expectedError)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient}
	exists, err := dashboardStore.SubExistsForActorType(context.Background(), "a-sub-id", actor.TypeDonor)

	assert.Equal(t, expectedError, err)
	assert.False(t, exists)
}

func TestLpaLinkUserSub(t *testing.T) {
	assert.Equal(t, "a-sub", lpaLink{SK: dynamo.SubKey("a-sub")}.UserSub())
	assert.Equal(t, "", lpaLink{}.UserSub())
}
