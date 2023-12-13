package app

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/dynamo"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestDashboardStoreGetAll(t *testing.T) {
	sessionID := "an-id"
	aTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	lpa0 := &actor.DonorProvidedDetails{LpaID: "0", LpaUID: "M", UpdatedAt: aTime, SK: donorKey(sessionID), PK: lpaKey("0")}
	lpa123 := &actor.DonorProvidedDetails{LpaID: "123", LpaUID: "M", UpdatedAt: aTime, SK: donorKey(sessionID), PK: lpaKey("123")}
	lpa456 := &actor.DonorProvidedDetails{LpaID: "456", LpaUID: "M", SK: donorKey("another-id"), PK: lpaKey("456")}
	lpa456CpProvidedDetails := &actor.CertificateProviderProvidedDetails{
		LpaID: "456", Tasks: actor.CertificateProviderTasks{ConfirmYourDetails: actor.TaskCompleted}, SK: certificateProviderKey(sessionID),
	}
	lpa789 := &actor.DonorProvidedDetails{LpaID: "789", LpaUID: "M", SK: donorKey("different-id"), PK: lpaKey("789")}
	lpa789AttorneyProvidedDetails := &actor.AttorneyProvidedDetails{
		LpaID: "789", Tasks: actor.AttorneyTasks{ConfirmYourDetails: actor.TaskInProgress}, SK: attorneyKey(sessionID),
	}
	lpaNoUID := &actor.DonorProvidedDetails{LpaID: "999", UpdatedAt: aTime, SK: donorKey(sessionID), PK: lpaKey("0")}
	lpaSignedByCp := &actor.DonorProvidedDetails{LpaID: "signed-by-cp", LpaUID: "M", SK: donorKey("another-id"), PK: lpaKey("signed-by-cp")}
	lpaSignedByCpProvidedDetails := &actor.CertificateProviderProvidedDetails{
		LpaID: "signed-by-cp", SK: certificateProviderKey(sessionID), Certificate: actor.Certificate{AgreeToStatement: true},
	}

	testCases := map[string][]map[string]types.AttributeValue{
		"details returned after lpas": {
			makeAttributeValueMap(lpa123),
			makeAttributeValueMap(lpa456),
			makeAttributeValueMap(lpa456CpProvidedDetails),
			makeAttributeValueMap(lpa789),
			makeAttributeValueMap(lpa789AttorneyProvidedDetails),
			makeAttributeValueMap(lpa0),
			makeAttributeValueMap(lpaSignedByCp),
			makeAttributeValueMap(lpaSignedByCpProvidedDetails),
		},
		"details returned before lpas": {
			makeAttributeValueMap(lpaNoUID),
			makeAttributeValueMap(lpa456CpProvidedDetails),
			makeAttributeValueMap(lpa789AttorneyProvidedDetails),
			makeAttributeValueMap(lpaSignedByCpProvidedDetails),
			makeAttributeValueMap(lpa123),
			makeAttributeValueMap(lpa456),
			makeAttributeValueMap(lpa789),
			makeAttributeValueMap(lpaSignedByCp),
			makeAttributeValueMap(lpa0),
		},
	}

	for name, attributeValues := range testCases {
		t.Run(name, func(t *testing.T) {
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: sessionID})

			dynamoClient := newMockDynamoClient(t)
			dynamoClient.ExpectAllForActor(ctx, "#SUB#an-id",
				[]lpaLink{
					{PK: "LPA#123", SK: "#SUB#an-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor},
					{PK: "LPA#456", SK: "#SUB#an-id", DonorKey: "#DONOR#another-id", ActorType: actor.TypeCertificateProvider},
					{PK: "LPA#789", SK: "#SUB#an-id", DonorKey: "#DONOR#different-id", ActorType: actor.TypeAttorney},
					{PK: "LPA#0", SK: "#SUB#an-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor},
					{PK: "LPA#999", SK: "#SUB#an-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor},
					{PK: "LPA#signed-by-cp", SK: "#SUB#an-id", DonorKey: "#DONOR#another-id", ActorType: actor.TypeCertificateProvider},
				}, nil)
			dynamoClient.ExpectAllByKeys(ctx, []dynamo.Key{
				{PK: "LPA#123", SK: "#DONOR#an-id"},
				{PK: "LPA#456", SK: "#DONOR#another-id"},
				{PK: "LPA#456", SK: "#CERTIFICATE_PROVIDER#an-id"},
				{PK: "LPA#789", SK: "#DONOR#different-id"},
				{PK: "LPA#789", SK: "#ATTORNEY#an-id"},
				{PK: "LPA#0", SK: "#DONOR#an-id"},
				{PK: "LPA#999", SK: "#DONOR#an-id"},
				{PK: "LPA#signed-by-cp", SK: "#DONOR#another-id"},
				{PK: "LPA#signed-by-cp", SK: "#CERTIFICATE_PROVIDER#an-id"},
			}, attributeValues, nil)

			dashboardStore := &dashboardStore{dynamoClient: dynamoClient}

			donor, attorney, certificateProvider, err := dashboardStore.GetAll(ctx)
			assert.Nil(t, err)

			assert.Equal(t, []page.LpaAndActorTasks{{Donor: lpa123}, {Donor: lpa0}}, donor)
			assert.Equal(t, []page.LpaAndActorTasks{{Donor: lpa456, CertificateProvider: lpa456CpProvidedDetails}}, certificateProvider)
			assert.Equal(t, []page.LpaAndActorTasks{{Donor: lpa789, Attorney: lpa789AttorneyProvidedDetails}}, attorney)
		})
	}
}

