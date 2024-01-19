package app

import (
	"context"
	"testing"

	"github.com/ministryofjustice/opg-modernising-lpa/internal/actor"
	"github.com/ministryofjustice/opg-modernising-lpa/internal/page"
	"github.com/stretchr/testify/assert"
)

func TestGroupStoreCreate(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	dynamoClient := newMockDynamoClient(t)
	dynamoClient.EXPECT().
		Create(ctx, &actor.Group{
			PK:        "GROUP#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
			SK:        "GROUP#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
			ID:        "0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
			CreatedAt: testNow,
			Name:      "A name",
		}).
		Return(nil)
	dynamoClient.EXPECT().
		Create(ctx, &actor.GroupMember{
			PK:        "GROUP#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
			SK:        "#SUB#an-id",
			CreatedAt: testNow,
		}).
		Return(nil)

	groupStore := &groupStore{dynamoClient: dynamoClient, now: testNowFn}

	err := groupStore.Create(ctx, "A name")
	assert.Nil(t, err)
}

func TestGroupStoreCreateWithSessionMissing(t *testing.T) {
	ctx := context.Background()
	groupStore := &groupStore{dynamoClient: nil, now: testNowFn}

	err := groupStore.Create(ctx, "A name")
	assert.Equal(t, page.SessionMissingError{}, err)
}

func TestGroupStoreCreateWhenErrors(t *testing.T) {
	ctx := page.ContextWithSessionData(context.Background(), &page.SessionData{SessionID: "an-id"})

	testcases := map[string]func(*testing.T) *mockDynamoClient{
		"group": func(t *testing.T) *mockDynamoClient {
			dynamoClient := newMockDynamoClient(t)
			dynamoClient.EXPECT().
				Create(ctx, &actor.Group{
					PK:        "GROUP#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					SK:        "GROUP#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
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
				Create(ctx, &actor.Group{
					PK:        "GROUP#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					SK:        "GROUP#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					ID:        "0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
					CreatedAt: testNow,
					Name:      "A name",
				}).
				Return(nil)
			dynamoClient.EXPECT().
				Create(ctx, &actor.GroupMember{
					PK:        "GROUP#0b7fbdb32a8bb3a8db34fd8d9b04bed390daff9869314240410c9af8b4578db1",
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
			groupStore := &groupStore{dynamoClient: dynamoClient, now: testNowFn}

			err := groupStore.Create(ctx, "A name")
			assert.ErrorIs(t, err, expectedError)
		})
	}
}
