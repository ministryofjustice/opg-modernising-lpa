package app

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestOrganisationStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Organisation{
			PK:        "ORGANISATION#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
			SK:        "ORGANISATION#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
			ID:        "0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
			CreatedAt: testNow,
			Name:      "A name",
		}).
		Return(nil)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Member{
			PK:        "ORGANISATION#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
			SK:        "#SUB#an-id",
			CreatedAt: testNow,
		}).
		Return(nil)

	organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn}

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
				Create(ctx, &actor.Organisation{
					PK:        "ORGANISATION#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					SK:        "ORGANISATION#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					ID:        "0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					CreatedAt: testNow,
					Name:      "A name",
				}).
				Return(expectedError)

			return dynamoClient
		},
		"member": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, &actor.Organisation{
					PK:        "ORGANISATION#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					SK:        "ORGANISATION#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					ID:        "0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					CreatedAt: testNow,
					Name:      "A name",
				}).
				Return(nil)
			dynamoClient.EXPECT().
				Create(ctx, &actor.Member{
					PK:        "ORGANISATION#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					SK:        "#SUB#an-id",
					CreatedAt: testNow,
				}).
				Return(expectedError)

			return dynamoClient
		},
	}

	for name, makeMockDynamoClient := range testcases {
		t.Run(name, func(t *testing.T) {
			dynamoClient := makeMockDynamoClient(t)
			organisationStore := &organisationStore{dynamoClient: dynamoClient, now: testNowFn}

			err := organisationStore.Create(ctx, "A name")
			assert.ErrorIs(t, err, expectedError)
		})
	}
}
