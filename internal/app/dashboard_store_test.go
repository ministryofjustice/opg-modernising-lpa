package app

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/attorney/attorneydata"
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
	lpa0Donor := &actor.DonorProvidedDetails{
		PK:        dynamo.LpaKey("0"),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey(sessionID)),
		LpaID:     "0",
		LpaUID:    "M",
		UpdatedAt: aTime,
	}
	lpa123 := &lpastore.Lpa{LpaID: "123", LpaUID: "M", UpdatedAt: aTime}
	lpa123Donor := &actor.DonorProvidedDetails{
		PK:        dynamo.LpaKey("123"),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey(sessionID)),
		LpaID:     "123",
		LpaUID:    "M",
		UpdatedAt: aTime,
	}
	lpa456 := &lpastore.Lpa{LpaID: "456", LpaUID: "M"}
	lpa456Donor := &actor.DonorProvidedDetails{
		PK:     dynamo.LpaKey("456"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")),
		LpaID:  "456",
		LpaUID: "M",
	}
	lpa456CertificateProvider := &actor.CertificateProviderProvidedDetails{
		PK:    dynamo.LpaKey("456"),
		SK:    dynamo.CertificateProviderKey(sessionID),
		LpaID: "456",
		Tasks: actor.CertificateProviderTasks{ConfirmYourDetails: actor.TaskCompleted},
	}
	lpa789 := &lpastore.Lpa{LpaID: "789", LpaUID: "M"}
	lpa789Donor := &actor.DonorProvidedDetails{
		PK:     dynamo.LpaKey("789"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("different-id")),
		LpaID:  "789",
		LpaUID: "M",
	}
	lpa789Attorney := &attorneydata.Provided{
		PK:    dynamo.LpaKey("789"),
		SK:    dynamo.AttorneyKey(sessionID),
		LpaID: "789",
		Tasks: attorneydata.Tasks{ConfirmYourDetails: actor.TaskInProgress},
	}
	lpaNoUIDDonor := &actor.DonorProvidedDetails{
		PK:        dynamo.LpaKey("0"),
		SK:        dynamo.LpaOwnerKey(dynamo.DonorKey(sessionID)),
		LpaID:     "999",
		UpdatedAt: aTime,
	}
	lpaCertified := &lpastore.Lpa{LpaID: "signed-by-cp", LpaUID: "M"}
	lpaCertifiedDonor := &actor.DonorProvidedDetails{
		PK:     dynamo.LpaKey("signed-by-cp"),
		SK:     dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")),
		LpaID:  "signed-by-cp",
		LpaUID: "M",
	}
	lpaCertifiedCertificateProvider := &actor.CertificateProviderProvidedDetails{
		PK:       dynamo.LpaKey("signed-by-cp"),
		SK:       dynamo.CertificateProviderKey(sessionID),
		LpaID:    "signed-by-cp",
		SignedAt: time.Now(),
	}
	lpaReferenced := &lpastore.Lpa{LpaID: "referenced", LpaUID: "X"}
	lpaReferencedLink := map[string]any{
		"PK":           dynamo.LpaKey("referenced"),
		"SK":           dynamo.DonorKey(sessionID),
		"ReferencedSK": dynamo.OrganisationKey("org-id"),
	}
	lpaReferencedDonor := &actor.DonorProvidedDetails{
		PK:     dynamo.LpaKey("referenced"),
		SK:     dynamo.LpaOwnerKey(dynamo.OrganisationKey("org-id")),
		LpaID:  "referenced",
		LpaUID: "X",
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
			makeAttributeValueMap(lpaReferencedLink),
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
			makeAttributeValueMap(lpaReferencedLink),
		},
	}

	for name, attributeValues := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: sessionID})

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
				[]lpaLink{
					{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor},
					{PK: dynamo.LpaKey("456"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")), ActorType: actor.TypeCertificateProvider},
					{PK: dynamo.LpaKey("789"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("different-id")), ActorType: actor.TypeAttorney},
					{PK: dynamo.LpaKey("0"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor},
					{PK: dynamo.LpaKey("999"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor},
					{PK: dynamo.LpaKey("signed-by-cp"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")), ActorType: actor.TypeCertificateProvider},
					{PK: dynamo.LpaKey("referenced"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor},
				}, nil)
			dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
				{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("456"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id"))},
				{PK: dynamo.LpaKey("456"), SK: dynamo.CertificateProviderKey("an-id")},
				{PK: dynamo.LpaKey("789"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("different-id"))},
				{PK: dynamo.LpaKey("789"), SK: dynamo.AttorneyKey("an-id")},
				{PK: dynamo.LpaKey("0"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("999"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
				{PK: dynamo.LpaKey("signed-by-cp"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id"))},
				{PK: dynamo.LpaKey("signed-by-cp"), SK: dynamo.CertificateProviderKey("an-id")},
				{PK: dynamo.LpaKey("referenced"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
			}, attributeValues, nil)
			dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
				{PK: dynamo.LpaKey("referenced"), SK: dynamo.OrganisationKey("org-id")},
			}, []map[string]types.AttributeValue{
				makeAttributeValueMap(lpaReferencedDonor),
			}, nil)

			lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
			lpaStoreResolvingService.EXPECT().
				ResolveList(ctx, []*actor.DonorProvidedDetails{lpa123Donor, lpa456Donor, lpa789Donor, lpa0Donor, lpaCertifiedDonor, lpaReferencedDonor}).
				Return([]*lpastore.Lpa{lpa123, lpa456, lpa789, lpa0, lpaCertified, lpaReferenced}, nil)

			dashboardStore := &dashboardStore{dynamoClient: dynamoClient, lpaStoreResolvingService: lpaStoreResolvingService}

			donor, attorney, certificateProvider, err := dashboardStore.GetAll(ctx)
			assert.Nil(t, err)

			assert.Equal(t, []page.LpaAndActorTasks{{Lpa: lpa123}, {Lpa: lpa0}, {Lpa: lpaReferenced}}, donor)
			assert.Equal(t, []page.LpaAndActorTasks{{Lpa: lpa456, CertificateProvider: lpa456CertificateProvider}}, certificateProvider)
			assert.Equal(t, []page.LpaAndActorTasks{{Lpa: lpa789, Attorney: lpa789Attorney}}, attorney)
		})
	}
}

func TestDashboardStoreGetAllSubmittedForAttorneys(t *testing.T) {
	sessionID := "an-id"
	aTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	lpaSubmitted := &lpastore.Lpa{LpaID: "submitted", LpaUID: "M", Submitted: true}
	lpaSubmittedDonor := &actor.DonorProvidedDetails{
		PK:          dynamo.LpaKey("submitted"),
		SK:          dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")),
		LpaID:       "submitted",
		LpaUID:      "M",
		SubmittedAt: aTime,
	}
	lpaSubmittedAttorney := &attorneydata.Provided{
		PK:    dynamo.LpaKey("submitted"),
		SK:    dynamo.AttorneyKey(sessionID),
		LpaID: "submitted",
	}
	lpaSubmittedReplacement := &lpastore.Lpa{LpaID: "submitted-replacement", LpaUID: "M", Submitted: true}
	lpaSubmittedReplacementDonor := &actor.DonorProvidedDetails{
		PK:          dynamo.LpaKey("submitted-replacement"),
		SK:          dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")),
		LpaID:       "submitted-replacement",
		LpaUID:      "M",
		SubmittedAt: aTime,
	}
	lpaSubmittedReplacementAttorney := &attorneydata.Provided{
		PK:            dynamo.LpaKey("submitted-replacement"),
		SK:            dynamo.AttorneyKey(sessionID),
		LpaID:         "submitted-replacement",
		IsReplacement: true,
	}

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: sessionID})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]lpaLink{
			{PK: dynamo.LpaKey("submitted"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")), ActorType: actor.TypeAttorney},
			{PK: dynamo.LpaKey("submitted-replacement"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id")), ActorType: actor.TypeAttorney},
		}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
		{PK: dynamo.LpaKey("submitted"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id"))},
		{PK: dynamo.LpaKey("submitted"), SK: dynamo.AttorneyKey("an-id")},
		{PK: dynamo.LpaKey("submitted-replacement"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("another-id"))},
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

	donor := &actor.DonorProvidedDetails{LpaID: "0", LpaUID: "M", UpdatedAt: aTime, SK: dynamo.LpaOwnerKey(dynamo.DonorKey(sessionID)), PK: dynamo.LpaKey("0")}

	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: sessionID})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]lpaLink{
			{PK: dynamo.LpaKey("0"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor},
		}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
		{PK: dynamo.LpaKey("0"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
	}, []map[string]types.AttributeValue{makeAttributeValueMap(donor)}, nil)

	lpaStoreResolvingService := newMockLpaStoreResolvingService(t)
	lpaStoreResolvingService.EXPECT().ResolveList(ctx, mock.Anything).Return(nil, expectedError)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient, lpaStoreResolvingService: lpaStoreResolvingService}

	_, _, _, err := dashboardStore.GetAll(ctx)
	if !assert.Equal(t, expectedError, err) {
		t.Log(err.Error())
	}
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
		[]lpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("an-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor}}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
		{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
	}, nil, expectedError)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient}

	_, _, _, err := dashboardStore.GetAll(ctx)
	assert.Equal(t, expectedError, err)
}

func TestDashboardStoreGetAllWhenReferenceGetErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllBySK(ctx, dynamo.SubKey("an-id"),
		[]lpaLink{{
			PK:        dynamo.LpaKey("123"),
			SK:        dynamo.SubKey("an-id"),
			DonorKey:  dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")),
			ActorType: actor.TypeDonor,
		}}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
		{PK: dynamo.LpaKey("123"), SK: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id"))},
	}, []map[string]types.AttributeValue{
		makeAttributeValueMap(map[string]any{
			"PK":           dynamo.LpaKey("123"),
			"SK":           dynamo.DonorKey("an-id"),
			"ReferencedSK": dynamo.OrganisationKey("org-id"),
		}),
	}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Keys{
		{PK: dynamo.LpaKey("123"), SK: dynamo.OrganisationKey("org-id")},
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
			lpas:           []lpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("a-sub-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor}},
			expectedExists: true,
			actorType:      actor.TypeDonor,
		},
		"lpas exist - incorrect actor": {
			lpas:           []lpaLink{{PK: dynamo.LpaKey("123"), SK: dynamo.SubKey("a-sub-id"), DonorKey: dynamo.LpaOwnerKey(dynamo.DonorKey("an-id")), ActorType: actor.TypeDonor}},
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
