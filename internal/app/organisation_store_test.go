package app

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
)

func TestOrganisationStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Organisation{
			PK:        "ORGANISATION#a-uuid",
			SK:        "ORGANISATION#a-uuid",
			ID:        "a-uuid",
			CreatedAt: testNow,
			Name:      "A name",
		}).
		Return(nil)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Member{
			PK:        "MEMBER#an-id",
			SK:        "ORGANISATION#a-uuid",
			CreatedAt: testNow,
		}).
		Return(nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	err := organisationStore.Create(ctx, "A name")
	assert.Nil(t, err)
}

func TestOrganisationStoreCreateWithSessionMissing(t *testing.T) {
	ctx := context.Background()
	organisationStore := &organisationStore{dynamoClient: nil, now: testNowFn}

	err := organisationStore.Create(ctx, "A name")
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestOrganisationStoreCreateWithMissingSessionID(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{})
	organisationStore := &organisationStore{dynamoClient: nil, now: testNowFn}

	err := organisationStore.Create(ctx, "A name")
	assert.Error(t, err)
}

func TestOrganisationStoreCreateWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"organisation": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
		"member": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(nil).
				Once()
			dynamoClient.EXPECT().
				Create(ctx, mock.Anything).
				Return(expectedError)

			return dynamoClient
		},
	}

	for name, makeMockDynamoClient := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := makeMockDynamoClient(t)
			organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			err := organisationStore.Create(ctx, "A name")
			assert.ErrorIs(t, err, expectedError)
		})
	}
}

func TestOrganisationStoreGet(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})
	organisation := &actor.Organisation{Name: "A name"}

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.
		ExpectOneByPartialSk(ctx, "MEMBER#an-id", "ORGANISATION#",
			&actor.Member{PK: "MEMBER#an-id", SK: "ORGANISATION#a-uuid"}, nil)
	dynamoClient.
		ExpectOne(ctx, "ORGANISATION#a-uuid", "ORGANISATION#a-uuid",
			organisation, nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

	result, err := organisationStore.Get(ctx)
	assert.Nil(t, err)
	assert.Equal(t, organisation, result)
}

func TestOrganisationStoreGetWithSessionMissing(t *testing.T) {
	testcases := map[string]context.Context{
		"no session id":   page.ContextWithSessionData(context.Background(), &page.SessionData{}),
		"no session data": context.Background(),
	}

	for name, ctx := range testcases {
		t.Run(name, func(t *testing.T) {
			organisationStore := &organisationStore{}

			_, err := organisationStore.Get(ctx)
			assert.Error(t, err)
		})
	}
}

func TestOrganisationStoreGetWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"member": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOneByPartialSk(ctx, "MEMBER#an-id", "ORGANISATION#",
					nil, expectedError)

			return dynamoClient
		},
		"organisation": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.
				ExpectOneByPartialSk(ctx, "MEMBER#an-id", "ORGANISATION#",
					&actor.Member{PK: "MEMBER#an-id", SK: "ORGANISATION#a-uuid"}, nil)
			dynamoClient.
				ExpectOne(ctx, "ORGANISATION#a-uuid", "ORGANISATION#a-uuid",
					nil, expectedError)

			return dynamoClient
		},
	}

	for name, makeMockDynamoClient := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := makeMockDynamoClient(t)
			organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn, uuidString: func() string { return "a-uuid" }}

			_, err := organisationStore.Get(ctx)
			assert.Equal(t, expectedError, err)
		})
	}
}