func TestDashboardStoreGetAllSubmittedForAttorneys(t *testing.T) {
	sessionID := "an-id"
	aTime := time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC)

	lpaSubmitted := &actor.DonorProvidedDetails{LpaID: "submitted", LpaUID: "M", SK: donorKey("another-id"), PK: lpaKey("submitted"), SubmittedAt: aTime}
	lpaSubmittedAttorneyDetails := &actor.AttorneyProvidedDetails{LpaID: "submitted", SK: attorneyKey(sessionID)}
	lpaSubmittedReplacement := &actor.DonorProvidedDetails{LpaID: "submitted-replacement", LpaUID: "M", SK: donorKey("another-id"), PK: lpaKey("submitted-replacement"), SubmittedAt: aTime}
	lpaSubmittedReplacementAttorneyDetails := &actor.AttorneyProvidedDetails{LpaID: "submitted-replacement", SK: attorneyKey(sessionID), IsReplacement: true}
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: sessionID})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllForActor(ctx, "#SUB#an-id",
		[]lpaLink{
			{PK: "LPA#submitted", SK: "#SUB#an-id", DonorKey: "#DONOR#another-id", ActorType: actor.TypeAttorney},
			{PK: "LPA#submitted-replacement", SK: "#SUB#an-id", DonorKey: "#DONOR#another-id", ActorType: actor.TypeAttorney},
		}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Key{
		{PK: "LPA#submitted", SK: "#DONOR#another-id"},
		{PK: "LPA#submitted", SK: "#ATTORNEY#an-id"},
		{PK: "LPA#submitted-replacement", SK: "#DONOR#another-id"},
		{PK: "LPA#submitted-replacement", SK: "#ATTORNEY#an-id"},
	}, []map[string]types.AttributeValue{
		makeAttributeValueMap(lpaSubmitted),
		makeAttributeValueMap(lpaSubmittedAttorneyDetails),
		makeAttributeValueMap(lpaSubmittedReplacement),
		makeAttributeValueMap(lpaSubmittedReplacementAttorneyDetails),
	}, nil)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient}

	_, attorney, _, err := dashboardStore.GetAll(ctx)
	assert.Nil(t, err)

	assert.Equal(t, []page.LpaAndActorTasks{
		{Donor: lpaSubmitted, Attorney: lpaSubmittedAttorneyDetails},
	}, attorney)
}

func makeAttributeValueMap(i interface{}) map[string]types.AttributeValue {
	result, _ := attributevalue.MarshalMap(i)
	return result
}

func TestDashboardStoreGetAllWhenNone(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllForActor(ctx, "#SUB#an-id",
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
	dynamoClient.ExpectAllForActor(ctx, "#SUB#an-id",
		[]lpaLink{}, expectedError)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient}

	_, _, _, err := dashboardStore.GetAll(ctx)
	assert.Equal(t, err, expectedError)
}

func TestDashboardStoreGetAllWhenAllByKeysErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.ExpectAllForActor(ctx, "#SUB#an-id",
		[]lpaLink{{PK: "LPA#123", SK: "#SUB#an-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor}}, nil)
	dynamoClient.ExpectAllByKeys(ctx, []dynamo.Key{
		{PK: "LPA#123", SK: "#DONOR#an-id"},
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
			lpas:           []lpaLink{{PK: "LPA#123", SK: "#SUB#a-sub-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor}},
			expectedExists: true,
			actorType:      actor.TypeDonor,
		},
		"lpas exist - incorrect actor": {
			lpas:           []lpaLink{{PK: "LPA#123", SK: "#SUB#a-sub-id", DonorKey: "#DONOR#an-id", ActorType: actor.TypeDonor}},
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
			dynamoClient.ExpectAllForActor(context.Background(), "#SUB#a-sub-id",
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
	dynamoClient.ExpectAllForActor(context.Background(), "#SUB#a-sub-id",
		[]lpaLink{}, expectedError)

	dashboardStore := &dashboardStore{dynamoClient: dynamoClient}
	exists, err := dashboardStore.SubExistsForActorType(context.Background(), "a-sub-id", actor.TypeDonor)

	assert.Equal(t, expectedError, err)
	assert.False(t, exists)
}
