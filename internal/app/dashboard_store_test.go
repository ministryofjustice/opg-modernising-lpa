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
	lpa0 := &page.Lpa{ID: "0", UID: "M", UpdatedAt: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), SK: donorKey("an-id"), PK: lpaKey("0")}
	lpa123 := &page.Lpa{ID: "123", UID: "M", UpdatedAt: time.Date(2001, time.January, 1, 0, 0, 0, 0, time.UTC), SK: donorKey("an-id"), PK: lpaKey("123")}
	lpa456 := &page.Lpa{ID: "456", UID: "M", SK: donorKey("another-id"), PK: lpaKey("456")}
	lpa456CpProvidedDetails := &actor.CertificateProviderProvidedDetails{
		LpaID: "456", Tasks: actor.CertificateProviderTasks{ConfirmYourDetails: actor.TaskCompleted}, SK: certificateProviderKey("an-id"),
	}
	lpa789 := &page.Lpa{ID: "789", UID: "M", SK: donorKey("different-id"), PK: lpaKey("789")}
	lpa789AttorneyProvidedDetails := &actor.AttorneyProvidedDetails{
		LpaID: "789", Tasks: actor.AttorneyTasks{ConfirmYourDetails: actor.TaskInProgress}, SK: attorneyKey("an-id"),
	}
	lpaNoUID := &page.Lpa{ID: "999", UpdatedAt: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC), SK: donorKey("an-id"), PK: lpaKey("0")}
	lpaSignedByCp := &page.Lpa{ID: "signed-by-cp", UID: "M", SK: donorKey("another-id"), PK: lpaKey("signed-by-cp")}
	lpaSignedByCpProvidedDetails := &actor.CertificateProviderProvidedDetails{
		LpaID: "signed-by-cp", SK: certificateProviderKey("an-id"), Certificate: actor.Certificate{AgreeToStatement: true},
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
			ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

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

			assert.Equal(t, []page.LpaAndActorTasks{{Lpa: lpa123}, {Lpa: lpa0}}, donor)
			assert.Equal(t, []page.LpaAndActorTasks{{Lpa: lpa456, CertificateProviderTasks: actor.CertificateProviderTasks{ConfirmYourDetails: actor.TaskCompleted}}}, certificateProvider)
			assert.Equal(t, []page.LpaAndActorTasks{{Lpa: lpa789, AttorneyTasks: actor.AttorneyTasks{ConfirmYourDetails: actor.TaskInProgress}}}, attorney)
		})
	}
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
